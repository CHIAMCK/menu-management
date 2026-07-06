package main

import (
	"context"
	"log"
	"os"

	"menu-management/internal/db"
	"menu-management/internal/routes"
)

func main() {
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

	r := routes.Setup(sqlDB)

	log.Printf("menu management server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
