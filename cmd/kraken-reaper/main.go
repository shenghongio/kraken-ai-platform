package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/shenghongio/kraken-ai-platform/internal/bootstrap"
	"github.com/shenghongio/kraken-ai-platform/orchestrator/lease"
	"github.com/shenghongio/kraken-ai-platform/platform/banner"
	"github.com/shenghongio/kraken-ai-platform/platform/shutdown"
)

const (
	serviceName     = "kraken-reaper"
	shutdownTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	reaperConfig, reaperConfigPath, err := bootstrap.LoadConfig(serviceName, "")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	if err := bootstrap.ValidateConfig(reaperConfig, serviceName); err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}
	banner.Print(serviceName, reaperConfig)
	reaperLogger := bootstrap.InitLogger(reaperConfig)
	reaperLogger.Info("starting kraken-reaper",
		slog.String("config_path", reaperConfigPath),
		slog.String("log_level", reaperConfig.Log.Level),
		slog.Duration("lease_ttl", reaperConfig.LeaseTTL),
		slog.Duration("interval", reaperConfig.LeaseReaperInterval),
		slog.Duration("shutdown_timeout", shutdownTimeout),
	)

	parentCtx, stop := context.WithCancel(context.Background())
	defer stop()
	signalCtx, cancelSignal, sigCh := shutdown.Context(parentCtx)
	defer cancelSignal()

	deps, err := bootstrap.InitDeps(signalCtx, reaperConfig, reaperLogger, serviceName)
	if err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	reaper := lease.NewReaper(deps.Pool, reaperLogger, reaperConfig.LeaseReaperInterval)
	errCh := make(chan error, 1)
	go func() {
		reaperLogger.Info("starting reaper")
		errCh <- reaper.Run(signalCtx)
	}()

	var runErr error
	select {
	case sig := <-sigCh:
		reaperLogger.Info("received shutdown signal", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			reaperLogger.Error("lease reaper error", slog.Any("error", err))
			runErr = fmt.Errorf("lease reaper error: %w", err)
		} else {
			reaperLogger.Info("lease reaper shutdown")
		}
	}
	stop()

	reaperLogger.Info("kraken-reaper shutting down", slog.Duration("shutdown_timeout", shutdownTimeout))
	if err := shutdown.DrainWithTimeout(shutdownTimeout,
		shutdown.Hook{Name: "deps", Fn: func(context.Context) error {
			bootstrap.CloseDeps(deps)
			return nil
		}},
	); err != nil {
		reaperLogger.Error("shutdown error", slog.Any("error", err))
	}
	return runErr
}
