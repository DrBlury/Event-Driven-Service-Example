package metrics

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestNewOtelMetricsConsole(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELMetricsExporter: "console",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	err := NewOtelMetrics(context.Background(), cfg, logger)
	if err != nil {
		t.Errorf("NewOtelMetrics failed for console exporter: %v", err)
	}
}

func TestNewOtelMetricsUnsupported(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELMetricsExporter: "unknown",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	err := NewOtelMetrics(context.Background(), cfg, logger)
	if err == nil {
		t.Error("NewOtelMetrics should fail for unknown exporter")
	}
}

func TestNewOtelMetricsEmptyExporter(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELMetricsExporter: "",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	err := NewOtelMetrics(context.Background(), cfg, logger)
	if err == nil {
		t.Error("NewOtelMetrics should fail for empty exporter")
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
		OtelEndpoint:        "http://localhost:4317",
		Headers:             "key=value",
		OTELMetricsExporter: "otlp",
		ServiceName:         "my-service",
		ServiceVersion:      "2.0.0",
		Enabled:             true,
	}

	if cfg.OtelEndpoint != "http://localhost:4317" {
		t.Errorf("OtelEndpoint = %q, want %q", cfg.OtelEndpoint, "http://localhost:4317")
	}
	if cfg.Headers != "key=value" {
		t.Errorf("Headers = %q, want %q", cfg.Headers, "key=value")
	}
	if cfg.OTELMetricsExporter != "otlp" {
		t.Errorf("OTELMetricsExporter = %q, want %q", cfg.OTELMetricsExporter, "otlp")
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
	if cfg.OTELMetricsExporter != "" {
		t.Errorf("zero value OTELMetricsExporter = %q, want empty", cfg.OTELMetricsExporter)
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

func TestNewMeterProviderConsole(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &Config{
		OTELMetricsExporter: "console",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	mp, err := newMeterProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newMeterProvider console failed: %v", err)
	}
	if mp == nil {
		t.Fatal("newMeterProvider console returned nil")
	}
}

func TestConfigWithHeaders(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OtelEndpoint:        "http://localhost:4317",
		Headers:             "Authorization=Bearer token",
		OTELMetricsExporter: "otlp",
		ServiceName:         "my-service",
		ServiceVersion:      "1.0.0",
		Enabled:             true,
	}

	if cfg.Headers != "Authorization=Bearer token" {
		t.Errorf("Headers = %q, want %q", cfg.Headers, "Authorization=Bearer token")
	}
}

func TestNewOtelMetricsMixedCase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELMetricsExporter: "CONSOLE",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	err := NewOtelMetrics(context.Background(), cfg, logger)
	if err != nil {
		t.Errorf("NewOtelMetrics should handle mixed case: %v", err)
	}
}

func TestNewMeterProviderUnsupported(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELMetricsExporter: "invalid",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	mp, err := newMeterProvider(context.Background(), cfg, logger)
	if err == nil {
		t.Error("newMeterProvider should fail for invalid exporter")
	}
	if mp != nil {
		t.Error("newMeterProvider should return nil for invalid exporter")
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

func TestNewMeterProviderConsoleUppercase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELMetricsExporter: "CONSOLE",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	mp, err := newMeterProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newMeterProvider CONSOLE failed: %v", err)
	}
	if mp == nil {
		t.Fatal("newMeterProvider CONSOLE returned nil")
	}
}

func TestNewMeterProviderConsoleMixedCase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELMetricsExporter: "Console",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	mp, err := newMeterProvider(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("newMeterProvider Console failed: %v", err)
	}
	if mp == nil {
		t.Fatal("newMeterProvider Console returned nil")
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
		OtelEndpoint:        "http://otel-collector:4317",
		Headers:             "x-api-key=secret123",
		OTELMetricsExporter: "otlp",
		ServiceName:         "complete-service",
		ServiceVersion:      "3.0.0-rc1",
		Enabled:             true,
	}

	// Verify all fields
	if cfg.OtelEndpoint != "http://otel-collector:4317" {
		t.Errorf("OtelEndpoint mismatch")
	}
	if cfg.Headers != "x-api-key=secret123" {
		t.Errorf("Headers mismatch")
	}
	if cfg.OTELMetricsExporter != "otlp" {
		t.Errorf("OTELMetricsExporter mismatch")
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

func TestNewOtelMetricsVariousExporters(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Exporters that should succeed (console in various cases)
	successExporters := []string{"console", "Console", "CONSOLE"}
	for _, exporter := range successExporters {
		t.Run("success_"+exporter, func(t *testing.T) {
			cfg := &Config{
				OTELMetricsExporter: exporter,
				ServiceName:         "test-service",
				ServiceVersion:      "1.0.0",
			}

			err := NewOtelMetrics(context.Background(), cfg, logger)
			if err != nil {
				t.Errorf("NewOtelMetrics(%q) failed: %v", exporter, err)
			}
		})
	}

	// Exporters that should fail (unsupported)
	failExporters := []string{"", "noop", "none", "default", "invalid"}
	for _, exporter := range failExporters {
		t.Run("fail_"+exporter, func(t *testing.T) {
			cfg := &Config{
				OTELMetricsExporter: exporter,
				ServiceName:         "test-service",
				ServiceVersion:      "1.0.0",
			}

			err := NewOtelMetrics(context.Background(), cfg, logger)
			if err == nil {
				t.Errorf("NewOtelMetrics(%q) should fail for unsupported exporter", exporter)
			}
		})
	}
}

func TestNewMeterProviderDefaultCase(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &Config{
		OTELMetricsExporter: "unsupported_exporter",
		ServiceName:         "test-service",
		ServiceVersion:      "1.0.0",
	}

	mp, err := newMeterProvider(context.Background(), cfg, logger)
	if err == nil {
		t.Error("newMeterProvider should error for unsupported exporter")
	}
	if mp != nil {
		t.Error("newMeterProvider should return nil for unsupported exporter")
	}
}
