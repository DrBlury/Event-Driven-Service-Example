package logging

import (
	"context"
	"errors"
	"os"
	"testing"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

type stubLogExporter struct{}

func (stubLogExporter) Export(context.Context, []log.Record) error { return nil }
func (stubLogExporter) Shutdown(context.Context) error             { return nil }
func (stubLogExporter) ForceFlush(context.Context) error           { return nil }

func resetOtelFactories() {
	logExporterFactory = autoexport.NewLogExporter
	resourceFactory = newResource
	logProviderFactory = getLogProvider
}

func TestCreateOtelHandlerNilConfig(t *testing.T) {
	t.Parallel()

	if handler := createOtelHandler(context.Background(), nil); handler != nil {
		t.Fatal("expected nil handler when config absent")
	}
}

func TestCreateOtelHandlerNilContext(t *testing.T) {
	t.Parallel()

	orig := logExporterFactory
	logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
		return nil, errors.New("ignore")
	}
	defer func() { logExporterFactory = orig }()

	var nilCtx context.Context
	if handler := createOtelHandler(nilCtx, &otelSettings{}); handler != nil {
		t.Fatal("expected nil handler due to exporter failure")
	}
}

func TestBuildHandlersFallbackOnOtelFailure(t *testing.T) {
	t.Parallel()

	origExporter := logExporterFactory
	t.Cleanup(func() {
		logExporterFactory = origExporter
	})
	logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
		return nil, errors.New("export")
	}

	s := defaultSettings()
	s.console.enabled = false
	s.otel.enabled = true

	handlers := buildHandlers(context.Background(), s)
	if len(handlers) != 1 {
		t.Fatalf("expected fallback console handler, got %d", len(handlers))
	}
	if !s.console.enabled {
		t.Fatal("expected console fallback to re-enable console output")
	}
}

func TestBuildHandlersIncludeOtelHandler(t *testing.T) {
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
	resourceFactory = func(serviceName, serviceVersion string) (*resource.Resource, error) {
		return resource.Empty(), nil
	}
	logProviderFactory = func(exp log.Exporter, res *resource.Resource) (*log.LoggerProvider, error) {
		return getLogProvider(exp, res)
	}

	s := defaultSettings()
	s.console.enabled = true
	s.otel.enabled = true
	s.otel.mirrorConsole = true

	handlers := buildHandlers(context.Background(), s)
	if len(handlers) != 2 {
		t.Fatalf("expected console and otel handlers, got %d", len(handlers))
	}
}

func TestCreateOtelHandlerErrorPaths(t *testing.T) {
	t.Parallel()

	defer resetOtelFactories()

	logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
		return nil, errors.New("export")
	}
	if handler := createOtelHandler(context.Background(), &otelSettings{}); handler != nil {
		t.Fatal("expected nil handler on exporter failure")
	}

	logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
		return stubLogExporter{}, nil
	}
	resourceFactory = func(string, string) (*resource.Resource, error) {
		return nil, errors.New("resource")
	}
	if handler := createOtelHandler(context.Background(), &otelSettings{}); handler != nil {
		t.Fatal("expected nil handler on resource failure")
	}

	resourceFactory = func(string, string) (*resource.Resource, error) {
		return resource.Empty(), nil
	}
	logProviderFactory = func(log.Exporter, *resource.Resource) (*log.LoggerProvider, error) {
		return nil, errors.New("provider")
	}
	if handler := createOtelHandler(context.Background(), &otelSettings{}); handler != nil {
		t.Fatal("expected nil handler on provider failure")
	}
}

func TestCreateOtelHandlerSuccess(t *testing.T) {
	t.Parallel()

	defer resetOtelFactories()
	logExporterFactory = func(ctx context.Context, _ ...autoexport.LogOption) (log.Exporter, error) {
		return stubLogExporter{}, nil
	}
	resourceFactory = func(string, string) (*resource.Resource, error) {
		return resource.Empty(), nil
	}
	logProviderFactory = func(exp log.Exporter, res *resource.Resource) (*log.LoggerProvider, error) {
		return getLogProvider(exp, res)
	}

	cfg := &otelSettings{serviceName: "svc", serviceVersion: "1", endpoint: "http://localhost:4318", headers: "a=b"}
	os.Unsetenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_LOGS_HEADERS")

	handler := createOtelHandler(context.Background(), cfg)
	if handler == nil {
		t.Fatal("expected handler to be created")
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") != "http://localhost:4318" {
		t.Fatal("expected endpoint environment variable to be set")
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_LOGS_HEADERS") != "a=b" {
		t.Fatal("expected headers environment variable to be set")
	}
}
