package tracing

// Config defines the OpenTelemetry tracing settings consumed by NewOtelTracer.
type Config struct {
	OtelEndpoint       string
	Headers            string
	OTELTracesExporter string
	ServiceName        string
	ServiceVersion     string
	Enabled            bool
}
