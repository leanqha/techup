package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv загружает переменные из .env
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// GetDBDSN формирует строку подключения для PostgreSQL
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

// GetJWTSecret возвращает секретный ключ для JWT
func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

// GetPort возвращает порт для сервера
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // дефолтный порт
	}
	return port
}
