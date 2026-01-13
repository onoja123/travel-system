package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onoja123/travel-companion-backend/internal/models"
	"github.com/onoja123/travel-companion-backend/internal/services"
	"github.com/onoja123/travel-companion-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FlightHandler struct {
	FlightService *services.FlightService
}

func NewFlightHandler(flightService *services.FlightService) *FlightHandler {
	return &FlightHandler{
		FlightService: flightService,
	}
}

// TrackFlight godoc
// @Summary Track a flight
// @Description Add a flight to user's tracking list
// @Tags flights
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.TrackFlightRequest true "Flight details"
// @Success 201 {object} utils.Response
// @Router /api/flights/track [post]
func (h *FlightHandler) TrackFlight(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, 401, "Unauthorized")
		return
	}

	var req models.TrackFlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate inputs
	if !utils.IsValidFlightNumber(req.FlightNumber) {
		utils.ErrorResponse(c, 400, "Invalid flight number format")
		return
	}

	if !utils.IsValidAirportCode(req.DepartureAirport) || !utils.IsValidAirportCode(req.ArrivalAirport) {
		utils.ErrorResponse(c, 400, "Invalid airport code")
		return
	}

	if !utils.IsValidDate(req.DepartureDate) {
		utils.ErrorResponse(c, 400, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(userID)
	flight, err := h.FlightService.TrackFlight(ctx, objID, req)
	if err != nil {
		utils.ErrorResponse(c, 400, err.Error())
		return
	}

	utils.SuccessResponse(c, 201, "Flight tracked successfully", flight)
}

// GetUserFlights godoc
// @Summary Get user's flights
// @Description Get all tracked flights for the authenticated user
// @Tags flights
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Success 200 {object} utils.Response
// @Router /api/flights/user/{userId} [get]
func (h *FlightHandler) GetUserFlights(c *gin.Context) {
	userID := c.Param("userId")
	authenticatedUserID := c.GetString("user_id")

	// Ensure user can only access their own flights
	if userID != authenticatedUserID {
		utils.ErrorResponse(c, 403, "Forbidden")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(userID)
	flights, err := h.FlightService.GetUserFlights(ctx, objID)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch flights")
		return
	}

	utils.SuccessResponse(c, 200, "Flights retrieved successfully", flights)
}

// GetFlightStatus godoc
// @Summary Get flight status
// @Description Get real-time status of a specific flight
// @Tags flights
// @Produce json
// @Param flightNumber path string true "Flight Number"
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} models.FlightStatusResponse
// @Router /api/flights/status/{flightNumber}/{date} [get]
func (h *FlightHandler) GetFlightStatus(c *gin.Context) {
	flightNumber := c.Param("flightNumber")
	date := c.Param("date")

	if !utils.IsValidFlightNumber(flightNumber) {
		utils.ErrorResponse(c, 400, "Invalid flight number")
		return
	}

	if !utils.IsValidDate(date) {
		utils.ErrorResponse(c, 400, "Invalid date format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	status, err := h.FlightService.GetFlightStatus(ctx, flightNumber, date)
	if err != nil {
		utils.ErrorResponse(c, 404, "Flight not found")
		return
	}

	utils.SuccessResponse(c, 200, "Flight status retrieved", status)
}

// DeleteTrackedFlight godoc
// @Summary Stop tracking a flight
// @Description Remove a flight from user's tracking list
// @Tags flights
// @Produce json
// @Security BearerAuth
// @Param id path string true "Flight ID"
// @Success 200 {object} utils.Response
// @Router /api/flights/{id} [delete]
func (h *FlightHandler) DeleteTrackedFlight(c *gin.Context) {
	flightID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		utils.ErrorResponse(c, 401, "Unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	flightObjID, err := primitive.ObjectIDFromHex(flightID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid flight ID")
		return
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	err = h.FlightService.DeleteTrackedFlight(ctx, flightObjID, userObjID)
	if err != nil {
		utils.ErrorResponse(c, 404, err.Error())
		return
	}

	utils.SuccessResponse(c, 200, "Flight tracking stopped", nil)
}
