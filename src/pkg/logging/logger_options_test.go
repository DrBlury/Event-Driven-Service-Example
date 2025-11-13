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

	cases := map[string]Format{
		"json":       FormatJSON,
		"Pretty":     FormatPretty,
		"text":       FormatText,
		"plain":      FormatText,
		"prettyjson": FormatPretty,
		"":           FormatText,
		"unknown":    "",
	}

	for input, expected := range cases {
		if got := ParseFormat(input); got != expected {
			t.Fatalf("format %q: expected %q, got %q", input, expected, got)
		}
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected slog.Level
		wantErr  bool
	}{
		{"debug", slog.LevelDebug, false},
		{"INFO", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"warning", slog.LevelWarn, false},
		{"error", slog.LevelError, false},
		{"invalid", slog.LevelInfo, true},
	}

	for _, tc := range cases {
		lvl, err := parseLevel(tc.input)
		if tc.wantErr && err == nil {
			t.Fatalf("expected error for %q", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.input, err)
		}
		if lvl != tc.expected {
			t.Fatalf("level %q: expected %v, got %v", tc.input, tc.expected, lvl)
		}
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

func TestSetLoggerConsoleFallbacks(t *testing.T) {
	t.Parallel()

	logger := SetLogger(context.Background(), WithoutConsole(), WithoutGlobal())
	if logger == nil {
		t.Fatal("expected logger instance")
	}
}

func TestSetLoggerWithOptions(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := SetLogger(context.Background(), WithoutGlobal(), WithLevel(slog.LevelDebug), WithJSONFormat(), WithConsoleWriter(buf), WithAttrs(slog.String("attr", "value")))

	logger.Debug("hello")
	out := buf.String()
	if !strings.Contains(out, "hello") || !strings.Contains(out, "attr") {
		t.Fatalf("expected output to include message and attribute, got %q", out)
	}
}

func TestNewConsoleHandlerFormats(t *testing.T) {
	t.Parallel()

	cfg := defaultSettings()
	cfg.console.format = FormatJSON
	if _, ok := newConsoleHandler(cfg).(*slog.JSONHandler); !ok {
		t.Fatal("expected JSON handler for JSON format")
	}

	cfg.console.format = FormatPretty
	if _, ok := newConsoleHandler(cfg).(*PrettyJsonHandler); !ok {
		t.Fatal("expected pretty handler for pretty format")
	}

	cfg.console.format = FormatText
	if _, ok := newConsoleHandler(cfg).(*slog.TextHandler); !ok {
		t.Fatal("expected text handler for text format")
	}
}

func TestSetLoggerWithNilConsoleWriter(t *testing.T) {
	t.Parallel()

	logger := SetLogger(context.Background(), WithoutGlobal(), WithConsoleWriter(nil))
	if logger == nil {
		t.Fatal("expected logger instance")
	}
}

func TestSetLoggerWithOtelSuccess(t *testing.T) {
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

	logger := SetLogger(context.Background(), WithoutGlobal(), WithOTel("service", "1", WithOTelConsoleMirror()))
	logger.Info("otel message")
}
