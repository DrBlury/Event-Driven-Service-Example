package events

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
)

var logLevelMapping = map[slog.Level]slog.Level{
	slog.LevelDebug: slog.LevelDebug,
	slog.LevelInfo:  slog.LevelInfo,
	slog.LevelWarn:  slog.LevelWarn,
	slog.LevelError: slog.LevelError,
}

type Service struct {
	Conf       *Config
	Publisher  message.Publisher
	Subscriber message.Subscriber
	Router     *message.Router
	Logger     watermill.LoggerAdapter
}

func NewService(conf *Config, log *slog.Logger, ctx context.Context) *Service {
	logger := watermill.NewSlogLoggerWithLevelMapping(log, logLevelMapping)
	publisher := createPublisher(conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)
	subscriber := createSubscriber(conf.KafkaConsumerGroup, conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(middleware.Recoverer)

	addAllHandlers(router, conf, publisher, subscriber)

	// Simulate producing events
	go simulateEvents(publisher, conf.ConsumeTopic)

	if err := router.Run(ctx); err != nil {
		panic(err)
	}

	return &Service{
		Conf:       conf,
		Publisher:  publisher,
		Subscriber: subscriber,
		Router:     router,
		Logger:     logger,
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

func addAllHandlers(
	router *message.Router,
	conf *Config,
	publisher message.Publisher,
	subscriber message.Subscriber,
) {
	// This is just for demonstration purposes.
	// In a real application, you would have different handlers for different topics.
	router.AddHandler(
		"demoHandler",     // handler name, must be unique
		conf.ConsumeTopic, // topic from which messages should be consumed
		subscriber,
		conf.PublishTopic, // topic to which messages should be published
		publisher,
		demoHandlerFunc(),
	)

	// Add the signup handler
	router.AddHandler(
		"signupHandler",         // handler name, must be unique
		conf.ConsumeTopicSignup, // topic from which messages should be consumed
		subscriber,
		conf.PublishTopicSignup, // topic to which messages should be published
		publisher,
		signupHandlerFunc(),
	)
}

// middleware to log all messages being processed
func logMessagesMiddleware(logger watermill.LoggerAdapter) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger.Info("Processing message", watermill.LogFields{
				"message_uuid": msg.UUID,
				"payload":      string(msg.Payload),
			})
			return h(msg)
		}
	}
}
