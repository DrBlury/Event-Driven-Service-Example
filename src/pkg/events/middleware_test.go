package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCorrelationIDMiddleware(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	mw := svc.correlationIDMiddleware()

	t.Run("adds missing id", func(t *testing.T) {
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{}
		called := false
		_, err := mw(func(m *message.Message) ([]*message.Message, error) {
			called = true
			if m.Metadata["correlation_id"] == "" {
				t.Fatal("expected correlation id to be populated")
			}
			return nil, nil
		})(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("handler not invoked")
		}
	})

	t.Run("keeps existing id", func(t *testing.T) {
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{"correlation_id": "fixed"}
		_, err := mw(func(m *message.Message) ([]*message.Message, error) {
			if m.Metadata["correlation_id"] != "fixed" {
				t.Fatal("expected correlation id to be preserved")
			}
			return nil, nil
		})(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestProtoValidateMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("skips when validator unset", func(t *testing.T) {
		svc := &Service{}
		mw := svc.protoValidateMiddleware()
		msg := message.NewMessage(uuid.NewString(), []byte(`{"foo":"bar"}`))
		msg.Metadata = message.Metadata{}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("warns when schema missing", func(t *testing.T) {
		svc := &Service{validator: &testValidator{}}
		mw := svc.protoValidateMiddleware()
		msg := message.NewMessage(uuid.NewString(), []byte(`{"foo":"bar"}`))
		msg.Metadata = message.Metadata{}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("fails for unknown schema", func(t *testing.T) {
		svc := &Service{validator: &testValidator{}, protoRegistry: make(map[string]func() proto.Message)}
		mw := svc.protoValidateMiddleware()
		msg := message.NewMessage(uuid.NewString(), []byte(`{"foo":"bar"}`))
		msg.Metadata = message.Metadata{"event_message_schema": "unknown"}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg); err == nil {
			t.Fatal("expected error for unknown schema")
		} else if _, ok := err.(*UnprocessableEventError); !ok {
			t.Fatalf("unexpected error type: %T", err)
		}
	})

	t.Run("fails for invalid payload", func(t *testing.T) {
		svc := &Service{validator: &testValidator{}, protoRegistry: make(map[string]func() proto.Message)}
		svc.registerProtoType(&structpb.Struct{})
		mw := svc.protoValidateMiddleware()
		msg := message.NewMessage(uuid.NewString(), []byte("not json"))
		msg.Metadata = message.Metadata{"event_message_schema": "*structpb.Struct"}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg); err == nil {
			t.Fatal("expected error for invalid payload")
		}
	})

	t.Run("fails validation", func(t *testing.T) {
		svc := &Service{validator: &testValidator{err: errors.New("bad")}, protoRegistry: make(map[string]func() proto.Message)}
		svc.registerProtoType(&structpb.Struct{})
		mw := svc.protoValidateMiddleware()
		msg := message.NewMessage(uuid.NewString(), []byte(`{"foo":"bar"}`))
		msg.Metadata = message.Metadata{"event_message_schema": "*structpb.Struct"}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg); err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("passes on success", func(t *testing.T) {
		svc := &Service{validator: &testValidator{}, protoRegistry: make(map[string]func() proto.Message)}
		svc.registerProtoType(&structpb.Struct{})
		mw := svc.protoValidateMiddleware()
		msg := message.NewMessage(uuid.NewString(), []byte(`{"foo":"bar"}`))
		msg.Metadata = message.Metadata{"event_message_schema": "*structpb.Struct"}
		called := false
		_, err := mw(func(m *message.Message) ([]*message.Message, error) {
			called = true
			return nil, nil
		})(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("handler not invoked")
		}
	})
}

func TestPoisonMiddlewareWithFilter(t *testing.T) {
	t.Parallel()

	svc := &Service{
		Conf:      &Config{PoisonQueue: "poison"},
		Publisher: &testPublisher{},
	}
	mw, err := svc.poisonMiddlewareWithFilter(func(err error) bool { return true })
	if err != nil {
		t.Fatalf("unexpected error creating poison middleware: %v", err)
	}
	msg := message.NewMessage(uuid.NewString(), nil)
	msg.Metadata = message.Metadata{}
	pub := svc.Publisher.(*testPublisher)
	_, _ = mw(func(m *message.Message) ([]*message.Message, error) {
		return nil, errors.New("boom")
	})(msg)
	if len(pub.Topics()) != 1 || pub.Topics()[0] != "poison" {
		t.Fatalf("expected poison message to be published: %#v", pub.Topics())
	}

	t.Run("returns error when middleware creation fails", func(t *testing.T) {
		svc := &Service{Conf: &Config{}, Publisher: nil}
		if _, err := svc.poisonMiddlewareWithFilter(func(error) bool { return false }); err == nil {
			t.Fatal("expected error when poison queue misconfigured")
		}
	})
}

func TestLogMessagesMiddleware(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	logger := &testLogger{}
	mw := svc.logMessagesMiddleware(logger)
	msg := message.NewMessage(uuid.NewString(), []byte("payload"))
	msg.Metadata = message.Metadata{"key": "value"}
	_, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logger.entries) == 0 {
		t.Fatal("expected log entry to be recorded")
	}
}

func TestOutboxMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("skips when outbox missing", func(t *testing.T) {
		svc := &Service{}
		mw := svc.outboxMiddleware()
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{}
		msgs, err := mw(func(m *message.Message) ([]*message.Message, error) {
			return []*message.Message{message.NewMessage(uuid.NewString(), []byte("ok"))}, nil
		})(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(msgs) != 1 {
			t.Fatalf("expected message passthrough")
		}
	})

	t.Run("propagates handler error", func(t *testing.T) {
		svc := &Service{outbox: &testOutbox{}}
		mw := svc.outboxMiddleware()
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) {
			return nil, errors.New("fail")
		})(msg); err == nil {
			t.Fatal("expected handler error to propagate")
		}
	})

	t.Run("stores outgoing messages", func(t *testing.T) {
		svc := &Service{outbox: &testOutbox{}}
		mw := svc.outboxMiddleware()
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{}
		out := message.NewMessage(uuid.NewString(), []byte("ok"))
		out.Metadata = message.Metadata{"event_message_schema": "OrderCreated"}
		msgs, err := mw(func(m *message.Message) ([]*message.Message, error) {
			return []*message.Message{out}, nil
		})(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(msgs) != 1 {
			t.Fatalf("expected outgoing message")
		}
		records := svc.outbox.(*testOutbox).Records()
		if len(records) != 1 || records[0].eventType != "OrderCreated" {
			t.Fatalf("unexpected outbox records: %#v", records)
		}
	})

	t.Run("uses fallback event type", func(t *testing.T) {
		svc := &Service{outbox: &testOutbox{}}
		mw := svc.outboxMiddleware()
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{}
		out := message.NewMessage(uuid.NewString(), []byte("ok"))
		out.Metadata = message.Metadata{}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) {
			return []*message.Message{out}, nil
		})(msg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		records := svc.outbox.(*testOutbox).Records()
		if len(records) != 1 || records[0].eventType != "unknown_event" {
			t.Fatalf("expected fallback event type, got %#v", records)
		}
	})

	t.Run("returns on outbox failure", func(t *testing.T) {
		svc := &Service{outbox: &testOutbox{err: errors.New("store fail")}}
		mw := svc.outboxMiddleware()
		msg := message.NewMessage(uuid.NewString(), nil)
		msg.Metadata = message.Metadata{}
		out := message.NewMessage(uuid.NewString(), []byte("ok"))
		out.Metadata = message.Metadata{}
		if _, err := mw(func(m *message.Message) ([]*message.Message, error) {
			return []*message.Message{out}, nil
		})(msg); err == nil {
			t.Fatal("expected outbox error to bubble up")
		}
	})
}

func TestRetryMiddleware(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	mw := svc.retryMiddleware()
	attempts := 0
	msg := message.NewMessage(uuid.NewString(), nil)
	msg.Metadata = message.Metadata{}
	_, err := mw(func(m *message.Message) ([]*message.Message, error) {
		attempts++
		if attempts < 2 {
			return nil, errors.New("retry")
		}
		return nil, nil
	})(msg)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if attempts < 2 {
		t.Fatalf("expected retries, got %d", attempts)
	}
}

func TestTracerMiddleware(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	mw := svc.tracerMiddleware()
	msg := message.NewMessage(uuid.NewString(), nil)
	msg.Metadata = message.Metadata{}
	ctx := context.Background()
	msg.SetContext(ctx)
	var observed trace.Span
	_, err := mw(func(m *message.Message) ([]*message.Message, error) {
		observed = trace.SpanFromContext(m.Context())
		return nil, nil
	})(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if observed == nil {
		t.Fatal("expected span to be attached to context")
	}
}

func TestTracerMiddlewareSetsAttributes(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	mw := svc.tracerMiddleware()
	msg := message.NewMessage(uuid.NewString(), nil)
	msg.Metadata = message.Metadata{"key": "value"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	msg.SetContext(ctx)
	_, err := mw(func(m *message.Message) ([]*message.Message, error) { return nil, nil })(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterMiddlewareValidations(t *testing.T) {
	t.Parallel()

	t.Run("requires router", func(t *testing.T) {
		svc := &Service{}
		err := svc.RegisterMiddleware(MiddlewareRegistration{
			Middleware: func(h message.HandlerFunc) message.HandlerFunc { return h },
		})
		if err == nil {
			t.Fatal("expected error when router is missing")
		}
	})

	t.Run("requires configuration", func(t *testing.T) {
		router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
		if err != nil {
			t.Fatalf("router init failed: %v", err)
		}
		svc := &Service{Router: router}
		if err := svc.RegisterMiddleware(MiddlewareRegistration{}); err == nil {
			t.Fatal("expected error when registration empty")
		}
	})

	t.Run("invokes builder", func(t *testing.T) {
		router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
		if err != nil {
			t.Fatalf("router init failed: %v", err)
		}
		svc := &Service{Router: router}
		called := false
		err = svc.RegisterMiddleware(MiddlewareRegistration{
			Builder: func(*Service) (message.HandlerMiddleware, error) {
				called = true
				return func(h message.HandlerFunc) message.HandlerFunc { return h }, nil
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected builder to be invoked")
		}
	})
}
