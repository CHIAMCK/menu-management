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
	channel   *amqp.Channel
	queueName string
}

func newConsumer(client *client) (*Consumer, error) {
	channel, err := client.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open rabbitmq consumer channel: %w", err)
	}

	if err := channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Consumer{
		channel:   channel,
		queueName: client.queueName,
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
		return c.channel.Close()
	}
	return nil
}
