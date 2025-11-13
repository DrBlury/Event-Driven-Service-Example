package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/pkg/events"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

// registerAppEventHandlers wires the demo handlers used by this application.
// In your own code base you can register entirely different handlers against
// the shared events.Service instance.
func registerAppEventHandlers(svc *events.Service) error {
	if err := svc.RegisterHandler(events.HandlerRegistration{
		Name:         "demoHandler",
		ConsumeQueue: svc.Conf.ConsumeQueue,
		PublishQueue: svc.Conf.PublishQueue,
		Handler:      demoHandler(svc),
	}); err != nil {
		return err
	}

	if err := svc.RegisterHandler(events.HandlerRegistration{
		ConsumeQueue:     svc.Conf.ConsumeQueueSignup,
		PublishQueue:     svc.Conf.PublishQueueSignup,
		Handler:          signupHandler(),
		MessagePrototype: &domain.Signup{},
	}); err != nil {
		return err
	}

	if err := svc.RegisterHandler(events.HandlerRegistration{
		ConsumeQueue:     svc.Conf.PublishQueueSignup,
		PublishQueue:     "signup_step_2_processed",
		Handler:          signupStepTwoHandler(),
		MessagePrototype: &domain.BillingAddress{},
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

			msgs, err := createNewProcessedEvent(e, map[string]string{})
			if err != nil {
				slog.Error("could not create event", "error", err)
				continue
			}

			if err := svc.Publisher.Publish(svc.Conf.ConsumeQueueSignup, msgs...); err != nil {
				slog.Error("could not publish event", "error", err)
				continue
			}

			i++
		}
	}
}

func demoHandler(svc *events.Service) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		type demoEvent struct {
			ID   int          `json:"id"`
			Date *domain.Date `json:"date"`
		}

		type processedDemoEvent struct {
			ID   int          `json:"id"`
			Time time.Time    `json:"time"`
			Date *domain.Date `json:"date"`
		}

		consumedPayload := &demoEvent{}
		if err := json.Unmarshal(msg.Payload, consumedPayload); err != nil {
			return nil, err
		}

		slog.Info("Received date",
			"year", consumedPayload.Date.Year,
			"month", consumedPayload.Date.Month,
			"day", consumedPayload.Date.Day,
		)

		newPayload, err := json.Marshal(
			processedDemoEvent{
				ID:   consumedPayload.ID,
				Time: time.Now(),
				Date: consumedPayload.Date,
			},
		)
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)
		newMessage.Metadata = msg.Metadata
		newMessage.Metadata["handler"] = "demoHandler"
		newMessage.Metadata["next_queue"] = svc.Conf.PublishQueue
		return []*message.Message{newMessage}, nil
	}
}

func signupHandler() message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &domain.Signup{}
		if err := readMessageToStruct(msg, consumedPayload); err != nil {
			return nil, err
		}

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

		return createNewProcessedEvent(newEvent, msg.Metadata)
	}
}

func signupStepTwoHandler() message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &domain.BillingAddress{}
		if err := readMessageToStruct(msg, consumedPayload); err != nil {
			return nil, err
		}

		newEvent := &domain.Date{
			Year:  int32(time.Now().Year()),
			Month: int32(time.Now().Month()),
			Day:   int32(time.Now().Day()),
		}

		if int(rand.Int64N(100))%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		return createNewProcessedEvent(newEvent, msg.Metadata)
	}
}

func readMessageToStruct(msg *message.Message, v interface{}) error {
	return json.Unmarshal(msg.Payload, v)
}

func createNewProcessedEvent(event proto.Message, metadata map[string]string) ([]*message.Message, error) {
	jsonPayload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	newMessage := message.NewMessage(watermill.NewUUID(), jsonPayload)
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["event_message_schema"] = fmt.Sprintf("%T", event)
	newMessage.Metadata = metadata
	return []*message.Message{newMessage}, nil
}
