package usecase

import (
	"context"
	"errors"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

// HandleExample persists the received example payload. Token handling is left as an
// exercise for service integrators so the sample stays vendor-neutral.
//
// NOTE: The token parameter is intentionally unused in this reference implementation.
// Integrators should implement proper token validation (e.g., JWT verification,
// API key validation) based on their authentication requirements.
func (a *AppLogic) HandleExample(ctx context.Context, record *domain.ExampleRecord, token string) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if record == nil {
		return errors.New("example payload is required")
	}
	// TODO: Implement token validation for your authentication scheme.
	// Example: if err := validateToken(ctx, token); err != nil { return err }
	_ = token

	if a.db == nil {
		return nil
	}

	// we can do something with the event here or just store it...
	err := a.emitExampleEvent(ctx, record)
	if err != nil {
		return err
	}

	return a.db.StoreExampleRecord(ctx, record)
}

// EmitExampleEvent publishes the example payload so downstream processors can pick it up.
// This method acquires a read lock to safely access shared configuration.
func (a *AppLogic) emitExampleEvent(ctx context.Context, record *domain.ExampleRecord) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if record == nil {
		return errors.New("example payload is required")
	}

	// Acquire read lock to safely access shared fields
	a.mu.RLock()
	topic := a.exampleTopic
	producer := a.eventProducer
	a.mu.RUnlock()

	if topic == "" {
		return errors.New("example topic not configured")
	}
	if producer == nil {
		return errors.New("event producer not configured")
	}

	metadata := protoflow.Metadata{
		"source": "api.examples",
	}

	return producer.PublishProto(ctx, topic, record, metadata)
}
