package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"drblury/event-driven-service/internal/database"

	"github.com/drblury/protoflow"
)

type AppLogic struct {
	db            *database.Database
	log           *slog.Logger
	eventProducer protoflow.Producer
	exampleTopic  string
	mu            sync.RWMutex // protects eventProducer and exampleTopic
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
// This method is thread-safe.
func (a *AppLogic) SetEventProducer(producer protoflow.Producer) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventProducer = producer
}

// SetExampleTopic configures the queue/topic used for outgoing example events.
// This method is thread-safe.
func (a *AppLogic) SetExampleTopic(topic string) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.exampleTopic = topic
}

// ExampleTopic returns the configured example topic.
// This method is thread-safe.
func (a *AppLogic) ExampleTopic() string {
	if a == nil {
		return ""
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
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
