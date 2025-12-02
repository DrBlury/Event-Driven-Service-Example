package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"drblury/event-driven-service/internal/domain"

	"github.com/drblury/protoflow"
)

// =============================================================================
// TEST FIXTURES
// =============================================================================

// TestFixtures provides factory methods for creating test data.
type TestFixtures struct{}

// NewTestFixtures creates a new TestFixtures instance.
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{}
}

// DemoEvent creates a demoEvent with the given parameters.
func (f *TestFixtures) DemoEvent(id int, year, month, day int32) *demoEvent {
	return &demoEvent{
		ID: id,
		Date: &domain.Date{
			Year:  year,
			Month: month,
			Day:   day,
		},
	}
}

// DemoEventNilDate creates a demoEvent with a nil date.
func (f *TestFixtures) DemoEventNilDate(id int) *demoEvent {
	return &demoEvent{
		ID:   id,
		Date: nil,
	}
}

// ExampleRecord creates a basic ExampleRecord with the given ID and title.
func (f *TestFixtures) ExampleRecord(recordID, title string) *domain.ExampleRecord {
	return &domain.ExampleRecord{
		RecordId: recordID,
		Title:    title,
	}
}

// ExampleRecordFull creates a full ExampleRecord with all fields populated.
func (f *TestFixtures) ExampleRecordFull(recordID, title, description string, tags []string, priority int32) *domain.ExampleRecord {
	return &domain.ExampleRecord{
		RecordId:    recordID,
		Title:       title,
		Description: description,
		Tags:        tags,
		Meta: &domain.ExampleMeta{
			Priority: priority,
		},
	}
}

// =============================================================================
// ASSERTION HELPERS
// =============================================================================

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertResultCount fails the test if got != want.
func AssertResultCount(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected %d results, got %d", want, got)
	}
}

// AssertEqual fails the test if got != want.
func AssertEqual[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: expected %v, got %v", msg, want, got)
	}
}

// AssertMetadataContains fails the test if the metadata doesn't contain the expected key-value pair.
func AssertMetadataContains(t *testing.T, metadata protoflow.Metadata, key, expectedValue string) {
	t.Helper()
	if metadata == nil {
		t.Errorf("metadata is nil, expected key %q with value %q", key, expectedValue)
		return
	}
	value, ok := metadata[key]
	if !ok {
		t.Errorf("metadata missing key %q", key)
		return
	}
	if value != expectedValue {
		t.Errorf("metadata[%q] = %q, expected %q", key, value, expectedValue)
	}
}

// AssertMetadataHasKey fails the test if the metadata doesn't contain the expected key.
func AssertMetadataHasKey(t *testing.T, metadata protoflow.Metadata, key string) {
	t.Helper()
	if metadata == nil {
		t.Errorf("metadata is nil, expected key %q", key)
		return
	}
	if _, ok := metadata[key]; !ok {
		t.Errorf("metadata missing key %q", key)
	}
}

// AssertTimeInRange fails the test if the time is not within the given range.
func AssertTimeInRange(t *testing.T, got, before, after time.Time) {
	t.Helper()
	if got.Before(before) || got.After(after) {
		t.Errorf("time %v is not in range [%v, %v]", got, before, after)
	}
}

// AssertValidStatus fails the test if the status is not one of the valid statuses.
func AssertValidStatus(t *testing.T, status string) {
	t.Helper()
	validStatuses := map[string]bool{
		"queued":      true,
		"in-progress": true,
		"completed":   true,
	}
	if !validStatuses[status] {
		t.Errorf("invalid status: %q", status)
	}
}

// =============================================================================
// CONCURRENCY TEST HELPERS
// =============================================================================

// RunConcurrentHandlerTest runs the handler concurrently with the given number of goroutines.
func RunConcurrentHandlerTest(
	t *testing.T,
	handler protoflow.JSONMessageHandler[*demoEvent, *processedDemoEvent],
	createEvent func(id int) *demoEvent,
	concurrency int,
) {
	t.Helper()

	ctx := context.Background()
	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			evt := protoflow.JSONMessageContext[*demoEvent]{
				Payload: createEvent(id),
			}
			_, err := handler(ctx, evt)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent handler call failed: %v", err)
	}
}
