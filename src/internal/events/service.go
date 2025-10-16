package events

import (
	"context"
	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/usecase"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
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
	DB      *database.Database
	Usecase *usecase.AppLogic
}

func NewService(conf *Config, log *slog.Logger, db *database.Database, usecase *usecase.AppLogic, ctx context.Context) *Service {
	logger := watermill.NewSlogLoggerWithLevelMapping(log, logLevelMapping)
	logger.Info("Creating event service",
		watermill.LogFields{
			"pubsub_system": conf.PubSubSystem,
			"config":        conf,
		})
	s := &Service{
		Conf:    conf,
		Logger:  logger,
		DB:      db,
		Usecase: usecase,
	}

	setupPubSub(s, conf, logger, ctx)

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
	s.Router.AddMiddleware(s.protoValidateMiddleware())
	s.Router.AddMiddleware(s.outboxMiddleware())
	s.Router.AddMiddleware(s.retryMiddleware()) // exponential backoff max 5 retries (1s, 2s, 4s, 8s, 16s)
	s.Router.AddMiddleware(s.poisonMiddlewareWithFilter(func(err error) bool {
		// Example: filter out certain errors from going to the poison queue
		if _, ok := err.(*UnprocessableEventError); ok {
			return true
		}
		return false
	}))
	s.Router.AddMiddleware(middleware.Recoverer) // built-in recoverer

	// Simulate producing events
	// go s.simulateEventsDemo()
	go s.simulateEventsSignup()

	if err := s.Router.Run(ctx); err != nil {
		panic(err)
	}

	return s
}

func setupPubSub(s *Service, conf *Config, logger watermill.LoggerAdapter, ctx context.Context) {
	switch conf.PubSubSystem {
	case "kafka":
		// Kafka setup
		s.createKafkaPublisher(conf.KafkaBrokers, logger)
		s.createKafkaSubscriber(conf.KafkaConsumerGroup, conf.KafkaBrokers, logger)
		return
	case "rabbitmq":
		// Rabbitmq setup
		ampqConn, ampqConfig := s.setupAmpq(conf, logger)
		s.createRabbitMQPublisher(ampqConfig, ampqConn, logger)
		s.createRabbitMQSubscriber(ampqConfig, ampqConn, logger)
		return
	case "aws":
		// AWS setup
		cfg := s.createAWSConfig(ctx)
		logger.Info("Created AWS config",
			watermill.LogFields{
				"AWSConfig": cfg,
			},
		)
		s.createAwsPublisher(logger, cfg)
		s.createAwsSubscriber(logger, cfg)
		return
	default:
		panic("unsupported PubSubSystem, must be 'kafka' or 'rabbitmq'")
	}
}
