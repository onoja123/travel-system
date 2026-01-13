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

type LocationHandler struct {
	LocationService *services.LocationService
}

func NewLocationHandler(locationService *services.LocationService) *LocationHandler {
	return &LocationHandler{
		LocationService: locationService,
	}
}

// UpdateLocation godoc
// @Summary Update user location
// @Description Update current location for walk time calculations
// @Tags location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.LocationUpdate true "Location data"
// @Success 200 {object} utils.Response
// @Router /api/location/update [post]
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	var req models.LocationUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate coordinates
	if req.Latitude < -90 || req.Latitude > 90 {
		utils.ErrorResponse(c, 400, "Invalid latitude")
		return
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		utils.ErrorResponse(c, 400, "Invalid longitude")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.LocationService.UpdateLocation(ctx, req)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to update location")
		return
	}

	utils.SuccessResponse(c, 200, "Location updated successfully", nil)
}

// GetWalkTime godoc
// @Summary Get walk time to gate
// @Description Calculate estimated walk time from current location to gate
// @Tags location
// @Produce json
// @Security BearerAuth
// @Param flightId path string true "Flight ID"
// @Success 200 {object} models.WalkTimeResponse
// @Router /api/location/walk-time/{flightId} [get]
func (h *LocationHandler) GetWalkTime(c *gin.Context) {
	flightID := c.Param("flightId")
	userID := c.GetString("user_id")

	if userID == "" {
		utils.ErrorResponse(c, 401, "Unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	walkTime, err := h.LocationService.GetWalkTime(ctx, flightID, userObjID)
	if err != nil {
		utils.ErrorResponse(c, 404, err.Error())
		return
	}

	utils.SuccessResponse(c, 200, "Walk time calculated", walkTime)
}
