package app

import (
	"context"
	"errors"
	"log/slog"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/usecase"
	"drblury/event-driven-service/pkg/events"
)

// buildEventService wires middleware, handlers, and dependencies for the event processing pipeline.
func buildEventService(
	ctx context.Context,
	cfg *Config,
	logger *slog.Logger,
	db *database.Database,
	appLogic *usecase.AppLogic,
) (*events.Service, error) {
	if cfg == nil || cfg.Events == nil {
		logger.Error("missing events configuration")
		return nil, errors.New("events configuration is required")
	}

	middlewares := composeEventMiddlewares(cfg.Events)

	svc := events.NewService(
		cfg.Events,
		logger,
		ctx,
		events.ServiceDependencies{
			Outbox:                    db,
			Validator:                 appLogic,
			DisableDefaultMiddlewares: true,
			Middlewares:               middlewares,
		},
	)

	if err := registerAppEventHandlers(svc); err != nil {
		logger.Error("failed to register event handlers", "error", err)
		return nil, err
	}

	return svc, nil
}

// composeEventMiddlewares returns the middleware chain enforced by this application.
func composeEventMiddlewares(cfg *events.Config) []events.MiddlewareRegistration {
	retryConfig := events.RetryMiddlewareConfig{
		MaxRetries:      cfg.RetryMaxRetries,
		InitialInterval: cfg.RetryInitialInterval,
		MaxInterval:     cfg.RetryMaxInterval,
	}

	return []events.MiddlewareRegistration{
		events.CorrelationIDMiddleware(),
		events.LogMessagesMiddleware(nil),
		events.ProtoValidateMiddleware(),
		events.OutboxMiddleware(),
		events.TracerMiddleware(),
		events.RetryMiddleware(retryConfig),
		events.PoisonQueueMiddleware(poisonQueueFilter()),
		events.RecovererMiddleware(),
	}
}

// poisonQueueFilter decides when an event should be redirected to the poison queue.
func poisonQueueFilter() func(error) bool {
	return func(err error) bool {
		var unprocessable *events.UnprocessableEventError
		if errors.As(err, &unprocessable) {
			return true
		}
		var validationErr domain.ErrValidations
		return errors.As(err, &validationErr)
	}
}

// startEventService runs the event consumer loop until the context is cancelled.
func startEventService(ctx context.Context, svc *events.Service, logger *slog.Logger) {
	if svc == nil {
		return
	}

	if err := svc.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("event service stopped", "error", err)
	}
}

// logEventServiceStartup records the event service configuration used at runtime.
func logEventServiceStartup(logger *slog.Logger, svc *events.Service) {
	if svc == nil || svc.Conf == nil {
		return
	}

	logger.With(
		"brokers", svc.Conf.KafkaBrokers,
		"consume_queue", svc.Conf.ConsumeQueue,
		"publish_queue", svc.Conf.PublishQueue,
	).Info("starting event service")
}
