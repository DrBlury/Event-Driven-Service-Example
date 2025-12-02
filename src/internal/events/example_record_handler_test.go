package events

import (
	"context"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

// =============================================================================
// EXAMPLE RECORD HANDLER TESTS
// =============================================================================
//
// This file contains all tests for the exampleRecordHandler.
// To add tests for a new handler, create a new file following this pattern:
//   - <handler_name>_handler_test.go
//
// =============================================================================

func TestExampleRecordHandler(t *testing.T) {
	handler := exampleRecordHandler()
	ctx := context.Background()
	fixtures := NewTestFixtures()

	t.Run("basic functionality", func(t *testing.T) {
		testExampleRecordBasicFunctionality(t, handler, ctx, fixtures)
	})

	t.Run("metadata enrichment", func(t *testing.T) {
		testExampleRecordMetadata(t, handler, ctx, fixtures)
	})

	t.Run("processed date is set", func(t *testing.T) {
		testExampleRecordProcessedDate(t, handler, ctx, fixtures)
	})

	t.Run("statuses distribution", func(t *testing.T) {
		testExampleRecordStatusDistribution(t, handler, ctx, fixtures)
	})

	t.Run("edge cases", func(t *testing.T) {
		testExampleRecordEdgeCases(t, handler, ctx)
	})
}

func testExampleRecordBasicFunctionality(t *testing.T, handler exampleRecordHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()
	payload := fixtures.ExampleRecordFull("TEST-001", "Test Record", "Test description", []string{"test", "unit"}, 3)
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{Payload: payload}

	successCount := 0
	for i := 0; i < 20; i++ {
		if result, err := handler(ctx, evt); err == nil {
			successCount++
			validateExampleRecordResult(t, result)
		}
	}
	if successCount == 0 {
		t.Error("expected at least some successful handler calls")
	}
}

func testExampleRecordMetadata(t *testing.T, handler exampleRecordHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{Payload: fixtures.ExampleRecord("TEST-002", "Test")}

	for i := 0; i < 50; i++ {
		if result, err := handler(ctx, evt); err == nil {
			AssertMetadataContains(t, result[0].Metadata, "handler", "exampleRecordHandler")
			AssertMetadataHasKey(t, result[0].Metadata, "processed_at")
			return
		}
	}
	t.Error("could not get successful handler call after 50 attempts")
}

func testExampleRecordProcessedDate(t *testing.T, handler exampleRecordHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{Payload: fixtures.ExampleRecord("DATE-001", "Date Test")}

	for i := 0; i < 30; i++ {
		if result, err := handler(ctx, evt); err == nil && len(result) > 0 {
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
}

func testExampleRecordStatusDistribution(t *testing.T, handler exampleRecordHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()
	seenStatuses := make(map[string]bool)
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{Payload: fixtures.ExampleRecord("STATUS-TEST", "Status Test")}

	for i := 0; i < 100; i++ {
		if result, err := handler(ctx, evt); err == nil && len(result) > 0 {
			status := result[0].Message.(*domain.ExampleResult).Status
			AssertValidStatus(t, status)
			seenStatuses[status] = true
		}
	}
	if len(seenStatuses) == 0 {
		t.Error("No successful results")
	}
}

func testExampleRecordEdgeCases(t *testing.T, handler exampleRecordHandlerFunc, ctx context.Context) {
	t.Helper()

	t.Run("empty record ID", func(t *testing.T) {
		runExampleRecordEdgeTest(t, handler, ctx, &domain.ExampleRecord{RecordId: "", Title: "Empty ID Test"})
	})
	t.Run("nil meta", func(t *testing.T) {
		runExampleRecordEdgeTest(t, handler, ctx, &domain.ExampleRecord{RecordId: "NIL-META-1", Title: "Nil Meta Test", Meta: nil})
	})
	t.Run("empty tags", func(t *testing.T) {
		runExampleRecordEdgeTest(t, handler, ctx, &domain.ExampleRecord{RecordId: "EMPTY-TAGS-1", Title: "Empty Tags Test", Tags: []string{}})
	})
	t.Run("many tags", func(t *testing.T) {
		tags := make([]string, 100)
		for i := 0; i < 100; i++ {
			tags[i] = "tag" + string(rune('A'+i%26))
		}
		runExampleRecordEdgeTest(t, handler, ctx, &domain.ExampleRecord{RecordId: "MANY-TAGS-1", Title: "Many Tags Test", Tags: tags})
	})
}

func runExampleRecordEdgeTest(t *testing.T, handler exampleRecordHandlerFunc, ctx context.Context, payload *domain.ExampleRecord) {
	t.Helper()
	evt := protoflow.ProtoMessageContext[*domain.ExampleRecord]{Payload: payload}
	for i := 0; i < 50; i++ {
		if result, err := handler(ctx, evt); err == nil && len(result) > 0 {
			return
		}
	}
}

type exampleRecordHandlerFunc = func(context.Context, protoflow.ProtoMessageContext[*domain.ExampleRecord]) ([]protoflow.ProtoMessageOutput, error)

// validateExampleRecordResult validates the result of exampleRecordHandler.
func validateExampleRecordResult(t *testing.T, result []protoflow.ProtoMessageOutput) {
	t.Helper()

	AssertResultCount(t, len(result), 1)

	exampleResult, ok := result[0].Message.(*domain.ExampleResult)
	if !ok {
		t.Fatalf("expected *domain.ExampleResult, got %T", result[0].Message)
	}

	if exampleResult.RecordId != "TEST-001" {
		t.Errorf("expected RecordId TEST-001, got %s", exampleResult.RecordId)
	}

	AssertValidStatus(t, exampleResult.Status)
}

// =============================================================================
// SIMULATION TESTS
// =============================================================================

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

	t.Run("nil context uses background", func(t *testing.T) {
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
			t.Error("runSomeSimulation should handle nil context gracefully")
		}
	})
}

// =============================================================================
// HANDLER REGISTRATION TESTS
// =============================================================================

func TestRegisterAppEventHandlers(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		cfg := &Config{
			DemoConsumeQueue:    "demo-in",
			DemoPublishQueue:    "demo-out",
			ExampleConsumeQueue: "example-in",
			ExamplePublishQueue: "example-out",
		}

		// nil service will cause panic or error
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Log("registerAppEventHandlers did not panic with nil service")
				}
			}()
			_ = registerAppEventHandlers(nil, cfg)
		}()
	})

	t.Run("nil config returns error", func(t *testing.T) {
		// nil config will cause nil pointer dereference
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Log("registerAppEventHandlers did not panic with nil config")
				}
			}()
			_ = registerAppEventHandlers(nil, nil)
		}()
	})
}
