package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/onoja123/travel-companion-backend/internal/config"
	"github.com/onoja123/travel-companion-backend/internal/utils"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, 401, "Authorization header required")
			c.Abort()
			return
		}

		// Expected format: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, 401, "Invalid authorization format")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
			utils.ErrorResponse(c, 401, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
