package apihandler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"drblury/event-driven-service/internal/domain"
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
