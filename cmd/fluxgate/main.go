package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/winds18/FluxGate/internal/config"
	"github.com/winds18/FluxGate/internal/httpapi"
	"github.com/winds18/FluxGate/internal/observability"
	"github.com/winds18/FluxGate/internal/stats"
	"github.com/winds18/FluxGate/internal/store"
	"github.com/winds18/FluxGate/internal/upstreamsync"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger, closeLogger, err := observability.NewLogger(cfg.LogDir)
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}
	defer closeLogger()

	db, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		logger.Info("database migrations completed")
		return
	}
	bootstrap, err := db.BootstrapAdmin(ctx, cfg.AdminBootstrapUsername, cfg.AdminBootstrapPassword)
	if err != nil {
		logger.Error("failed to bootstrap admin", "error", err)
		os.Exit(1)
	}
	if bootstrap.Created {
		logger.Info("bootstrap admin created", "username", bootstrap.Admin.Username)
	} else if bootstrap.Skipped {
		logger.Info("bootstrap admin skipped")
	}

	refresher := upstreamsync.Refresher{Store: db}
	go refresher.RunScheduler(ctx, cfg.SourceSyncPollInterval, cfg.SourceSyncBatchLimit, logger)

	if cfg.SingBoxV2RayAPIAddr != "" && cfg.StatsPollInterval > 0 {
		collector := stats.V2RayGRPCCollector{
			Addr:           cfg.SingBoxV2RayAPIAddr,
			QueryPattern:   cfg.SingBoxV2RayStatsPattern,
			RequestTimeout: cfg.SingBoxV2RayAPITimeout,
			DialTimeout:    cfg.SingBoxV2RayAPITimeout,
		}
		poller := stats.Poller{Collector: collector, Recorder: db, Logger: logger}
		go poller.RunScheduler(ctx, cfg.StatsPollInterval)
		logger.Info("stats poller enabled", "interval", cfg.StatsPollInterval.String())
	} else {
		logger.Info("stats poller disabled")
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewServer(cfg, db, logger),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("fluxgate server starting", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("fluxgate server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("fluxgate server shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("fluxgate server stopped")
}
