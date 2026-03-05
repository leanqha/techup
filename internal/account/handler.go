package account

import (
	"context"
	"net/http"
	"strconv"
	"strings"
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
	Logout(ctx context.Context, userID int, refreshToken string) error
	DeleteAccount(ctx context.Context, userID int) error
	UpdateAccountAdmin(ctx context.Context, userID int, req *AdminUpdateAccountRequest) (*Account, error)
	ListAccounts(ctx context.Context, f AdminAccountsFilter) ([]Account, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
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
	AdminListAccounts(c *gin.Context)
	AdminUpdateAccount(c *gin.Context)
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(s ServiceInterface) *Handler {
	return &Handler{service: s}
}

// Register godoc
// @Summary      Register a new user account
// @Description  Create a new account with email/password and profile details. Sets HTTP-only cookies (access_token, refresh_token).
// @Tags         Account
// @Accept       json
// @Produce      json
// @Param        register body RegisterRequest true "Registration details"
// @Success      200 {object} map[string]interface{} "Message plus user profile"
// @Failure      400 {object} ErrorResponse "Invalid input or registration error"
// @Router       /account/register [post]
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
// @Summary      Authenticate user
// @Description  Verify credentials and set HTTP-only cookies (access_token, refresh_token). Returns a confirmation message.
// @Tags         Account
// @Accept       json
// @Produce      json
// @Param        login body LoginRequest true "Login credentials"
// @Success      200 {object} map[string]string "Login confirmation message"
// @Failure      400 {object} ErrorResponse "Invalid input"
// @Failure      401 {object} ErrorResponse "Invalid credentials"
// @Router       /account/login [post]
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
// @Summary      Get current user profile
// @Description  Return profile data for the authenticated user.
// @Tags         Account
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200 {object} ProfileResponse "User profile"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /account/secure/profile [get]
func (h *Handler) Profile(c *gin.Context) {
	userID := c.GetInt("user_id")

	acc, err := h.service.GetByID(c, userID)
	if err != nil {
		logger.Log.Err(err).Msg("failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		ID:         acc.ID,
		UID:        acc.UID,
		Email:      acc.Email,
		FirstName:  acc.FirstName,
		MiddleName: acc.MiddleName,
		LastName:   acc.LastName,
		GroupID:    acc.GroupID,
		GroupName:  acc.GroupName,
		Role:       acc.Role,
	})
}

// Refresh godoc
// @Summary      Refresh authentication tokens
// @Description  Issue new access and refresh tokens using the refresh_token cookie; sets new HTTP-only cookies.
// @Tags         Account
// @Produce      json
// @Success      200 {object} map[string]string "Refresh confirmation message"
// @Failure      400 {object} ErrorResponse "Refresh token not provided"
// @Failure      401 {object} ErrorResponse "Invalid or expired refresh token"
// @Router       /account/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token not provided"})
		return
	}

	accessToken, newRefreshToken, err :=
		h.service.RefreshTokens(c.Request.Context(), refreshToken)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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
		Value:    newRefreshToken,
		Path:     "/",
		MaxAge:   config.GetRefreshTokenTTLSeconds(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	c.JSON(http.StatusOK, gin.H{"message": "tokens refreshed"})
}

// ChangePassword godoc
// @Summary      Change user password
// @Description  Change the password for the authenticated user. Requires current and new passwords.
// @Tags         Account
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        change body ChangePasswordRequest true "Password change payload"
// @Success      200 {object} map[string]string "Password change confirmation message"
// @Failure      400 {object} ErrorResponse "Invalid input or password change error"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /account/secure/change-password [post]
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
// @Summary      Update current user profile
// @Description  Update profile fields for the authenticated user and return the updated profile.
// @Tags         Account
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        update body UpdateProfileRequest true "Profile update payload"
// @Success      200 {object} ProfileResponse "Updated user profile"
// @Failure      400 {object} ErrorResponse "Invalid input or update error"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /account/secure/update [put]
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
// @Summary      Set user role (admin only)
// @Description  Change a user's role. Requires admin privileges.
// @Tags         Account
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        role body SetRoleRequest true "Role update payload"
// @Success      200 {object} map[string]string "Role update confirmation message"
// @Failure      400 {object} ErrorResponse "Invalid input or update error"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Router       /admin/set-role [post]
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

// Logout godoc
// @Summary      Logout user
// @Description  Revoke refresh tokens, clear auth cookies, and return a confirmation message.
// @Tags         Account
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200 {object} map[string]string "Logout confirmation message"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Logout error"
// @Router       /account/secure/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
		return
	}

	refreshToken, _ := c.Cookie("refresh_token")

	if err := h.service.Logout(c.Request.Context(), userID, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearAuthCookie(c, "access_token")
	clearAuthCookie(c, "refresh_token")

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func clearAuthCookie(c *gin.Context, name string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

// ForgotPassword godoc
// @Summary      Request password reset
// @Description  Send a password reset link/token to the user's email.
// @Tags         Account
// @Accept       json
// @Produce      json
// @Param        body body ForgotPasswordRequest true "Password reset request"
// @Success      200 {object} map[string]string "Reset request accepted"
// @Failure      400 {object} ErrorResponse "Invalid input"
// @Failure      500 {object} ErrorResponse "Reset request error"
// @Router       /account/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.service.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reset request accepted"})
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset the user's password using a valid reset token.
// @Tags         Account
// @Accept       json
// @Produce      json
// @Param        body body ResetPasswordRequest true "Password reset payload"
// @Success      200 {object} map[string]string "Password reset confirmation"
// @Failure      400 {object} ErrorResponse "Invalid input or token"
// @Failure      500 {object} ErrorResponse "Reset error"
// @Router       /account/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}

// DeleteAccount godoc
// @Summary      Delete a user account
// @Description  Delete an account by ID. Admins can delete any account; users can delete their own.
// @Tags         Account
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]string "Deletion confirmation message"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden"
// @Failure      404 {object} ErrorResponse "Account not found"
// @Failure      500 {object} ErrorResponse "Deletion error"
// @Router       /admin/account/{id} [delete]
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
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}

