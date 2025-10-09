package events

import (
	"drblury/poc-event-signup/internal/domain"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

// protoTypeRegistry maps event type names to functions returning new proto.Message instances
var protoTypeRegistry = map[string]func() proto.Message{}

func (s *Service) addHandlerWithType(
	consumeEventType proto.Message,
	consumeQueue string,
	subscriber message.Subscriber,
	produceQueue string,
	publisher message.Publisher,
	handlerFunc message.HandlerFunc,
) {
	protoTypeRegistry[fmt.Sprintf("%T", consumeEventType)] = func() proto.Message { return consumeEventType }
	s.Router.AddHandler(
		fmt.Sprintf("%T-Handler", consumeEventType),
		consumeQueue,
		subscriber,
		produceQueue,
		publisher,
		handlerFunc,
	)
}

func (s *Service) addAllHandlers() {
	// This is just for demonstration purposes.
	// In a real application, you would have different handlers for different Queues.
	s.Router.AddHandler(
		"demoHandler",       // handler name, must be unique
		s.Conf.ConsumeQueue, // Queue from which messages should be consumed
		s.Subscriber,
		s.Conf.PublishQueue, // Queue to which messages should be published
		s.Publisher,
		s.demoHandlerFunc(),
	)

	s.addHandlerWithType(
		&domain.Signup{},
		s.Conf.ConsumeQueueSignup,
		s.Subscriber,
		s.Conf.PublishQueueSignup,
		s.Publisher,
		s.signupHandlerFunc(),
	)

	s.addHandlerWithType(
		&domain.BillingAddress{},
		s.Conf.PublishQueueSignup, // Here its the consume queue because its step 2
		s.Subscriber,
		"signup_step_2_processed", // This is just for demo
		s.Publisher,
		s.signupStepTwoHandlerFunc(),
	)
}
