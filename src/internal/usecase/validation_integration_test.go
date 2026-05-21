package usecase

// validation_integration_test.go demonstrates that protovalidate CEL rules
// are enforced through the full usecase layer, not just at the domain boundary.
//
// By running validation inside HandleExample, invalid domain messages are
// rejected before any I/O (database writes, event publishing) takes place.
// This means:
//   - A record with an invalid UUID never reaches the database.
//   - A record with duplicate tags never hits the message broker.
//   - Cross-field violations (follow-up without sufficient priority, failed
//     result without a note) are caught at the earliest possible point.
//
// The test does NOT require a real database or broker — the validation fires
// before those subsystems are consulted.

import (
	"context"
	"strings"
	"testing"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/events"
)

// newLogicForValidation returns an AppLogic with a real protovalidate validator
// but without a database or event producer.
func newLogicForValidation(t *testing.T) *AppLogic {
	t.Helper()
	logic, err := NewAppLogic(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewAppLogic error: %v", err)
	}
	return logic
}

func assertHandleExampleValid(t *testing.T, logic *AppLogic, record *domain.ExampleRecord) {
	t.Helper()
	err := logic.HandleExample(context.Background(), record, "")
	if err != nil {
		t.Errorf("expected HandleExample to succeed but got: %v", err)
	}
}

func assertHandleExampleInvalid(t *testing.T, logic *AppLogic, record *domain.ExampleRecord, wantFragment string) {
	t.Helper()
	err := logic.HandleExample(context.Background(), record, "")
	if err == nil {
		t.Error("expected HandleExample to return a validation error but got nil")
		return
	}
	if wantFragment != "" && !strings.Contains(err.Error(), wantFragment) {
		t.Errorf("expected error to contain %q but got: %v", wantFragment, err)
	}
}

// ─── UUID validation ──────────────────────────────────────────────────────────

func TestHandleExample_RejectsNonUUID(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleInvalid(t, logic, &domain.ExampleRecord{
		RecordId: "plain-string-not-uuid",
		Title:    "Valid Title",
	}, "record_id must be a valid UUID")
}

func TestHandleExample_AcceptsValidUUID(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleValid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440000",
		Title:    "Valid Title",
	})
}

// ─── Title length ─────────────────────────────────────────────────────────────

func TestHandleExample_RejectsShortTitle(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleInvalid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440001",
		Title:    "Hi",
	}, "value length must be at least 3")
}

// ─── Tag duplicate detection ─────────────────────────────────────────────────

func TestHandleExample_RejectsDuplicateTags(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleInvalid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440002",
		Title:    "Duplicate Tag Record",
		Tags:     []string{"backend", "api", "backend"}, // "backend" twice
		Meta:     &domain.ExampleMeta{RequestedBy: "dev", Priority: 1},
	}, "tags must be unique")
}

// ─── Too many tags ───────────────────────────────────────────────────────────

func TestHandleExample_RejectsTooManyTags(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	tags := make([]string, 11)
	for i := range tags {
		tags[i] = string(rune('a' + i)) // a, b, c, …, k
	}
	assertHandleExampleInvalid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440003",
		Title:    "Too Many Tags",
		Tags:     tags,
		Meta:     &domain.ExampleMeta{RequestedBy: "dev", Priority: 1},
	}, "value must contain no more than 10 item")
}

// ─── Cross-field: tags without meta.requested_by ──────────────────────────────

func TestHandleExample_RejectsTagsWithoutMeta(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleInvalid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440004",
		Title:    "Tagged But No Owner",
		Tags:     []string{"critical"},
		// Meta is nil — cross-field rule fires
	}, "meta.requested_by must be set")
}

// ─── Cross-field: priority when follow-up required ───────────────────────────

func TestHandleExample_RejectsLowPriorityWithFollowUp(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleInvalid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440005",
		Title:    "Follow-Up Record",
		Meta: &domain.ExampleMeta{
			RequestedBy:      "manager",
			RequiresFollowUp: true,
			Priority:         2, // too low — must be ≥ 3 when follow-up is required
		},
	}, "priority must be at least 3 when follow-up is required")
}

func TestHandleExample_AcceptsSufficientPriorityWithFollowUp(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	assertHandleExampleValid(t, logic, &domain.ExampleRecord{
		RecordId: "550e8400-e29b-41d4-a716-446655440006",
		Title:    "High-Priority Follow-Up",
		Meta: &domain.ExampleMeta{
			RequestedBy:      "manager",
			RequiresFollowUp: true,
			Priority:         4,
		},
	})
}

// ─── ValidationError surfaces as domain.ErrValidations ───────────────────────

// The API layer maps domain.ErrValidations to HTTP 400 via the error classifier.
// This test confirms the error type is correct so the end-to-end mapping works.

func TestHandleExample_ValidationErrorType(t *testing.T) {
	t.Parallel()
	logic := newLogicForValidation(t)
	err := logic.HandleExample(context.Background(), &domain.ExampleRecord{
		RecordId: "bad-id",
		Title:    "Hi",
	}, "")
	if err == nil {
		t.Fatal("expected error")
	}
	var valErr domain.ErrValidations
	if _, ok := err.(domain.ErrValidations); !ok {
		_ = valErr
		t.Errorf("expected domain.ErrValidations but got %T: %v", err, err)
	}
}

// ─── Valid record with all constraints satisfied ──────────────────────────────

func TestHandleExample_AcceptsFullyValidRecord(t *testing.T) {
	t.Parallel()
	producer := &mockProducer{}
	logic, err := NewAppLogic(nil, nil, &events.Config{ExampleConsumeQueue: "topic"}, producer)
	if err != nil {
		t.Fatalf("NewAppLogic: %v", err)
	}

	assertHandleExampleValid(t, logic, &domain.ExampleRecord{
		RecordId:    "550e8400-e29b-41d4-a716-446655440099",
		Title:       "Fully Valid Record",
		Description: "All constraints satisfied",
		Tags:        []string{"production", "v2", "validated"},
		Meta: &domain.ExampleMeta{
			RequestedBy:      "service-account.api",
			RequiresFollowUp: true,
			Priority:         4,
		},
	})
}
