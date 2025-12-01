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

	t.Run("basic functionality", func(t *testing.T) {
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
	})

	t.Run("metadata enrichment", func(t *testing.T) {
		evt := protoflow.JSONMessageContext[*demoEvent]{
			Payload: &demoEvent{
				ID:   456,
				Date: &domain.Date{Year: 2023, Month: 12, Day: 25},
			},
		}

		result, err := handler(ctx, evt)
		if err != nil {
			t.Fatalf("demoHandler returned error: %v", err)
		}

		if result[0].Metadata["handler"] != "exampleRecordHandler" {
			t.Errorf("expected handler metadata, got %v", result[0].Metadata)
		}

		if result[0].Metadata["next_queue"] != "demo_processed_events" {
			t.Errorf("expected next_queue metadata, got %v", result[0].Metadata)
		}

		expectedKeys := []string{"handler", "processed_at", "next_queue"}
		for _, key := range expectedKeys {
			if _, ok := result[0].Metadata[key]; !ok {
				t.Errorf("expected metadata key %q to be present", key)
			}
		}
	})

	t.Run("time is set", func(t *testing.T) {
		before := time.Now()
		evt := protoflow.JSONMessageContext[*demoEvent]{
			Payload: &demoEvent{
				ID:   200,
				Date: &domain.Date{Year: 2024, Month: 6, Day: 1},
			},
		}

		result, err := handler(ctx, evt)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}

		after := time.Now()

		if result[0].Message.Time.Before(before) || result[0].Message.Time.After(after) {
			t.Error("Time should be set to approximately now")
		}
	})

	t.Run("preserves ID", func(t *testing.T) {
		testIDs := []int{0, 1, 100, -1, 999999}
		for _, id := range testIDs {
			evt := protoflow.JSONMessageContext[*demoEvent]{
				Payload: &demoEvent{
					ID:   id,
					Date: &domain.Date{Year: 2024, Month: 1, Day: 1},
				},
			}

			result, err := handler(ctx, evt)
			if err != nil {
				t.Fatalf("handler returned error for ID %d: %v", id, err)
			}

			if result[0].Message.ID != id {
				t.Errorf("ID mismatch: got %d, want %d", result[0].Message.ID, id)
			}
		}
	})

	t.Run("varying dates", func(t *testing.T) {
		dates := []*domain.Date{
			{Year: 2020, Month: 1, Day: 1},
			{Year: 2024, Month: 6, Day: 15},
			{Year: 2030, Month: 12, Day: 31},
		}

		for i, date := range dates {
			evt := protoflow.JSONMessageContext[*demoEvent]{
				Payload: &demoEvent{ID: i, Date: date},
			}

			result, err := handler(ctx, evt)
			if err != nil {
				t.Fatalf("handler returned error for date %d: %v", i, err)
			}

			if result[0].Message.Date.Year != date.Year {
				t.Errorf("Date year mismatch for case %d", i)
			}
		}
	})

	t.Run("concurrent calls", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				evt := protoflow.JSONMessageContext[*demoEvent]{
					Payload: &demoEvent{
						ID:   id,
						Date: &domain.Date{Year: 2024, Month: 1, Day: 1},
					},
				}
				_, err := handler(ctx, evt)
				done <- (err == nil)
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestExampleRecordHandler(t *testing.T) {
	handler := exampleRecordHandler()
	ctx := context.Background()

	t.Run("basic functionality", func(t *testing.T) {
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

				exampleResult, ok := result[0].Message.(*domain.ExampleResult)
				if !ok {
					t.Fatalf("expected *domain.ExampleResult, got %T", result[0].Message)
				}

				if exampleResult.RecordId != "TEST-001" {
					t.Errorf("expected RecordId TEST-001, got %s", exampleResult.RecordId)
				}

				validStatuses := map[string]bool{"queued": true, "in-progress": true, "completed": true}
				if !validStatuses[exampleResult.Status] {
					t.Errorf("unexpected status: %s", exampleResult.Status)
				}
			}
		}

		if successCount == 0 {
			t.Error("expected at least some successful handler calls")
		}
	})

	t.Run("metadata enrichment", func(t *testing.T) {
		evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{
			Payload: &domain.ExampleRecord{RecordId: "TEST-002", Title: "Test"},
		}

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
	})

	t.Run("processed date is set", func(t *testing.T) {
		evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{
			Payload: &domain.ExampleRecord{RecordId: "DATE-001", Title: "Date Test"},
		}

		for i := 0; i < 30; i++ {
			result, err := handler(ctx, evt)
			if err == nil && len(result) > 0 {
				exampleResult := result[0].Message.(*domain.ExampleResult)
				if exampleResult.ProcessedOn == nil {
					t.Error("ProcessedOn should be set")
					return
				}
				if exampleResult.ProcessedOn.Year < 2020 || exampleResult.ProcessedOn.Year > 2100 {
					t.Errorf("ProcessedOn Year out of range: %d", exampleResult.ProcessedOn.Year)
				}
				return
			}
		}
	})

	t.Run("statuses distribution", func(t *testing.T) {
		seenStatuses := make(map[string]bool)
		validStatuses := map[string]bool{"queued": true, "in-progress": true, "completed": true}

		evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{
			Payload: &domain.ExampleRecord{RecordId: "STATUS-TEST", Title: "Status Test"},
		}

		for i := 0; i < 100; i++ {
			result, err := handler(ctx, evt)
			if err == nil && len(result) > 0 {
				status := result[0].Message.(*domain.ExampleResult).Status
				if !validStatuses[status] {
					t.Errorf("Invalid status: %s", status)
				}
				seenStatuses[status] = true
			}
		}

		if len(seenStatuses) == 0 {
			t.Error("No successful results")
		}
	})
}

