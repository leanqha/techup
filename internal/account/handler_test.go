package account

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type MockService struct {
	registerErr error
	loginErr    error
	getByIDErr  error
	updateErr   error
	changeErr   error
	setRoleErr  error
	refreshErr  error
	logoutErr   error
	deleteErr   error
}

func (m *MockService) Register(_ context.Context, req RegisterRequest) (*Account, string, string, error) {
	if m.registerErr != nil {
		return nil, "", "", m.registerErr
	}
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
	if m.loginErr != nil {
		return "", "", m.loginErr
	}
	if email == "fail@example.com" {
		return "", "", errors.New("invalid credentials")
	}
	return "access-token", "refresh-token", nil
}

func (m *MockService) GetByID(ctx context.Context, id int) (*Account, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
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
	if m.updateErr != nil {
		return nil, m.updateErr
	}
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
	if m.changeErr != nil {
		return m.changeErr
	}
	if req.OldPassword == "wrong" {
		return errors.New("wrong password")
	}
	return nil
}

func (m *MockService) SetRole(ctx context.Context, adminID int, req *SetRoleRequest) error {
	if m.setRoleErr != nil {
		return m.setRoleErr
	}
	if adminID != 1 {
		return errors.New("forbidden")
	}
	return nil
}

func (m *MockService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	if m.refreshErr != nil {
		return "", "", m.refreshErr
	}
	if refreshToken == "invalid" {
		return "", "", errors.New("invalid refresh token")
	}
	return "new-access", "new-refresh", nil
}

func (m *MockService) Logout(ctx context.Context, userID int) error {
	if m.logoutErr != nil {
		return m.logoutErr
	}
	return nil
}

func (m *MockService) DeleteAccount(ctx context.Context, userID int) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if userID == 999 {
		return errors.New("account not found")
	}
	return nil
}

func setupRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	gofakeit.Seed(0)
	mockSvc := &MockService{}
	h := NewHandler(mockSvc)
	r := gin.New()
	return r, h
}

func setupRouterWithService(svc ServiceInterface) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	r := gin.New()
	return r, h
}

func TestRegisterHandler(t *testing.T) {
	r, h := setupRouter()
	r.POST("/register", h.Register)

	reqBody := RegisterRequest{
		Email:     gofakeit.Email(),
		Password:  "pass1234",
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем тело ответа
	assert.Contains(t, w.Body.String(), "registration successful")
	assert.Contains(t, w.Body.String(), reqBody.Email)

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
		Password:  gofakeit.Password(true, true, true, false, false, 8),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "registration failed")
}

func TestRegisterHandlerInvalidJSON(t *testing.T) {
	r, h := setupRouter()
	r.POST("/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid input")
}

func TestLoginHandler(t *testing.T) {
	r, h := setupRouter()
	r.POST("/login", h.Login)

	password := gofakeit.Password(true, true, true, false, false, 12)
	reqBody := LoginRequest{
		Email:    gofakeit.Email(),
		Password: password,
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "login successful")
}

func TestLoginHandlerInvalidJSON(t *testing.T) {
	r, h := setupRouter()
	r.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid input")
}

func TestLoginHandlerInvalidCredentials(t *testing.T) {
	r, h := setupRouter()
	r.POST("/login", h.Login)

	reqBody := LoginRequest{
		Email:    "fail@example.com",
		Password: "whatever",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

func TestProfileHandler(t *testing.T) {
	r, h := setupRouter()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", 1)
		// MockService returns fixed email, override here for test
		h.Profile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
}

func TestProfileHandlerNotFound(t *testing.T) {
	r, h := setupRouter()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", 0)
		h.Profile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
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

func TestRefreshMissingCookie(t *testing.T) {
	r, h := setupRouter()
	r.POST("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "refresh token not provided")
}

func TestRefreshInvalidToken(t *testing.T) {
	r, h := setupRouter()
	r.POST("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid refresh token")
}

func TestChangePasswordUnauthorized(t *testing.T) {
	r, h := setupRouter()
	r.POST("/change", h.ChangePassword)

	req := httptest.NewRequest(http.MethodPost, "/change", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "no claims found")
}

func TestChangePasswordInvalidInput(t *testing.T) {
	r, h := setupRouter()
	r.POST("/change", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{"user_id": float64(1)})
		h.ChangePassword(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/change", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid input")
}

func TestUpdateProfileUnauthorized(t *testing.T) {
	r, h := setupRouter()
	r.PUT("/update", h.UpdateProfile)

	req := httptest.NewRequest(http.MethodPut, "/update", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "no claims found")
}

func TestUpdateProfileInvalidInput(t *testing.T) {
	r, h := setupRouter()
	r.PUT("/update", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{"user_id": float64(1)})
		h.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid input")
}

func TestSetRoleForbidden(t *testing.T) {
	r, h := setupRouter()
	r.POST("/set-role", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{"user_id": float64(2)})
		h.SetRole(c)
	})

	payload := `{"user_id":2,"role":"teacher"}`
	req := httptest.NewRequest(http.MethodPost, "/set-role", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestLogoutUnauthorized(t *testing.T) {
	r, h := setupRouter()
	r.POST("/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not logged in")
}

func TestLogoutServiceError(t *testing.T) {
	mockSvc := &MockService{logoutErr: errors.New("logout failed")}
	r, h := setupRouterWithService(mockSvc)
	r.POST("/logout", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{"user_id": float64(1)})
		h.Logout(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "logout failed")
}

func TestDeleteAccountInvalidID(t *testing.T) {
	r, h := setupRouter()
	r.DELETE("/account/:id", func(c *gin.Context) {
		c.Set("claims", jwt.MapClaims{
			"user_id": float64(1),
			"role":    "admin",
		})
		h.DeleteAccount(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/account/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid user id")
}
