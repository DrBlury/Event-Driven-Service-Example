package events

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"

	"drblury/event-driven-service/pkg/jsonutil"
)

// ProtoHandlerRegistration configures a typed protobuf handler that automatically
// unmarshals incoming payloads and marshals emitted events.
type ProtoHandlerRegistration[T proto.Message] struct {
	Name             string
	ConsumeQueue     string
	Subscriber       message.Subscriber
	PublishQueue     string
	Publisher        message.Publisher
	MessagePrototype T
	Handler          ProtoMessageHandler[T]
}

// ProtoMessageContext provides strongly typed access to the incoming message payload
// while preserving the underlying Watermill message for advanced scenarios.
type ProtoMessageContext[T proto.Message] struct {
	Payload  T
	Metadata message.Metadata
	Message  *message.Message
}

// CloneMetadata returns a copy of the current metadata map so handlers can safely
// mutate headers for outgoing events without touching the original map.
func (c ProtoMessageContext[T]) CloneMetadata() map[string]string {
	return cloneMetadata(c.Metadata)
}

// ProtoMessageOutput describes an event that should be emitted after the handler succeeds.
type ProtoMessageOutput struct {
	Message  proto.Message
	Metadata map[string]string
}

// ProtoMessageHandler processes a typed protobuf payload and returns the events to emit.
type ProtoMessageHandler[T proto.Message] func(ctx context.Context, event ProtoMessageContext[T]) ([]ProtoMessageOutput, error)

// RegisterProtoHandler converts the typed handler into a Watermill handler and registers it on the Service router.
func RegisterProtoHandler[T proto.Message](svc *Service, cfg ProtoHandlerRegistration[T]) error {
	if svc == nil {
		return errors.New("event service is required")
	}

	wrapped, err := buildProtoHandler(cfg.MessagePrototype, cfg.Handler)
	if err != nil {
		return err
	}

	return svc.RegisterHandler(HandlerRegistration{
		Name:             cfg.Name,
		ConsumeQueue:     cfg.ConsumeQueue,
		Subscriber:       cfg.Subscriber,
		PublishQueue:     cfg.PublishQueue,
		Publisher:        cfg.Publisher,
		Handler:          wrapped,
		MessagePrototype: cfg.MessagePrototype,
	})
}

func buildProtoHandler[T proto.Message](prototype T, handler ProtoMessageHandler[T]) (message.HandlerFunc, error) {
	if handler == nil {
		return nil, errors.New("proto handler function is required")
	}
	if isNilProto(prototype) {
		return nil, errors.New("message prototype is required")
	}

	return func(msg *message.Message) ([]*message.Message, error) {
		typed, err := clonePrototype(prototype)
		if err != nil {
			return nil, err
		}

		if err := jsonutil.Unmarshal(msg.Payload, typed); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %T payload: %w", prototype, err)
		}

		ctx := ProtoMessageContext[T]{
			Payload:  typed,
			Metadata: msg.Metadata,
			Message:  msg,
		}

		outgoing, err := handler(msg.Context(), ctx)
		if err != nil {
			return nil, err
		}

		return convertProtoOutputs(outgoing, ctx.Metadata)
	}, nil
}

func clonePrototype[T proto.Message](prototype T) (T, error) {
	if isNilProto(prototype) {
		var zero T
		return zero, errors.New("message prototype is required")
	}

	cloned := proto.Clone(prototype)
	proto.Reset(cloned)

	typed, ok := cloned.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("unexpected prototype type %T", cloned)
	}

	return typed, nil
}

func convertProtoOutputs(outputs []ProtoMessageOutput, fallback message.Metadata) ([]*message.Message, error) {
	if len(outputs) == 0 {
		return nil, nil
	}

	result := make([]*message.Message, len(outputs))
	for i, out := range outputs {
		if out.Message == nil {
			return nil, errors.New("proto handler emitted nil message")
		}

		metadata := out.Metadata
		if metadata == nil {
			metadata = fallback
		}

		msg, err := NewMessageFromProto(out.Message, metadata)
		if err != nil {
			return nil, err
		}
		result[i] = msg
	}

	return result, nil
}

func isNilProto[T proto.Message](prototype T) bool {
	msg := proto.Message(prototype)
	if msg == nil {
		return true
	}

	val := reflect.ValueOf(msg)
	switch val.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map, reflect.Func:
		return val.IsNil()
	default:
		return false
	}
}
