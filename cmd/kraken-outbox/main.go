package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/shenghongio/kraken-ai-platform/internal/bootstrap"
	"github.com/shenghongio/kraken-ai-platform/internal/dispatch"
	"github.com/shenghongio/kraken-ai-platform/orchestrator/persistence/postgres"
	"github.com/shenghongio/kraken-ai-platform/platform/banner"
	"github.com/shenghongio/kraken-ai-platform/platform/shutdown"
)

const (
	serviceName     = "kraken-outbox"
	shutdownTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, configPath, err := bootstrap.LoadConfig(serviceName, "")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := bootstrap.ValidateConfig(cfg, serviceName); err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}
	banner.Print(serviceName, cfg)
	logger := bootstrap.InitLogger(cfg)
	logger.Info("starting", slog.String("service", serviceName),
		slog.String("config_path", configPath),
		slog.String("log_level", cfg.Log.Level),
		slog.Duration("outbox_interval", cfg.OutboxInterval),
		slog.Duration("shutdown_timeout", shutdownTimeout),
	)

	parentCtx, stopParent := context.WithCancel(context.Background())
	defer stopParent()
	ctx, cancelSignal, sigChan := shutdown.Context(parentCtx)
	defer cancelSignal()

	deps, err := bootstrap.InitDeps(ctx, cfg, logger, serviceName)
	if err != nil {
		return fmt.Errorf("init dependencies: %w", err)
	}

	outboxRepo := postgres.NewOutboxRepository(deps.Pool)
	dispatcher := dispatch.NewOutboxDispatcher(outboxRepo, deps.Redis, logger, cfg.OutboxInterval)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("outbox dispatcher started")
		errCh <- dispatcher.Run(ctx)
	}()

	var runErr error
	select {
	case sig := <-sigChan:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			logger.Error("outbox dispatcher error", slog.Any("error", err))
			runErr = fmt.Errorf("outbox dispatcher error: %w", err)
		} else {
			logger.Info("outbox dispatcher stopped")
		}
	}
	stopParent()

	logger.Info("stopped", slog.String("service", serviceName), slog.Duration("shutdown_timeout", shutdownTimeout))
	if err := shutdown.DrainWithTimeout(shutdownTimeout,
		shutdown.Hook{Name: "deps", Fn: func(context.Context) error {
			bootstrap.CloseDeps(deps)
			return nil
		}},
	); err != nil {
		logger.Error("shutdown error", slog.Any("error", err))
	}
	return runErr
}
