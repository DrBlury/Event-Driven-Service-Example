package usecase

import (
	"context"
	"errors"
	"log/slog"

	"drblury/event-driven-service/internal/database"

	"github.com/drblury/protoflow"
)

type AppLogic struct {
	db            *database.Database
	log           *slog.Logger
	eventProducer protoflow.Producer
	signupTopic   string
}

func NewAppLogic(
	db *database.Database,
	logger *slog.Logger,
) (*AppLogic, error) {

	return &AppLogic{
		db:  db,
		log: logger,
	}, nil
}

// SetEventProducer wires the event producer used by PublishEvent.
func (a *AppLogic) SetEventProducer(producer protoflow.Producer) {
	if a == nil {
		return
	}
	a.eventProducer = producer
}

func (a *AppLogic) SignupTopic() string {
	if a == nil {
		return ""
	}
	return a.signupTopic
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
