package main

import (
	"context"
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
	"gorm.io/gorm"

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

func setupDatabase(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	if err := db.RunMigrations(databaseURL); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	gormDB, err := db.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	return gormDB, nil
}

func setupMessaging(ctx context.Context) (*messagingComponents, error) {
	components, err := messaging.NewComponents(ctx, messaging.ConfigFromEnv(os.Getenv))
	if err != nil {
		return nil, fmt.Errorf("messaging setup failed: %w", err)
	}

	return &messagingComponents{components: components}, nil
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

	gormDB, err := setupDatabase(context.Background(), cfg.databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := gormDB.DB()
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var workerWG sync.WaitGroup
	msg.components.Consumer.RunInBackground(ctx, logger, &workerWG)

	userLocker := lock.NewInMemoryUserLocker(5 * time.Second)
	router := routes.Setup(gormDB, msg.components.Publisher, userLocker)
	if err := runServer(ctx, cfg.port, router); err != nil {
		log.Fatal(err)
	}

	workerWG.Wait()
}
