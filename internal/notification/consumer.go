package notification

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const retryCountHeader = "x-retry-count"

type Consumer struct {
	conn            *amqp.Connection
	ch              *amqp.Channel
	queueName       string
	exchange        string
	retryRoutingKey string
	dlqRoutingKey   string
	retryLimit      int
	prefetch        int
	consumerTag     string
}

type ConsumerOptions struct {
	Topology    TopologyConfig
	RetryLimit  int
	Prefetch    int
	ConsumerTag string
}

func NewConsumer(queueName, rabbitMQURL string) (*Consumer, error) {
	conn, ch, err := connect(rabbitMQURL, queueName)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:       conn,
		ch:         ch,
		queueName:  queueName,
		prefetch:   10,
		retryLimit: 5,
	}, nil
}

func NewConsumerWithOptions(rabbitMQURL string, opts ConsumerOptions) (*Consumer, error) {
	conn, ch, cfg, err := connectWithTopology(rabbitMQURL, opts.Topology)
	if err != nil {
		return nil, err
	}

	prefetch := opts.Prefetch
	if prefetch <= 0 {
		prefetch = 10
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	retryLimit := opts.RetryLimit
	if retryLimit <= 0 {
		retryLimit = 5
	}

	return &Consumer{
		conn:            conn,
		ch:              ch,
		queueName:       cfg.MainQueue,
		exchange:        cfg.Exchange,
		retryRoutingKey: cfg.RetryRoutingKey,
		dlqRoutingKey:   cfg.DLQRoutingKey,
		retryLimit:      retryLimit,
		prefetch:        prefetch,
		consumerTag:     opts.ConsumerTag,
	}, nil
}

func (c *Consumer) Close() error {
	if err := c.ch.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	if c.prefetch > 0 {
		if err := c.ch.Qos(c.prefetch, 0, false); err != nil {
			return nil, err
		}
	}

	msgs, err := c.ch.Consume(
		c.queueName,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (c *Consumer) ConsumeWithHandler(ctx context.Context, handler func(context.Context, amqp.Delivery) error) error {
	if handler == nil {
		return errors.New("handler is nil")
	}

	msgs, err := c.Consume()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("delivery channel is closed")
			}
			if err := handler(ctx, msg); err == nil {
				if ackErr := msg.Ack(false); ackErr != nil {
					return ackErr
				}
				continue
			}

			if failErr := c.handleFailure(ctx, msg); failErr != nil {
				_ = msg.Nack(false, true)
				return failErr
			}
		}
	}
}

func (c *Consumer) handleFailure(ctx context.Context, msg amqp.Delivery) error {
	retryCount := headerRetryCount(msg.Headers)
	publishing := amqp.Publishing{
		ContentType:  msg.ContentType,
		Type:         msg.Type,
		Body:         msg.Body,
		Headers:      copyHeaders(msg.Headers),
		MessageId:    msg.MessageId,
		Timestamp:    msg.Timestamp,
		DeliveryMode: amqp.Persistent,
	}

	routingKey := c.retryRoutingKey
	if retryCount >= c.retryLimit {
		routingKey = c.dlqRoutingKey
	} else {
		publishing.Headers[retryCountHeader] = retryCount + 1
	}

	if routingKey == "" {
		return errors.New("routing key is empty")
	}

	if err := c.ch.PublishWithContext(ctx, c.exchange, routingKey, false, false, publishing); err != nil {
		return err
	}

	return msg.Ack(false)
}

func headerRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}

	v, ok := headers[retryCountHeader]
	if !ok {
		return 0
	}

	switch t := v.(type) {
	case int:
		return t
	case int8:
		return int(t)
	case int16:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint:
		return int(t)
	case uint8:
		return int(t)
	case uint16:
		return int(t)
	case uint32:
		return int(t)
	case uint64:
		return int(t)
	default:
		return 0
	}
}

func copyHeaders(headers amqp.Table) amqp.Table {
	if headers == nil {
		return amqp.Table{}
	}

	out := amqp.Table{}
	for k, v := range headers {
		out[k] = v
	}
	return out
}
