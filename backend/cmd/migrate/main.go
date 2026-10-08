package main

import (
	"context"
	"log/slog"
	"os"

	"awesome-chances/backend/internal/adapters/postgres"
	"awesome-chances/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := migrate(); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database migrations applied")
}

func migrate() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := postgres.Connect(context.Background(), cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Migrate(context.Background())
}
