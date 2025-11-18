package events

import (
	"context"
	"errors"
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
		Name:               "demoHandler",
		ConsumeQueue:       cfg.DemoConsumeQueue,
		ConsumeMessageType: &demoEvent{},
		PublishQueue:       cfg.DemoPublishQueue,
		Handler:            demoHandler(),
	}); err != nil {
		return err
	}

	if err := protoflow.RegisterProtoHandler(svc, protoflow.ProtoHandlerRegistration[*domain.Signup]{
		Name:               "signupHandler",
		ConsumeQueue:       cfg.SomeConsumeQueue,
		PublishQueue:       cfg.SomePublishQueue,
		Handler:            signupHandler(),
		ConsumeMessageType: &domain.Signup{},
	}); err != nil {
		return err
	}

	return nil
}

func RunSignupSimulation(ctx context.Context, svc *protoflow.Service, cfg *Config) {
	runSomeSimulation(ctx, svc, cfg.SomeConsumeQueue)
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
			e := &domain.Signup{
				SignupMeta: &domain.SignupMeta{
					StartOfServiceDate: &domain.Date{
						Year:  int32(rand.IntN(5) + 2020),
						Month: int32(i%12 + 1),
						Day:   int32((i % 28) + 1),
					},
				},
			}

			if err := svc.PublishProto(ctx, queueName, e, protoflow.Metadata{}); err != nil {
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

		metadata := evt.CloneMetadata()
		metadata["handler"] = "demoHandler"
		metadata["next_queue"] = "processed_demo_events"

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

func signupHandler() protoflow.ProtoMessageHandler[*domain.Signup] {
	return func(ctx context.Context, e protoflow.ProtoMessageContext[*domain.Signup]) ([]protoflow.ProtoMessageOutput, error) {
		postOfficeBox := "POB 1234"
		newEvent := &domain.BillingAddress{
			City:          "Cologne",
			Country:       "DE",
			Zip:           "50667",
			PostOfficeBox: &postOfficeBox,
		}

		if rand.IntN(10)%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		// This will clone the existing metadata and add new entries
		newMetadata := e.CloneMetadata()
		newMetadata["processed_by"] = "signupHandler"
		newMetadata["some_extra_info"] = "additional_value"

		return []protoflow.ProtoMessageOutput{{Message: newEvent, Metadata: newMetadata}}, nil
	}
}
