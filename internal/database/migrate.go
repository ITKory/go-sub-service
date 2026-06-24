package database

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Apply migrations on start
func RunMigrations(databaseURL string, logger *slog.Logger) error {
	instance, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer instance.Close()

	if err := instance.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("database schema is up to date")
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	logger.Info("migrations applied")
	return nil
}
