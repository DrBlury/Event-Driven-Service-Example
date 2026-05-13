package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"

	"github.com/drblury/protoflow"
	"github.com/samber/oops"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	datepb "google.golang.org/genproto/googleapis/type/date"
)

// registerAppEventHandlers wires the demo handlers used by this application.
// In your own code base you can register entirely different handlers against
// the shared protoflow.Service instance.
var tracer = otel.Tracer("drblury/event-driven-service/internal/events")

func registerAppEventHandlers(svc *protoflow.Service, cfg *Config) error {
	if err := protoflow.RegisterJSONHandler(svc, protoflow.JSONHandlerRegistration[*demoEvent, *processedDemoEvent]{
		Name:         "demoHandler",
		ConsumeQueue: cfg.DemoConsumeQueue,
		PublishQueue: cfg.DemoPublishQueue,
		Handler:      demoHandler(),
	}); err != nil {
		return observability.Builder(context.Background(), "events.register_handlers", "demo_handler_registration_failed").
			Public("demo event handler could not be registered").
			With("consume_queue", cfg.DemoConsumeQueue, "publish_queue", cfg.DemoPublishQueue).
			Wrap(err)
	}

	if err := protoflow.RegisterProtoHandler(svc, protoflow.ProtoHandlerRegistration[*domain.ExampleRecord]{
		Name:         "exampleRecordHandler",
		ConsumeQueue: cfg.ExampleConsumeQueue,
		PublishQueue: cfg.ExamplePublishQueue,
		Handler:      exampleRecordHandler(),
	}); err != nil {
		return observability.Builder(context.Background(), "events.register_handlers", "example_handler_registration_failed").
			Public("example event handler could not be registered").
			With("consume_queue", cfg.ExampleConsumeQueue, "publish_queue", cfg.ExamplePublishQueue).
			Wrap(err)
	}

	return nil
}

func RunExampleSimulation(ctx context.Context, svc *protoflow.Service, cfg *Config) {
	runSomeSimulation(ctx, svc, cfg.ExampleConsumeQueue)
}

// runSomeSimulation produces demo events so the application can be
// exercised locally without external publishers.
func runSomeSimulation(ctx context.Context, svc *protoflow.Service, queueName string) {
	if ctx == nil {
		ctx = context.Background()
	}
	if svc == nil {
		observability.Logger(ctx, slog.Default()).Warn("simulation skipped because protoflow service is nil", "queue", queueName)
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tags := []string{"demo", fmt.Sprintf("batch-%d", i%3)}
			followUp := i%2 == 0
			if followUp {
				tags = append(tags, "follow-up")
			}

			e := &domain.ExampleRecord{
				RecordId:    fmt.Sprintf("EX-%04d", i+1),
				Title:       fmt.Sprintf("Example payload %d", i+1),
				Description: "Auto-generated sample data",
				Tags:        tags,
				Meta: &domain.ExampleMeta{
					RequestedBy:      "simulation-bot",
					RequiresFollowUp: followUp,
					Priority:         int32(rand.IntN(5) + 1), // #nosec G404 G115 -- non-security simulation data with bounded values
					DesiredStartDate: &datepb.Date{
						Year:  int32(rand.IntN(5) + 2020), // #nosec G404 G115 -- non-security simulation data with bounded values
						Month: int32(i%12 + 1),            // #nosec G115 -- bounded value 1-12
						Day:   int32((i % 28) + 1),        // #nosec G115 -- bounded value 1-28
					},
				},
			}

			if err := svc.PublishProto(ctx, queueName, e, protoflow.Metadata{"source": "simulation"}); err != nil {
				wrapped := observability.Builder(ctx, "events.simulation", "simulation_publish_failed").
					Public("simulation event could not be published").
					With("queue", queueName, "record_id", e.GetRecordId()).
					Wrap(err)
				observability.Logger(ctx, slog.Default()).Error(
					"could not publish simulation event",
					"error", wrapped,
					"queue", queueName,
					"record_id", e.GetRecordId(),
				)
				continue
			}

			i++
		}
	}
}

type demoEvent struct {
	ID   int          `json:"id"`
	Date *domain.Date `json:"date"`
}

type processedDemoEvent struct {
	ID   int          `json:"id"`
	Time time.Time    `json:"time"`
	Date *domain.Date `json:"date"`
}

