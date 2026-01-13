package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/onoja123/travel-companion-backend/internal/config"
	"github.com/onoja123/travel-companion-backend/internal/models"
)

type AviationService struct {
	Config *config.Config
	Client *http.Client
}

func NewAviationService(cfg *config.Config) *AviationService {
	return &AviationService{
		Config: cfg,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// AviationStack Response Structure
type AviationStackResponse struct {
	Data []AviationStackFlight `json:"data"`
}

type AviationStackFlight struct {
	FlightDate   string                   `json:"flight_date"`
	FlightStatus string                   `json:"flight_status"`
	Departure    AviationStackAirportInfo `json:"departure"`
	Arrival      AviationStackAirportInfo `json:"arrival"`
	Airline      AviationStackAirline     `json:"airline"`
	Flight       AviationStackFlightInfo  `json:"flight"`
}

type AviationStackAirportInfo struct {
	Airport   string `json:"airport"`
	Timezone  string `json:"timezone"`
	Iata      string `json:"iata"`
	Terminal  string `json:"terminal"`
	Gate      string `json:"gate"`
	Delay     int    `json:"delay"`
	Scheduled string `json:"scheduled"`
	Estimated string `json:"estimated"`
	Actual    string `json:"actual"`
}

type AviationStackAirline struct {
	Name string `json:"name"`
	Iata string `json:"iata"`
}

type AviationStackFlightInfo struct {
	Number string `json:"number"`
	Iata   string `json:"iata"`
}

func (s *AviationService) GetFlightStatus(flightNumber, date string) (*models.FlightStatus, error) {
	switch s.Config.Aviation.Provider {
	case "aviationstack":
		return s.getFromAviationStack(flightNumber, date)
	case "flightaware":
		return s.getFromFlightAware(flightNumber, date)
	case "amadeus":
		return s.getFromAmadeus(flightNumber, date)
	default:
		return nil, fmt.Errorf("unsupported aviation API provider: %s", s.Config.Aviation.Provider)
	}
}

func (s *AviationService) getFromAviationStack(flightNumber, date string) (*models.FlightStatus, error) {
	url := fmt.Sprintf("http://api.aviationstack.com/v1/flights?access_key=%s&flight_iata=%s&flight_date=%s",
		s.Config.Aviation.AviationStackKey,
		flightNumber,
		date,
	)

	resp, err := s.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call AviationStack API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp AviationStackResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("flight not found")
	}

	flight := apiResp.Data[0]

	// Parse times
	departureTime, _ := time.Parse(time.RFC3339, flight.Departure.Scheduled)
	arrivalTime, _ := time.Parse(time.RFC3339, flight.Arrival.Scheduled)

	// Calculate boarding time (typically 40 minutes before departure)
	boardingTime := departureTime.Add(-40 * time.Minute)

	flightStatus := &models.FlightStatus{
		FlightKey:     fmt.Sprintf("%s_%s", flightNumber, date),
		FlightNumber:  flightNumber,
		AirlineCode:   flight.Airline.Iata,
		Status:        s.mapStatus(flight.FlightStatus),
		Gate:          flight.Departure.Gate,
		Terminal:      flight.Departure.Terminal,
		BoardingTime:  boardingTime,
		DepartureTime: departureTime,
		ArrivalTime:   arrivalTime,
		DelayMinutes:  flight.Departure.Delay,
		LastUpdated:   time.Now(),
		RawData:       flight,
	}

	return flightStatus, nil
}

func (s *AviationService) getFromFlightAware(flightNumber, date string) (*models.FlightStatus, error) {
	// TODO: Implement FlightAware API integration
	return nil, fmt.Errorf("FlightAware integration not implemented yet")
}

func (s *AviationService) getFromAmadeus(flightNumber, date string) (*models.FlightStatus, error) {
	// TODO: Implement Amadeus API integration
	return nil, fmt.Errorf("Amadeus integration not implemented yet")
}

func (s *AviationService) mapStatus(apiStatus string) string {
	statusMap := map[string]string{
		"scheduled": "On Time",
		"active":    "Boarding Soon",
		"landed":    "Arrived",
		"cancelled": "Cancelled",
		"incident":  "Delayed",
		"diverted":  "Diverted",
	}

	if mapped, ok := statusMap[apiStatus]; ok {
		return mapped
	}
	return "Unknown"
}
