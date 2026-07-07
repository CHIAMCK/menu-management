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
	channel   *amqp.Channel
	queueName string
}

func newPublisher(client *client) (OrderEventPublisher, error) {
	channel, err := client.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open rabbitmq publisher channel: %w", err)
	}

	return &rabbitMQPublisher{
		channel:   channel,
		queueName: client.queueName,
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
		return p.channel.Close()
	}
	return nil
}
