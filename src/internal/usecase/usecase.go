package usecase

import (
	"context"
	"log/slog"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/internal/observability"

	"github.com/drblury/protoflow"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type AppLogic struct {
	db            *database.Database
	log           *slog.Logger
	eventProducer protoflow.Producer
	exampleTopic  string
}

var tracer = otel.Tracer("drblury/event-driven-service/internal/usecase")

func NewAppLogic(
	db *database.Database,
	logger *slog.Logger,
	eventsCfg *events.Config,
	producer protoflow.Producer,
) (*AppLogic, error) {
	appLogic := &AppLogic{
		db:            db,
		log:           logger,
		eventProducer: producer,
	}
	if eventsCfg != nil {
		appLogic.exampleTopic = eventsCfg.ExampleConsumeQueue
	}
	return appLogic, nil
}

// ExampleTopic returns the configured example topic.
func (a *AppLogic) ExampleTopic() string {
	if a == nil {
		return ""
	}
	return a.exampleTopic
}

// DatabaseProbe ensures the backing database remains reachable for readiness checks.
func (a *AppLogic) DatabaseProbe(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "usecase.database_probe")
	defer span.End()

	if a == nil {
		err := observability.Builder(ctx, "usecase.database_probe", "app_logic_nil").
			Public("application is not ready").
			New("applogic is nil")
		observability.RecordError(span, err)
		return err
	}
	if a.db == nil {
		err := observability.Builder(ctx, "usecase.database_probe", "database_not_configured").
			Public("database is not configured").
			New("database not configured")
		observability.RecordError(span, err)
		return err
	}

	span.SetAttributes(attribute.String("check.name", "database"))

	err := a.db.Ping(ctx)
	if err == nil {
		return nil
	}

	wrapped := observability.Builder(ctx, "usecase.database_probe", "database_probe_failed").
		Public("database health check failed").
		Wrap(err)
	observability.RecordError(span, wrapped)
	return wrapped
}
