package events

import (
	"encoding/json"
	"fmt"
	"log/slog"

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

// protoValidateMiddleware is a middleware to validate protobuf messages based on event type in metadata
func (s *Service) protoValidateMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			eventType, ok := msg.Metadata["event_message_schema"]
			if !ok {
				slog.Warn("missing event_message_schema in metadata")
				return h(msg) // just pass the message to the next handler - no validation
			}
			newProtoFunc, ok := protoTypeRegistry[eventType]
			if !ok {
				slog.Error("unknown event type", "event_message_schema", eventType)
				return nil, &UnprocessableEventError{
					eventMessage: string(msg.Payload),
					err:          fmt.Errorf("unknown event type: %s", eventType),
				}
			}
			protoMsg := newProtoFunc()
			if err := json.Unmarshal(msg.Payload, protoMsg); err != nil {
				slog.Error("failed to unmarshal protobuf message", "error", err, "event_message_schema", eventType)
				return nil, &UnprocessableEventError{
					eventMessage: string(msg.Payload),
					err:          err,
				}
			}
			err := s.Usecase.Validate(protoMsg)
			if err != nil {
				slog.Error("failed to validate protobuf message", "error", err, "event_message_schema", eventType)
				return nil, &UnprocessableEventError{
					eventMessage: string(msg.Payload),
					err:          err,
				}
			}
			return h(msg)
		}
	}
}

// poisonMiddlewareWithFilter is a middleware to handle poison messages (Dead letter queue) with a filter
func (s *Service) poisonMiddlewareWithFilter(filter func(err error) bool) message.HandlerMiddleware {
	mw, err := middleware.PoisonQueueWithFilter(
		s.Publisher,
		s.Conf.PoisonQueue,
		filter,
	)

	if err != nil {
		panic(err)
	}

	return mw
}

// logMessagesMiddleware to log all string messages being processed (json or string)
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

// outboxMiddleware is for an outbox pattern implementation.
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
				event_message_schema := "unknown_event"
				if outMsg.Metadata["event_message_schema"] != "" {
					event_message_schema = outMsg.Metadata["event_message_schema"]
				}
				err = s.DB.StoreOutgoingMessage(msg.Context(), event_message_schema, outMsg.UUID, string(outMsg.Payload))
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
