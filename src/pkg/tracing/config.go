package tracing

type Config struct {
	OtelEndpoint       string
	Headers            string
	OTELTracesExporter string
	ServiceName        string
	ServiceVersion     string
	Enabled            bool
}
