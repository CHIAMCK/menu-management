package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"menu-management/internal/messaging"
)

func main() {
	_ = godotenv.Load()

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("ORDER_QUEUE_NAME")
	if queueName == "" {
		queueName = messaging.DefaultOrderQueueName
	}

	consumer, err := messaging.NewOrderConsumer(context.Background(), rabbitMQURL, queueName)
	if err != nil {
		log.Fatalf("order consumer setup failed: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	log.Printf("order worker listening on queue %q", queueName)
	if err := consumer.Run(ctx, func(event messaging.OrderPlacedEvent) {
		messaging.LogKitchenDisplayNotification(logger, event)
	}); err != nil && ctx.Err() == nil {
		log.Fatalf("order consumer stopped with error: %v", err)
	}
}
