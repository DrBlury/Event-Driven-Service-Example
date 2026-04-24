package tracing

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func NewTracerProvider(ctx context.Context, config *Config, logger *slog.Logger) (*sdktrace.TracerProvider, error) {
	var (
		exporter sdktrace.SpanExporter
		err      error
	)

	switch strings.ToLower(config.OTELTracesExporter) {
	case "console":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithWriter(os.Stdout),
		)
	case "otlp":
		exporter, err = autoexport.NewSpanExporter(ctx)
		if err != nil {
			logger.With(
				"exporter", "autoexport",
			).Error("failed to create autoexport span exporter")
			return nil, err
		}
	default:
		exporter = tracetest.NewNoopExporter()
	}
	if err != nil {
		logger.With(
			"exporter", config.OTELTracesExporter,
		).Error("failed to create trace exporter")
		return nil, err
	}

	r, err := newResource(config.ServiceName, config.ServiceVersion)
	if err != nil {
		logger.With(
			"resource", "trace",
		).Error("failed to create trace resource")
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	)

	return tracerProvider, nil
}

func newTracerProvider(ctx context.Context, config *Config, logger *slog.Logger) (*sdktrace.TracerProvider, error) {
	return NewTracerProvider(ctx, config, logger)
}

// NewOtelTracer initialises the global OpenTelemetry tracer provider using the
// supplied configuration and logger.
func NewOtelTracer(ctx context.Context, logger *slog.Logger, cfg *Config) error {
	tp, err := NewTracerProvider(ctx, cfg, logger)
	if err != nil {
		return err
	}

	otel.SetTracerProvider(tp)
	return nil
}

func newResource(serviceName string, serviceVer string) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVer),
		))
}
