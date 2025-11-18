package usecase

import (
	"context"
	"errors"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

// HandleExample persists the received example payload. Token handling is left as an
// exercise for service integrators so the sample stays vendor-neutral.
func (a *AppLogic) HandleExample(ctx context.Context, record *domain.ExampleRecord, token string) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if record == nil {
		return errors.New("example payload is required")
	}
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
func (a *AppLogic) emitExampleEvent(ctx context.Context, record *domain.ExampleRecord) error {
	if a == nil {
		return errors.New("applogic is nil")
	}
	if record == nil {
		return errors.New("example payload is required")
	}
	if a.exampleTopic == "" {
		return errors.New("example topic not configured")
	}
	if a.eventProducer == nil {
		return errors.New("event producer not configured")
	}

	metadata := protoflow.Metadata{
		"source": "api.examples",
	}

	return a.eventProducer.PublishProto(ctx, a.exampleTopic, record, metadata)
}
