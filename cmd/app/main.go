package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	_ "subscription-service/docs"

	"github.com/joho/godotenv"

	"subscription-service/internal/config"
	"subscription-service/internal/database"
	"subscription-service/internal/dependencies"
)

// @title           Subscription Service API
// @version         1.0
// @description     API for managing subscriptions
// @host            xgw8so00sggcg88ogok0s80c.95.79.96.242.sslip.io
// @BasePath        /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env not found, using environment variables")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPostgresDB(ctx, cfg.DatabaseURL(), logger)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(cfg.DatabaseURL(), logger); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	srv := dependencies.WireDependencies(pool, cfg, logger)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
