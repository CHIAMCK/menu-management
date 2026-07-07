package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultDialMaxAttempts = 30
	defaultDialRetryDelay    = 2 * time.Second
)

func dialRabbitMQ(ctx context.Context, rabbitMQURL string) (*amqp.Connection, error) {
	var lastErr error

	for attempt := 1; attempt <= defaultDialMaxAttempts; attempt++ {
		conn, err := amqp.Dial(rabbitMQURL)
		if err == nil {
			if attempt > 1 {
				slog.Info("connected to rabbitmq", "attempt", attempt)
			}
			return conn, nil
		}

		lastErr = err
		if attempt == defaultDialMaxAttempts {
			break
		}

		slog.Warn("rabbitmq not ready, retrying", "attempt", attempt, "error", err)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect to rabbitmq: %w", ctx.Err())
		case <-time.After(defaultDialRetryDelay):
		}
	}

	return nil, fmt.Errorf("connect to rabbitmq after %d attempts: %w", defaultDialMaxAttempts, lastErr)
}
