package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
)

type Service struct {
	conf       *Config
	publisher  message.Publisher
	subscriber message.Subscriber
	router     *message.Router
	logger     watermill.LoggerAdapter
}

func NewService(conf *Config) *Service {
	logger := watermill.NewStdLogger(false, false)
	publisher := createPublisher(conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)
	subscriber := createSubscriber(conf.KafkaConsumerGroup, conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(middleware.Recoverer)

	addHandlers(router, conf, publisher, subscriber)

	return &Service{
		conf:       conf,
		publisher:  publisher,
		subscriber: subscriber,
		router:     router,
		logger:     logger,
	}
}

// Start starts the Kafka service by running the message router.
func addHandlers(router *message.Router, conf *Config, publisher message.Publisher, subscriber message.Subscriber) {
	router.AddHandler(
		"handlerName1",    // handler name, must be unique
		conf.ConsumeTopic, // topic from which messages should be consumed
		subscriber,
		conf.PublishTopic, // topic to which messages should be published
		publisher,
		handlerName1Func(),
	)

}

// handlerName1Func is an example of a message handler function.
func handlerName1Func() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		consumedPayload := event{}
		err := json.Unmarshal(msg.Payload, &consumedPayload)
		if err != nil {

			return nil, err
		}

		fmt.Printf("received event %+v\n", consumedPayload)

		newPayload, err := json.Marshal(processedEvent{
			ProcessedID: consumedPayload.ID,
			Time:        time.Now(),
		})
		if err != nil {
			return nil, err
		}

		newMessage := message.NewMessage(watermill.NewUUID(), newPayload)

		return []*message.Message{newMessage}, nil
	}
}

// createPublisher is a helper function that creates a Publisher, in this case - the Kafka Publisher.
func createPublisher(brokers []string, marshaler kafka.Marshaler, logger watermill.LoggerAdapter) message.Publisher {
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: marshaler,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	return kafkaPublisher
}

// createSubscriber is a helper function similar to the previous one, but in this case it creates a Subscriber.
func createSubscriber(consumerGroup string, brokers []string, unmarshaler kafka.Unmarshaler, logger watermill.LoggerAdapter) message.Subscriber {
	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       brokers,
			Unmarshaler:   unmarshaler,
			ConsumerGroup: consumerGroup, // every handler will use a separate consumer group
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	return kafkaSubscriber
}
