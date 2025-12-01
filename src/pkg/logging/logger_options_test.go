package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestParseFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected Format
	}{
		// Basic cases
		{"json", FormatJSON},
		{"Pretty", FormatPretty},
		{"text", FormatText},
		{"plain", FormatText},
		{"prettyjson", FormatPretty},
		{"", FormatText},
		{"unknown", ""},
		// Case insensitivity
		{"JSON", FormatJSON},
		{"Json", FormatJSON},
		{"PRETTY", FormatPretty},
		{"PrettyJson", FormatPretty},
		{"TEXT", FormatText},
		{"Text", FormatText},
		{"CONSOLE", FormatText},
		{"PLAIN", FormatText},
		// Whitespace handling
		{"  json  ", FormatJSON},
		{"\tpretty\t", FormatPretty},
		{" text ", FormatText},
		{"\t\n", FormatText},
	}

	for _, tc := range testCases {
		got := ParseFormat(tc.input)
		if got != tc.expected {
			t.Errorf("ParseFormat(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	t.Run("valid levels", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected slog.Level
		}{
			// Basic cases
			{"debug", slog.LevelDebug},
			{"INFO", slog.LevelInfo},
			{"warn", slog.LevelWarn},
			{"warning", slog.LevelWarn},
			{"error", slog.LevelError},
			// Case insensitivity
			{"DEBUG", slog.LevelDebug},
			{"Debug", slog.LevelDebug},
			{"Info", slog.LevelInfo},
			{"WARN", slog.LevelWarn},
			{"Warn", slog.LevelWarn},
			{"WARNING", slog.LevelWarn},
			{"Warning", slog.LevelWarn},
			{"ERROR", slog.LevelError},
			{"Error", slog.LevelError},
			// Whitespace handling
			{"  debug  ", slog.LevelDebug},
			{"\tinfo\t", slog.LevelInfo},
			{" warn ", slog.LevelWarn},
			{" error ", slog.LevelError},
			{"\n\r", slog.LevelInfo},
		}

		for _, tc := range testCases {
			got, err := parseLevel(tc.input)
			if err != nil {
				t.Errorf("parseLevel(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		_, err := parseLevel("invalid")
		if err == nil {
			t.Error("expected error for invalid level")
		}
	})
}

func TestDefaultSettings(t *testing.T) {
	t.Parallel()

	s := defaultSettings()
	if s.level != slog.LevelInfo {
		t.Errorf("default level = %v, want %v", s.level, slog.LevelInfo)
	}
	if !s.addSource {
		t.Error("default addSource should be true")
	}
	if !s.console.enabled {
		t.Error("default console.enabled should be true")
	}
	if s.console.format != FormatText {
		t.Errorf("default console.format = %q, want %q", s.console.format, FormatText)
	}
	if !s.setAsDefault {
		t.Error("default setAsDefault should be true")
	}
}

func TestOptionsMutateSettings(t *testing.T) {
	t.Parallel()

	s := defaultSettings()
	WithLevel(slog.LevelDebug)(s)
	WithJSONFormat()(s)
	WithAddSource(false)(s)
	WithReplaceAttr(func([]string, slog.Attr) slog.Attr { return slog.Attr{} })(s)
	WithAttrs(slog.String("key", "value"))(s)
	WithoutGlobal()(s)
	WithConsoleWriter(nil)(s)
	WithTextFormat()(s)
	WithPrettyFormat()(s)
	WithLevelString("warn")(s)
	WithoutConsole()(s)

	cfg := &Config{
		Level:          "error",
		Format:         FormatJSON,
		AddSource:      true,
		ConsoleEnabled: true,
		SetAsDefault:   true,
		OTel: OTelConfig{
			Enabled:         true,
			MirrorToConsole: true,
			ServiceName:     "svc",
			ServiceVersion:  "1",
			Endpoint:        "endpoint",
			Headers:         "hdr",
		},
	}
	WithConfig(cfg)(s)
	WithOTel("service", "v", WithOTelEndpoint("ep"), WithOTelHeaders("hdr"), WithOTelConsoleMirror())(s)

	if !s.console.enabled {
		t.Fatal("expected console to be enabled")
	}
	if s.otel.serviceName != "service" || s.otel.serviceVersion != "v" {
		t.Fatalf("unexpected otel config: %#v", s.otel)
	}
	if len(s.attrs) != 1 {
		t.Fatalf("expected attributes to be recorded")
	}
	if s.console.format != FormatJSON {
		t.Fatalf("expected json format, got %v", s.console.format)
	}
}

func TestWithConfig(t *testing.T) {
	t.Parallel()

	t.Run("nil config does not panic", func(t *testing.T) {
		s := defaultSettings()
		opt := WithConfig(nil)
		opt(s)
		if s.level != slog.LevelInfo {
			t.Error("Level should be unchanged with nil config")
		}
	})

	t.Run("partial config", func(t *testing.T) {
		s := defaultSettings()
		cfg := &Config{Level: "debug"}
		WithConfig(cfg)(s)
		if s.level != slog.LevelDebug {
			t.Errorf("Level = %v, want %v", s.level, slog.LevelDebug)
		}
	})

	t.Run("otel disabled", func(t *testing.T) {
		s := defaultSettings()
		cfg := &Config{Level: "debug", OTel: OTelConfig{Enabled: false}}
		WithConfig(cfg)(s)
		if s.otel.enabled {
			t.Error("OTel should be disabled")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		cfg := &Config{}
		logger := SetLogger(context.Background(), WithoutGlobal(), WithConfig(cfg))
		if logger == nil {
			t.Fatal("expected logger instance")
		}
	})
}

func TestWithOTel(t *testing.T) {
	t.Parallel()

	t.Run("empty service name", func(t *testing.T) {
		s := defaultSettings()
		WithOTel("", "", WithOTelEndpoint("http://localhost:4317"))(s)
		if !s.otel.enabled {
			t.Error("OTel should be enabled")
		}
		if s.otel.endpoint != "http://localhost:4317" {
			t.Errorf("Endpoint = %q, want 'http://localhost:4317'", s.otel.endpoint)
		}
	})
}

func TestWithLevelStringInvalid(t *testing.T) {
	t.Parallel()

	s := defaultSettings()
	originalLevel := s.level
	WithLevelString("invalid_level")(s)
	if s.level != originalLevel {
		t.Errorf("Level should remain unchanged for invalid input, got %v", s.level)
	}
}

func TestSetLogger(t *testing.T) {
	t.Parallel()

	t.Run("console fallbacks", func(t *testing.T) {
		logger := SetLogger(context.Background(), WithoutConsole(), WithoutGlobal())
		if logger == nil {
			t.Fatal("expected logger instance")
		}
	})

	t.Run("with options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := SetLogger(context.Background(), WithoutGlobal(), WithLevel(slog.LevelDebug), WithJSONFormat(), WithConsoleWriter(buf), WithAttrs(slog.String("attr", "value")))

		logger.Debug("hello")
		out := buf.String()
		if !strings.Contains(out, "hello") || !strings.Contains(out, "attr") {
			t.Fatalf("expected output to include message and attribute, got %q", out)
		}
	})

	t.Run("with nil console writer", func(t *testing.T) {
		logger := SetLogger(context.Background(), WithoutGlobal(), WithConsoleWriter(nil))
		if logger == nil {
			t.Fatal("expected logger instance")
		}
	})

	t.Run("TODO context", func(t *testing.T) {
		logger := SetLogger(context.TODO(), WithoutGlobal())
		if logger == nil {
			t.Fatal("expected logger instance with TODO context")
		}
	})

	t.Run("nil context", func(t *testing.T) {
		//nolint:staticcheck // Testing nil context handling deliberately
		var nilCtx context.Context = nil
		logger := SetLogger(nilCtx, WithoutGlobal())
		if logger == nil {
			t.Fatal("expected logger instance with nil context")
		}
	})

	t.Run("empty format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := SetLogger(context.Background(), WithoutGlobal(), WithConsoleWriter(buf), func(s *settings) {
			s.console.format = ""
		})
		if logger == nil {
			t.Fatal("expected logger instance")
		}
	})

	t.Run("nil options", func(t *testing.T) {
		logger := SetLogger(context.Background(), nil, WithoutGlobal(), nil)
		if logger == nil {
			t.Fatal("expected logger instance")
		}
	})

	t.Run("multiple attrs", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := SetLogger(context.Background(),
			WithoutGlobal(),
			WithJSONFormat(),
			WithConsoleWriter(buf),
			WithAttrs(
				slog.String("key1", "value1"),
				slog.Int("key2", 42),
				slog.Bool("key3", true),
			),
		)

		logger.Info("multi attr test")
		out := buf.String()
		if !strings.Contains(out, "key1") {
			t.Error("expected key1 in output")
		}
	})
}

