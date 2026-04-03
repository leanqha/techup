package account_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"techup/config"
	"techup/internal/account"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var svc *account.Service
var repo *account.Repository

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Printf("No .env.test file found: %v", err)
	}

	if os.Getenv("DB_HOST") == "db" {
		_ = os.Setenv("DB_HOST", "localhost")
	}

	db, err := config.NewPostgresPool()
	if err != nil {
		fmt.Println("Failed to connect to DB:", err)
		os.Exit(1)
	}
	defer db.Close()

	repo = account.NewRepository(db)
	svc = account.NewService(repo)

	os.Exit(m.Run())
}

func TestRegisterLoginRefreshLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := account.NewHandler(svc)
	r := gin.New()
	r.POST("/api/v1/account/register", h.Register)
	r.POST("/api/v1/account/login", h.Login)
	r.POST("/api/v1/account/refresh", h.Refresh)
	r.POST("/api/v1/account/logout", account.AuthMiddleware(), h.Logout)

	email := fmt.Sprintf("local.account.%d@example.com", time.Now().UnixNano())
	password := "password123"
	firstName := "Local"
	lastName := "Smoke"

	// -----------------------------
	// REGISTER
	// -----------------------------
	regBody := fmt.Sprintf(`{
        "email": "%s",
        "password": "%s",
        "first_name": "%s",
        "last_name": "%s"
			}`, email, password, firstName, lastName)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/account/register",
		bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies, "cookies should be set on registration")

	registerAccessToken, registerRefreshToken := extractTokensFromResponse(w.Result())
	require.NotEmpty(t, registerAccessToken)
	require.NotEmpty(t, registerRefreshToken)

	// -----------------------------
	// LOGIN
	// -----------------------------
	loginBody := fmt.Sprintf(`{
        "email":"%s",
        "password":"%s"
    }`, email, password)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/account/login",
		bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		assert.Contains(t, w.Body.String(), "login successful")
	}

	// -----------------------------
	// REFRESH TOKENS
	// -----------------------------
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/account/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: registerRefreshToken})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	newAccessToken, newRefreshToken := extractTokensFromResponse(w.Result())
	require.NotEmpty(t, newAccessToken)
	require.NotEmpty(t, newRefreshToken)

	// -----------------------------
	// LOGOUT
	// -----------------------------
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/account/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: newAccessToken})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: newRefreshToken})

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func extractTokensFromResponse(resp *http.Response) (string, string) {
	var accessToken, refreshToken string

	for _, c := range resp.Cookies() {
		switch c.Name {
		case "access_token":
			accessToken = c.Value
		case "refresh_token":
			refreshToken = c.Value
		}
	}

	if accessToken != "" || refreshToken != "" {
		return accessToken, refreshToken
	}

	for _, raw := range resp.Header.Values("Set-Cookie") {
		parts := strings.SplitN(raw, ";", 2)
		if len(parts) == 0 {
			continue
		}
		kv := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "access_token":
			accessToken = kv[1]
		case "refresh_token":
			refreshToken = kv[1]
		}
	}

	return accessToken, refreshToken
}

func TestLoginWithWrongPassword(t *testing.T) {
	_ = gofakeit.Seed(0)
	ctx := context.Background()
	email := gofakeit.Email()

	_, _, err := svc.Login(ctx, email, "wrongpassword")
	assert.Error(t, err)
}

func TestCleanup(t *testing.T) {
	_ = gofakeit.Seed(0)
	ctx := context.Background()
	email := gofakeit.Email()

	acc, err := repo.GetByEmail(ctx, email)
	if err == nil {
		_ = repo.DeleteRefreshTokens(ctx, acc.ID)
		_ = repo.DeleteAccount(ctx, acc.ID)
	}
}
