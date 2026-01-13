package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	FlightKey string             `bson:"flight_key" json:"flight_key"`
	Type      string             `bson:"type" json:"type"` // "gate_change", "boarding_soon", "urgent", "critical", "delay"
	Title     string             `bson:"title" json:"title"`
	Body      string             `bson:"body" json:"body"`
	Priority  string             `bson:"priority" json:"priority"` // "normal", "high"
	SentAt    time.Time          `bson:"sent_at" json:"sent_at"`
	ReadAt    *time.Time         `bson:"read_at,omitempty" json:"read_at,omitempty"`
}

type NotificationPreferencesRequest struct {
	NotifyGateChange   bool `json:"notify_gate_change"`
	NotifyBoarding     bool `json:"notify_boarding"`
	NotifyDelay        bool `json:"notify_delay"`
	BoardingReminder40 bool `json:"boarding_reminder_40"`
	BoardingReminder20 bool `json:"boarding_reminder_20"`
	BoardingReminder10 bool `json:"boarding_reminder_10"`
}