// AdminListAccounts godoc
// @Summary      List accounts (admin only)
// @Description  Return accounts filtered by optional criteria.
// @Tags         Account
// @Security     ApiKeyAuth
// @Produce      json
// @Param        role        query string false "Role"
// @Param        group_id    query int    false "Group ID"
// @Param        email       query string false "Email (partial match)"
// @Param        uid         query string false "UID (partial match)"
// @Param        name        query string false "Name (partial match)"
// @Param        is_verified query bool   false "Verification status"
// @Success      200 {array} AdminAccountResponse "Accounts list"
// @Failure      400 {object} ErrorResponse "Invalid filter value"
// @Failure      500 {object} ErrorResponse "Failed to load accounts"
// @Router       /admin/accounts [get]
func (h *Handler) AdminListAccounts(c *gin.Context) {
	var f AdminAccountsFilter

	if v := strings.TrimSpace(c.Query("role")); v != "" {
		f.Role = &v
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		f.GroupID = &id
	}
	if v := strings.TrimSpace(c.Query("email")); v != "" {
		f.Email = &v
	}
	if v := strings.TrimSpace(c.Query("uid")); v != "" {
		f.UID = &v
	}
	if v := strings.TrimSpace(c.Query("name")); v != "" {
		f.Name = &v
	}
	if v := strings.TrimSpace(c.Query("is_verified")); v != "" {
		val, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_verified"})
			return
		}
		f.IsVerified = &val
	}

	accounts, err := h.service.ListAccounts(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]AdminAccountResponse, 0, len(accounts))
	for _, acc := range accounts {
		resp = append(resp, AdminAccountResponse{
			ID:         acc.ID,
			UID:        acc.UID,
			Email:      acc.Email,
			FirstName:  acc.FirstName,
			MiddleName: acc.MiddleName,
			LastName:   acc.LastName,
			Role:       acc.Role,
			IsVerified: acc.IsVerified,
			GroupID:    acc.GroupID,
			GroupName:  acc.GroupName,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// AdminUpdateAccount godoc
// @Summary      Update account info (admin only)
// @Description  Update account fields by ID.
// @Tags         Account
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "User ID"
// @Param        body body AdminUpdateAccountRequest true "Account update payload"
// @Success      200 {object} AdminAccountResponse "Updated account"
// @Failure      400 {object} ErrorResponse "Invalid input"
// @Failure      404 {object} ErrorResponse "Account not found"
// @Failure      500 {object} ErrorResponse "Failed to update account"
// @Router       /admin/account/{id} [put]
func (h *Handler) AdminUpdateAccount(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req AdminUpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if req.Email == nil && req.FirstName == nil && req.MiddleName == nil && req.LastName == nil &&
		req.UID == nil && req.Role == nil && req.GroupID == nil && req.IsVerified == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	acc, err := h.service.UpdateAccountAdmin(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AdminAccountResponse{
		ID:         acc.ID,
		UID:        acc.UID,
		Email:      acc.Email,
		FirstName:  acc.FirstName,
		MiddleName: acc.MiddleName,
		LastName:   acc.LastName,
		Role:       acc.Role,
		IsVerified: acc.IsVerified,
		GroupID:    acc.GroupID,
		GroupName:  acc.GroupName,
	})
}
