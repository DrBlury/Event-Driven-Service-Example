package events

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"

	"github.com/drblury/protoflow"
	"go.uber.org/fx"
)

// BuildEventService wires middleware, handlers, and dependencies for the event processing pipeline.
func BuildEventService(
	cfg *Config,
	logger *slog.Logger,
	db *database.Database,
	protoflowCfg *protoflow.Config,
) (*protoflow.Service, error) {
	if cfg == nil || protoflowCfg == nil {
		err := observability.Builder(context.Background(), "events.build_service", "events_config_missing").
			Public("event configuration is required").
			New("events configuration is required")
		logger.Error("missing events configuration", "error", err)
		return nil, err
	}

	middlewares := composeEventMiddlewares(protoflowCfg)

	validator, err := NewValidator()
	if err != nil {
		logger.Error("failed to create proto validator", "error", err)
		return nil, observability.Builder(context.Background(), "events.build_service", "validator_init_failed").
			Public("event validation could not be initialized").
			Wrap(err)
	}

	svc := protoflow.NewService(
		protoflowCfg,
		protoflow.NewSlogServiceLogger(logger),
		context.Background(),
		protoflow.ServiceDependencies{
			Outbox:                    db,
			Validator:                 validator,
			DisableDefaultMiddlewares: true,
			Middlewares:               middlewares,
		},
	)

	if err := registerAppEventHandlers(svc, cfg); err != nil {
		wrapped := observability.Builder(context.Background(), "events.build_service", "handler_registration_failed").
			Public("event handlers could not be registered").
			Wrap(err)
		logger.Error("failed to register event handlers", "error", wrapped)
		return nil, wrapped
	}

	return svc, nil
}

// composeEventMiddlewares returns the middleware chain enforced by this application.
func composeEventMiddlewares(cfg *protoflow.Config) []protoflow.MiddlewareRegistration {
	retryConfig := protoflow.RetryMiddlewareConfig{
		MaxRetries:      cfg.RetryMaxRetries,
		InitialInterval: cfg.RetryInitialInterval,
		MaxInterval:     cfg.RetryMaxInterval,
	}

	return []protoflow.MiddlewareRegistration{
		protoflow.CorrelationIDMiddleware(),
		protoflow.LogMessagesMiddleware(nil),
		protoflow.ProtoValidateMiddleware(),
		protoflow.OutboxMiddleware(),
		protoflow.TracerMiddleware(),
		protoflow.RetryMiddleware(retryConfig),
		protoflow.PoisonQueueMiddleware(poisonQueueFilter()),
		protoflow.RecovererMiddleware(),
	}
}

// poisonQueueFilter decides when an event should be redirected to the poison queue.
func poisonQueueFilter() func(error) bool {
	return func(err error) bool {
		// Check for UnprocessableEventError struct type
		var unprocessable *protoflow.UnprocessableEventError
		if errors.As(err, &unprocessable) {
			return true
		}
		// Check for ErrUnprocessable sentinel error
		if errors.Is(err, protoflow.ErrUnprocessable) {
			return true
		}
		var validationErr domain.ErrValidations
		return errors.As(err, &validationErr)
	}
}

func RegisterLifecycle(lc fx.Lifecycle, svc *protoflow.Service, logger *slog.Logger, cfg *Config) {
	if svc == nil {
		return
	}

	var (
		runCtx context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup
	)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			runCtx, cancel = newLifecycleContext()
			logEventServiceStartup(logger, svc)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := svc.Start(runCtx); err != nil && !errors.Is(err, context.Canceled) {
					logger.Error("event service stopped", "error", observability.Builder(runCtx, "events.lifecycle", "event_service_stopped").Wrap(err))
				}
			}()

			if cfg != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					RunExampleSimulation(runCtx, svc, cfg)
				}()
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel != nil {
				cancel()
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}

func newLifecycleContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// StartEventService runs the event consumer loop until the context is cancelled.
func StartEventService(ctx context.Context, svc *protoflow.Service, logger *slog.Logger) {
	if svc == nil {
		return
	}
	logEventServiceStartup(logger, svc)

	if err := svc.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("event service stopped", "error", observability.Builder(ctx, "events.lifecycle", "event_service_stopped").Wrap(err))
	}
}

// logEventServiceStartup records the event service configuration used at runtime.
func logEventServiceStartup(logger *slog.Logger, svc *protoflow.Service) {
	if svc == nil || svc.Conf == nil {
		return
	}

	logger.With(
		"config", svc.Conf,
	).Info("starting event service")
}
