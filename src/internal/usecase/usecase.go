package usecase

import (
	"context"
	"errors"
	"log/slog"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/events"

	"github.com/drblury/protoflow"
)

type AppLogic struct {
	db            *database.Database
	log           *slog.Logger
	eventProducer protoflow.Producer
	exampleTopic  string
}

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
	if a == nil {
		return errors.New("applogic is nil")
	}
	if a.db == nil {
		return errors.New("database not configured")
	}
	return a.db.Ping(ctx)
}
