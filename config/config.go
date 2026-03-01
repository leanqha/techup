package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env: %w", err)
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

func GetDomain() string {
	return os.Getenv("DOMAIN")
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
