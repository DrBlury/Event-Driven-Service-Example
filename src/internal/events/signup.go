package events

import (
	"drblury/event-driven-service/internal/domain"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

// signupHandlerFunc is an example of a message handler function.
func (s *Service) signupHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		// Deserialize the incoming message to a Signup struct
		consumedPayload := &domain.Signup{}
		err := readMessageToStruct(msg, consumedPayload)
		if err != nil {
			return nil, err
		}

		// Create processed event
		postOfficeBox := "POB 1234"
		newEvent := &domain.BillingAddress{
			City:          "Cologne",
			Country:       "DE",
			Zip:           "50667",
			PostOfficeBox: &postOfficeBox,
		}

		// make 10% of the events fail fatally
		if rand.IntN(10)%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		return createNewProcessedEvent(newEvent, msg.Metadata)
	}
}

// simulateEventsSignup produces events that will be later consumed.
func (s *Service) simulateEventsSignup() {
	i := 0
	for {
		e := &domain.Signup{
			SignupMeta: &domain.SignupMeta{
				StartOfServiceDate: &domain.Date{
					Year:  int32(rand.IntN(5) + 2020),
					Month: int32(i%12 + 1),
					Day:   int32((i % 28) + 1),
				},
			},
		}

		msgs, err := createNewProcessedEvent(e, map[string]string{
			"event_type": "Signup",
		})
		if err != nil {
			slog.Error("could not create event", "error", err)
			time.Sleep(10 * time.Second)
			panic(err)
		}

		err = s.Publisher.Publish(s.Conf.ConsumeQueueSignup, msgs...)
		if err != nil {
			slog.Error("could not publish event", "error", err)
			time.Sleep(10 * time.Second)
			panic(err)
		}

		time.Sleep(2 * time.Second)
		i++
	}
}
