package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrackedFlight struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	FlightNumber     string             `bson:"flight_number" json:"flight_number"`
	AirlineCode      string             `bson:"airline_code" json:"airline_code"`
	DepartureDate    time.Time          `bson:"departure_date" json:"departure_date"`
	DepartureAirport string             `bson:"departure_airport" json:"departure_airport"`
	ArrivalAirport   string             `bson:"arrival_airport" json:"arrival_airport"`
	IsActive         bool               `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type FlightStatus struct {
	FlightKey     string      `bson:"flight_key" json:"flight_key"`
	FlightNumber  string      `bson:"flight_number" json:"flight_number"`
	AirlineCode   string      `bson:"airline_code" json:"airline_code"`
	Status        string      `bson:"status" json:"status"`
	Gate          string      `bson:"gate,omitempty" json:"gate,omitempty"`
	Terminal      string      `bson:"terminal,omitempty" json:"terminal,omitempty"`
	BoardingTime  time.Time   `bson:"boarding_time,omitempty" json:"boarding_time,omitempty"`
	DepartureTime time.Time   `bson:"departure_time" json:"departure_time"`
	ArrivalTime   time.Time   `bson:"arrival_time" json:"arrival_time"`
	DelayMinutes  int         `bson:"delay_minutes" json:"delay_minutes"`
	GateChange    *GateChange `bson:"gate_change,omitempty" json:"gate_change,omitempty"`
	LastUpdated   time.Time   `bson:"last_updated" json:"last_updated"`
	RawData       interface{} `bson:"raw_data,omitempty" json:"raw_data,omitempty"`
}

type GateChange struct {
	OldGate    string    `bson:"old_gate" json:"old_gate"`
	NewGate    string    `bson:"new_gate" json:"new_gate"`
	Reason     string    `bson:"reason,omitempty" json:"reason,omitempty"`
	TimeImpact string    `bson:"time_impact" json:"time_impact"` // "none", "minor", "major"
	ChangedAt  time.Time `bson:"changed_at" json:"changed_at"`
}

type TrackFlightRequest struct {
	FlightNumber     string `json:"flight_number" binding:"required"`
	DepartureDate    string `json:"departure_date" binding:"required"` // YYYY-MM-DD
	DepartureAirport string `json:"departure_airport" binding:"required,len=3"`
	ArrivalAirport   string `json:"arrival_airport" binding:"required,len=3"`
}

type FlightStatusResponse struct {
	Flight       FlightStatus `json:"flight"`
	TimeUntil    TimeUntil    `json:"time_until"`
	UrgencyLevel string       `json:"urgency_level"` // "calm", "moderate", "urgent", "critical"
}

type TimeUntil struct {
	BoardingMinutes  int `json:"boarding_minutes"`
	DepartureMinutes int `json:"departure_minutes"`
}
