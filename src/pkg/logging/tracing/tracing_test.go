package tracing

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestNewOtelTracerConsole(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELTracesExporter: "console",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	err := NewOtelTracer(context.Background(), logger, cfg)
	if err != nil {
		t.Errorf("NewOtelTracer failed for console exporter: %v", err)
	}
}

func TestNewOtelTracerDefault(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELTracesExporter: "unknown",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	err := NewOtelTracer(context.Background(), logger, cfg)
	if err != nil {
		t.Errorf("NewOtelTracer failed for unknown exporter: %v", err)
	}
}

func TestNewOtelTracerEmptyExporter(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELTracesExporter: "",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	err := NewOtelTracer(context.Background(), logger, cfg)
	if err != nil {
		t.Errorf("NewOtelTracer failed for empty exporter: %v", err)
	}
}

func TestNewResource(t *testing.T) {
	t.Parallel()

	res, err := newResource("test-service", "1.0.0")
	if err != nil {
		t.Fatalf("newResource failed: %v", err)
	}
	if res == nil {
		t.Fatal("newResource returned nil")
	}
}

func TestConfigFields(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OtelEndpoint:       "http://localhost:4317",
		Headers:            "key=value",
		OTELTracesExporter: "otlp",
		ServiceName:        "my-service",
		ServiceVersion:     "2.0.0",
		Enabled:            true,
	}

	if cfg.OtelEndpoint != "http://localhost:4317" {
		t.Errorf("OtelEndpoint = %q, want %q", cfg.OtelEndpoint, "http://localhost:4317")
	}
	if cfg.Headers != "key=value" {
		t.Errorf("Headers = %q, want %q", cfg.Headers, "key=value")
	}
	if cfg.OTELTracesExporter != "otlp" {
		t.Errorf("OTELTracesExporter = %q, want %q", cfg.OTELTracesExporter, "otlp")
	}
	if cfg.ServiceName != "my-service" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "my-service")
	}
	if cfg.ServiceVersion != "2.0.0" {
		t.Errorf("ServiceVersion = %q, want %q", cfg.ServiceVersion, "2.0.0")
	}
	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestConfigZeroValue(t *testing.T) {
	t.Parallel()

	var cfg Config

	if cfg.OtelEndpoint != "" {
		t.Errorf("zero value OtelEndpoint = %q, want empty", cfg.OtelEndpoint)
	}
	if cfg.Headers != "" {
		t.Errorf("zero value Headers = %q, want empty", cfg.Headers)
	}
	if cfg.OTELTracesExporter != "" {
		t.Errorf("zero value OTELTracesExporter = %q, want empty", cfg.OTELTracesExporter)
	}
	if cfg.Enabled {
		t.Error("zero value Enabled should be false")
	}
}

func TestNewResourceEmpty(t *testing.T) {
	t.Parallel()

	res, err := newResource("", "")
	if err != nil {
		t.Fatalf("newResource with empty strings failed: %v", err)
	}
	if res == nil {
		t.Fatal("newResource with empty strings returned nil")
	}
}

func TestNewTracerProviderConsole(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELTracesExporter: "console",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	tp, err := newTracerProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newTracerProvider console failed: %v", err)
	}
	if tp == nil {
		t.Fatal("newTracerProvider console returned nil")
	}
}

func TestNewTracerProviderDefault(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELTracesExporter: "",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	tp, err := newTracerProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newTracerProvider default failed: %v", err)
	}
	if tp == nil {
		t.Fatal("newTracerProvider default returned nil")
	}
}

func TestConfigWithHeaders(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OtelEndpoint:       "http://localhost:4317",
		Headers:            "Authorization=Bearer token",
		OTELTracesExporter: "otlp",
		ServiceName:        "my-service",
		ServiceVersion:     "1.0.0",
		Enabled:            true,
	}

	if cfg.Headers != "Authorization=Bearer token" {
		t.Errorf("Headers = %q, want %q", cfg.Headers, "Authorization=Bearer token")
	}
}

