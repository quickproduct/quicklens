// Command server runs the consolidated QuickLens backend: a single binary
// serving the embedded SvelteKit dashboard, the REST/WebSocket API, the
// transparent LLM proxy, and the background workers.
package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"quicklens/backend/internal/config"
	"quicklens/backend/internal/db"
	"quicklens/backend/internal/handlers"
	"quicklens/backend/internal/logging"
	"quicklens/backend/internal/router"
	"quicklens/backend/internal/workers"
	"quicklens/backend/internal/ws"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.Env)
	zap.ReplaceGlobals(logger)
	defer func() { _ = logger.Sync() }()

	logger.Info("starting quicklens",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
		zap.String("db", cfg.DatabasePath),
	)

	// ── Persistence ─────────────────────────────────────────────────────────
	if err := db.InitDB(db.Config{
		Path:          cfg.DatabasePath,
		AdminEmail:    cfg.AdminEmail,
		AdminPassword: cfg.AdminPassword,
	}, logger); err != nil {
		logger.Fatal("init database", zap.Error(err))
	}
	db.SeedMockDataIfEmpty(logger)

	// ── Background workers ────────────────────────────────────────────────────
	ws.Manager.StartHeartbeat()
	handlers.StartIngestWorker()
	workers.StartOllamaWorker()
	workers.StartMetricsWorker()
	workers.StartJanitorWorker()

	// ── HTTP server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(logger),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Fatal("server error", zap.Error(err))
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
	logger.Info("shutdown complete")
}
