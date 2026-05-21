package apihandler

// validation_api_test.go demonstrates the full API-to-domain validation pipeline.
//
// These tests show that protovalidate CEL rules propagate all the way from the
// domain .proto definition back to the HTTP response code, with no extra
// validation code needed in the handler or the responder.
//
// The flow is:
//   HTTP POST /examples
//     → JSON decoded into *domain.ExampleRecord
//       → usecase.HandleExample() calls protovalidate.Validate()
//         → domain.ErrValidations returned
//           → error classifier maps it to HTTP 400
//
// Contrast with a typical OpenAPI-only approach where you need:
//   1. JSON Schema validation in the middleware (catches structural issues)
//   2. Custom handler code for every cross-field or format rule
//   3. Manual duplication of those rules in every consumer (gRPC, CLI, workers)
// With protovalidate the rules are declared once in the proto and enforced
// at every layer automatically.

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/usecase"
)

func newHandlerWithLogic(t *testing.T) *APIHandler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	info := &domain.Info{Version: "1.0.0"}
	appLogic, err := usecase.NewAppLogic(nil, logger, nil, nil)
	if err != nil {
		t.Fatalf("NewAppLogic error: %v", err)
	}
	return NewAPIHandler(appLogic, info, logger, "", "")
}

func postJSON(t *testing.T, handler *APIHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.CreateExampleRecord(rec, req)
	return rec
}

// ─── Valid request ────────────────────────────────────────────────────────────

func TestAPIValidation_ValidRecord_Returns201(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440200",
		"title": "Valid API Record"
	}`)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 for valid record, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ─── UUID format (CEL regex rule) ────────────────────────────────────────────

func TestAPIValidation_InvalidUUID_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "this-is-not-a-uuid",
		"title": "Bad UUID"
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

// ─── Title length ─────────────────────────────────────────────────────────────

func TestAPIValidation_ShortTitle_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440201",
		"title": "Hi"
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short title, got %d", rec.Code)
	}
}

func TestAPIValidation_LongTitle_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440202",
		"title": "`+strings.Repeat("x", 101)+`"
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for title > 100 chars, got %d", rec.Code)
	}
}

// ─── Duplicate tags (CEL uniqueness rule) ────────────────────────────────────

func TestAPIValidation_DuplicateTags_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440203",
		"title": "Duplicate Tags",
		"tags": ["a", "b", "a"],
		"meta": {"requested_by": "owner", "priority": 1}
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for duplicate tags, got %d", rec.Code)
	}
}

// ─── Cross-field: tags require meta.requested_by (CEL message rule) ──────────

func TestAPIValidation_TagsWithoutMeta_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440204",
		"title": "Tagged No Meta",
		"tags": ["critical"]
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when tags provided without meta, got %d", rec.Code)
	}
}

// ─── Cross-field: follow-up requires priority ≥ 3 (CEL message rule) ─────────

func TestAPIValidation_FollowUpLowPriority_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440205",
		"title": "Follow-Up Low Priority",
		"meta": {
			"requested_by": "manager",
			"requires_follow_up": true,
			"priority": 2
		}
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for follow-up with priority=2, got %d", rec.Code)
	}
}

func TestAPIValidation_FollowUpSufficientPriority_Returns201(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440206",
		"title": "Follow-Up High Priority",
		"meta": {
			"requested_by": "manager",
			"requires_follow_up": true,
			"priority": 3
		}
	}`)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 for follow-up with priority=3, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ─── Priority range ───────────────────────────────────────────────────────────

func TestAPIValidation_InvalidPriority_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440207",
		"title": "Invalid Priority",
		"meta": {"requested_by": "user", "priority": 10}
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for priority=10, got %d", rec.Code)
	}
}

// ─── Invalid requested_by format (CEL regex rule) ────────────────────────────

func TestAPIValidation_RequestedByWithSpace_Returns400(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440208",
		"title": "Bad Owner",
		"meta": {"requested_by": "john doe", "priority": 1}
	}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for requested_by with space, got %d", rec.Code)
	}
}

// ─── Fully valid complex record ───────────────────────────────────────────────

func TestAPIValidation_FullyValidComplexRecord_Returns201(t *testing.T) {
	t.Parallel()
	handler := newHandlerWithLogic(t)
	rec := postJSON(t, handler, `{
		"record_id": "550e8400-e29b-41d4-a716-446655440209",
		"title": "Full Validation Showcase",
		"description": "Demonstrating protovalidate CEL rules at the API boundary",
		"tags": ["production", "validated", "v2"],
		"meta": {
			"requested_by": "service-account.api",
			"requires_follow_up": true,
			"priority": 5,
			"desired_start_date": {"year": 2025, "month": 3, "day": 15}
		}
	}`)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 for fully valid record, got %d: %s", rec.Code, rec.Body.String())
	}
}
