package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
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

func GetJWTAccessTTL() time.Duration {
	minutesStr := os.Getenv("JWT_ACCESS_TOKEN_TTL")
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}

func GetJWTRefreshTTL() time.Duration {
	minutesStr := os.Getenv("JWT_REFRESH_TOKEN_TTL")
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		minutes = 7 * 24 * 60
	}
	return time.Duration(minutes) * time.Minute
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
		minutes = 10080 // 7 days
	}
	return minutes * int(time.Minute/time.Second)
}
