package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Firebase FirebaseConfig
	Aviation AviationConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	URL      string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type FirebaseConfig struct {
	CredentialsPath string
}

type AviationConfig struct {
	Provider            string
	AviationStackKey    string
	FlightAwareKey      string
	AmadeusClientID     string
	AmadeusClientSecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	expiryDuration, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "travel_companion"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-this-secret"),
			Expiry: expiryDuration,
		},
		Firebase: FirebaseConfig{
			CredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./firebase-credentials.json"),
		},
		Aviation: AviationConfig{
			Provider:            getEnv("AVIATION_API_PROVIDER", "aviationstack"),
			AviationStackKey:    getEnv("AVIATIONSTACK_API_KEY", ""),
			FlightAwareKey:      getEnv("FLIGHTAWARE_API_KEY", ""),
			AmadeusClientID:     getEnv("AMADEUS_CLIENT_ID", ""),
			AmadeusClientSecret: getEnv("AMADEUS_CLIENT_SECRET", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
