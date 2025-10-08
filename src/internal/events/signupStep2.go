package events

import (
	"drblury/poc-event-signup/internal/domain"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// signupStepTwoHandlerFunc is an example of a message handler function.
func (s *Service) signupStepTwoHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &domain.BillingAddress{}
		err := json.Unmarshal(msg.Payload, consumedPayload)
		if err != nil {
			return nil, err
		}

		// Create processed event
		type processedSignupEvent struct {
			ID             string    `json:"id"`
			Time           time.Time `json:"time"`
			SuccessMessage string    `json:"success_message"`
			ErrorMessage   string    `json:"error_message"`
		}

		newEvent := processedSignupEvent{
			ID:             "A-" + watermill.NewUUID(),
			Time:           time.Now(),
			SuccessMessage: "Signup successful",
			ErrorMessage:   "",
		}

		// make 50% of the events encounter a non-fatal error
		if rand.IntN(2)%2 == 0 {
			newEvent.SuccessMessage = ""
			newEvent.ErrorMessage = "Signup failed due to some error"
		}

		// make 10% of the events fail fatally
		if rand.IntN(100)%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		newPayload, err := json.Marshal(newEvent)
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)
		newMessage.Metadata = msg.Metadata // propagate metadata
		newMessage.Metadata["handler"] = "signupStepTwoHandler"
		newMessage.Metadata["next_queue"] = "signup-processable_2"
		return []*message.Message{newMessage}, nil
	}
}
