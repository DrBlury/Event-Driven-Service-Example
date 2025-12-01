package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

// registerAppEventHandlers wires the demo handlers used by this application.
// In your own code base you can register entirely different handlers against
// the shared protoflow.Service instance.
func registerAppEventHandlers(svc *protoflow.Service, cfg *Config) error {

	if err := protoflow.RegisterJSONHandler(svc, protoflow.JSONHandlerRegistration[*demoEvent, *processedDemoEvent]{
		Name:         "demoHandler",
		ConsumeQueue: cfg.DemoConsumeQueue,
		PublishQueue: cfg.DemoPublishQueue,
		Handler:      demoHandler(),
	}); err != nil {
		return err
	}

	if err := protoflow.RegisterProtoHandler(svc, protoflow.ProtoHandlerRegistration[*domain.ExampleRecord]{
		Name:         "exampleRecordHandler",
		ConsumeQueue: cfg.ExampleConsumeQueue,
		PublishQueue: cfg.ExamplePublishQueue,
		Handler:      exampleRecordHandler(),
	}); err != nil {
		return err
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
					DesiredStartDate: &domain.Date{
						Year:  int32(rand.IntN(5) + 2020), // #nosec G404 G115 -- non-security simulation data with bounded values
						Month: int32(i%12 + 1),            // #nosec G115 -- bounded value 1-12
						Day:   int32((i % 28) + 1),        // #nosec G115 -- bounded value 1-28
					},
				},
			}

			if err := svc.PublishProto(ctx, queueName, e, protoflow.Metadata{"source": "simulation"}); err != nil {
				slog.Error("could not publish event", "error", err)
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
		_ = ctx

		slog.Info("Received date",
			"year", evt.Payload.Date.Year,
			"month", evt.Payload.Date.Month,
			"day", evt.Payload.Date.Day,
		)

		metadata := evt.Metadata.WithAll(
			protoflow.Metadata{
				"handler":      "exampleRecordHandler",
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
		// #nosec G404 -- non-security simulation for random failures
		if rand.IntN(10) == 0 {
			return nil, errors.New("fatal error processing example event")
		}

		statuses := []string{"queued", "in-progress", "completed"}
		status := statuses[rand.IntN(len(statuses))] // #nosec G404 -- non-security simulation data

		now := time.Now()
		result := &domain.ExampleResult{
			RecordId: e.Payload.GetRecordId(),
			Status:   status,
			Note:     fmt.Sprintf("processed %s", e.Payload.GetTitle()),
			ProcessedOn: &domain.Date{
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

		return []protoflow.ProtoMessageOutput{{Message: result, Metadata: metadata}}, nil
	}
}
