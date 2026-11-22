package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zagvozdeen/again/internal/api"
	"github.com/zagvozdeen/again/internal/config"
	"github.com/zagvozdeen/again/internal/db"
	"github.com/zagvozdeen/again/internal/logger"
	"github.com/zagvozdeen/again/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Failed to run application", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log, stop, err := logger.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer stop()

	pool, err := db.New(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()
	storage := store.New(pool, log)

	return api.NewApplication(cfg, log, storage).Run(ctx)
}
