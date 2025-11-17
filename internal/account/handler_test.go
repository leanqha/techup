package account

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type MockService struct{}

func (m *MockService) Register(_ context.Context, req RegisterRequest) (*Account, string, string, error) {
	if req.Email == "fail@example.com" {
		return nil, "", "", errors.New("registration failed")
	}
	return &Account{
		ID:        1,
		UID:       "uid1",
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "student",
	}, "access-token", "refresh-token", nil
}

func (m *MockService) Login(_ context.Context, email, password string) (string, string, error) {
	if email == "fail@example.com" {
		return "", "", errors.New("invalid credentials")
	}
	return "access-token", "refresh-token", nil
}

func (m *MockService) GetByID(ctx context.Context, id int) (*Account, error) {
	if id == 0 {
		return nil, errors.New("not found")
	}
	return &Account{
		ID:        id,
		UID:       "uid1",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "student",
	}, nil
}

func (m *MockService) UpdateProfile(ctx context.Context, userID int, req *UpdateProfileRequest) (*Account, error) {
	return &Account{
		ID:        userID,
		UID:       "uid1",
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "student",
	}, nil
}

func (m *MockService) ChangePassword(ctx context.Context, userID int, req *ChangePasswordRequest) error {
	if req.OldPassword == "wrong" {
		return errors.New("wrong password")
	}
	return nil
}

func (m *MockService) SetRole(ctx context.Context, adminID int, req *SetRoleRequest) error {
	if adminID != 1 {
		return errors.New("forbidden")
	}
	return nil
}

func (m *MockService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	if refreshToken == "invalid" {
		return "", "", errors.New("invalid refresh token")
	}
	return "new-access", "new-refresh", nil
}

func (m *MockService) Logout(ctx context.Context, userID int) error {
	return nil
}

func (m *MockService) DeleteAccount(ctx context.Context, userID int) error {
	if userID == 999 {
		return errors.New("account not found")
	}
	return nil
}

func setupRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	mockSvc := &MockService{}
	h := NewHandler(mockSvc)
	r := gin.New()
	return r, h
}

func TestRegisterHandler(t *testing.T) {
	r, h := setupRouter()
	r.POST("/register", h.Register)

	reqBody := RegisterRequest{
		Email:     "test@example.com",
		Password:  "pass1234",
		FirstName: "John",
		LastName:  "Doe",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем тело ответа
	assert.Contains(t, w.Body.String(), "registration successful")
	assert.Contains(t, w.Body.String(), "test@example.com")

	// Проверяем куки
	cookies := w.Result().Cookies()
	var hasAccess, hasRefresh bool
	for _, c := range cookies {
		if c.Name == "access_token" {
			hasAccess = true
		}
		if c.Name == "refresh_token" {
			hasRefresh = true
		}
	}
	assert.True(t, hasAccess, "access_token cookie should be set")
	assert.True(t, hasRefresh, "refresh_token cookie should be set")
}

func TestRegisterHandlerFail(t *testing.T) {
	r, h := setupRouter()
	r.POST("/register", h.Register)

	reqBody := RegisterRequest{
		Email:     "fail@example.com",
		Password:  "pass1234",
		FirstName: "John",
		LastName:  "Doe",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "registration failed")
}

func TestLoginHandler(t *testing.T) {
	r, h := setupRouter()
	r.POST("/login", h.Login)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "pass1234",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "login successful")
}

func TestProfileHandler(t *testing.T) {
	r, h := setupRouter()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", 1)
		h.Profile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
}

func TestDeleteAccountHandler(t *testing.T) {
	r, h := setupRouter()
	r.DELETE("/account/:id", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{
			"user_id": float64(1), // <- здесь float64
			"role":    "admin",
		})
		h.DeleteAccount(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/account/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "account deleted")
}

func TestDeleteAccountHandlerForbidden(t *testing.T) {
	r, h := setupRouter()
	r.DELETE("/account/:id", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{
			"user_id": float64(2), // <- здесь float64
			"role":    "student",
		})
		h.DeleteAccount(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/account/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "access denied")
}
