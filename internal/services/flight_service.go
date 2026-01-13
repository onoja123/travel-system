package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/onoja123/travel-companion-backend/internal/database"
	"github.com/onoja123/travel-companion-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FlightService struct {
	MongoDB         *database.MongoDB
	Redis           *database.RedisClient
	AviationService *AviationService
	NotificationSvc *NotificationService
}

func NewFlightService(
	db *database.MongoDB,
	redis *database.RedisClient,
	aviationSvc *AviationService,
	notifSvc *NotificationService,
) *FlightService {
	return &FlightService{
		MongoDB:         db,
		Redis:           redis,
		AviationService: aviationSvc,
		NotificationSvc: notifSvc,
	}
}

func (s *FlightService) TrackFlight(ctx context.Context, userID primitive.ObjectID, req models.TrackFlightRequest) (*models.TrackedFlight, error) {
	// Validate flight exists by calling aviation API
	flightStatus, err := s.AviationService.GetFlightStatus(req.FlightNumber, req.DepartureDate)
	if err != nil {
		return nil, fmt.Errorf("flight not found or invalid: %w", err)
	}

	// Parse departure date
	departureDate, err := time.Parse("2006-01-02", req.DepartureDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create tracked flight
	trackedFlight := &models.TrackedFlight{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		FlightNumber:     req.FlightNumber,
		AirlineCode:      flightStatus.AirlineCode,
		DepartureDate:    departureDate,
		DepartureAirport: req.DepartureAirport,
		ArrivalAirport:   req.ArrivalAirport,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to MongoDB
	_, err = s.MongoDB.TrackedFlights().InsertOne(ctx, trackedFlight)
	if err != nil {
		return nil, fmt.Errorf("failed to save tracked flight: %w", err)
	}

	// Cache flight status in Redis
	s.cacheFlightStatus(flightStatus)

	// Save flight status to MongoDB
	_, err = s.MongoDB.FlightStatus().InsertOne(ctx, flightStatus)
	if err != nil {
		log.Printf("Warning: Failed to save flight status: %v", err)
	}

	return trackedFlight, nil
}

func (s *FlightService) GetUserFlights(ctx context.Context, userID primitive.ObjectID) ([]models.TrackedFlight, error) {
	cursor, err := s.MongoDB.TrackedFlights().Find(ctx, bson.M{
		"user_id":   userID,
		"is_active": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var flights []models.TrackedFlight
	if err := cursor.All(ctx, &flights); err != nil {
		return nil, err
	}

	return flights, nil
}

func (s *FlightService) GetFlightStatus(ctx context.Context, flightNumber, date string) (*models.FlightStatusResponse, error) {
	flightKey := fmt.Sprintf("%s_%s", flightNumber, date)

	// Try Redis cache first
	cachedStatus, err := s.getFlightStatusFromCache(flightKey)
	if err == nil && cachedStatus != nil {
		return s.buildFlightStatusResponse(cachedStatus), nil
	}

	// Try MongoDB
	var status models.FlightStatus
	err = s.MongoDB.FlightStatus().FindOne(ctx, bson.M{"flight_key": flightKey}).Decode(&status)
	if err == nil {
		s.cacheFlightStatus(&status)
		return s.buildFlightStatusResponse(&status), nil
	}

	// Fetch from aviation API
	status2, err := s.AviationService.GetFlightStatus(flightNumber, date)
	if err != nil {
		return nil, err
	}

	// Cache and save
	s.cacheFlightStatus(status2)
	s.MongoDB.FlightStatus().InsertOne(ctx, status2)

	return s.buildFlightStatusResponse(status2), nil
}

func (s *FlightService) DeleteTrackedFlight(ctx context.Context, flightID primitive.ObjectID, userID primitive.ObjectID) error {
	result, err := s.MongoDB.TrackedFlights().UpdateOne(
		ctx,
		bson.M{"_id": flightID, "user_id": userID},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("flight not found or unauthorized")
	}

	return nil
}

func (s *FlightService) cacheFlightStatus(status *models.FlightStatus) {
	data, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal flight status: %v", err)
		return
	}

	key := fmt.Sprintf("flight:%s", status.FlightKey)
	s.Redis.Client.Set(context.Background(), key, data, 5*time.Minute)
}

func (s *FlightService) getFlightStatusFromCache(flightKey string) (*models.FlightStatus, error) {
	key := fmt.Sprintf("flight:%s", flightKey)
	data, err := s.Redis.Client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	var status models.FlightStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (s *FlightService) buildFlightStatusResponse(status *models.FlightStatus) *models.FlightStatusResponse {
	now := time.Now()
	boardingMinutes := int(status.BoardingTime.Sub(now).Minutes())
	departureMinutes := int(status.DepartureTime.Sub(now).Minutes())

	urgencyLevel := "calm"
	if boardingMinutes <= 10 {
		urgencyLevel = "critical"
	} else if boardingMinutes <= 20 {
		urgencyLevel = "urgent"
	} else if boardingMinutes <= 40 {
		urgencyLevel = "moderate"
	}

	return &models.FlightStatusResponse{
		Flight: *status,
		TimeUntil: models.TimeUntil{
			BoardingMinutes:  boardingMinutes,
			DepartureMinutes: departureMinutes,
		},
		UrgencyLevel: urgencyLevel,
	}
}

// Background polling service
func (s *FlightService) StartPollingService(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	log.Println("🔄 Started flight polling service")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping flight polling service")
			return
		case <-ticker.C:
			s.pollAllActiveFlights(ctx)
		}
	}
}

func (s *FlightService) pollAllActiveFlights(ctx context.Context) {
	cursor, err := s.MongoDB.TrackedFlights().Find(ctx, bson.M{"is_active": true})
	if err != nil {
		log.Printf("Error fetching active flights: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var flights []models.TrackedFlight
	if err := cursor.All(ctx, &flights); err != nil {
		log.Printf("Error decoding flights: %v", err)
		return
	}

	for _, flight := range flights {
		s.checkFlightUpdates(ctx, &flight)
	}
}

func (s *FlightService) checkFlightUpdates(ctx context.Context, flight *models.TrackedFlight) {
	dateStr := flight.DepartureDate.Format("2006-01-02")

	// Get current status from cache/db
	flightKey := fmt.Sprintf("%s_%s", flight.FlightNumber, dateStr)
	var oldStatus models.FlightStatus
	err := s.MongoDB.FlightStatus().FindOne(ctx, bson.M{"flight_key": flightKey}).Decode(&oldStatus)

	// Fetch latest status from API
	newStatus, err := s.AviationService.GetFlightStatus(flight.FlightNumber, dateStr)
	if err != nil {
		log.Printf("Error fetching flight status for %s: %v", flight.FlightNumber, err)
		return
	}

	// Check for changes
	changes := s.detectChanges(&oldStatus, newStatus)
	if len(changes) > 0 {
		// Update database
		s.MongoDB.FlightStatus().UpdateOne(
			ctx,
			bson.M{"flight_key": flightKey},
			bson.M{"$set": newStatus},
		)

		// Update cache
		s.cacheFlightStatus(newStatus)

		// Send notifications
		s.NotificationSvc.HandleFlightChanges(ctx, flight.UserID, newStatus, changes)

		log.Printf("✈️  Flight %s updated: %v", flight.FlightNumber, changes)
	}
}

func (s *FlightService) detectChanges(old, new *models.FlightStatus) map[string]interface{} {
	changes := make(map[string]interface{})

	if old.Gate != "" && old.Gate != new.Gate {
		changes["gate"] = map[string]string{
			"old": old.Gate,
			"new": new.Gate,
		}
	}

	if old.Status != new.Status {
		changes["status"] = map[string]string{
			"old": old.Status,
			"new": new.Status,
		}
	}

	if old.DelayMinutes != new.DelayMinutes {
		changes["delay"] = map[string]int{
			"old": old.DelayMinutes,
			"new": new.DelayMinutes,
		}
	}

	return changes
}
