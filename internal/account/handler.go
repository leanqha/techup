package account

import (
	"fmt"
	"net/http"
	"techup/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with role "student"
// @Tags account
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration info"
// @Success 200 {object} map[string]interface{} "Successfully created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/v1/account/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	acc, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": acc.ID, "email": acc.Email})
}

// Login godoc
// @Summary User login
// @Description Authenticates user and returns JWT tokens. Sets access and refresh tokens in HTTP-only cookies.
// @Tags account
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User login info"
// @Success 200 {object} map[string]string "Login successful message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /api/v1/account/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(req)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	fmt.Println(req)

	accessToken, refreshToken, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.SetCookie(
		"access_token",
		accessToken,
		config.GetAccessTokenTTLSeconds(),
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		config.GetRefreshTokenTTLSeconds(),
		"/",
		"",
		false,
		true,
	)
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

// Profile godoc
// @Summary Get user profile
// @Description Returns current user info based on JWT token passed in Authorization header as Bearer token.
// @Tags account
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProfileResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/account/secure/profile [get]
func (h *Handler) Profile(c *gin.Context) {
	claimsRaw, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims found"})
		return
	}

	claims, ok := claimsRaw.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in claims"})
		return
	}
	userID := int(userIDFloat)

	acc, err := h.service.GetByID(c, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	resp := ProfileResponse{
		ID:        acc.ID,
		UID:       acc.UID,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Role:      acc.Role,
	}
	c.JSON(http.StatusOK, resp)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Allows a logged-in user to change their password. Authentication required via Bearer token.
// @Tags account
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Old and new password"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/account/secure/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims found"})
		return
	}
	userClaims := claims.(jwt.MapClaims)
	userID := int(userClaims["user_id"].(float64))

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.service.ChangePassword(c, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Allows a logged-in user to update profile information. Authentication required via Bearer token.
// @Tags account
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update info"
// @Success 200 {object} ProfileResponse "Updated profile"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/account/secure/update [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims found"})
		return
	}
	userClaims := claims.(jwt.MapClaims)
	userID := int(userClaims["user_id"].(float64))

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	acc, err := h.service.UpdateProfile(c, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		ID:        acc.ID,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Role:      acc.Role,
	})
}

// SetRole godoc
// @Summary Set user role
// @Description Allows admin to set the role of another user. Authentication required via Bearer token.
// @Tags account
// @Accept json
// @Produce json
// @Param request body SetRoleRequest true "User ID and new role"
// @Success 200 {object} map[string]string "Role updated successfully"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /api/v1/account/set-role [post]
func (h *Handler) SetRole(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims found"})
		return
	}
	userClaims := claims.(jwt.MapClaims)
	userID := int(userClaims["user_id"].(float64))

	var req SetRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.service.SetRole(c, userID, &req)
	if err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

// Refresh godoc
// @Summary Refresh JWT tokens
// @Description Refreshes access and refresh tokens using a valid refresh token sent in request body.
// @Tags account
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} LoginResponse "New tokens"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/account/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	access, refresh, err := h.service.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}