func TestRunExampleSimulation(t *testing.T) {
	t.Run("cancelled context returns immediately", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})
		go func() {
			RunExampleSimulation(ctx, nil, &Config{
				ExampleConsumeQueue: "test-queue",
			})
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("RunExampleSimulation did not return after context cancellation")
		}
	})
}

func TestDemoEventStruct(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		evt := &demoEvent{}
		if evt.ID != 0 {
			t.Errorf("zero value ID should be 0, got %d", evt.ID)
		}
		if evt.Date != nil {
			t.Error("zero value Date should be nil")
		}
	})

	t.Run("with values", func(t *testing.T) {
		evt := &demoEvent{
			ID:   42,
			Date: &domain.Date{Year: 2024, Month: 1, Day: 15},
		}

		if evt.ID != 42 {
			t.Errorf("expected ID 42, got %d", evt.ID)
		}
		if evt.Date.Year != 2024 {
			t.Errorf("expected year 2024, got %d", evt.Date.Year)
		}
	})
}

func TestProcessedDemoEventStruct(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		evt := &processedDemoEvent{}
		if evt.ID != 0 {
			t.Errorf("zero value ID should be 0, got %d", evt.ID)
		}
		if !evt.Time.IsZero() {
			t.Error("zero value Time should be zero time")
		}
		if evt.Date != nil {
			t.Error("zero value Date should be nil")
		}
	})

	t.Run("with values", func(t *testing.T) {
		now := time.Now()
		evt := &processedDemoEvent{
			ID:   123,
			Time: now,
			Date: &domain.Date{Year: 2024, Month: 6, Day: 15},
		}

		if evt.ID != 123 {
			t.Errorf("expected ID 123, got %d", evt.ID)
		}
		if evt.Time != now {
			t.Errorf("expected time %v, got %v", now, evt.Time)
		}
	})
}

func TestRunSomeSimulation(t *testing.T) {
	t.Run("cancelled context returns immediately", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})
		go func() {
			runSomeSimulation(ctx, nil, "test-queue")
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("runSomeSimulation should exit immediately when context is cancelled")
		}
	})
}
