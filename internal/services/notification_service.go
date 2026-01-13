package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onoja123/travel-companion-backend/internal/database"
	"github.com/onoja123/travel-companion-backend/internal/models"
	"github.com/onoja123/travel-companion-backend/pkg/fcm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	MongoDB    *database.MongoDB
	FCMService *fcm.FCMService
}

func NewNotificationService(db *database.MongoDB, fcmService *fcm.FCMService) *NotificationService {
	return &NotificationService{
		MongoDB:    db,
		FCMService: fcmService,
	}
}

func (s *NotificationService) HandleFlightChanges(ctx context.Context, userID primitive.ObjectID, flight *models.FlightStatus, changes map[string]interface{}) {
	// Get user to check FCM token and preferences
	var user models.User
	err := s.MongoDB.Users().FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		return
	}

	if user.FCMToken == "" {
		return // No FCM token, skip notification
	}

	// Handle gate change
	if gateChange, ok := changes["gate"].(map[string]string); ok && user.Preferences.NotifyGateChange {
		s.sendGateChangeNotification(ctx, &user, flight, gateChange)
	}

	// Handle status change
	if statusChange, ok := changes["status"].(map[string]string); ok && user.Preferences.NotifyBoarding {
		s.sendStatusChangeNotification(ctx, &user, flight, statusChange)
	}

	// Handle delay
	if delayChange, ok := changes["delay"].(map[string]int); ok && user.Preferences.NotifyDelay {
		s.sendDelayNotification(ctx, &user, flight, delayChange)
	}
}

func (s *NotificationService) sendGateChangeNotification(ctx context.Context, user *models.User, flight *models.FlightStatus, change map[string]string) {
	title := "⚡ Gate Changed"
	body := fmt.Sprintf("%s moved from Gate %s to Gate %s", flight.FlightNumber, change["old"], change["new"])

	notification := &models.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		FlightKey: flight.FlightKey,
		Type:      "gate_change",
		Title:     title,
		Body:      body,
		Priority:  "high",
		SentAt:    time.Now(),
	}

	// Save to database
	s.MongoDB.Notifications().InsertOne(ctx, notification)

	// Send via FCM
	s.FCMService.SendNotification(user.FCMToken, title, body, map[string]string{
		"type":       "gate_change",
		"flight_key": flight.FlightKey,
		"new_gate":   change["new"],
	})
}

func (s *NotificationService) sendStatusChangeNotification(ctx context.Context, user *models.User, flight *models.FlightStatus, change map[string]string) {
	title := fmt.Sprintf("Flight %s", change["new"])
	body := fmt.Sprintf("%s status changed to %s", flight.FlightNumber, change["new"])

	notification := &models.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		FlightKey: flight.FlightKey,
		Type:      "status_change",
		Title:     title,
		Body:      body,
		Priority:  "normal",
		SentAt:    time.Now(),
	}

	s.MongoDB.Notifications().InsertOne(ctx, notification)
	s.FCMService.SendNotification(user.FCMToken, title, body, map[string]string{
		"type":       "status_change",
		"flight_key": flight.FlightKey,
	})
}

func (s *NotificationService) sendDelayNotification(ctx context.Context, user *models.User, flight *models.FlightStatus, change map[string]int) {
	title := "⏰ Flight Delayed"
	body := fmt.Sprintf("%s is delayed by %d minutes", flight.FlightNumber, change["new"])

	notification := &models.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		FlightKey: flight.FlightKey,
		Type:      "delay",
		Title:     title,
		Body:      body,
		Priority:  "high",
		SentAt:    time.Now(),
	}

	s.MongoDB.Notifications().InsertOne(ctx, notification)
	s.FCMService.SendNotification(user.FCMToken, title, body, map[string]string{
		"type":       "delay",
		"flight_key": flight.FlightKey,
		"delay":      fmt.Sprintf("%d", change["new"]),
	})
}

