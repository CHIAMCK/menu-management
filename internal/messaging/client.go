package messaging

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type client struct {
	conn      *amqp.Connection
	queueName string
}

func newClient(ctx context.Context, cfg Config) (*client, error) {
	queueName := cfg.QueueName
	if queueName == "" {
		queueName = DefaultOrderQueueName
	}

	conn, err := dialRabbitMQ(ctx, cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	if _, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue %q: %w", queueName, err)
	}

	if err := channel.Close(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("close setup channel: %w", err)
	}

	return &client{
		conn:      conn,
		queueName: queueName,
	}, nil
}

type Components struct {
	Publisher OrderEventPublisher
	Consumer  *Consumer
	client    *client
}

func NewComponents(ctx context.Context, cfg Config) (*Components, error) {
	rabbitClient, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	publisher, err := newPublisher(rabbitClient)
	if err != nil {
		_ = rabbitClient.Close()
		return nil, err
	}

	consumer, err := newConsumer(rabbitClient)
	if err != nil {
		_ = publisher.Close()
		_ = rabbitClient.Close()
		return nil, err
	}

	return &Components{
		Publisher: publisher,
		Consumer:  consumer,
		client:    rabbitClient,
	}, nil
}

func (c *Components) Close() error {
	var firstErr error

	if c.Publisher != nil {
		if err := c.Publisher.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if c.Consumer != nil {
		if err := c.Consumer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if c.client != nil {
		if err := c.client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
