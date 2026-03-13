package config

import (
	"os"
	"strconv"
	"time"
)

func GetRabbitMQURL() string {
	return os.Getenv("RABBITMQ_URL")
}

func GetRabbitMQQueue() string {
	return os.Getenv("RABBITMQ_QUEUE")
}

func GetRabbitMQRoutingKey() string {
	return os.Getenv("RABBITMQ_ROUTING_KEY")
}

func GetRabbitMQReconnectBaseDelay() time.Duration {
	seconds, _ := strconv.Atoi(os.Getenv("RABBITMQ_RECONNECT_BASE_DELAY_SECONDS"))
	if seconds <= 0 {
		seconds = 2
	}
	return time.Duration(seconds) * time.Second
}

func GetRabbitMQReconnectMaxDelay() time.Duration {
	seconds, _ := strconv.Atoi(os.Getenv("RABBITMQ_RECONNECT_MAX_DELAY_SECONDS"))
	if seconds <= 0 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}
