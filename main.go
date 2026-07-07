package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"menu-management/internal/db"
	"menu-management/internal/lock"
	"menu-management/internal/messaging"
	"menu-management/internal/routes"
)

type config struct {
	databaseURL string
	port        string
}

type messagingComponents struct {
	components *messaging.Components
}

func loadConfig() config {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/menu_management?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return config{
		databaseURL: databaseURL,
		port:        port,
	}
}

func setupDatabase(ctx context.Context, databaseURL string) (*sql.DB, error) {
	if err := db.RunMigrations(databaseURL); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	sqlDB, err := db.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	return sqlDB, nil
}

func setupMessaging(ctx context.Context) (*messagingComponents, error) {
	components, err := messaging.NewComponents(ctx, messaging.ConfigFromEnv(os.Getenv))
	if err != nil {
		return nil, fmt.Errorf("messaging setup failed: %w", err)
	}

	return &messagingComponents{components: components}, nil
}

func setupRedis(ctx context.Context) (*redis.Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}

func runServer(ctx context.Context, port string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("menu management server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func main() {
	cfg := loadConfig()

	sqlDB, err := setupDatabase(context.Background(), cfg.databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	msg, err := setupMessaging(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer msg.components.Close()

	redisClient, err := setupRedis(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var workerWG sync.WaitGroup
	msg.components.Consumer.RunInBackground(ctx, logger, &workerWG)

	userLocker := lock.NewRedisUserLocker(redisClient, 5*time.Second)
	router := routes.Setup(sqlDB, msg.components.Publisher, userLocker)
	if err := runServer(ctx, cfg.port, router); err != nil {
		log.Fatal(err)
	}

	workerWG.Wait()
}
