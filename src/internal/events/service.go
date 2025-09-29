package events

import (
	"context"
	"drblury/poc-event-signup/internal/database"
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
	// add database or other dependencies here as needed
	DB *database.Database
}

func NewService(conf *Config, log *slog.Logger, db *database.Database, ctx context.Context) *Service {
	logger := watermill.NewSlogLoggerWithLevelMapping(log, logLevelMapping)
	s := &Service{
		Conf:   conf,
		Logger: logger,
		DB:     db,
	}
	s.createPublisher(conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)
	s.createSubscriber(conf.KafkaConsumerGroup, conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	s.Router = router

	s.Router.AddPlugin(plugin.SignalsHandler)
	s.Router.AddMiddleware(middleware.Recoverer)

	s.addAllHandlers()

	s.Router.AddMiddleware(logMessagesMiddleware(logger))

	// Simulate producing events
	go s.simulateEvents()

	if err := s.Router.Run(ctx); err != nil {
		panic(err)
	}

	return s
}

// createPublisher is a helper function that creates a Publisher, in this case - the Kafka Publisher.
func (s *Service) createPublisher(brokers []string, marshaler kafka.Marshaler, logger watermill.LoggerAdapter) {
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

	s.Publisher = kafkaPublisher
}

// createSubscriber is a helper function similar to the previous one, but in this case it creates a Subscriber.
func (s *Service) createSubscriber(consumerGroup string, brokers []string, unmarshaler kafka.Unmarshaler, logger watermill.LoggerAdapter) {
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

	s.Subscriber = kafkaSubscriber
}

func (s *Service) addAllHandlers() {
	// This is just for demonstration purposes.
	// In a real application, you would have different handlers for different topics.
	s.Router.AddHandler(
		"demoHandler",       // handler name, must be unique
		s.Conf.ConsumeTopic, // topic from which messages should be consumed
		s.Subscriber,
		s.Conf.PublishTopic, // topic to which messages should be published
		s.Publisher,
		s.demoHandlerFunc(),
	)

	// Add the signup handler
	s.Router.AddHandler(
		"signupHandler",
		s.Conf.ConsumeTopicSignup,
		s.Subscriber,
		s.Conf.PublishTopicSignup,
		s.Publisher,
		s.signupHandlerFunc(),
	)
}

// middleware to log all messages being processed
func logMessagesMiddleware(logger watermill.LoggerAdapter) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger.Info("Processing message", watermill.LogFields{
				"message_uuid": msg.UUID,
				"payload":      string(msg.Payload),
				"metadata":     msg.Metadata,
			})
			return h(msg)
		}
	}
}
