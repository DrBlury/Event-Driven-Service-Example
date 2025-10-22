package logging

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
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

func createOtelHandler(ctx context.Context, loggerConfig *Config) *otelslog.Handler {
	exporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		slog.With(
			"exporter", "autoexport",
		).Error("failed to create log exporter")
		return nil
	}

	res, err := newResource(loggerConfig.ServiceName, loggerConfig.ServiceVersion)
	if err != nil {
		slog.With(
			"resource", "log",
		).Error("failed to create log resource")
		return nil
	}

	provider, err := getLogProvider(exporter, res)
	if err != nil {
		slog.With(
			"provider", "log",
		).Error("failed to create log provider")
		return nil
	}

	otelHandler := otelslog.NewHandler(loggerConfig.ServiceName, otelslog.WithLoggerProvider(provider))
	return otelHandler
}
