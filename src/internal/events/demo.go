package events

import (
	"drblury/poc-event-signup/internal/domain"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// demoHandlerFunc is an example of a message handler function.
func (s *Service) demoHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := demoEvent{}
		err := json.Unmarshal(msg.Payload, &consumedPayload)
		if err != nil {
			return nil, err
		}

		slog.Info("Received date",
			"year", consumedPayload.Date.Year,
			"month", consumedPayload.Date.Month,
			"day", consumedPayload.Date.Day,
		)

		newPayload, err := json.Marshal(processedDemoEvent{
			ID:   consumedPayload.ID,
			Time: time.Now(),
			Date: consumedPayload.Date, // Example usage
		})
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)

		return []*message.Message{newMessage}, nil
	}
}

// simulateEvents produces events that will be later consumed.
func (s *Service) simulateEvents() {
	i := 0
	for {
		e := demoEvent{
			ID: i,
			Date: &domain.Date{
				Year:  int32(rand.IntN(5) + 2020),
				Month: int32(i%12 + 1),
				Day:   int32(i%28 + 1),
			},
		}

		payload, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}

		err = s.Publisher.Publish(s.Conf.ConsumeTopic, message.NewMessage(
			watermill.NewUUID(), // internal uuid of the message, useful for debugging
			payload,
		))
		if err != nil {
			slog.Error("could not publish event", "error", err)
			time.Sleep(10 * time.Second)
			panic(err)
		}

		time.Sleep(5 * time.Second)
		i++
		slog.Info("published new event", "id", e.ID, "date", fmt.Sprintf("%d-%d-%d", e.Date.Year, e.Date.Month, e.Date.Day))
	}
}
