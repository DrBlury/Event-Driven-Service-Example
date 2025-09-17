package logging

import "time"

type Config struct {
	// Open Telemetry
	OtelEndpoint               string
	OtelAuthorization          string
	OtelUseBasicAuth           bool
	OtelBasicAuthUsername      string
	OtelBasicAuthPassword      string
	OtelShowConsoleLogs        bool
	OtelExporterType           string
	OtelLogQueueSize           int
	OtelLogQueueIdleFlushTimer time.Duration
	ServiceName                string
	ServiceVersion             string

	// Logging basic configuration
	LogLevel string
	Logger   string
}

type ExporterType string

const (
	ExporterTypeConsole  ExporterType = "console"
	ExporterTypeOtelHttp ExporterType = "otel_http"
	ExporterTypeOtelGrpc ExporterType = "otel_grpc"
)
