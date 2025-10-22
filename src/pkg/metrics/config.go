package metrics

type Config struct {
	OtelEndpoint        string
	Headers             string
	OTELMetricsExporter string
	ServiceName         string
	ServiceVersion      string
	Enabled             bool
}
