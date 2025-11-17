package account_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"techup/config"
	"techup/internal/account"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var svc *account.Service
var repo *account.Repository

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Printf("No .env.test file found: %v", err)
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
	ctx := context.Background()

	email := "integration_test@example.com"
	password := "password123"

	if acc, _ := repo.GetByEmail(ctx, email); acc != nil {
		_ = repo.DeleteRefreshTokens(ctx, acc.ID)
		_ = repo.DeleteAccount(ctx, acc.ID)
	}

	h := account.NewHandler(svc)
	r := gin.New()
	r.POST("/api/v1/account/register", h.Register)
	r.POST("/api/v1/account/login", h.Login)
	r.POST("/api/v1/account/refresh", h.Refresh)
	r.POST("/api/v1/account/logout", account.AuthMiddleware(), h.Logout)

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &http.Client{}
	var cookies []*http.Cookie

	registerBody := map[string]string{
		"email":      email,
		"password":   password,
		"first_name": "Test",
		"last_name":  "User",
	}
	b, _ := json.Marshal(registerBody)
	resp, _ := client.Post(ts.URL+"/api/v1/account/register", "application/json", bytes.NewBuffer(b))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookies = resp.Cookies()
	assert.NotEmpty(t, cookies, "cookies should be set on registration")

	accessToken, refreshToken := "", ""
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessToken = c.Value
		}
		if c.Name == "refresh_token" {
			refreshToken = c.Value
		}
	}
	assert.NotEmpty(t, accessToken, "access_token cookie should be set")
	assert.NotEmpty(t, refreshToken, "refresh_token cookie should be set")

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	b, _ = json.Marshal(loginBody)
	resp, _ = client.Post(ts.URL+"/api/v1/account/login", "application/json", bytes.NewBuffer(b))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookies = resp.Cookies()
	assert.NotEmpty(t, cookies, "cookies should be set on login")
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessToken = c.Value
		}
		if c.Name == "refresh_token" {
			refreshToken = c.Value
		}
	}
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/account/refresh", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	newAccessToken, newRefreshToken := "", ""
	for _, c := range resp.Cookies() {
		if c.Name == "access_token" {
			newAccessToken = c.Value
		}
		if c.Name == "refresh_token" {
			newRefreshToken = c.Value
		}
	}
	assert.NotEmpty(t, newAccessToken, "new access_token cookie should be set")
	assert.NotEmpty(t, newRefreshToken, "new refresh_token cookie should be set")

	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/account/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: newAccessToken})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: newRefreshToken})
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginWithWrongPassword(t *testing.T) {
	ctx := context.Background()
	email := "integration_test@example.com"

	_, _, err := svc.Login(ctx, email, "wrongpassword")
	assert.Error(t, err)
}

func TestCleanup(t *testing.T) {
	ctx := context.Background()
	email := "integration_test@example.com"

	acc, err := repo.GetByEmail(ctx, email)
	if err == nil {
		_ = repo.DeleteRefreshTokens(ctx, acc.ID)
		_ = repo.DeleteAccount(ctx, acc.ID)
	}
}
