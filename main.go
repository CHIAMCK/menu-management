package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"menu-management/internal/db"
	"menu-management/internal/messaging"
	"menu-management/internal/routes"
)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/menu_management?sslmode=disable"
	}

	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	sqlDB, err := db.Connect(context.Background())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer sqlDB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	publisher, err := newOrderEventPublisher(context.Background())
	if err != nil {
		log.Fatalf("order event publisher setup failed: %v", err)
	}
	defer publisher.Close()

	r := routes.Setup(sqlDB, publisher)

	log.Printf("menu management server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func newOrderEventPublisher(ctx context.Context) (messaging.OrderEventPublisher, error) {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	queueName := os.Getenv("ORDER_QUEUE_NAME")
	if queueName == "" {
		queueName = messaging.DefaultOrderQueueName
	}

	return messaging.NewRabbitMQPublisher(ctx, rabbitMQURL, queueName)
}
