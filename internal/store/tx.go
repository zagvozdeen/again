package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (s *Store) Begin(ctx context.Context) (context.Context, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return context.WithValue(ctx, "tx", tx), nil
}

func (s *Store) Commit(ctx context.Context) error {
	tx, ok := ctx.Value("tx").(pgx.Tx)
	if !ok {
		return errors.New("commit failed: tx not found in context")
	}
	err := tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *Store) Rollback(ctx context.Context) {
	tx, ok := ctx.Value("tx").(pgx.Tx)
	if !ok {
		s.log.ErrorContext(ctx, "Rollback failed: tx not found in context")
		return
	}
	err := tx.Rollback(context.WithoutCancel(ctx))
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		s.log.ErrorContext(ctx, "Failed to rollback transaction", slog.Any("error", err))
	}
}

func (s *Store) querier(ctx context.Context) querier {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return tx
	}
	return s.pool
}
