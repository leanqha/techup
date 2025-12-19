package account

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"techup/config"
	"techup/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ServiceInterface interface {
	Register(ctx context.Context, req RegisterRequest) (*Account, string, string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	GetByID(ctx context.Context, id int) (*Account, error)
	UpdateProfile(ctx context.Context, userID int, req *UpdateProfileRequest) (*Account, error)
	ChangePassword(ctx context.Context, userID int, req *ChangePasswordRequest) error
	SetRole(ctx context.Context, adminID int, req *SetRoleRequest) error
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID int) error
	DeleteAccount(ctx context.Context, userID int) error
}

type HandlerInterface interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Profile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ChangePassword(c *gin.Context)
	SetRole(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
	DeleteAccount(c *gin.Context)
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(s ServiceInterface) *Handler {
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

	acc, accessToken, refreshToken, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   config.GetAccessTokenTTLSeconds(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   config.GetRefreshTokenTTLSeconds(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	// Response
	c.JSON(http.StatusOK, gin.H{
		"message": "registration successful",
		"user": ProfileResponse{
			ID:        acc.ID,
			UID:       acc.UID,
			Email:     acc.Email,
			FirstName: acc.FirstName,
			LastName:  acc.LastName,
			Role:      acc.Role,
		},
	})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	accessToken, refreshToken, err := h.service.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   config.GetAccessTokenTTLSeconds(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   config.GetRefreshTokenTTLSeconds(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

// Profile godoc
// @Summary Get user profile
// @Description Returns current user info based on JWT token passed in Authorization header as Bearer token.
// @Tags account
// @Produce json
// @Security cookieAuth
// @Success 200 {object} ProfileResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/account/secure/profile [get]
func (h *Handler) Profile(c *gin.Context) {
	userID := c.GetInt("user_id")

	acc, err := h.service.GetByID(c, userID)
	if err != nil {
		logger.Log.Err(err).Msg("failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		ID:        acc.ID,
		UID:       acc.UID,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		GroupName: acc.GroupName,
		Role:      acc.Role,
	})
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
// @Security cookieAuth
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
// @Security cookieAuth
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
// @Security cookieAuth
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
	cookie, err := c.Cookie("refresh_token")
	if err != nil || cookie == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token not provided"})
		return
	}
	refreshToken := cookie

	accessToken, newRefreshToken, err := h.service.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"access_token",
		accessToken,
		config.GetAccessTokenTTLSeconds(),
		"/",
		config.GetDomain(),
		false,
		true,
	)
	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		config.GetRefreshTokenTTLSeconds(),
		"/",
		config.GetDomain(),
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "tokens refreshed"})
}

// Logout godoc
// @Summary Logout user
// @Description Clears access and refresh tokens (HTTP-only cookies)
// @Tags account
// @Produce json
// @Success 200 {object} map[string]string "Logout successful"
// @Router /api/v1/account/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	claimsRaw, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
		return
	}

	claims := claimsRaw.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	if err := h.service.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("access_token", "", -1, "/", config.GetDomain(), false, true)
	c.SetCookie("refresh_token", "", -1, "/", config.GetDomain(), false, true)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	claimsRaw, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims"})
		return
	}
	claims := claimsRaw.(jwt.MapClaims)
	role := claims["role"].(string)
	claimsUserID := int(claims["user_id"].(float64))

	if role != "admin" && claimsUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.service.DeleteAccount(c, userID)
	if err != nil {
		if err == errors.New("account not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
