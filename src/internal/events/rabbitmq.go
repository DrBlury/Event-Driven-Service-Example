package events

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
)

func (s *Service) setupAmpq(conf *Config, logger watermill.LoggerAdapter) (*amqp.ConnectionWrapper, amqp.Config) {
	ampqConfig := amqp.NewDurablePubSubConfig(
		conf.RabbitMQURL,
		// Rabbit's queue name in this example is based on Watermill's topic passed to Subscribe
		// plus provided suffix.
		// Exchange is Rabbit's "fanout", so when subscribing with suffix other than "test_consumer_group",
		// it will also receive all messages. It will work like separate consumer groups in Kafka.
		amqp.GenerateQueueNameTopicNameWithSuffix("-queueSuffix"),
	)
	ampqConn, err := amqp.NewConnection(amqp.ConnectionConfig{
		AmqpURI:   conf.RabbitMQURL,
		TLSConfig: nil,
		Reconnect: amqp.DefaultReconnectConfig(),
	}, logger)
	if err != nil {
		panic(err)
	}
	return ampqConn, ampqConfig
}

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
