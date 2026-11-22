package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/zagvozdeen/again/internal/config"
	"github.com/zagvozdeen/again/internal/httpx"
	"github.com/zagvozdeen/again/internal/mcpserver"
	"github.com/zagvozdeen/again/internal/store"
)

type API struct {
	cfg   *config.Config
	log   *slog.Logger
	store *store.Store
}

func NewApplication(cfg *config.Config, log *slog.Logger, storage *store.Store) *API {
	return &API{cfg: cfg, log: log, store: storage}
}

func (a *API) Run(ctx context.Context) error {
	addr := net.JoinHostPort(a.cfg.App.Host, a.cfg.App.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	server := &http.Server{
		Handler:           a.Handler(),
		ErrorLog:          slog.NewLogLogger(a.log.Handler(), slog.LevelError),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		a.log.InfoContext(ctx, "Server started", slog.String("addr", addr))
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		err = server.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	case err = <-errCh:
		return fmt.Errorf("failed to serve: %w", err)
	}

	a.log.InfoContext(ctx, "Application has successfully stopped")
	return nil
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/mcp", mcpserver.New(a.log, a.store))

	return a.withAccessLog(a.withRecover(mux))
}

func (a *API) handle(mux *http.ServeMux, pattern string, h httpx.Handler) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			httpx.WriteError(w, r, a.log, err)
		}
	})
}
