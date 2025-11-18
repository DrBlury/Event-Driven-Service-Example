package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/usecase"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/metrics"
	"drblury/event-driven-service/pkg/tracing"

	events "github.com/drblury/protoflow"
)

// createAppContext builds a cancellable context reacting to OS interrupts and optional external shutdown signals.
func createAppContext(shutdownChannel chan os.Signal) (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	if shutdownChannel != nil {
		go func() {
			select {
			case <-shutdownChannel:
				stop()
			case <-ctx.Done():
			}
		}()
	}
	return ctx, stop
}

// initializeLogger configures the structured logger according to the supplied configuration.
func initializeLogger(ctx context.Context, cfg *Config) *slog.Logger {
	if cfg == nil {
		return logging.SetLogger(ctx)
	}
	return logging.SetLogger(ctx, logging.WithConfig(cfg.Logger))
}

// initializeTracing wires OpenTelemetry tracing when enabled.
func initializeTracing(ctx context.Context, logger *slog.Logger, cfg *Config) error {
	if cfg == nil || cfg.Tracing == nil {
		return nil
	}
	if err := tracing.NewOtelTracer(ctx, logger, cfg.Tracing); err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		return err
	}
	return nil
}

// initializeMetrics sets up metrics exporters and returns a descriptive error when it fails.
func initializeMetrics(ctx context.Context, cfg *Config, logger *slog.Logger) error {
	if cfg == nil || cfg.Metrics == nil {
		return nil
	}
	if !cfg.Metrics.Enabled {
		logger.Debug("metrics disabled, skipping initialization")
		return nil
	}
	if err := metrics.NewOtelMetrics(ctx, cfg.Metrics, logger); err != nil {
		logger.Error("failed to initialize metrics", "error", err)
		return err
	}
	return nil
}

// connectToDatabase initialises the database connection pool.
func connectToDatabase(ctx context.Context, cfg *Config, logger *slog.Logger) (*database.Database, error) {
	if cfg == nil || cfg.Database == nil {
		logger.Error("missing database configuration")
		return nil, errors.New("database configuration is required")
	}

	db, err := database.NewDatabase(cfg.Database, logger, ctx)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return nil, err
	}
	return db, nil
}

// initializeAppLogic constructs the core application use cases.
func initializeAppLogic(db *database.Database, logger *slog.Logger, eventsCfg *events.Config) *usecase.AppLogic {
	return usecase.NewAppLogic(db, logger, eventsCfg)
}