func TestSetLoggerWithOtel(t *testing.T) {
	t.Parallel()

	origExporter := logExporterFactory
	origResource := resourceFactory
	origProvider := logProviderFactory
	t.Cleanup(func() {
		logExporterFactory = origExporter
		resourceFactory = origResource
		logProviderFactory = origProvider
	})

	logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
		return stubLogExporter{}, nil
	}
	resourceFactory = func(service, version string) (*resource.Resource, error) {
		return resource.Empty(), nil
	}
	logProviderFactory = func(exp log.Exporter, res *resource.Resource) (*log.LoggerProvider, error) {
		return getLogProvider(exp, res)
	}

	t.Run("with console mirror", func(t *testing.T) {
		logger := SetLogger(context.Background(), WithoutGlobal(), WithOTel("service", "1", WithOTelConsoleMirror()))
		logger.Info("otel message")
	})

	t.Run("multiple handlers", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := SetLogger(context.Background(), WithoutGlobal(), WithConsoleWriter(buf), WithOTel("service", "1"))
		logger.Info("test message")
		if !strings.Contains(buf.String(), "test message") {
			t.Error("expected message in console output")
		}
	})
}

func TestNewConsoleHandlerFormats(t *testing.T) {
	t.Parallel()

	cfg := defaultSettings()

	t.Run("JSON format", func(t *testing.T) {
		cfg.console.format = FormatJSON
		if _, ok := newConsoleHandler(cfg).(*slog.JSONHandler); !ok {
			t.Fatal("expected JSON handler for JSON format")
		}
	})

	t.Run("Pretty format", func(t *testing.T) {
		cfg.console.format = FormatPretty
		if _, ok := newConsoleHandler(cfg).(*PrettyJsonHandler); !ok {
			t.Fatal("expected pretty handler for pretty format")
		}
	})

	t.Run("Text format", func(t *testing.T) {
		cfg.console.format = FormatText
		if _, ok := newConsoleHandler(cfg).(*slog.TextHandler); !ok {
			t.Fatal("expected text handler for text format")
		}
	})
}

