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

type demoEvent struct {
	ID   int          `json:"id"`
	Date *domain.Date `json:"date"`
}

type processedDemoEvent struct {
	ID   int          `json:"id"`
	Time time.Time    `json:"time"`
	Date *domain.Date `json:"date"`
}

// demoHandlerFunc is an example of a message handler function.
func (s *Service) demoHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &demoEvent{}
		err := json.Unmarshal(msg.Payload, consumedPayload)
		if err != nil {
			return nil, err
		}

		// Do something with the payload, for demo purposes we just log it
		slog.Info("Received date",
			"year", consumedPayload.Date.Year,
			"month", consumedPayload.Date.Month,
			"day", consumedPayload.Date.Day,
		)

		// create the new payload
		newPayload, err := json.Marshal(
			processedDemoEvent{
				ID:   consumedPayload.ID,
				Time: time.Now(),
				Date: consumedPayload.Date, // Example usage
			},
		)
		if err != nil {
			return nil, err
		}

		// create the new message to be published
		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)
		newMessage.Metadata = msg.Metadata // propagate metadata = time.Now().Format(time.RFC3339)
		newMessage.Metadata["handler"] = "demoHandler"
		newMessage.Metadata["next_topic"] = s.Conf.PublishTopic
		return []*message.Message{newMessage}, nil
	}
}

// simulateEventsDemo produces events that will be later consumed.
func (s *Service) simulateEventsDemo() {
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