func TestNewOtelTracerMixedCase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELTracesExporter: "CONSOLE",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	err := NewOtelTracer(context.Background(), logger, cfg)
	if err != nil {
		t.Errorf("NewOtelTracer should handle mixed case: %v", err)
	}
}

func TestNewTracerProviderWithResource(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELTracesExporter: "console",
		ServiceName:        "my-service",
		ServiceVersion:     "2.0.0",
	}

	tp, err := newTracerProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newTracerProvider failed: %v", err)
	}
	if tp == nil {
		t.Fatal("newTracerProvider returned nil")
	}
}

func TestNewResourceWithVersion(t *testing.T) {
	t.Parallel()

	versions := []string{"1.0.0", "2.0.0-beta", "v3.0.0", ""}
	for _, version := range versions {
		res, err := newResource("test-service", version)
		if err != nil {
			t.Errorf("newResource with version %q failed: %v", version, err)
		}
		if res == nil {
			t.Errorf("newResource with version %q returned nil", version)
		}
	}
}

func TestNewTracerProviderUnknownExporter(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELTracesExporter: "invalidexporter",
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
	}

	// Unknown exporter should fall through to default (noop)
	tp, err := newTracerProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newTracerProvider should not error for unknown exporter: %v", err)
	}
	if tp == nil {
		t.Fatal("newTracerProvider should return noop provider for unknown exporter")
	}
}

func TestNewOtelTracerVariousExporters(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	exporters := []string{"console", "Console", "CONSOLE", "", "noop", "none", "default"}
	for _, exporter := range exporters {
		t.Run("exporter_"+exporter, func(t *testing.T) {
			cfg := &Config{
				OTELTracesExporter: exporter,
				ServiceName:        "test-service",
				ServiceVersion:     "1.0.0",
			}

			err := NewOtelTracer(context.Background(), logger, cfg)
			if err != nil {
				t.Errorf("NewOtelTracer(%q) failed: %v", exporter, err)
			}
		})
	}
}

func TestNewResourceWithServiceNames(t *testing.T) {
	t.Parallel()

	serviceNames := []string{
		"",
		"simple",
		"my-service",
		"my_service",
		"my.service.name",
		"service-with-numbers-123",
	}

	for _, name := range serviceNames {
		t.Run("name_"+name, func(t *testing.T) {
			res, err := newResource(name, "1.0.0")
			if err != nil {
				t.Errorf("newResource with name %q failed: %v", name, err)
			}
			if res == nil {
				t.Errorf("newResource with name %q returned nil", name)
			}
		})
	}
}

func TestConfigFieldsComplete(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OtelEndpoint:       "http://otel-collector:4317",
		Headers:            "x-api-key=secret123",
		OTELTracesExporter: "otlp",
		ServiceName:        "complete-service",
		ServiceVersion:     "3.0.0-rc1",
		Enabled:            true,
	}

	// Verify all fields
	if cfg.OtelEndpoint != "http://otel-collector:4317" {
		t.Errorf("OtelEndpoint mismatch")
	}
	if cfg.Headers != "x-api-key=secret123" {
		t.Errorf("Headers mismatch")
	}
	if cfg.OTELTracesExporter != "otlp" {
		t.Errorf("OTELTracesExporter mismatch")
	}
	if cfg.ServiceName != "complete-service" {
		t.Errorf("ServiceName mismatch")
	}
	if cfg.ServiceVersion != "3.0.0-rc1" {
		t.Errorf("ServiceVersion mismatch")
	}
	if !cfg.Enabled {
		t.Error("Enabled mismatch")
	}
}

func TestNewTracerProviderConsoleExporterLowercase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELTracesExporter: "console",
		ServiceName:        "test",
		ServiceVersion:     "1.0.0",
	}

	tp, err := newTracerProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if tp == nil {
		t.Fatal("returned nil")
	}
}
