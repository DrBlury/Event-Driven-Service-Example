package events

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
)

func (s *Service) createRabbitMQPublisher(config amqp.Config, conn *amqp.ConnectionWrapper, logger watermill.LoggerAdapter) {
	publisher, err := amqp.NewPublisherWithConnection(
		config,
		logger,
		conn,
	)
	if err != nil {
		panic(err)
	}
	s.Publisher = publisher
}

func (s *Service) createRabbitMQSubscriber(config amqp.Config, conn *amqp.ConnectionWrapper, logger watermill.LoggerAdapter) {
	subscriber, err := amqp.NewSubscriberWithConnection(
		config,
		logger,
		conn,
	)
	if err != nil {
		panic(err)
	}
	s.Subscriber = subscriber
}
