package notification

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TopologyConfig struct {
	Exchange        string
	MainQueue       string
	RetryQueue      string
	DLQQueue        string
	MainRoutingKey  string
	RetryRoutingKey string
	DLQRoutingKey   string
	RetryDelay      time.Duration
}

func (c TopologyConfig) withDefaults() TopologyConfig {
	if c.Exchange == "" {
		c.Exchange = "notifications"
	}
	if c.MainQueue == "" {
		c.MainQueue = "notifications.main"
	}
	if c.RetryQueue == "" {
		c.RetryQueue = "notifications.retry"
	}
	if c.DLQQueue == "" {
		c.DLQQueue = "notifications.dlq"
	}
	if c.MainRoutingKey == "" {
		c.MainRoutingKey = c.MainQueue
	}
	if c.RetryRoutingKey == "" {
		c.RetryRoutingKey = c.RetryQueue
	}
	if c.DLQRoutingKey == "" {
		c.DLQRoutingKey = c.DLQQueue
	}
	if c.RetryDelay <= 0 {
		c.RetryDelay = 15 * time.Second
	}

	return c
}

func connect(url, queueName string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func connectWithTopology(url string, cfg TopologyConfig) (*amqp.Connection, *amqp.Channel, TopologyConfig, error) {
	cfg = cfg.withDefaults()

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, TopologyConfig{}, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, TopologyConfig{}, err
	}

	if err := declareTopology(ch, cfg); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, TopologyConfig{}, err
	}

	return conn, ch, cfg, nil
}

func declareTopology(ch *amqp.Channel, cfg TopologyConfig) error {
	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	if _, err := ch.QueueDeclare(cfg.MainQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare main queue: %w", err)
	}

	retryArgs := amqp.Table{
		"x-message-ttl":             int32(cfg.RetryDelay / time.Millisecond),
		"x-dead-letter-exchange":    cfg.Exchange,
		"x-dead-letter-routing-key": cfg.MainRoutingKey,
	}
	if _, err := ch.QueueDeclare(cfg.RetryQueue, true, false, false, false, retryArgs); err != nil {
		return fmt.Errorf("declare retry queue: %w", err)
	}

	if _, err := ch.QueueDeclare(cfg.DLQQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq queue: %w", err)
	}

	if err := ch.QueueBind(cfg.MainQueue, cfg.MainRoutingKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind main queue: %w", err)
	}
	if err := ch.QueueBind(cfg.RetryQueue, cfg.RetryRoutingKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind retry queue: %w", err)
	}
	if err := ch.QueueBind(cfg.DLQQueue, cfg.DLQRoutingKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq queue: %w", err)
	}

	return nil
}
