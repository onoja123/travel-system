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

type NotificationHandler struct {
	NotificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		NotificationService: notificationService,
	}
}

// GetNotifications godoc
// @Summary Get user notifications
// @Description Get all notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Success 200 {object} utils.Response
// @Router /api/notifications/{userId} [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.Param("userId")
	authenticatedUserID := c.GetString("user_id")

	if userID != authenticatedUserID {
		utils.ErrorResponse(c, 403, "Forbidden")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(userID)
	notifications, err := h.NotificationService.GetUserNotifications(ctx, objID)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch notifications")
		return
	}

	utils.SuccessResponse(c, 200, "Notifications retrieved", notifications)
}

// UpdatePreferences godoc
// @Summary Update notification preferences
// @Description Update user's notification settings
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.NotificationPreferencesRequest true "Preferences"
// @Success 200 {object} utils.Response
// @Router /api/notifications/preferences [post]
func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, 401, "Unauthorized")
		return
	}

	var req models.NotificationPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(userID)
	err := h.NotificationService.UpdatePreferences(ctx, objID, req)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to update preferences")
		return
	}

	utils.SuccessResponse(c, 200, "Preferences updated successfully", nil)
}
