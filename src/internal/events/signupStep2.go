package events

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// signupStepTwoHandlerFunc is an example of a message handler function.
func (s *Service) signupStepTwoHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &signupEvent{}
		err := json.Unmarshal(msg.Payload, consumedPayload)
		if err != nil {
			return nil, err
		}

		// Create processed event
		newEvent := processedSignupEvent{
			ID:             consumedPayload.ID,
			Time:           time.Now(),
			SuccessMessage: "Signup successful",
			ErrorMessage:   "",
		}

		// make 50% of the events encounter a non-fatal error
		if consumedPayload.ID%2 == 0 {
			newEvent.SuccessMessage = ""
			newEvent.ErrorMessage = "Signup failed due to some error"
		}

		// make 10% of the events fail fatally
		if consumedPayload.ID%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		newPayload, err := json.Marshal(newEvent)
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)
		newMessage.Metadata = msg.Metadata // propagate metadata
		newMessage.Metadata["handler"] = "signupStepTwoHandler"
		newMessage.Metadata["next_topic"] = "signup-processable_2"
		return []*message.Message{newMessage}, nil
	}
}
