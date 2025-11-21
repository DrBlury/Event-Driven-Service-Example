package events

import (
	"context"
	"errors"
	"log/slog"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/usecase"

	"github.com/drblury/protoflow"
)

// BuildEventService wires middleware, handlers, and dependencies for the event processing pipeline.
func BuildEventService(
	ctx context.Context,
	cfg *Config,
	logger *slog.Logger,
	db *database.Database,
	appLogic *usecase.AppLogic,
	protoflowCfg *protoflow.Config,
) (*protoflow.Service, error) {
	if cfg == nil || protoflowCfg == nil {
		logger.Error("missing events configuration")
		return nil, errors.New("events configuration is required")
	}

	middlewares := composeEventMiddlewares(protoflowCfg)

	validator, err := NewValidator()
	if err != nil {
		logger.Error("failed to create proto validator", "error", err)
		return nil, err
	}

	svc := protoflow.NewService(
		protoflowCfg,
		protoflow.NewSlogServiceLogger(logger),
		ctx,
		protoflow.ServiceDependencies{
			Outbox:                    db,
			Validator:                 validator,
			DisableDefaultMiddlewares: true,
			Middlewares:               middlewares,
		},
	)

	if err := registerAppEventHandlers(svc, cfg); err != nil {
		logger.Error("failed to register event handlers", "error", err)
		return nil, err
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
		var unprocessable *protoflow.UnprocessableEventError
		if errors.As(err, &unprocessable) {
			return true
		}
		var validationErr domain.ErrValidations
		return errors.As(err, &validationErr)
	}
}

// StartEventService runs the event consumer loop until the context is cancelled.
func StartEventService(ctx context.Context, svc *protoflow.Service, logger *slog.Logger) {
	if svc == nil {
		return
	}
	logEventServiceStartup(logger, svc)

	if err := svc.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("event service stopped", "error", err)
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
