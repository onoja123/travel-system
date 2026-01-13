package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(uri, database string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("✅ Connected to MongoDB")

	db := client.Database(database)

	return &MongoDB{
		Client:   client,
		Database: db,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

// Collection helpers
func (m *MongoDB) Users() *mongo.Collection {
	return m.Database.Collection("users")
}

func (m *MongoDB) TrackedFlights() *mongo.Collection {
	return m.Database.Collection("tracked_flights")
}

func (m *MongoDB) FlightStatus() *mongo.Collection {
	return m.Database.Collection("flight_status")
}

func (m *MongoDB) Notifications() *mongo.Collection {
	return m.Database.Collection("notifications")
}

func (m *MongoDB) Airports() *mongo.Collection {
	return m.Database.Collection("airports")
}
