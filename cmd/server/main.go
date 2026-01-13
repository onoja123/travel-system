package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/onoja123/travel-companion-backend/internal/config"
	handlers "github.com/onoja123/travel-companion-backend/internal/controllers"
	"github.com/onoja123/travel-companion-backend/internal/database"
	"github.com/onoja123/travel-companion-backend/internal/middleware"
	"github.com/onoja123/travel-companion-backend/internal/routes"
	"github.com/onoja123/travel-companion-backend/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize MongoDB
	db, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	log.Println("✅ Connected to MongoDB")

	// Initialize Redis
	redisClient, err := database.NewRedisClient(cfg.Redis.URL, cfg.Redis.Password)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()
	log.Println("✅ Connected to Redis")

	// Initialize all services in one place
	// Initialize all controllers
	aviationService := services.NewAviationService(cfg)
	notificationService := services.NewNotificationService(db, nil) // FCMService can be added
	flightService := services.NewFlightService(db, redisClient, aviationService, notificationService)
	locationService := services.NewLocationService(db, redisClient)

	airportController := handlers.NewAirportController(db, locationService)
	authController := handlers.NewAuthHandler(db, cfg)
	flightController := handlers.NewFlightHandler(flightService)
	locationController := handlers.NewLocationHandler(locationService)
	notificationController := handlers.NewNotificationHandler(notificationService)

	// Setup Gin router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Register all API routes in a separate function for cleaner code

	// Register all API routes
	routes.RegisterRoutes(router, airportController, authController, flightController, locationController, notificationController)

	// Start server
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := db.Close(); err != nil {
		log.Fatal("Error disconnecting from MongoDB:", err)
	}

	log.Println("Server exited")
}