func TestBuildHandlers(t *testing.T) {
	t.Parallel()

	t.Run("no handlers fallback", func(t *testing.T) {
		cfg := defaultSettings()
		cfg.console.enabled = false
		cfg.otel.enabled = false

		handlers := buildHandlers(context.Background(), cfg)
		if len(handlers) == 0 {
			t.Fatal("expected at least one fallback handler")
		}
	})

	t.Run("otel without console", func(t *testing.T) {
		origExporter := logExporterFactory
		origResource := resourceFactory
		origProvider := logProviderFactory
		t.Cleanup(func() {
			logExporterFactory = origExporter
			resourceFactory = origResource
			logProviderFactory = origProvider
		})

		logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
			return stubLogExporter{}, nil
		}
		resourceFactory = func(service, version string) (*resource.Resource, error) {
			return resource.Empty(), nil
		}
		logProviderFactory = func(exp log.Exporter, res *resource.Resource) (*log.LoggerProvider, error) {
			return getLogProvider(exp, res)
		}

		cfg := defaultSettings()
		cfg.console.enabled = false
		cfg.otel.enabled = true
		cfg.otel.serviceName = "test"
		cfg.otel.serviceVersion = "1"

		handlers := buildHandlers(context.Background(), cfg)
		if len(handlers) == 0 {
			t.Fatal("expected at least one handler")
		}
	})
}
