package events

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

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
			// Do something before processing the message

			// Process the message
			outgoingMessages, err := h(msg)
			if err != nil {
				return nil, err
			}

			if len(outgoingMessages) == 0 {
				// nothing to store
				return nil, nil
			}

			// Write it to the outbox table as well
			for _, outMsg := range outgoingMessages {
				outTopic := "unknown_topic"
				if outMsg.Metadata["next_topic"] != "" {
					outTopic = outMsg.Metadata["next_topic"]
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
