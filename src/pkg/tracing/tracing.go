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
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func newTracerProvider(ctx context.Context, config *Config, logger *slog.Logger) (*trace.TracerProvider, error) {
	var (
		exporter trace.SpanExporter
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

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(r),
	)

	return tracerProvider, nil
}

// Updated NewOtelTracer to pass logger to newTracerProvider
func NewOtelTracer(ctx context.Context, logger *slog.Logger, cfg *Config) error {
	tp, err := newTracerProvider(ctx, cfg, logger)
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
