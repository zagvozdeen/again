package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/zagvozdeen/again/internal/config"
)

//go:embed migrations
var migrations embed.FS

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	connCfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Database,
	))
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	err = migrate(ctx, cfg, log, pool)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("run postgres migrations: %w", err)
	}
	return pool, nil
}

func migrate(ctx context.Context, cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) (err error) {
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys, goose.WithSlog(log), goose.WithVerbose(true))
	if err != nil {
		return err
	}
	if cfg.App.DownMigrations && !cfg.App.IsProduction {
		if _, err = provider.Down(ctx); err != nil {
			return err
		}
	}
	_, err = provider.Up(ctx)
	return err
}
