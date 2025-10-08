package events

import (
	"drblury/poc-event-signup/internal/domain"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

// signupStepTwoHandlerFunc is an example of a message handler function.
func (s *Service) signupStepTwoHandlerFunc() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := &domain.BillingAddress{}
		err := readMessageToStruct(msg, consumedPayload)
		if err != nil {
			return nil, err
		}

		newEvent := &domain.Date{
			Year:  int32(time.Now().Year()),
			Month: int32(time.Now().Month()),
			Day:   int32(time.Now().Day()),
		}

		// make 10% of the events fail fatally
		if rand.IntN(100)%10 == 0 {
			return nil, errors.New("fatal error processing signup event")
		}

		return createNewProcessedEvent(newEvent, msg.Metadata)
	}
}
