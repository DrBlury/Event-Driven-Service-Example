package logging

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type OtelLog struct {
	ctx         context.Context
	config      *Config
	logProvider *log.LoggerProvider
}

func newResource(serviceName string, serviceVer string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
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

func NewOtelLog(ctx context.Context, config *Config) (*OtelLog, error) {
	oe := &OtelLog{
		ctx:    ctx,
		config: config,
	}

	logExporter, err := oe.newOtelGRPCLogExporter()
	if err != nil {
		return nil, err
	}

	res, err := newResource(config.ServiceName, config.ServiceVersion)
	if err != nil {
		return nil, err
	}

	logProvider, err := getLogProvider(logExporter, res)
	if err != nil {
		return nil, err
	}

	oe.logProvider = logProvider

	return oe, nil
}

// === LOGS ===
func (oe *OtelLog) newOtelGRPCLogExporter() (log.Exporter, error) {
	return otlploggrpc.New(oe.ctx,
		otlploggrpc.WithEndpointURL(oe.config.OtelEndpoint),
		otlploggrpc.WithHeaders(
			map[string]string{
				"Authorization": oe.config.OtelAuthorization,
				"stream-name":   oe.config.ServiceName,
				"organization":  "default",
			},
		),
		otlploggrpc.WithInsecure(),
	)
}
