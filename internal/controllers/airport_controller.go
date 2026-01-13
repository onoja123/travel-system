package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onoja123/travel-companion-backend/internal/database"
	"github.com/onoja123/travel-companion-backend/internal/models"
	"github.com/onoja123/travel-companion-backend/internal/services"
	"github.com/onoja123/travel-companion-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
)

type AirportController struct {
	MongoDB         *database.MongoDB
	LocationService *services.LocationService
}

func NewAirportController(db *database.MongoDB, locationService *services.LocationService) *AirportController {
	return &AirportController{
		MongoDB:         db,
		LocationService: locationService,
	}
}

// GetAirport godoc
// @Summary Get airport information
// @Description Get details about a specific airport
// @Tags airports
// @Produce json
// @Param code path string true "Airport Code (IATA)"
// @Success 200 {object} models.Airport
// @Router /api/airports/{code} [get]
func (h *AirportController) GetAirport(c *gin.Context) {
	code := c.Param("code")

	if !utils.IsValidAirportCode(code) {
		utils.ErrorResponse(c, 400, "Invalid airport code")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var airport models.Airport
	err := h.MongoDB.Airports().FindOne(ctx, bson.M{"code": code}).Decode(&airport)
	if err != nil {
		utils.ErrorResponse(c, 404, "Airport not found")
		return
	}

	utils.SuccessResponse(c, 200, "Airport information retrieved", airport)
}

// GetSecurityWaitTime godoc
// @Summary Get TSA wait time
// @Description Get current security checkpoint wait time for an airport
// @Tags airports
// @Produce json
// @Param code path string true "Airport Code (IATA)"
// @Success 200 {object} models.SecurityWaitTime
// @Router /api/airports/{code}/security-wait [get]
func (h *AirportController) GetSecurityWaitTime(c *gin.Context) {
	code := c.Param("code")

	if !utils.IsValidAirportCode(code) {
		utils.ErrorResponse(c, 400, "Invalid airport code")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	waitTime, err := h.LocationService.GetSecurityWaitTime(ctx, code)
	if err != nil {
		utils.ErrorResponse(c, 404, "Security wait time not available")
		return
	}

	utils.SuccessResponse(c, 200, "Security wait time retrieved", waitTime)
}
