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

	s.addAllHandlers()

	// Order: correlationID -> logging -> outbox -> retry -> poison -> recoverer -> custom recover
	s.Router.AddMiddleware(s.correlationIDMiddleware())     // add correlation ID if not present
	s.Router.AddMiddleware(s.logMessagesMiddleware(logger)) // log all messages being processed
	s.Router.AddMiddleware(s.outboxMiddleware())
	s.Router.AddMiddleware(s.retryMiddleware())  // exponential backoff max 5 retries (1s, 2s, 4s, 8s, 16s)
	s.Router.AddMiddleware(s.poisonMiddleware()) // this is a dead letter queue
	s.Router.AddMiddleware(middleware.Recoverer) // built-in recoverer

	// Simulate producing events
	// go s.simulateEventsDemo()
	go s.simulateEventsSignup()

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
	// s.Router.AddHandler(
	// 	"demoHandler",       // handler name, must be unique
	// 	s.Conf.ConsumeTopic, // topic from which messages should be consumed
	// 	s.Subscriber,
	// 	s.Conf.PublishTopic, // topic to which messages should be published
	// 	s.Publisher,
	// 	s.demoHandlerFunc(),
	// )

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

// correlationIDMiddleware is a middleware to add correlation IDs to messages
func (s *Service) correlationIDMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			if _, ok := msg.Metadata["correlation_id"]; !ok {
				msg.Metadata["correlation_id"] = watermill.NewUUID()
			}
			return h(msg)
		}
	}
}

// poisonmiddleware is a middleware to handle poison messages (Dead letter queue)
func (s *Service) poisonMiddleware() message.HandlerMiddleware {
	mw, err := middleware.PoisonQueue(
		s.Publisher,
		s.Conf.PoisonTopic,
	)

	if err != nil {
		panic(err)
	}

	return mw
}

// middleware to log all messages being processed
func (s *Service) logMessagesMiddleware(logger watermill.LoggerAdapter) message.HandlerMiddleware {
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

// outboxMiddleware is a placeholder for an outbox pattern implementation.
// We want to store the incoming message in the database before processing it.
// And then after processing, we want to mark it as processed.
func (s *Service) outboxMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			// BEFORE processing the message, store it in the outbox
			// turn message into json
			topic := "unknown_topic"
			// if no topic is set, use a default one
			if msg.Metadata["consumed_topic"] != "" {
				topic = msg.Metadata["consumed_topic"]
			}

			err := s.DB.StoreIncomingMessage(msg.Context(), topic, msg.UUID, string(msg.Payload))
			if err != nil {
				return nil, err
			}

			// Process the message
			outgoingMessages, err := h(msg)
			if err != nil {
				return nil, err
			}

			// AFTER processing the message, mark it as processed in the outbox
			err = s.DB.SetIncomingMessageProcessed(msg.Context(), topic, msg.UUID)
			if err != nil {
				return nil, err
			}

			// Write it to the outbox table as well
			for _, outMsg := range outgoingMessages {
				outTopic := "unknown_topic"
				if outMsg.Metadata["published_topic"] != "" {
					outTopic = outMsg.Metadata["published_topic"]
				}
				err = s.DB.StoreOutgoingMessage(msg.Context(), outTopic, outMsg.UUID, string(outMsg.Payload))
				if err != nil {
					return nil, err
				}
			}

			return outgoingMessages, nil
		}
	}
}

// retryMiddleware is a middleware that will use exponential backoff to retry message processing.
func (s *Service) retryMiddleware() message.HandlerMiddleware {
	return middleware.Retry{
		MaxRetries:      5,        // 1, 2, 4, 8, 16
		InitialInterval: 1 * 1e9,  // 1s
		MaxInterval:     16 * 1e9, // 16s
	}.Middleware
}
