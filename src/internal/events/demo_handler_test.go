package events

import (
	"context"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

// =============================================================================
// DEMO HANDLER TESTS
// =============================================================================
//
// This file contains all tests for the demoHandler.
// To add tests for a new handler, create a new file following this pattern:
//   - <handler_name>_handler_test.go
//
// =============================================================================

func TestDemoHandler(t *testing.T) {
	handler := demoHandler()
	ctx := context.Background()
	fixtures := NewTestFixtures()

	t.Run("core functionality", func(t *testing.T) {
		testDemoHandlerCoreFunctionality(t, handler, ctx, fixtures)
	})

	t.Run("ID preservation", func(t *testing.T) {
		testDemoHandlerIDPreservation(t, handler, ctx, fixtures)
	})

	t.Run("date handling", func(t *testing.T) {
		testDemoHandlerDateHandling(t, handler, ctx, fixtures)
	})

	t.Run("edge cases", func(t *testing.T) {
		testDemoHandlerEdgeCases(t, handler, ctx, fixtures)
	})

	t.Run("concurrent calls", func(t *testing.T) {
		RunConcurrentHandlerTest(t, handler, func(id int) *demoEvent {
			return fixtures.DemoEvent(id, 2024, 1, 1)
		}, 10)
	})
}

func testDemoHandlerCoreFunctionality(t *testing.T, handler demoHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()

	t.Run("basic functionality", func(t *testing.T) {
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEvent(123, 2024, 6, 15)}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "demoHandler returned error")
		AssertResultCount(t, len(result), 1)
		AssertEqual(t, result[0].Message.ID, 123, "ID mismatch")
		AssertEqual(t, result[0].Message.Date.Year, int32(2024), "Year mismatch")
	})

	t.Run("metadata enrichment", func(t *testing.T) {
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEvent(456, 2023, 12, 25)}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "demoHandler returned error")
		AssertMetadataContains(t, result[0].Metadata, "handler", "exampleRecordHandler")
		AssertMetadataContains(t, result[0].Metadata, "next_queue", "demo_processed_events")
		AssertMetadataHasKey(t, result[0].Metadata, "processed_at")
	})

	t.Run("time is set to now", func(t *testing.T) {
		before := time.Now()
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEvent(200, 2024, 6, 1)}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "handler returned error")
		AssertTimeInRange(t, result[0].Message.Time, before, time.Now())
	})
}

func testDemoHandlerIDPreservation(t *testing.T, handler demoHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()
	for _, id := range []int{0, 1, 100, -1, 999999} {
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEvent(id, 2024, 1, 1)}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "handler returned error for ID")
		AssertEqual(t, result[0].Message.ID, id, "ID mismatch")
	}
}

func testDemoHandlerDateHandling(t *testing.T, handler demoHandlerFunc, ctx context.Context, _ *TestFixtures) {
	t.Helper()

	t.Run("varying dates", func(t *testing.T) {
		dates := []*domain.Date{{Year: 2020, Month: 1, Day: 1}, {Year: 2024, Month: 6, Day: 15}, {Year: 2030, Month: 12, Day: 31}}
		for i, date := range dates {
			evt := protoflow.JSONMessageContext[*demoEvent]{Payload: &demoEvent{ID: i, Date: date}}
			result, err := handler(ctx, evt)
			AssertNoError(t, err, "handler returned error for date")
			AssertEqual(t, result[0].Message.Date.Year, date.Year, "Date year mismatch")
		}
	})

	t.Run("zero date values", func(t *testing.T) {
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: &demoEvent{ID: 0, Date: &domain.Date{Year: 0, Month: 0, Day: 0}}}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "handler returned error")
		AssertResultCount(t, len(result), 1)
	})
}

func testDemoHandlerEdgeCases(t *testing.T, handler demoHandlerFunc, ctx context.Context, fixtures *TestFixtures) {
	t.Helper()

	t.Run("negative ID", func(t *testing.T) {
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEvent(-999, 2024, 1, 1)}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "handler returned error")
		AssertEqual(t, result[0].Message.ID, -999, "ID mismatch")
	})

	t.Run("large ID", func(t *testing.T) {
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEvent(999999999, 2024, 1, 1)}
		result, err := handler(ctx, evt)
		AssertNoError(t, err, "handler returned error")
		AssertEqual(t, result[0].Message.ID, 999999999, "ID mismatch")
	})

	t.Run("nil date handling", func(t *testing.T) {
		defer func() { _ = recover() }()
		evt := protoflow.JSONMessageContext[*demoEvent]{Payload: fixtures.DemoEventNilDate(1)}
		_, _ = handler(ctx, evt)
	})
}

type demoHandlerFunc = func(context.Context, protoflow.JSONMessageContext[*demoEvent]) ([]protoflow.JSONMessageOutput[*processedDemoEvent], error)

// =============================================================================
// DEMO EVENT STRUCT TESTS
// =============================================================================

func TestDemoEventStruct(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		evt := &demoEvent{}
		AssertEqual(t, evt.ID, 0, "zero value ID should be 0")
		if evt.Date != nil {
			t.Error("zero value Date should be nil")
		}
	})

	t.Run("with values", func(t *testing.T) {
		evt := &demoEvent{
			ID:   42,
			Date: &domain.Date{Year: 2024, Month: 1, Day: 15},
		}
		AssertEqual(t, evt.ID, 42, "ID mismatch")
		AssertEqual(t, evt.Date.Year, int32(2024), "Year mismatch")
	})

	t.Run("various ID values", func(t *testing.T) {
		ids := []int{0, 1, -1, 100, -100, 1000000, -1000000}
		for _, id := range ids {
			evt := &demoEvent{ID: id, Date: nil}
			AssertEqual(t, evt.ID, id, "ID mismatch")
		}
	})

	t.Run("various date values", func(t *testing.T) {
		dates := []*domain.Date{
			nil,
			{Year: 0, Month: 0, Day: 0},
			{Year: 2024, Month: 1, Day: 1},
			{Year: 9999, Month: 12, Day: 31},
		}

		for i, date := range dates {
			evt := &demoEvent{ID: i, Date: date}
			if evt.Date != date {
				t.Errorf("Date mismatch at index %d", i)
			}
		}
	})
}

// =============================================================================
// PROCESSED DEMO EVENT STRUCT TESTS
// =============================================================================

func TestProcessedDemoEventStruct(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		evt := &processedDemoEvent{}
		AssertEqual(t, evt.ID, 0, "zero value ID should be 0")
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
		AssertEqual(t, evt.ID, 123, "ID mismatch")
		if evt.Time != now {
			t.Errorf("expected time %v, got %v", now, evt.Time)
		}
	})

	t.Run("time handling", func(t *testing.T) {
		times := []time.Time{
			{},
			time.Now(),
			time.Now().Add(-24 * time.Hour),
			time.Now().Add(24 * time.Hour),
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		for i, tm := range times {
			evt := &processedDemoEvent{ID: i, Time: tm, Date: nil}
			if evt.Time != tm {
				t.Errorf("Time mismatch at index %d", i)
			}
		}
	})
}