func demoHandler() protoflow.JSONMessageHandler[*demoEvent, *processedDemoEvent] {
	return func(ctx context.Context, evt protoflow.JSONMessageContext[*demoEvent]) ([]protoflow.JSONMessageOutput[*processedDemoEvent], error) {
		ctx, span := tracer.Start(ctx, "events.demo_handler")
		defer span.End()

		// Context is available for cancellation checks and passing to downstream calls
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if evt.Payload == nil || evt.Payload.Date == nil {
			err := observability.Builder(ctx, "events.demo_handler", "demo_event_invalid").
				Public("demo event payload is invalid").
				With("has_payload", evt.Payload != nil).
				New("demo event date is required")
			observability.RecordError(span, err)
			return nil, errors.Join(protoflow.ErrUnprocessable, err)
		}

		span.SetAttributes(
			attribute.Int("example.id", evt.Payload.ID),
			attribute.Int64("example.date.year", int64(evt.Payload.Date.Year)),
			attribute.Int64("example.date.month", int64(evt.Payload.Date.Month)),
			attribute.Int64("example.date.day", int64(evt.Payload.Date.Day)),
		)

		observability.Logger(ctx, slog.Default()).Debug(
			"received demo date event",
			"handler", "demoHandler",
			"id", evt.Payload.ID,
			"year", evt.Payload.Date.Year,
			"month", evt.Payload.Date.Month,
			"day", evt.Payload.Date.Day,
		)

		metadata := evt.Metadata.WithAll(
			protoflow.Metadata{
				"handler":      "demoHandler",
				"processed_at": time.Now().Format(time.RFC3339),
				"next_queue":   "demo_processed_events",
			},
		)

		return []protoflow.JSONMessageOutput[*processedDemoEvent]{
			{
				Message: &processedDemoEvent{
					ID:   evt.Payload.ID,
					Time: time.Now(),
					Date: evt.Payload.Date,
				},
				Metadata: metadata,
			},
		}, nil
	}
}

func exampleRecordHandler() protoflow.ProtoMessageHandler[*domain.ExampleRecord] {
	return func(ctx context.Context, e protoflow.ProtoMessageContext[*domain.ExampleRecord]) ([]protoflow.ProtoMessageOutput, error) {
		ctx, span := tracer.Start(ctx, "events.example_record_handler")
		defer span.End()

		if e.Payload == nil {
			err := observability.Builder(ctx, "events.example_record_handler", "example_payload_required").
				Public("example event payload is required").
				New("example payload is required")
			observability.RecordError(span, err)
			return nil, errors.Join(protoflow.ErrUnprocessable, err)
		}

		span.SetAttributes(attribute.String("example.record_id", e.Payload.GetRecordId()))

		// #nosec G404 -- non-security simulation for random failures
		if rand.IntN(10) == 0 {
			err := observability.Builder(ctx, "events.example_record_handler", "simulated_processing_failure").
				Public("example event processing failed").
				With("record_id", e.Payload.GetRecordId(), "title", e.Payload.GetTitle()).
				Wrapf(oops.New("fatal error processing example event"), "simulate failure for example record %q", e.Payload.GetRecordId())
			observability.RecordError(span, err)
			observability.Logger(ctx, slog.Default()).Warn(
				"example event processing failed",
				"error", err,
				"record_id", e.Payload.GetRecordId(),
			)
			return nil, err
		}

		statuses := []string{"queued", "in-progress", "completed"}
		status := statuses[rand.IntN(len(statuses))] // #nosec G404 -- non-security simulation data

		now := time.Now()
		result := &domain.ExampleResult{
			RecordId: e.Payload.GetRecordId(),
			Status:   status,
			Note:     fmt.Sprintf("processed %s", e.Payload.GetTitle()),
			ProcessedOn: &datepb.Date{
				Year:  int32(now.Year()),  // #nosec G115 -- year is bounded to reasonable values
				Month: int32(now.Month()), // #nosec G115 -- month is bounded 1-12
				Day:   int32(now.Day()),   // #nosec G115 -- day is bounded 1-31
			},
		}

		metadata := e.Metadata.WithAll(
			protoflow.Metadata{
				"handler":      "exampleRecordHandler",
				"processed_at": now.Format(time.RFC3339),
			},
		)

		observability.Logger(ctx, slog.Default()).Debug(
			"processed example record event",
			"handler", "exampleRecordHandler",
			"record_id", e.Payload.GetRecordId(),
			"status", status,
		)

		return []protoflow.ProtoMessageOutput{{Message: result, Metadata: metadata}}, nil
	}
}
