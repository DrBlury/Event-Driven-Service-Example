package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
	"google.golang.org/protobuf/proto"
)

// mockProducer implements protoflow.Producer for testing
type mockProducer struct {
	published     []proto.Message
	publishErr    error
	publishCalled bool
}

func (m *mockProducer) PublishProto(ctx context.Context, topic string, msg proto.Message, md protoflow.Metadata) error {
	m.publishCalled = true
	if m.publishErr != nil {
		return m.publishErr
	}
	m.published = append(m.published, msg)
	return nil
}

func TestNewAppLogic(t *testing.T) {
	logger := slog.Default()

	logic, err := NewAppLogic(nil, logger)
	if err != nil {
		t.Fatalf("NewAppLogic returned error: %v", err)
	}
	if logic == nil {
		t.Error("NewAppLogic returned nil")
	}
	if logic.log != logger {
		t.Error("Logger not set correctly")
	}
}

func TestNewAppLogicWithNilLogger(t *testing.T) {
	logic, err := NewAppLogic(nil, nil)
	if err != nil {
		t.Fatalf("NewAppLogic returned error: %v", err)
	}
	if logic == nil {
		t.Error("NewAppLogic returned nil")
	}
}

func TestSetEventProducer(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	producer := &mockProducer{}

	logic.SetEventProducer(producer)

	if logic.eventProducer == nil {
		t.Error("Event producer not set")
	}
}

func TestSetEventProducerNilReceiver(t *testing.T) {
	var logic *AppLogic
	producer := &mockProducer{}

	// Should not panic
	logic.SetEventProducer(producer)
}

func TestSetExampleTopic(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	topic := "test-topic"

	logic.SetExampleTopic(topic)

	if logic.exampleTopic != topic {
		t.Errorf("Expected topic %q, got %q", topic, logic.exampleTopic)
	}
}

func TestSetExampleTopicNilReceiver(t *testing.T) {
	var logic *AppLogic

	// Should not panic
	logic.SetExampleTopic("test")
}

func TestExampleTopic(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	topic := "test-topic"
	logic.SetExampleTopic(topic)

	got := logic.ExampleTopic()
	if got != topic {
		t.Errorf("ExampleTopic() = %q, want %q", got, topic)
	}
}

func TestExampleTopicNilReceiver(t *testing.T) {
	var logic *AppLogic

	got := logic.ExampleTopic()
	if got != "" {
		t.Errorf("ExampleTopic() on nil receiver = %q, want empty string", got)
	}
}

func TestHandleExampleNilReceiver(t *testing.T) {
	var logic *AppLogic
	record := &domain.ExampleRecord{}

	err := logic.HandleExample(context.Background(), record, "token")
	if err == nil {
		t.Error("Expected error for nil receiver")
	}
	if err.Error() != "applogic is nil" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestHandleExampleNilRecord(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)

	err := logic.HandleExample(context.Background(), nil, "token")
	if err == nil {
		t.Error("Expected error for nil record")
	}
	if err.Error() != "example payload is required" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestHandleExampleWithoutDatabase(t *testing.T) {
	// When db is nil, HandleExample returns early without emitting or storing
	logic, _ := NewAppLogic(nil, nil)
	record := &domain.ExampleRecord{
		RecordId: "test-123",
		Title:    "Test Record",
	}

	err := logic.HandleExample(context.Background(), record, "token")
	if err != nil {
		t.Errorf("HandleExample with nil db should succeed: %v", err)
	}
}

func TestHandleExampleWithProducerNoDb(t *testing.T) {
	// Even with producer configured, if db is nil, it returns early
	logic, _ := NewAppLogic(nil, nil)
	producer := &mockProducer{}
	logic.SetEventProducer(producer)
	logic.SetExampleTopic("test-topic")

	record := &domain.ExampleRecord{
		RecordId: "test-123",
		Title:    "Test Record",
	}

	err := logic.HandleExample(context.Background(), record, "token")
	if err != nil {
		t.Errorf("HandleExample with nil db should succeed even with producer: %v", err)
	}
	// Producer should NOT be called because db is nil (early return)
	if producer.publishCalled {
		t.Error("Producer should not be called when db is nil")
	}
}

func TestDatabaseProbeNilReceiver(t *testing.T) {
	var logic *AppLogic

	err := logic.DatabaseProbe(context.Background())
	if err == nil {
		t.Error("Expected error for nil receiver")
	}
	if err.Error() != "applogic is nil" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestDatabaseProbeNilDatabase(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)

	err := logic.DatabaseProbe(context.Background())
	if err == nil {
		t.Error("Expected error for nil database")
	}
	if err.Error() != "database not configured" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestEmitExampleEventNilReceiver(t *testing.T) {
	var logic *AppLogic
	record := &domain.ExampleRecord{}

	err := logic.emitExampleEvent(context.Background(), record)
	if err == nil {
		t.Error("Expected error for nil receiver")
	}
}

func TestEmitExampleEventNilRecord(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)

	err := logic.emitExampleEvent(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil record")
	}
}

func TestEmitExampleEventNoTopic(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	record := &domain.ExampleRecord{}

	err := logic.emitExampleEvent(context.Background(), record)
	if err == nil {
		t.Error("Expected error for missing topic")
	}
	if err.Error() != "example topic not configured" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEmitExampleEventNoProducer(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	logic.SetExampleTopic("test-topic")
	record := &domain.ExampleRecord{}

	err := logic.emitExampleEvent(context.Background(), record)
	if err == nil {
		t.Error("Expected error for missing producer")
	}
	if err.Error() != "event producer not configured" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEmitExampleEventSuccess(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	producer := &mockProducer{}
	logic.SetEventProducer(producer)
	logic.SetExampleTopic("test-topic")

	record := &domain.ExampleRecord{
		RecordId: "test-123",
		Title:    "Test Record",
	}

	err := logic.emitExampleEvent(context.Background(), record)
	if err != nil {
		t.Errorf("emitExampleEvent failed: %v", err)
	}
	if !producer.publishCalled {
		t.Error("Producer.PublishProto was not called")
	}
	if len(producer.published) != 1 {
		t.Errorf("Expected 1 published message, got %d", len(producer.published))
	}
}

func TestEmitExampleEventPublishError(t *testing.T) {
	logic, _ := NewAppLogic(nil, nil)
	producer := &mockProducer{publishErr: errors.New("publish failed")}
	logic.SetEventProducer(producer)
	logic.SetExampleTopic("test-topic")

	record := &domain.ExampleRecord{
		RecordId: "test-123",
		Title:    "Test Record",
	}

	err := logic.emitExampleEvent(context.Background(), record)
	if err == nil {
		t.Error("Expected error when publish fails")
	}
	if err.Error() != "publish failed" {
		t.Errorf("Unexpected error: %v", err)
	}
}
