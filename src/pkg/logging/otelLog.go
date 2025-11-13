package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var (
	logExporterFactory = autoexport.NewLogExporter
	resourceFactory    = newResource
	logProviderFactory = getLogProvider
)

func newResource(serviceName string, serviceVer string) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVer),
		))
}

func getLogProvider(logExporter log.Exporter, res *resource.Resource) (*log.LoggerProvider, error) {
	logProvider := log.NewLoggerProvider(
		log.WithProcessor(
			log.NewBatchProcessor(logExporter,
				log.WithMaxQueueSize(10),
				log.WithExportInterval(time.Second),
			),
		),
		log.WithResource(res),
	)
	return logProvider, nil
}

func createOtelHandler(ctx context.Context, cfg *otelSettings) slog.Handler {
	if cfg == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if cfg.endpoint != "" {
		_ = os.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", cfg.endpoint)
	}
	if cfg.headers != "" {
		_ = os.Setenv("OTEL_EXPORTER_OTLP_LOGS_HEADERS", cfg.headers)
	}

	exporter, err := logExporterFactory(ctx)
	if err != nil {
		slog.With(
			"exporter", "autoexport",
		).Error("failed to create log exporter")
		return nil
	}

	res, err := resourceFactory(cfg.serviceName, cfg.serviceVersion)
	if err != nil {
		slog.With(
			"resource", "log",
		).Error("failed to create log resource")
		return nil
	}

	provider, err := logProviderFactory(exporter, res)
	if err != nil {
		slog.With(
			"provider", "log",
		).Error("failed to create log provider")
		return nil
	}

	serviceName := cfg.serviceName
	if serviceName == "" {
		serviceName = "service"
	}
	otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(provider))
	return otelHandler
}
