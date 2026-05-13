package usecase

import (
	"context"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"

	"github.com/drblury/protoflow"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HandleExample persists the received example payload. Token handling is left as an
// exercise for service integrators so the sample stays vendor-neutral.
//
// NOTE: The token parameter is intentionally unused in this reference implementation.
// Integrators should implement proper token validation (e.g., JWT verification,
// API key validation) based on their authentication requirements.
func (a *AppLogic) HandleExample(ctx context.Context, record *domain.ExampleRecord, token string) error {
	ctx, span := tracer.Start(ctx, "usecase.handle_example")
	defer span.End()

	if a == nil {
		err := observability.Builder(ctx, "usecase.handle_example", "app_logic_nil").
			Public("example processing is unavailable").
			New("applogic is nil")
		observability.RecordError(span, err)
		return err
	}
	if record == nil {
		err := observability.Builder(ctx, "usecase.handle_example", "example_payload_required").
			Public("example payload is required").
			New("example payload is required")
		observability.RecordError(span, err)
		return err
	}

	span.SetAttributes(
		attribute.String("example.record_id", record.GetRecordId()),
		attribute.String("example.title", record.GetTitle()),
	)
	// TODO: Implement token validation for your authentication scheme.
	// Example: if err := validateToken(ctx, token); err != nil { return err }
	_ = token

	if a.db == nil {
		observability.Logger(ctx, a.log).Warn(
			"database not configured, skipping example persistence",
			"record_id", record.GetRecordId(),
		)
		return nil
	}

	// we can do something with the event here or just store it...
	err := a.emitExampleEvent(ctx, record)
	if err != nil {
		wrapped := observability.Builder(ctx, "usecase.handle_example", "example_event_publish_failed").
			Public("example event could not be queued").
			With("record_id", record.GetRecordId()).
			Wrapf(err, "publish example event for record %q", record.GetRecordId())
		observability.RecordError(span, wrapped)
		return wrapped
	}

	err = a.db.StoreExampleRecord(ctx, record)
	if err == nil {
		return nil
	}

	wrapped := observability.Builder(ctx, "usecase.handle_example", "example_record_store_failed").
		Public("example record could not be stored").
		With("record_id", record.GetRecordId()).
		Wrapf(err, "store example record %q", record.GetRecordId())
	observability.RecordError(span, wrapped)
	return wrapped
}

// EmitExampleEvent publishes the example payload so downstream processors can pick it up.
func (a *AppLogic) emitExampleEvent(ctx context.Context, record *domain.ExampleRecord) error {
	ctx, span := tracer.Start(ctx, "usecase.emit_example_event")
	defer span.End()

	topic, producer, err := a.validateEmitExampleEvent(ctx, span, record)
	if err != nil {
		return err
	}

	span.SetAttributes(
		attribute.String("example.record_id", record.GetRecordId()),
		attribute.String("messaging.destination.name", topic),
	)

	metadata := buildExampleEventMetadata(ctx)

	err = producer.PublishProto(ctx, topic, record, metadata)
	if err == nil {
		return nil
	}

	wrapped := observability.Builder(ctx, "usecase.emit_example_event", "event_publish_failed").
		Public("example event could not be queued").
		With("record_id", record.GetRecordId(), "topic", topic).
		Wrap(err)
	observability.RecordError(span, wrapped)
	return wrapped
}

func (a *AppLogic) validateEmitExampleEvent(
	ctx context.Context,
	span trace.Span,
	record *domain.ExampleRecord,
) (string, protoflow.Producer, error) {
	if a == nil {
		err := observability.Builder(ctx, "usecase.emit_example_event", "app_logic_nil").
			Public("example event publishing is unavailable").
			New("applogic is nil")
		observability.RecordError(span, err)
		return "", nil, err
	}
	if record == nil {
		err := observability.Builder(ctx, "usecase.emit_example_event", "example_payload_required").
			Public("example payload is required").
			New("example payload is required")
		observability.RecordError(span, err)
		return "", nil, err
	}

	topic := a.exampleTopic
	if topic == "" {
		err := observability.Builder(ctx, "usecase.emit_example_event", "example_topic_missing").
			Public("example event topic is not configured").
			With("record_id", record.GetRecordId()).
			New("example topic not configured")
		observability.RecordError(span, err)
		return "", nil, err
	}

	if a.eventProducer == nil {
		err := observability.Builder(ctx, "usecase.emit_example_event", "event_producer_missing").
			Public("event producer is not configured").
			With("record_id", record.GetRecordId(), "topic", topic).
			New("event producer not configured")
		observability.RecordError(span, err)
		return "", nil, err
	}

	return topic, a.eventProducer, nil
}

func buildExampleEventMetadata(ctx context.Context) protoflow.Metadata {
	metadata := protoflow.Metadata{
		"source": "api.examples",
	}

	traceID, spanID := observability.TraceAndSpanIDs(ctx)
	if traceID != "" {
		metadata["trace_id"] = traceID
	}
	if spanID != "" {
		metadata["span_id"] = spanID
	}

	return metadata
}
