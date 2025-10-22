package metrics

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func newMeterProvider(ctx context.Context, config *Config, logger *slog.Logger) (*metric.MeterProvider, error) {
	var metricReader metric.Reader
	var exporter metric.Exporter
	var err error

	switch strings.ToLower(config.OTELMetricsExporter) {
	case "console":
		exporter, err = stdoutmetric.New(
			stdoutmetric.WithPrettyPrint(),
			stdoutmetric.WithWriter(os.Stdout),
		)
		if err != nil {
			logger.With(
				"exporter", config.OTELMetricsExporter,
			).Error("failed to create metric exporter")
			return nil, err
		}
		metricReader = metric.NewPeriodicReader(exporter)
	case "otlp":
		metricReader, err = autoexport.NewMetricReader(ctx)
	default:
		// no exporter
	}
	if err != nil {
		logger.With(
			"exporter", config.OTELMetricsExporter,
		).Error("failed to create metric exporter")
		return nil, err
	}

	r, err := newResource(config.ServiceName, config.ServiceVersion)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metricReader),
		metric.WithResource(r),
	)

	return meterProvider, nil
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

func NewOtelMetrics(ctx context.Context, cfg *Config, log *slog.Logger) error {
	mp, err := newMeterProvider(ctx, cfg, log)
	if err != nil {
		return err
	}
	// Set global meter provider
	otel.SetMeterProvider(mp)
	return nil
}
