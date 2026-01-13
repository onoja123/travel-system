package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onoja123/travel-companion-backend/internal/config"
	"github.com/onoja123/travel-companion-backend/internal/database"
	"github.com/onoja123/travel-companion-backend/internal/models"
	"github.com/onoja123/travel-companion-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	MongoDB *database.MongoDB
	Config  *config.Config
}

func NewAuthHandler(db *database.MongoDB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		MongoDB: db,
		Config:  cfg,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details"
// @Success 201 {object} utils.Response
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingUser models.User
	err := h.MongoDB.Users().FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, 400, "Email already registered")
		return
	}
	if err != mongo.ErrNoDocuments {
		utils.ErrorResponse(c, 500, "Database error")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to hash password")
		return
	}

	// Create user
	user := models.User{
		ID:       primitive.NewObjectID(),
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Preferences: models.UserPreferences{
			NotifyGateChange:   true,
			NotifyBoarding:     true,
			NotifyDelay:        true,
			BoardingReminder40: true,
			BoardingReminder20: true,
			BoardingReminder10: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = h.MongoDB.Users().InsertOne(ctx, user)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to create user")
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Email, h.Config.JWT.Secret, h.Config.JWT.Expiry)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to generate token")
		return
	}

	user.Password = "" // Don't send password back

	utils.SuccessResponse(c, 201, "User registered successfully", gin.H{
		"token": token,
		"user":  user,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user
	var user models.User
	err := h.MongoDB.Users().FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, 401, "Invalid email or password")
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		utils.ErrorResponse(c, 401, "Invalid email or password")
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Email, h.Config.JWT.Secret, h.Config.JWT.Expiry)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to generate token")
		return
	}

	user.Password = "" // Don't send password back

	utils.SuccessResponse(c, 200, "Login successful", gin.H{
		"token": token,
		"user":  user,
	})
}

// UpdateFCMToken godoc
// @Summary Update FCM token
// @Description Update user's Firebase Cloud Messaging token for push notifications
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.FCMTokenRequest true "FCM Token"
// @Success 200 {object} utils.Response
// @Router /api/auth/fcm-token [post]
func (h *AuthHandler) UpdateFCMToken(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, 401, "Unauthorized")
		return
	}

	var req models.FCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(userID)
	_, err := h.MongoDB.Users().UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"fcm_token":  req.FCMToken,
				"updated_at": time.Now(),
			},
		},
	)

	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to update FCM token")
		return
	}

	utils.SuccessResponse(c, 200, "FCM token updated successfully", nil)
}
