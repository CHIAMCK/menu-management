package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderEventPublisher interface {
	PublishOrderPlaced(ctx context.Context, event OrderPlacedEvent) error
	Close() error
}

type NoOpPublisher struct{}

func (NoOpPublisher) PublishOrderPlaced(_ context.Context, _ OrderPlacedEvent) error {
	return nil
}

func (NoOpPublisher) Close() error {
	return nil
}

type rabbitMQPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewPublisher(ctx context.Context, cfg Config) (OrderEventPublisher, error) {
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

	return &rabbitMQPublisher{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (p *rabbitMQPublisher) PublishOrderPlaced(_ context.Context, event OrderPlacedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal order placed event: %w", err)
	}

	if err := p.channel.Publish(
		"",
		p.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	); err != nil {
		return fmt.Errorf("publish order placed event: %w", err)
	}

	return nil
}

func (p *rabbitMQPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return err
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
