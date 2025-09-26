package events

import (
	"context"
	"drblury/poc-event-signup/internal/domain"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
)

type Service struct {
	Conf       *Config
	Publisher  message.Publisher
	Subscriber message.Subscriber
	Router     *message.Router
	Logger     watermill.LoggerAdapter
}

func NewService(conf *Config, ctx context.Context) *Service {
	logger := watermill.NewStdLogger(false, false)
	publisher := createPublisher(conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)
	subscriber := createSubscriber(conf.KafkaConsumerGroup, conf.KafkaBrokers, kafka.DefaultMarshaler{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(middleware.Recoverer)

	router.AddHandler(
		"demoHandler",     // handler name, must be unique
		conf.ConsumeTopic, // topic from which messages should be consumed
		subscriber,
		conf.PublishTopic, // topic to which messages should be published
		publisher,
		handlerName1Func(),
	)

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

// handlerName1Func is an example of a message handler function.
func handlerName1Func() func(msg *message.Message) ([]*message.Message, error) {
	return func(msg *message.Message) ([]*message.Message, error) {
		slog.Info("Processing message", "message_uuid", msg.UUID, "payload", string(msg.Payload))
		consumedPayload := event{}
		err := json.Unmarshal(msg.Payload, &consumedPayload)
		if err != nil {
			return nil, err
		}

		slog.Info("Received date",
			"year", consumedPayload.Date.Year,
			"month", consumedPayload.Date.Month,
			"day", consumedPayload.Date.Day,
		)

		newPayload, err := json.Marshal(processedEvent{
			ProcessedID: consumedPayload.ID,
			Time:        time.Now(),
			Date:        consumedPayload.Date, // Example usage
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

// simulateEvents produces events that will be later consumed.
func simulateEvents(publisher message.Publisher, consumeTopic string) {
	i := 0
	for {
		e := event{
			ID: i,
			Date: &domain.Date{
				Year:  int32(rand.IntN(5) + 2020),
				Month: int32(i%12 + 1),
				Day:   int32(i%28 + 1),
			},
		}

		payload, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}

		err = publisher.Publish(consumeTopic, message.NewMessage(
			watermill.NewUUID(), // internal uuid of the message, useful for debugging
			payload,
		))
		if err != nil {
			slog.Error("could not publish event", "error", err)
			time.Sleep(10 * time.Second)
			panic(err)
		}

		time.Sleep(5 * time.Second)
		i++
		slog.Info("published new event", "id", e.ID, "date", fmt.Sprintf("%d-%d-%d", e.Date.Year, e.Date.Month, e.Date.Day))
	}
}
