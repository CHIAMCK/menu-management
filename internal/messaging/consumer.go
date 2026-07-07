package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewConsumer(ctx context.Context, cfg Config) (*Consumer, error) {
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

	if err := channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Consumer{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (c *Consumer) RunInBackground(ctx context.Context, logger *slog.Logger, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("order worker listening on queue %q", c.queueName)
		if err := c.run(ctx, logger); err != nil && ctx.Err() == nil {
			log.Printf("order consumer stopped with error: %v", err)
		}
	}()
}

func (c *Consumer) run(ctx context.Context, logger *slog.Logger) error {
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

			LogKitchenDisplayNotification(logger, event)

			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack delivery: %w", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
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
