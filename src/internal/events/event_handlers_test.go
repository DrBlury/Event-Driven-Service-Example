package events

import (
	"context"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

func TestDemoHandler(t *testing.T) {
	handler := demoHandler()

	ctx := context.Background()
	evt := protoflow.JSONMessageContext[*demoEvent]{
		Payload: &demoEvent{
			ID: 123,
			Date: &domain.Date{
				Year:  2024,
				Month: 6,
				Day:   15,
			},
		},
	}

	result, err := handler(ctx, evt)

	if err != nil {
		t.Fatalf("demoHandler returned error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	if result[0].Message.ID != 123 {
		t.Errorf("expected ID 123, got %d", result[0].Message.ID)
	}

	if result[0].Message.Date.Year != 2024 {
		t.Errorf("expected year 2024, got %d", result[0].Message.Date.Year)
	}
}

func TestDemoHandler_WithMetadata(t *testing.T) {
	handler := demoHandler()

	ctx := context.Background()
	evt := protoflow.JSONMessageContext[*demoEvent]{
		Payload: &demoEvent{
			ID: 456,
			Date: &domain.Date{
				Year:  2023,
				Month: 12,
				Day:   25,
			},
		},
	}

	result, err := handler(ctx, evt)

	if err != nil {
		t.Fatalf("demoHandler returned error: %v", err)
	}

	// Check metadata was enriched
	if result[0].Metadata["handler"] != "exampleRecordHandler" {
		t.Errorf("expected handler metadata, got %v", result[0].Metadata)
	}

	if result[0].Metadata["next_queue"] != "demo_processed_events" {
		t.Errorf("expected next_queue metadata, got %v", result[0].Metadata)
	}
}

func TestExampleRecordHandler(t *testing.T) {
	handler := exampleRecordHandler()

	ctx := context.Background()
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{
		Payload: &domain.ExampleRecord{
			RecordId:    "TEST-001",
			Title:       "Test Record",
			Description: "Test description",
			Tags:        []string{"test", "unit"},
			Meta: &domain.ExampleMeta{
				RequestedBy:      "test-user",
				RequiresFollowUp: true,
				Priority:         3,
				DesiredStartDate: &domain.Date{
					Year:  2024,
					Month: 7,
					Day:   1,
				},
			},
		},
	}

	// Run multiple times to test random behavior
	successCount := 0
	for i := 0; i < 20; i++ {
		result, err := handler(ctx, evt)
		if err == nil {
			successCount++
			if len(result) != 1 {
				t.Fatalf("expected 1 result, got %d", len(result))
			}

			// Check result is ExampleResult
			exampleResult, ok := result[0].Message.(*domain.ExampleResult)
			if !ok {
				t.Fatalf("expected *domain.ExampleResult, got %T", result[0].Message)
			}

			if exampleResult.RecordId != "TEST-001" {
				t.Errorf("expected RecordId TEST-001, got %s", exampleResult.RecordId)
			}

			// Check status is one of the valid values
			validStatuses := map[string]bool{"queued": true, "in-progress": true, "completed": true}
			if !validStatuses[exampleResult.Status] {
				t.Errorf("unexpected status: %s", exampleResult.Status)
			}
		}
	}

	// Should have some successes (90% success rate expected)
	if successCount == 0 {
		t.Error("expected at least some successful handler calls")
	}
}

func TestExampleRecordHandler_MetadataEnrichment(t *testing.T) {
	handler := exampleRecordHandler()

	ctx := context.Background()
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{
		Payload: &domain.ExampleRecord{
			RecordId: "TEST-002",
			Title:    "Test",
		},
	}

	// Keep trying until we get a success
	for i := 0; i < 50; i++ {
		result, err := handler(ctx, evt)
		if err == nil {
			if result[0].Metadata["handler"] != "exampleRecordHandler" {
				t.Errorf("expected handler metadata")
			}
			if result[0].Metadata["processed_at"] == "" {
				t.Error("expected processed_at metadata")
			}
			return
		}
	}
	t.Error("could not get successful handler call after 50 attempts")
}

func TestRunExampleSimulation_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should return immediately without blocking
	done := make(chan struct{})
	go func() {
		RunExampleSimulation(ctx, nil, &Config{
			ExampleConsumeQueue: "test-queue",
		})
		close(done)
	}()

	select {
	case <-done:
		// Success - returned quickly
	case <-time.After(1 * time.Second):
		t.Error("RunExampleSimulation did not return after context cancellation")
	}
}

func TestDemoEventStruct(t *testing.T) {
	evt := &demoEvent{
		ID: 42,
		Date: &domain.Date{
			Year:  2024,
			Month: 1,
			Day:   15,
		},
	}

	if evt.ID != 42 {
		t.Errorf("expected ID 42, got %d", evt.ID)
	}
	if evt.Date.Year != 2024 {
		t.Errorf("expected year 2024, got %d", evt.Date.Year)
	}
}

func TestProcessedDemoEventStruct(t *testing.T) {
	now := time.Now()
	evt := &processedDemoEvent{
		ID:   123,
		Time: now,
		Date: &domain.Date{
			Year:  2024,
			Month: 6,
			Day:   15,
		},
	}

	if evt.ID != 123 {
		t.Errorf("expected ID 123, got %d", evt.ID)
	}
	if evt.Time != now {
		t.Errorf("expected time %v, got %v", now, evt.Time)
	}
}
