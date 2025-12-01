package apihandler

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

func TestCreateExampleRecordNilHandler(t *testing.T) {
	var handler *APIHandler

	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestCreateExampleRecordNilAppLogic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	body := `{"record_id": "test-123", "title": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestCreateExampleRecordInvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	if rec.Code == http.StatusOK || rec.Code == http.StatusCreated {
		t.Errorf("Invalid JSON should not result in success, got status %d", rec.Code)
	}
}

func TestCreateExampleRecordEmptyBody(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	// Should fail with empty body (either bad request or internal server error due to nil AppLogic)
	if rec.Code == http.StatusCreated {
		t.Errorf("Empty body should not result in created, got status %d", rec.Code)
	}
}

func TestCreateExampleRecordWithAuthHeader(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	body := `{"record_id": "test-123", "title": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	// Should fail because AppLogic is nil
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestCreateExampleRecordWithFullBody(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "/api", "")

	body := `{
		"record_id": "test-456",
		"title": "Full Test",
		"description": "Test description",
		"tags": ["test", "unit"],
		"meta": {
			"requested_by": "test-user",
			"requires_follow_up": true,
			"priority": 5,
			"desired_start_date": {
				"year": 2024,
				"month": 6,
				"day": 15
			}
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	// Should fail because AppLogic is nil
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestCreateExampleRecordMinimalPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	body := `{"record_id": "min-001"}`
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	// Should fail because AppLogic is nil
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestCreateExampleRecordWithContentType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	contentTypes := []string{
		"application/json",
		"application/json; charset=utf-8",
		"application/json;charset=utf-8",
	}

	for _, ct := range contentTypes {
		t.Run(ct, func(t *testing.T) {
			body := `{"record_id": "ct-test"}`
			req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
			req.Header.Set("Content-Type", ct)
			rec := httptest.NewRecorder()

			handler.CreateExampleRecord(rec, req)

			// Should fail because AppLogic is nil, not because of content type
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("Content-Type %q: Status code = %d, want %d", ct, rec.Code, http.StatusInternalServerError)
			}
		})
	}
}

func TestCreateExampleRecordWithMalformedJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	malformedBodies := []string{
		`{`,
		`}`,
		`{"key": }`,
		`{"record_id": "unclosed`,
		`not json at all`,
		`<xml>not json</xml>`,
	}

	for _, body := range malformedBodies {
		t.Run(body, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.CreateExampleRecord(rec, req)

			// Should fail with bad request or error status
			if rec.Code == http.StatusCreated || rec.Code == http.StatusOK {
				t.Errorf("Malformed JSON should not succeed: %s", body)
			}
		})
	}
}

func TestCreateExampleRecordNilBody(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{Version: "1.0.0"}

	handler := NewAPIHandler(nil, info, logger, "", "")

	req := httptest.NewRequest(http.MethodPost, "/examples", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	// Should fail with some error status
	if rec.Code == http.StatusCreated {
		t.Error("Nil body should not result in created")
	}
}

func TestCreateExampleRecordWithAppLogic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	info := &domain.Info{Version: "1.0.0"}

	// Create AppLogic without database (will return early without error)
	appLogic, _ := usecase.NewAppLogic(nil, logger)
	handler := NewAPIHandler(appLogic, info, logger, "", "")

	body := `{"record_id": "test-with-logic", "title": "Test with AppLogic"}`
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	// Should succeed because AppLogic.HandleExample returns nil when db is nil
	if rec.Code != http.StatusCreated {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestCreateExampleRecordWithAppLogicFullPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	info := &domain.Info{Version: "1.0.0"}

	appLogic, _ := usecase.NewAppLogic(nil, logger)
	handler := NewAPIHandler(appLogic, info, logger, "/api", "")

	body := `{
		"record_id": "full-payload-test",
		"title": "Full Payload Test",
		"description": "Testing with full payload",
		"tags": ["test", "full", "payload"],
		"meta": {
			"requested_by": "test-user",
			"requires_follow_up": true,
			"priority": 5
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestCreateExampleRecordSuccessResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	info := &domain.Info{Version: "1.0.0"}

	appLogic, _ := usecase.NewAppLogic(nil, logger)
	handler := NewAPIHandler(appLogic, info, logger, "", "")

	body := `{"record_id": "response-test", "title": "Response Test"}`
	req := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateExampleRecord(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Status code = %d, want %d", rec.Code, http.StatusCreated)
	}

	// Check response body contains expected fields
	responseBody := rec.Body.String()
	if !strings.Contains(responseBody, "queued") {
		t.Error("Response should contain 'queued' status")
	}
	if !strings.Contains(responseBody, "example event accepted") {
		t.Error("Response should contain success message")
	}
}
