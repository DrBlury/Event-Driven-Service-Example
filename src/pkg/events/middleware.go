package events

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// correlationIDMiddleware injects a correlation ID into the message metadata when missing.
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

// protoValidateMiddleware validates protobuf payloads using the registered message prototypes.
func (s *Service) protoValidateMiddleware() message.HandlerMiddleware {
	if s.validator == nil {
		return func(h message.HandlerFunc) message.HandlerFunc {
			return h
		}
	}

	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			eventType, ok := msg.Metadata["event_message_schema"]
			if !ok {
				slog.Warn("missing event_message_schema in metadata")
				return h(msg)
			}

			s.protoRegistryMu.RLock()
			newProtoFunc, ok := s.protoRegistry[eventType]
			s.protoRegistryMu.RUnlock()
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
			if err := s.validator.Validate(protoMsg); err != nil {
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

// poisonMiddlewareWithFilter publishes poison messages based on the provided filter.
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

// logMessagesMiddleware logs all processed messages with their metadata.
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

// outboxMiddleware stores outgoing messages in the configured OutboxStore, if present.
func (s *Service) outboxMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			if s.outbox == nil {
				return h(msg)
			}

			outgoingMessages, err := h(msg)
			if err != nil {
				return nil, err
			}

			if len(outgoingMessages) == 0 {
				return outgoingMessages, nil
			}

			for _, outMsg := range outgoingMessages {
				eventType := outMsg.Metadata["event_message_schema"]
				if eventType == "" {
					eventType = "unknown_event"
				}
				if err := s.outbox.StoreOutgoingMessage(msg.Context(), eventType, outMsg.UUID, string(outMsg.Payload)); err != nil {
					return nil, err
				}
			}

			return outgoingMessages, nil
		}
	}
}

// retryMiddleware retries message processing with exponential backoff.
func (s *Service) retryMiddleware() message.HandlerMiddleware {
	return middleware.Retry{
		MaxRetries:      5,
		InitialInterval: 1 * 1e9,
		MaxInterval:     16 * 1e9,
	}.Middleware
}

// tracerMiddleware wraps message handling with an OpenTelemetry span.
func (s *Service) tracerMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			tracer := otel.Tracer("events-service-tracer")
			ctx, span := tracer.Start(
				msg.Context(),
				"ProcessMessage",
			)
			defer span.End()
			msg.SetContext(ctx)

			span.SetAttributes(
				attribute.String("message.uuid", msg.UUID),
				attribute.String("message.metadata", fmt.Sprintf("%v", msg.Metadata)),
			)
			return h(msg)
		}

	}
}
