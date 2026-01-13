package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/onoja123/travel-companion-backend/internal/controllers"
)

// RegisterRoutes registers all API routes to the Gin router
func RegisterRoutes(router *gin.Engine,
	airportController *handlers.AirportController,
	authController *handlers.AuthHandler,
	flightController *handlers.FlightHandler,
	locationController *handlers.LocationHandler,
	notificationController *handlers.NotificationHandler,
) {
	// Auth routes
	router.POST("/api/auth/register", authController.Register)
	router.POST("/api/auth/login", authController.Login)
	router.POST("/api/auth/fcm-token", authController.UpdateFCMToken)

	// Airport routes
	router.GET("/api/airports/:code", airportController.GetAirport)
	router.GET("/api/airports/:code/security-wait", airportController.GetSecurityWaitTime)

	// Flight routes
	router.POST("/api/flights/track", flightController.TrackFlight)
	router.GET("/api/flights/user/:userId", flightController.GetUserFlights)
	router.GET("/api/flights/status/:flightNumber/:date", flightController.GetFlightStatus)
	router.DELETE("/api/flights/:id", flightController.DeleteTrackedFlight)

	// Location routes
	router.POST("/api/location/update", locationController.UpdateLocation)
	router.GET("/api/location/walk-time/:flightId", locationController.GetWalkTime)

	// Notification routes
	router.GET("/api/notifications/:userId", notificationController.GetNotifications)
	router.POST("/api/notifications/preferences", notificationController.UpdatePreferences)
}
