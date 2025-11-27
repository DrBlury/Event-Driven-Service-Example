package apihandler

import (
	"log/slog"
	"os"
	"testing"

	"drblury/event-driven-service/internal/domain"
)

func TestOpenAPISpecURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{"empty base URL", "", "/info/openapi.json"},
		{"root base URL", "/", "/info/openapi.json"},
		{"simple base URL", "/api", "/api/info/openapi.json"},
		{"base URL with trailing slash", "/api/v1/", "/api/v1/info/openapi.json"},
		{"whitespace only", "   ", "/info/openapi.json"},
		{"base URL with spaces", "  /api  ", "/api/info/openapi.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := openAPISpecURL(tt.baseURL)
			if got != tt.want {
				t.Errorf("openAPISpecURL(%q) = %q, want %q", tt.baseURL, got, tt.want)
			}
		})
	}
}

func TestParseDocsTemplateEmptyPath(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	result := parseDocsTemplate("", logger)
	if result != nil {
		t.Error("parseDocsTemplate with empty path should return nil")
	}

	result = parseDocsTemplate("   ", logger)
	if result != nil {
		t.Error("parseDocsTemplate with whitespace path should return nil")
	}
}

func TestParseDocsTemplateInvalidPath(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	result := parseDocsTemplate("/nonexistent/path/template.html", logger)
	if result != nil {
		t.Error("parseDocsTemplate with nonexistent path should return nil")
	}
}

func TestNewAPIHandlerNilLogger(t *testing.T) {
	info := &domain.Info{
		Version:   "1.0.0",
		BuildDate: "2024-01-01",
	}

	// This should not panic even with nil logger
	handler := NewAPIHandler(nil, info, nil, "", "")

	if handler == nil {
		t.Error("NewAPIHandler should not return nil")
	}
}

func TestAPIHandlerFields(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	info := &domain.Info{
		Version:   "1.0.0",
		BuildDate: "2024-01-01",
	}

	handler := NewAPIHandler(nil, info, logger, "/api", "")

	if handler.AppLogic != nil {
		t.Error("AppLogic should be nil when passed nil")
	}
	if handler.log != logger {
		t.Error("logger not set correctly")
	}
}
