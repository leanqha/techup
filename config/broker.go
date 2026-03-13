package config

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func newConnection(amqpURL string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}
	return nil
}

func initBroker(amqpURL, queueName string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := newConnection(amqpURL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	if err := newQueue(ch, queueName); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func NewProducer(amqpURL, queueName string) (*Producer, error) {
	conn, ch, err := initBroker(amqpURL, queueName)
	if err != nil {
		return nil, err
	}

	return &Producer{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

func NewConsumer(amqpURL, queueName string) (*Consumer, error) {
	conn, ch, err := initBroker(amqpURL, queueName)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

func (p *Producer) Close() error {
	if p == nil {
		return nil
	}
	return closeAMQP(p.channel, p.conn)
}

func (c *Consumer) Close() error {
	if c == nil {
		return nil
	}
	return closeAMQP(c.channel, c.conn)
}

func closeAMQP(ch *amqp.Channel, conn *amqp.Connection) error {
	var errs []error
	if ch != nil {
		if err := ch.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
