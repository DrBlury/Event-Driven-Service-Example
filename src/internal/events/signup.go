package events

import (
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// signupHandlerFunc is an example of a message handler function.
func signupHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := signupEvent{}
		err := json.Unmarshal(msg.Payload, &consumedPayload)
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

		// make 10% of the events fail
		if consumedPayload.ID%10 == 0 {
			newEvent.SuccessMessage = ""
			newEvent.ErrorMessage = "Signup failed due to some error"
		}

		newPayload, err := json.Marshal(newEvent)
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)

		// if there was an error, send the message to a different topic
		if newEvent.ErrorMessage != "" {
			// Send to error topic
			return []*message.Message{newMessage}, nil
		}

		return []*message.Message{newMessage}, nil
	}
}
