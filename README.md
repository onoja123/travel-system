# Backend Travel API

This project is a backend API for a travel-related application, built with Go. It provides endpoints and services for managing airports, flights, locations, notifications, authentication, and more. The architecture follows a modular structure with clear separation of concerns.

## Features
- User authentication and authorization (JWT-based)
- Airport, flight, and location management
- Real-time notifications via WebSocket and FCM
- Integration with external aviation APIs
- Redis caching
- RESTful API design
- Middleware for CORS and authentication

## Project Structure
```
cmd/server/main.go           # Application entry point
internal/
  config/                    # Configuration management
  controllers/               # HTTP handlers/controllers
  database/                  # Database and Redis setup
  middleware/                # Middleware (auth, CORS)
  models/                    # Data models
  routes/                    # API route definitions
  services/                  # Business logic and external API integrations
  utils/                     # Utility functions (JWT, validation, response)
  websocket/                 # WebSocket real-time communication
pkg/fcm/                     # Firebase Cloud Messaging integration
```

## Getting Started

### Prerequisites
- Go 1.18+
- Redis server
- (Optional) Firebase Cloud Messaging credentials

### Installation
1. Clone the repository:
   ```sh
   git clone <repo-url>
   cd backend-travel
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Set up environment variables (see `internal/config/config.go` for required variables).
4. Start the server:
   ```sh
   make run
   # or
   go run cmd/server/main.go
   ```

## API Endpoints
- Authentication: `/api/auth/*`
- Airports: `/api/airports/*`
- Flights: `/api/flights/*`
- Locations: `/api/locations/*`
- Notifications: `/api/notifications/*`
- WebSocket: `/ws`

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](LICENSE)
