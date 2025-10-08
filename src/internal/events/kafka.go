package events

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
)

// createKafkaPublisher is a helper function that creates a Publisher, in this case - the Kafka Publisher.
func (s *Service) createKafkaPublisher(brokers []string, logger watermill.LoggerAdapter) {
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	s.Publisher = kafkaPublisher
}

// createKafkaSubscriber is a helper function similar to the previous one, but in this case it creates a Subscriber.
func (s *Service) createKafkaSubscriber(consumerGroup string, brokers []string, logger watermill.LoggerAdapter) {
	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       brokers,
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: consumerGroup, // every handler will use a separate consumer group
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	s.Subscriber = kafkaSubscriber
}
