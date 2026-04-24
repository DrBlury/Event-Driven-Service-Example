package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/internal/usecase"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/logging/metrics"
	"drblury/event-driven-service/pkg/logging/tracing"

	"github.com/drblury/apiweaver/router"
	"github.com/drblury/protoflow"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type ShutdownChannel <-chan os.Signal

type configSections struct {
	fx.Out

	Config    *Config
	Info      *domain.Info
	Router    *router.Config
	Server    *server.Config
	Database  *database.Config
	Logger    *logging.Config
	Tracing   *tracing.Config
	Metrics   *metrics.Config
	Protoflow *protoflow.Config
	Events    *events.Config
}

func splitConfig(cfg *Config) configSections {
	if cfg == nil {
		return configSections{}
	}

	return configSections{
		Config:    cfg,
		Info:      cfg.Info,
		Router:    cfg.Router,
		Server:    cfg.Server,
		Database:  cfg.Database,
		Logger:    cfg.Logger,
		Tracing:   cfg.Tracing,
		Metrics:   cfg.Metrics,
		Protoflow: cfg.Protoflow,
		Events:    cfg.Events,
	}
}

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

func provideLogger(cfg *Config) *slog.Logger {
	return initializeLogger(context.Background(), cfg)
}

func provideFXLogger(logger *slog.Logger) fxevent.Logger {
	if logger == nil {
		return fxevent.NopLogger
	}
	return &fxevent.SlogLogger{Logger: logger}
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

func registerTelemetryHooks(lc fx.Lifecycle, cfg *Config, logger *slog.Logger) {
	if cfg == nil {
		return
	}

	var tracerProvider *sdktrace.TracerProvider
	var meterProvider *sdkmetric.MeterProvider

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cfg.Tracing != nil && cfg.Tracing.Enabled {
				tp, err := tracing.NewTracerProvider(ctx, cfg.Tracing, logger)
				if err != nil {
					logger.Error("failed to initialize tracer", "error", err)
					return err
				}
				tracerProvider = tp
				otel.SetTracerProvider(tp)
			}

			if cfg.Metrics == nil {
				return nil
			}
			if !cfg.Metrics.Enabled {
				logger.Debug("metrics disabled, skipping initialization")
				return nil
			}

			mp, err := metrics.NewMeterProvider(ctx, cfg.Metrics, logger)
			if err != nil {
				logger.Error("failed to initialize metrics", "error", err)
				return err
			}
			meterProvider = mp
			otel.SetMeterProvider(mp)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			var shutdownErr error
			if meterProvider != nil {
				shutdownErr = errors.Join(shutdownErr, meterProvider.Shutdown(ctx))
			}
			if tracerProvider != nil {
				shutdownErr = errors.Join(shutdownErr, tracerProvider.Shutdown(ctx))
			}
			return shutdownErr
		},
	})
}

// connectToDatabase initialises the database connection pool and ensures it is closed on shutdown.
func connectToDatabase(ctx context.Context, cfg *Config, logger *slog.Logger) (*database.Database, error) {
	if cfg == nil || cfg.Database == nil {
		if logger != nil {
			logger.Error("missing database configuration")
		}
		return nil, errors.New("database configuration is required")
	}

	db, err := database.NewDatabase(cfg.Database, logger, ctx)
	if err != nil {
		if logger != nil {
			logger.Error("failed to connect to database", "error", err)
		}
		return nil, err
	}
	return db, nil
}

func provideDatabase(lc fx.Lifecycle, cfg *Config, logger *slog.Logger) (*database.Database, error) {
	db, err := connectToDatabase(context.Background(), cfg, logger)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := db.Close(ctx); err != nil {
				logger.Error("failed to disconnect from database", "error", err)
				return err
			}
			return nil
		},
	})

	return db, nil
}

func provideEventProducer(svc *protoflow.Service) protoflow.Producer {
	return svc
}

// initializeAppLogic constructs the core application use cases.
func initializeAppLogic(db *database.Database, logger *slog.Logger) (*usecase.AppLogic, error) {
	appLogic, err := usecase.NewAppLogic(db, logger)
	if err != nil {
		if logger != nil {
			logger.Error("failed to initialize app logic", "error", err)
		}
		return nil, err
	}
	return appLogic, nil
}

func provideAppLogic(db *database.Database, logger *slog.Logger, cfg *events.Config, producer protoflow.Producer) (*usecase.AppLogic, error) {
	appLogic, err := usecase.NewConfiguredAppLogic(db, logger, cfg, producer)
	if err != nil {
		if logger != nil {
			logger.Error("failed to initialize app logic", "error", err)
		}
		return nil, err
	}
	return appLogic, nil
}

func registerShutdownChannelHook(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *slog.Logger, shutdownChannel ShutdownChannel) {
	if shutdownChannel == nil {
		return
	}

	stopWatcher := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				select {
				case sig, ok := <-shutdownChannel:
					if !ok {
						return
					}
					logger.Info("received external shutdown signal", "signal", sig)
					if err := shutdowner.Shutdown(); err != nil {
						logger.Error("failed to trigger fx shutdown", "error", err)
					}
				case <-stopWatcher:
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			close(stopWatcher)
			return nil
		},
	})
}
