package events

import (
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

// HandlerRegistration configures a Watermill handler before it is registered on the router.
type HandlerRegistration struct {
	Name             string
	ConsumeQueue     string
	Subscriber       message.Subscriber
	PublishQueue     string
	Publisher        message.Publisher
	Handler          message.HandlerFunc
	MessagePrototype proto.Message
}

// RegisterHandler attaches a handler to the router. The subscriber and publisher default to the
// Service-wide instances when omitted. Provide MessagePrototype when you want protoValidateMiddleware
// to unmarshal and validate payloads using the supplied type.
func (s *Service) RegisterHandler(cfg HandlerRegistration) error {
	if cfg.Handler == nil {
		return errors.New("handler function is required")
	}
	if cfg.ConsumeQueue == "" {
		return errors.New("consume queue is required")
	}
	if cfg.Subscriber == nil {
		cfg.Subscriber = s.Subscriber
	}
	if cfg.Publisher == nil {
		cfg.Publisher = s.Publisher
	}
	if cfg.MessagePrototype != nil {
		s.registerProtoType(cfg.MessagePrototype)
		if cfg.Name == "" {
			cfg.Name = fmt.Sprintf("%T-Handler", cfg.MessagePrototype)
		}
	}
	if cfg.Name == "" {
		return errors.New("handler name is required")
	}

	s.Router.AddHandler(
		cfg.Name,
		cfg.ConsumeQueue,
		cfg.Subscriber,
		cfg.PublishQueue,
		cfg.Publisher,
		cfg.Handler,
	)

	return nil
}

// RegisterProtoMessage exposes a proto message type for validation without registering a handler.
func (s *Service) RegisterProtoMessage(msg proto.Message) {
	s.registerProtoType(msg)
}

func (s *Service) registerProtoType(msg proto.Message) {
	if msg == nil {
		return
	}

	typeName := fmt.Sprintf("%T", msg)

	s.protoRegistryMu.Lock()
	s.protoRegistry[typeName] = func() proto.Message {
		clone := proto.Clone(msg)
		proto.Reset(clone)
		return clone
	}
	s.protoRegistryMu.Unlock()
}
