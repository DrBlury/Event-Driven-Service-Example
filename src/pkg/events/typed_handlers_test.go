package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/encoding/protojson"

	"drblury/event-driven-service/internal/domain"
)

func TestBuildProtoHandlerUnmarshalsPayload(t *testing.T) {
	prototype := &domain.AddAService{}
	handler, err := buildProtoHandler(prototype, func(ctx context.Context, evt ProtoMessageContext[*domain.AddAService]) ([]ProtoMessageOutput, error) {
		if ctx == nil {
			t.Fatalf("context should not be nil")
		}
		if evt.Payload == nil {
			t.Fatalf("expected payload instance")
		}
		if evt.Metadata == nil {
			t.Fatalf("expected metadata")
		}
		evt.Metadata["processed_by"] = "typed_handler"
		return []ProtoMessageOutput{{
			Message: &domain.AddAService{CustomerId: "processed"},
		}}, nil
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error building handler: %v", err)
	}

	payload, err := protojson.Marshal(&domain.AddAService{CustomerId: "cust-123"})
	if err != nil {
		t.Fatalf("failed to marshal proto: %v", err)
	}
	msg := message.NewMessage(CreateULID(), payload)
	msg.Metadata = message.Metadata{"origin": "test"}

	produced, err := handler(msg)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(produced) != 1 {
		t.Fatalf("expected single outgoing message, got %d", len(produced))
	}
	if produced[0].Metadata["processed_by"] != "typed_handler" {
		t.Fatalf("missing propagated metadata")
	}
}

func TestBuildProtoHandlerHonoursCustomMetadata(t *testing.T) {
	prototype := &domain.AddAService{}
	handler, err := buildProtoHandler(prototype, func(ctx context.Context, evt ProtoMessageContext[*domain.AddAService]) ([]ProtoMessageOutput, error) {
		md := evt.CloneMetadata()
		md["origin"] = "cloned"
		return []ProtoMessageOutput{{
			Message:  &domain.AddAService{},
			Metadata: md,
		}}, nil
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error building handler: %v", err)
	}

	msg := message.NewMessage(CreateULID(), []byte(`{}`))
	produced, err := handler(msg)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if produced[0].Metadata["origin"] != "cloned" {
		t.Fatalf("custom metadata not applied")
	}
}

func TestRegisterProtoHandlerValidations(t *testing.T) {
	svc := newTestService(t)
	err := RegisterProtoHandler[*domain.AddAService](nil, ProtoHandlerRegistration[*domain.AddAService]{
		ConsumeMessageType: &domain.AddAService{},
		Handler: func(context.Context, ProtoMessageContext[*domain.AddAService]) ([]ProtoMessageOutput, error) {
			return nil, nil
		},
	})
	if err == nil {
		t.Fatalf("expected error when service nil")
	}

	err = RegisterProtoHandler(svc, ProtoHandlerRegistration[*domain.AddAService]{
		ConsumeQueue:       "queue",
		ConsumeMessageType: nil,
		Handler: func(context.Context, ProtoMessageContext[*domain.AddAService]) ([]ProtoMessageOutput, error) {
			return nil, nil
		},
	})
	if err == nil {
		t.Fatalf("expected error when prototype nil")
	}

	err = RegisterProtoHandler(svc, ProtoHandlerRegistration[*domain.AddAService]{
		ConsumeQueue:       "queue",
		ConsumeMessageType: &domain.AddAService{},
		Handler:            nil,
	})
	if err == nil {
		t.Fatalf("expected error when handler nil")
	}

	if err := RegisterProtoHandler(svc, ProtoHandlerRegistration[*domain.AddAService]{
		ConsumeQueue:       "queue",
		PublishQueue:       "out",
		ConsumeMessageType: &domain.AddAService{},
		Handler: func(context.Context, ProtoMessageContext[*domain.AddAService]) ([]ProtoMessageOutput, error) {
			return nil, nil
		},
	}); err != nil {
		t.Fatalf("unexpected error registering handler: %v", err)
	}
	if _, ok := svc.router.Handlers()["*domain.AddAService-Handler"]; !ok {
		t.Fatalf("typed handler not registered")
	}
}

func TestRegisterProtoHandlerRegistersPublishTypes(t *testing.T) {
	svc := newTestService(t)
	primary := &domain.Signup{}
	extra := &domain.BillingAddress{}

	if err := RegisterProtoHandler(svc, ProtoHandlerRegistration[*domain.AddAService]{
		Name:               "typed",
		ConsumeQueue:       "queue",
		PublishQueue:       "out",
		ConsumeMessageType: &domain.AddAService{},
		PublishMessageType: primary,
		Options: []ProtoHandlerOption{
			WithPublishMessageTypes(extra),
		},
		Handler: func(context.Context, ProtoMessageContext[*domain.AddAService]) ([]ProtoMessageOutput, error) {
			return nil, nil
		},
	}); err != nil {
		t.Fatalf("unexpected error registering handler: %v", err)
	}

	if _, ok := svc.protoRegistry[fmt.Sprintf("%T", primary)]; !ok {
		t.Fatalf("primary publish type not registered")
	}
	if _, ok := svc.protoRegistry[fmt.Sprintf("%T", extra)]; !ok {
		t.Fatalf("option publish type not registered")
	}
}
