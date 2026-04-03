package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

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

func GetRabbitMQURL() string {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	return url
}

func GetRabbitMQExchange() string {
	v := os.Getenv("RABBITMQ_EXCHANGE")
	if v == "" {
		v = "notifications"
	}
	return v
}

func GetRabbitMQMainQueue() string {
	v := os.Getenv("RABBITMQ_MAIN_QUEUE")
	if v == "" {
		v = "notifications.main"
	}
	return v
}

func GetRabbitMQRetryQueue() string {
	v := os.Getenv("RABBITMQ_RETRY_QUEUE")
	if v == "" {
		v = "notifications.retry"
	}
	return v
}

func GetRabbitMQDLQ() string {
	v := os.Getenv("RABBITMQ_DLQ")
	if v == "" {
		v = "notifications.dlq"
	}
	return v
}

func GetRabbitMQMainRoutingKey() string {
	v := os.Getenv("RABBITMQ_MAIN_ROUTING_KEY")
	if v == "" {
		v = GetRabbitMQMainQueue()
	}
	return v
}

func GetRabbitMQRetryRoutingKey() string {
	v := os.Getenv("RABBITMQ_RETRY_ROUTING_KEY")
	if v == "" {
		v = GetRabbitMQRetryQueue()
	}
	return v
}

func GetRabbitMQDLQRoutingKey() string {
	v := os.Getenv("RABBITMQ_DLQ_ROUTING_KEY")
	if v == "" {
		v = GetRabbitMQDLQ()
	}
	return v
}

func GetRabbitMQPrefetch() int {
	v, _ := strconv.Atoi(os.Getenv("RABBITMQ_PREFETCH"))
	if v <= 0 {
		v = 10
	}
	return v
}

func GetRabbitMQRetryDelay() time.Duration {
	v, _ := strconv.Atoi(os.Getenv("RABBITMQ_RETRY_DELAY_MS"))
	if v <= 0 {
		v = 15000
	}
	return time.Duration(v) * time.Millisecond
}

func GetRabbitMQMaxRetries() int {
	v, _ := strconv.Atoi(os.Getenv("RABBITMQ_MAX_RETRIES"))
	if v <= 0 {
		v = 5
	}
	return v
}
