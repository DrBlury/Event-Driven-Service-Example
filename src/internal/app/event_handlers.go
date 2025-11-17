package app

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/pkg/events"
)

// registerAppEventHandlers wires the demo handlers used by this application.
// In your own code base you can register entirely different handlers against
// the shared events.Service instance.
func registerAppEventHandlers(svc *events.Service) error {
	if err := events.RegisterJSONHandler(svc, events.JSONHandlerRegistration[*demoEvent, *processedDemoEvent]{
		Name:               "demoHandler",
		ConsumeQueue:       svc.Conf.ConsumeQueue,
		PublishQueue:       svc.Conf.PublishQueue,
		ConsumeMessageType: &demoEvent{},
		Handler:            demoHandler(svc),
	}); err != nil {
		return err
	}

	if err := events.RegisterProtoHandler(svc, events.ProtoHandlerRegistration[*domain.Signup]{
		ConsumeQueue:       svc.Conf.ConsumeQueueSignup,
		PublishQueue:       svc.Conf.PublishQueueSignup,
		Handler:            signupHandler(),
		ConsumeMessageType: &domain.Signup{},
	}); err != nil {
		return err
	}

	if err := events.RegisterProtoHandler(svc, events.ProtoHandlerRegistration[*domain.BillingAddress]{
		ConsumeQueue:       svc.Conf.PublishQueueSignup,
		PublishQueue:       "signup_step_2_processed",
		Handler:            signupStepTwoHandler(),
		ConsumeMessageType: &domain.BillingAddress{},
	}); err != nil {
		return err
	}

	return nil
}

// runSignupSimulation produces demo signup events so the application can be
// exercised locally without external publishers.
func runSignupSimulation(ctx context.Context, svc *events.Service) {
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

			if err := svc.PublishProto(ctx, svc.Conf.ConsumeQueueSignup, e, events.Metadata{}); err != nil {
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

func demoHandler(svc *events.Service) events.JSONMessageHandler[*demoEvent, *processedDemoEvent] {
	return func(ctx context.Context, evt events.JSONMessageContext[*demoEvent]) ([]events.JSONMessageOutput[*processedDemoEvent], error) {
		_ = ctx

		slog.Info("Received date",
			"year", evt.Payload.Date.Year,
			"month", evt.Payload.Date.Month,
			"day", evt.Payload.Date.Day,
		)

		metadata := evt.CloneMetadata()
		metadata["handler"] = "demoHandler"
		metadata["next_queue"] = svc.Conf.PublishQueue

		return []events.JSONMessageOutput[*processedDemoEvent]{
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

func signupHandler() events.ProtoMessageHandler[*domain.Signup] {
	return func(ctx context.Context, e events.ProtoMessageContext[*domain.Signup]) ([]events.ProtoMessageOutput, error) {
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

		return []events.ProtoMessageOutput{{Message: newEvent, Metadata: newMetadata}}, nil
	}
}

func signupStepTwoHandler() events.ProtoMessageHandler[*domain.BillingAddress] {
	return func(_ context.Context, _ events.ProtoMessageContext[*domain.BillingAddress]) ([]events.ProtoMessageOutput, error) {
		newEvent := &domain.Date{
			Year:  int32(time.Now().Year()),
			Month: int32(time.Now().Month()),
			Day:   int32(time.Now().Day()),
		}

		if int(rand.Int64N(100))%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		return []events.ProtoMessageOutput{{Message: newEvent}}, nil
	}
}
