package main

import (
	"techup/config"
	"techup/internal/logger"
)

func main() {
	logger.Init()

	if err := config.LoadEnv(); err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot load environment")
	}

	// Wiring plan:
	// 1) Build RabbitMQ config from config.GetRabbitMQ* helpers.
	// 2) Create AMQP consumer adapter implementing notification.Consumer.
	// 3) Create router with use-case handlers.
	// 4) Start notification.Worker.Run(ctx, queue).
	logger.Log.Info().Msg("notification worker bootstrap is ready for adapter wiring")
}
