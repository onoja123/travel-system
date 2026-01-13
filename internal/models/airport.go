package models

type Airport struct {
	Code            string  `bson:"code" json:"code"`
	Name            string  `bson:"name" json:"name"`
	City            string  `bson:"city" json:"city"`
	Country         string  `bson:"country" json:"country"`
	Timezone        string  `bson:"timezone" json:"timezone"`
	SecurityWaitAvg int     `bson:"security_wait_avg" json:"security_wait_avg"` // minutes
	Latitude        float64 `bson:"latitude" json:"latitude"`
	Longitude       float64 `bson:"longitude" json:"longitude"`
}

type SecurityWaitTime struct {
	AirportCode     string `json:"airport_code"`
	CurrentWaitTime int    `json:"current_wait_time"` // minutes
	Timestamp       string `json:"timestamp"`
}

type LocationUpdate struct {
	UserID    string  `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Timestamp string  `json:"timestamp" binding:"required"`
}

type WalkTimeResponse struct {
	FlightID          string  `json:"flight_id"`
	DistanceMeters    float64 `json:"distance_meters"`
	WalkTimeMinutes   int     `json:"walk_time_minutes"`
	UrgencyLevel      string  `json:"urgency_level"`
	RecommendedAction string  `json:"recommended_action"`
}
