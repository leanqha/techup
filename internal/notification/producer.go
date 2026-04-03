package notification

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	exchange   string
	routingKey string
	confirmCh  <-chan amqp.Confirmation
	timeout    time.Duration
	mu         sync.Mutex
}

type ProducerOptions struct {
	Topology       TopologyConfig
	PublishTimeout time.Duration
}

func NewProducer(queueName, rabbitMQURL string) (*Producer, error) {
	conn, ch, err := connect(rabbitMQURL, queueName)
	if err != nil {
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &Producer{
		conn:       conn,
		ch:         ch,
		exchange:   "",
		routingKey: queueName,
		confirmCh:  confirmCh,
		timeout:    5 * time.Second,
	}, nil
}

func NewProducerWithOptions(rabbitMQURL string, opts ProducerOptions) (*Producer, error) {
	conn, ch, cfg, err := connectWithTopology(rabbitMQURL, opts.Topology)
	if err != nil {
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	timeout := opts.PublishTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &Producer{
		conn:       conn,
		ch:         ch,
		exchange:   cfg.Exchange,
		routingKey: cfg.MainRoutingKey,
		confirmCh:  confirmCh,
		timeout:    timeout,
	}, nil
}

func (c *Producer) Close() error {
	if err := c.ch.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Producer) Publish(topic string, data []byte) error {
	return c.PublishWithConfirm(context.Background(), topic, data, "")
}

func (c *Producer) PublishWithConfirm(ctx context.Context, topic string, data []byte, messageID string) error {
	if messageID == "" {
		messageID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	publishCtx := ctx
	if c.timeout > 0 {
		var cancel context.CancelFunc
		publishCtx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.ch.PublishWithContext(
		publishCtx,
		c.exchange,
		c.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Type:         topic,
			Body:         data,
			MessageId:    messageID,
			Timestamp:    time.Now().UTC(),
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return err
	}

	select {
	case confirm, ok := <-c.confirmCh:
		if !ok {
			return fmt.Errorf("publisher confirm channel closed")
		}
		if !confirm.Ack {
			return fmt.Errorf("broker nacked message %s", messageID)
		}
		return nil
	case <-publishCtx.Done():
		return publishCtx.Err()
	}
}
