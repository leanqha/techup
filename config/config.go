package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// IsDevMode reports whether the app runs in development (GIN_MODE != release).
func IsDevMode() bool {
	return !strings.EqualFold(strings.TrimSpace(os.Getenv("GIN_MODE")), "release")
}

// CookieSecure returns the Secure flag for auth cookies: off in dev so they
// work over http://localhost, on in production (HTTPS).
func CookieSecure() bool {
	return !IsDevMode()
}

// CookieSameSite returns the SameSite policy for auth cookies. Production uses
// None (cross-site over HTTPS); dev uses Lax, which Secure=false requires and
// which works for the same-origin dev proxy.
func CookieSameSite() http.SameSite {
	if IsDevMode() {
		return http.SameSiteLaxMode
	}
	return http.SameSiteNoneMode
}

func LoadEnv() error {
	envFile := strings.TrimSpace(os.Getenv("ENV_FILE"))
	if envFile == "" {
		envFile = ".env"
	}

	if _, err := os.Stat(envFile); err != nil {
		if os.IsNotExist(err) {
			// In CI/prod we rely on process environment variables.
			return nil
		}
		return fmt.Errorf("failed to access env file %q: %w", envFile, err)
	}

	if err := godotenv.Overload(envFile); err != nil {
		return fmt.Errorf("failed to load env file %q: %w", envFile, err)
	}

	return nil
}

func GetDBDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
}

func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func GetAccessTokenTTLSeconds() int {
	minutes, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TOKEN_TTL"))
	if minutes == 0 {
		minutes = 15
	}
	return minutes * int(time.Minute/time.Second)
}

func GetRefreshTokenTTLSeconds() int {
	minutes, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TOKEN_TTL"))
	if minutes == 0 {
		minutes = 10080
	}
	return minutes * int(time.Minute/time.Second)
}

func GetAppBaseURL() string {
	return os.Getenv("APP_BASE_URL")
}

func GetPasswordResetTokenTTLSeconds() int {
	minutes, _ := strconv.Atoi(os.Getenv("PASSWORD_RESET_TTL_MINUTES"))
	if minutes == 0 {
		minutes = 30
	}
	return minutes * int(time.Minute/time.Second)
}

func GetSMTPHost() string {
	return os.Getenv("SMTP_HOST")
}

func GetSMTPPort() int {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}
	return port
}

func GetSMTPUser() string {
	return os.Getenv("SMTP_USER")
}

func GetSMTPPassword() string {
	return os.Getenv("SMTP_PASS")
}

func GetSMTPFrom() string {
	return os.Getenv("SMTP_FROM")
}

func GetSMTPUseTLS() bool {
	val, _ := strconv.ParseBool(os.Getenv("SMTP_USE_TLS"))
	return val
}

func GetSMTPSkipVerify() bool {
	val, _ := strconv.ParseBool(os.Getenv("SMTP_SKIP_VERIFY"))
	return val
}
