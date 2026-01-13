package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email       string             `bson:"email" json:"email" binding:"required,email"`
	Phone       string             `bson:"phone,omitempty" json:"phone,omitempty"`
	Password    string             `bson:"password" json:"-"`
	FCMToken    string             `bson:"fcm_token,omitempty" json:"fcm_token,omitempty"`
	Preferences UserPreferences    `bson:"preferences" json:"preferences"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type UserPreferences struct {
	NotifyGateChange   bool `bson:"notify_gate_change" json:"notify_gate_change"`
	NotifyBoarding     bool `bson:"notify_boarding" json:"notify_boarding"`
	NotifyDelay        bool `bson:"notify_delay" json:"notify_delay"`
	BoardingReminder40 bool `bson:"boarding_reminder_40" json:"boarding_reminder_40"`
	BoardingReminder20 bool `bson:"boarding_reminder_20" json:"boarding_reminder_20"`
	BoardingReminder10 bool `bson:"boarding_reminder_10" json:"boarding_reminder_10"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type FCMTokenRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}
