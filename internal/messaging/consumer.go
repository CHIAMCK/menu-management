package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderEventHandler func(event OrderPlacedEvent)

type OrderConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewOrderConsumer(ctx context.Context, rabbitMQURL, queueName string) (*OrderConsumer, error) {
	if queueName == "" {
		queueName = DefaultOrderQueueName
	}

	conn, err := dialRabbitMQ(ctx, rabbitMQURL)
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

	if err := channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &OrderConsumer{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (c *OrderConsumer) Run(ctx context.Context, handler OrderEventHandler) error {
	deliveries, err := c.channel.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue %q: %w", c.queueName, err)
	}

	slog.Info("order consumer started", "queue", c.queueName)

	for {
		select {
		case <-ctx.Done():
			slog.Info("order consumer shutting down", "queue", c.queueName)
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}

			var event OrderPlacedEvent
			if err := json.Unmarshal(delivery.Body, &event); err != nil {
				slog.Error("failed to unmarshal order placed event", "error", err)
				_ = delivery.Nack(false, false)
				continue
			}

			handler(event)

			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack delivery: %w", err)
			}
		}
	}
}

func (c *OrderConsumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