func (s *NotificationService) CheckBoardingReminders(ctx context.Context) {
	// Find all active flights
	cursor, err := s.MongoDB.TrackedFlights().Find(ctx, bson.M{"is_active": true})
	if err != nil {
		log.Printf("Error fetching flights for reminders: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var flights []models.TrackedFlight
	cursor.All(ctx, &flights)

	for _, flight := range flights {
		s.checkBoardingReminderForFlight(ctx, &flight)
	}
}

func (s *NotificationService) checkBoardingReminderForFlight(ctx context.Context, flight *models.TrackedFlight) {
	// Get flight status
	flightKey := fmt.Sprintf("%s_%s", flight.FlightNumber, flight.DepartureDate.Format("2006-01-02"))
	var status models.FlightStatus
	err := s.MongoDB.FlightStatus().FindOne(ctx, bson.M{"flight_key": flightKey}).Decode(&status)
	if err != nil {
		return
	}

	// Get user
	var user models.User
	err = s.MongoDB.Users().FindOne(ctx, bson.M{"_id": flight.UserID}).Decode(&user)
	if err != nil || user.FCMToken == "" {
		return
	}

	// Calculate minutes until boarding
	minutesUntil := int(time.Until(status.BoardingTime).Minutes())

	// Check if we should send reminders
	if minutesUntil == 40 && user.Preferences.BoardingReminder40 {
		s.sendBoardingReminder(ctx, &user, &status, 40)
	} else if minutesUntil == 20 && user.Preferences.BoardingReminder20 {
		s.sendBoardingReminder(ctx, &user, &status, 20)
	} else if minutesUntil == 10 && user.Preferences.BoardingReminder10 {
		s.sendBoardingReminder(ctx, &user, &status, 10)
	}
}

func (s *NotificationService) sendBoardingReminder(ctx context.Context, user *models.User, flight *models.FlightStatus, minutes int) {
	var title, priority string

	switch minutes {
	case 40:
		title = "⏰ Start Heading to Gate"
		priority = "normal"
	case 20:
		title = "🚨 Boarding Soon"
		priority = "high"
	case 10:
		title = "🔴 FINAL CALL"
		priority = "high"
	}

	body := fmt.Sprintf("%s boards in %d minutes - Gate %s", flight.FlightNumber, minutes, flight.Gate)

	notification := &models.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		FlightKey: flight.FlightKey,
		Type:      fmt.Sprintf("boarding_%d", minutes),
		Title:     title,
		Body:      body,
		Priority:  priority,
		SentAt:    time.Now(),
	}

	s.MongoDB.Notifications().InsertOne(ctx, notification)
	s.FCMService.SendNotification(user.FCMToken, title, body, map[string]string{
		"type":       fmt.Sprintf("boarding_%d", minutes),
		"flight_key": flight.FlightKey,
		"gate":       flight.Gate,
	})
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID primitive.ObjectID) ([]models.Notification, error) {
	cursor, err := s.MongoDB.Notifications().Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *NotificationService) UpdatePreferences(ctx context.Context, userID primitive.ObjectID, prefs models.NotificationPreferencesRequest) error {
	_, err := s.MongoDB.Users().UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"preferences.notify_gate_change":   prefs.NotifyGateChange,
			"preferences.notify_boarding":      prefs.NotifyBoarding,
			"preferences.notify_delay":         prefs.NotifyDelay,
			"preferences.boarding_reminder_40": prefs.BoardingReminder40,
			"preferences.boarding_reminder_20": prefs.BoardingReminder20,
			"preferences.boarding_reminder_10": prefs.BoardingReminder10,
			"updated_at":                       time.Now(),
		}},
	)
	return err
}

// Background service for boarding reminders
func (s *NotificationService) StartReminderService(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("🔔 Started boarding reminder service")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping reminder service")
			return
		case <-ticker.C:
			s.CheckBoardingReminders(ctx)
		}
	}
}
