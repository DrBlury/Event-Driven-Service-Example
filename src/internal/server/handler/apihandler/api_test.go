package apihandler

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/usecase"
)

func TestOpenAPISpecURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		// Basic cases
		{"empty base URL", "", "/info/openapi.json"},
		{"root base URL", "/", "/info/openapi.json"},
		{"simple base URL", "/api", "/api/info/openapi.json"},
		{"base URL with trailing slash", "/api/v1/", "/api/v1/info/openapi.json"},
		// Whitespace handling
		{"whitespace only", "   ", "/info/openapi.json"},
		{"base URL with spaces", "  /api  ", "/api/info/openapi.json"},
		{"just tabs", "\t\t", "/info/openapi.json"},
		{"newlines", "\n\n", "/info/openapi.json"},
		{"mixed whitespace", "  \t\n  ", "/info/openapi.json"},
		// Complex paths
		{"versioned", "/api/v1", "/api/v1/info/openapi.json"},
		{"trailing slash", "/api/", "/api/info/openapi.json"},
		{"deep path", "/api/v1/service/module", "/api/v1/service/module/info/openapi.json"},
		{"very deep path", "/very/deep/nested/path/", "/very/deep/nested/path/info/openapi.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := openAPISpecURL(tt.baseURL)
			if got != tt.want {
				t.Errorf("openAPISpecURL(%q) = %q, want %q", tt.baseURL, got, tt.want)
			}
			// Verify the result always ends with /info/openapi.json
			if !strings.HasSuffix(got, "/info/openapi.json") {
				t.Errorf("openAPISpecURL(%q) = %q, should end with /info/openapi.json", tt.baseURL, got)
			}
		})
	}
}

func TestParseDocsTemplate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	t.Run("empty and whitespace paths return nil", func(t *testing.T) {
		paths := []string{"", "   ", "\t\t", "\n\n"}
		for _, path := range paths {
			result := parseDocsTemplate(path, logger)
			if result != nil {
				t.Errorf("parseDocsTemplate(%q) should return nil", path)
			}
		}
	})

	t.Run("nonexistent path returns nil", func(t *testing.T) {
		result := parseDocsTemplate("/nonexistent/path/template.html", logger)
		if result != nil {
			t.Error("parseDocsTemplate with nonexistent path should return nil")
		}
	})

	t.Run("valid template file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "template*.html")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`<!DOCTYPE html><html><body>{{.BaseURL}}</body></html>`)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		result := parseDocsTemplate(tmpFile.Name(), logger)
		if result == nil {
			t.Error("parseDocsTemplate should return a template for a valid file")
		}
	})

	t.Run("complex template file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "complex_template*.html")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		complexTemplate := `<!DOCTYPE html>
<html>
<head><title>API Docs</title></head>
<body>
<h1>Base URL: {{.BaseURL}}</h1>
<h2>Spec URL: {{.SpecURL}}</h2>
</body>
</html>`
		_, err = tmpFile.WriteString(complexTemplate)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		result := parseDocsTemplate(tmpFile.Name(), logger)
		if result == nil {
			t.Error("parseDocsTemplate should return a template for a complex valid file")
		}
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "invalid_template*.html")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`{{.Invalid`)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		result := parseDocsTemplate(tmpFile.Name(), logger)
		if result != nil {
			t.Error("parseDocsTemplate should return nil for invalid template syntax")
		}
	})
}

func TestNewAPIHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	t.Run("nil logger does not panic", func(t *testing.T) {
		info := &domain.Info{Version: "1.0.0", BuildDate: "2024-01-01"}
		handler := NewAPIHandler(nil, info, nil, "", "")
		if handler == nil {
			t.Error("NewAPIHandler should not return nil")
		}
	})

	t.Run("various info configurations", func(t *testing.T) {
		testCases := []struct {
			name string
			info *domain.Info
		}{
			{"minimal info", &domain.Info{}},
			{"version only", &domain.Info{Version: "1.0.0"}},
			{"full info", &domain.Info{
				Version:    "1.0.0",
				BuildDate:  "2024-01-01",
				Details:    "Test",
				CommitHash: "abc123",
				CommitDate: "2024-01-01T00:00:00Z",
			}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				handler := NewAPIHandler(nil, tc.info, logger, "/api", "")
				if handler == nil {
					t.Fatal("handler should not be nil")
				}
				if handler.InfoHandler == nil {
					t.Error("InfoHandler should not be nil")
				}
			})
		}
	})

	t.Run("various baseURL configurations", func(t *testing.T) {
		info := &domain.Info{Version: "1.0.0"}
		baseURLs := []string{
			"",
			"/",
			"/api",
			"/api/",
			"/api/v1",
			"/api/v1/",
			"  /api  ",
			"/api/v2/nested/path",
		}

		for _, baseURL := range baseURLs {
			t.Run("baseURL_"+baseURL, func(t *testing.T) {
				handler := NewAPIHandler(nil, info, logger, baseURL, "")
				if handler == nil {
					t.Fatalf("NewAPIHandler(%q) should not return nil", baseURL)
				}
			})
		}
	})

	t.Run("with AppLogic", func(t *testing.T) {
		info := &domain.Info{Version: "1.0.0"}
		appLogic, _ := usecase.NewAppLogic(nil, logger)
		handler := NewAPIHandler(appLogic, info, logger, "/api", "")

		if handler == nil {
			t.Fatal("handler should not be nil")
		}
		if handler.AppLogic != appLogic {
			t.Error("AppLogic should be set correctly")
		}
	})

	t.Run("with valid docs template", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "api_docs*.html")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`<!DOCTYPE html><html><body>{{.BaseURL}} - {{.SpecURL}}</body></html>`)
		if err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		info := &domain.Info{Version: "1.0.0"}
		handler := NewAPIHandler(nil, info, logger, "/api", tmpFile.Name())
		if handler == nil {
			t.Fatal("handler should not be nil with valid template")
		}
	})

	t.Run("with invalid docs template path", func(t *testing.T) {
		info := &domain.Info{Version: "1.0.0"}
		handler := NewAPIHandler(nil, info, logger, "", "/nonexistent/template.html")
		if handler == nil {
			t.Fatal("handler should not be nil even with invalid template path")
		}
	})

	t.Run("fields are set correctly", func(t *testing.T) {
		info := &domain.Info{Version: "1.0.0", BuildDate: "2024-01-01"}
		handler := NewAPIHandler(nil, info, logger, "/api", "")

		if handler.AppLogic != nil {
			t.Error("AppLogic should be nil when passed nil")
		}
		if handler.log != logger {
			t.Error("logger not set correctly")
		}
		if handler.InfoHandler == nil {
			t.Error("InfoHandler should not be nil")
		}
	})
}
