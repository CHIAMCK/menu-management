package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const DefaultOrderQueueName = "order.placed"

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

type RabbitMQPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQPublisher(ctx context.Context, rabbitMQURL, queueName string) (*RabbitMQPublisher, error) {
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

	return &RabbitMQPublisher{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (p *RabbitMQPublisher) PublishOrderPlaced(_ context.Context, event OrderPlacedEvent) error {
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

func (p *RabbitMQPublisher) Close() error {
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
