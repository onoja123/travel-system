package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/onoja123/travel-companion-backend/internal/database"
	"github.com/onoja123/travel-companion-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LocationService struct {
	MongoDB *database.MongoDB
	Redis   *database.RedisClient
}

func NewLocationService(db *database.MongoDB, redis *database.RedisClient) *LocationService {
	return &LocationService{
		MongoDB: db,
		Redis:   redis,
	}
}

type UserLocation struct {
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *LocationService) UpdateLocation(ctx context.Context, update models.LocationUpdate) error {
	location := UserLocation{
		UserID:    update.UserID,
		Latitude:  update.Latitude,
		Longitude: update.Longitude,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(location)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("user:location:%s", update.UserID)
	return s.Redis.Client.Set(ctx, key, data, 10*time.Minute).Err()
}

func (s *LocationService) GetWalkTime(ctx context.Context, flightID string, userID primitive.ObjectID) (*models.WalkTimeResponse, error) {
	// Get user location from Redis
	key := fmt.Sprintf("user:location:%s", userID.Hex())
	data, err := s.Redis.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("user location not found")
	}

	var userLoc UserLocation
	if err := json.Unmarshal([]byte(data), &userLoc); err != nil {
		return nil, err
	}

	// Get flight details
	flightObjID, _ := primitive.ObjectIDFromHex(flightID)
	var flight models.TrackedFlight
	err = s.MongoDB.TrackedFlights().FindOne(ctx, bson.M{"_id": flightObjID}).Decode(&flight)
	if err != nil {
		return nil, fmt.Errorf("flight not found")
	}

	// Get airport/gate location
	var airport models.Airport
	err = s.MongoDB.Airports().FindOne(ctx, bson.M{"code": flight.DepartureAirport}).Decode(&airport)
	if err != nil {
		return nil, fmt.Errorf("airport not found")
	}

	// Calculate distance
	distance := s.calculateDistance(userLoc.Latitude, userLoc.Longitude, airport.Latitude, airport.Longitude)

	// Estimate walk time (average walking speed: 1.4 m/s or ~5 km/h)
	walkTimeMinutes := int(distance / 1.4 / 60) // Convert to minutes

	// Get flight status to determine urgency
	flightKey := fmt.Sprintf("%s_%s", flight.FlightNumber, flight.DepartureDate.Format("2006-01-02"))
	var status models.FlightStatus
	s.MongoDB.FlightStatus().FindOne(ctx, bson.M{"flight_key": flightKey}).Decode(&status)

	minutesUntilBoarding := int(time.Until(status.BoardingTime).Minutes())

	urgencyLevel := "calm"
	recommendedAction := "You have plenty of time"

	if walkTimeMinutes > minutesUntilBoarding {
		urgencyLevel = "critical"
		recommendedAction = "Head to gate immediately!"
	} else if walkTimeMinutes+10 > minutesUntilBoarding {
		urgencyLevel = "urgent"
		recommendedAction = "Start heading to gate now"
	} else if walkTimeMinutes+20 > minutesUntilBoarding {
		urgencyLevel = "moderate"
		recommendedAction = "Consider heading to gate soon"
	}

	return &models.WalkTimeResponse{
		FlightID:          flightID,
		DistanceMeters:    distance,
		WalkTimeMinutes:   walkTimeMinutes,
		UrgencyLevel:      urgencyLevel,
		RecommendedAction: recommendedAction,
	}, nil
}

// Haversine formula to calculate distance between two coordinates
func (s *LocationService) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func (s *LocationService) GetSecurityWaitTime(ctx context.Context, airportCode string) (*models.SecurityWaitTime, error) {
	// Try Redis cache first
	key := fmt.Sprintf("airport:wait:%s", airportCode)
	cachedData, err := s.Redis.Client.Get(ctx, key).Result()
	if err == nil {
		var waitTime models.SecurityWaitTime
		if json.Unmarshal([]byte(cachedData), &waitTime) == nil {
			return &waitTime, nil
		}
	}

	// Get from database (average)
	var airport models.Airport
	err = s.MongoDB.Airports().FindOne(ctx, bson.M{"code": airportCode}).Decode(&airport)
	if err != nil {
		return nil, err
	}

	waitTime := &models.SecurityWaitTime{
		AirportCode:     airportCode,
		CurrentWaitTime: airport.SecurityWaitAvg,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	// Cache for 15 minutes
	data, _ := json.Marshal(waitTime)
	s.Redis.Client.Set(ctx, key, data, 15*time.Minute)

	return waitTime, nil
}
