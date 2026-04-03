package main

import (
	"context"
	"os/signal"
	"syscall"
	"techup/config"
	"techup/internal/logger"
	"techup/internal/notification"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	logger.Init()

	if err := config.LoadEnv(); err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot load environment")
	}

	topology := notification.TopologyConfig{
		Exchange:        config.GetRabbitMQExchange(),
		MainQueue:       config.GetRabbitMQMainQueue(),
		RetryQueue:      config.GetRabbitMQRetryQueue(),
		DLQQueue:        config.GetRabbitMQDLQ(),
		MainRoutingKey:  config.GetRabbitMQMainRoutingKey(),
		RetryRoutingKey: config.GetRabbitMQRetryRoutingKey(),
		DLQRoutingKey:   config.GetRabbitMQDLQRoutingKey(),
		RetryDelay:      config.GetRabbitMQRetryDelay(),
	}

	consumer, err := notification.NewConsumerWithOptions(config.GetRabbitMQURL(), notification.ConsumerOptions{
		Topology:    topology,
		RetryLimit:  config.GetRabbitMQMaxRetries(),
		Prefetch:    config.GetRabbitMQPrefetch(),
		ConsumerTag: "notification-worker",
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot initialize rabbitmq consumer")
	}
	defer func() {
		if closeErr := consumer.Close(); closeErr != nil {
			logger.Log.Error().Err(closeErr).Msg("failed to close consumer")
		}
	}()

	deduplicator := notification.NewInMemoryDeduplicator()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Log.Info().
		Str("queue", topology.MainQueue).
		Int("retry_limit", config.GetRabbitMQMaxRetries()).
		Int("prefetch", config.GetRabbitMQPrefetch()).
		Msg("notification worker started")

	err = consumer.ConsumeWithHandler(ctx, func(ctx context.Context, msg amqp.Delivery) error {
		return notification.HandleWithDedup(ctx, deduplicator, msg, func(ctx context.Context, msg amqp.Delivery) error {
			// TODO: route by msg.Type and call the concrete notifier.
			logger.Log.Info().
				Str("message_id", msg.MessageId).
				Str("type", msg.Type).
				Msg("notification message handled")
			return nil
		})
	})
	if err != nil && err != context.Canceled {
		logger.Log.Fatal().Err(err).Msg("consumer stopped with error")
	}

	logger.Log.Info().Msg("notification worker stopped")
}
