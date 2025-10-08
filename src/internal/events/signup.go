package events

import (
	"drblury/poc-event-signup/internal/domain"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// signupHandlerFunc is an example of a message handler function.
func (s *Service) signupHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &domain.Signup{}
		err := json.Unmarshal(msg.Payload, consumedPayload)
		if err != nil {
			return nil, err
		}

		// Create processed event
		newEvent := &domain.BillingAddress{
			City:    "Cologne",
			Country: "DE",
			Zip:     "50667",
		}

		// make 10% of the events fail fatally
		if rand.IntN(10)%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		newPayload, err := json.Marshal(newEvent)
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)
		newMessage.Metadata = msg.Metadata // propagate metadata
		newMessage.Metadata["handler"] = "signupHandler"
		newMessage.Metadata["next_queue"] = s.Conf.PublishQueueSignup
		return []*message.Message{newMessage}, nil
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

		payload, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}

		err = s.Publisher.Publish(s.Conf.ConsumeQueueSignup, message.NewMessage(
			watermill.NewUUID(), // internal uuid of the message, useful for debugging
			payload,
		))
		if err != nil {
			slog.Error("could not publish event", "error", err)
			time.Sleep(10 * time.Second)
			panic(err)
		}

		time.Sleep(2 * time.Second)
		i++
	}
}
