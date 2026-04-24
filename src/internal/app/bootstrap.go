package app

import (
	"context"
	"errors"
	"log/slog"
	"os"

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

func provideLogger(cfg *logging.Config) *slog.Logger {
	if cfg == nil {
		return logging.SetLogger(context.Background())
	}
	return logging.SetLogger(context.Background(), logging.WithConfig(cfg))
}

func provideFXLogger(cfg *logging.Config) fxevent.Logger {
	return &fxevent.SlogLogger{Logger: provideLogger(cfg)}
}

func registerTelemetryHooks(
	lc fx.Lifecycle,
	tracingCfg *tracing.Config,
	metricsCfg *metrics.Config,
	logger *slog.Logger,
) {
	log := fallbackLogger(logger)

	var tracerProvider *sdktrace.TracerProvider
	var meterProvider *sdkmetric.MeterProvider

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if tracingCfg != nil && tracingCfg.Enabled {
				tp, err := tracing.NewTracerProvider(ctx, tracingCfg, log)
				if err != nil {
					log.Error("failed to initialize tracer", "error", err)
					return err
				}
				tracerProvider = tp
				otel.SetTracerProvider(tp)
			}

			if metricsCfg == nil {
				return nil
			}
			if !metricsCfg.Enabled {
				log.Debug("metrics disabled, skipping initialization")
				return nil
			}

			mp, err := metrics.NewMeterProvider(ctx, metricsCfg, log)
			if err != nil {
				log.Error("failed to initialize metrics", "error", err)
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

func provideDatabase(lc fx.Lifecycle, cfg *database.Config, logger *slog.Logger) (*database.Database, error) {
	log := fallbackLogger(logger)
	if cfg == nil {
		log.Error("missing database configuration")
		return nil, errors.New("database configuration is required")
	}

	db, err := database.NewDatabase(cfg, log, context.Background())
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := db.Close(ctx); err != nil {
				log.Error("failed to disconnect from database", "error", err)
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

func provideAppLogic(
	db *database.Database,
	logger *slog.Logger,
	cfg *events.Config,
	producer protoflow.Producer,
) (*usecase.AppLogic, error) {
	appLogic, err := usecase.NewAppLogic(db, logger, cfg, producer)
	if err != nil {
		fallbackLogger(logger).Error("failed to initialize app logic", "error", err)
		return nil, err
	}
	return appLogic, nil
}

func registerShutdownChannelHook(
	lc fx.Lifecycle,
	shutdowner fx.Shutdowner,
	logger *slog.Logger,
	shutdownChannel ShutdownChannel,
) {
	if shutdownChannel == nil {
		return
	}

	log := fallbackLogger(logger)
	stopWatcher := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				select {
				case sig, ok := <-shutdownChannel:
					if !ok {
						return
					}
					log.Info("received external shutdown signal", "signal", sig)
					if err := shutdowner.Shutdown(); err != nil {
						log.Error("failed to trigger fx shutdown", "error", err)
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

func fallbackLogger(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.Default()
}
