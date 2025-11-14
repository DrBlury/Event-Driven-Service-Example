package metrics

// Config captures the knobs required to initialise OpenTelemetry metrics
// emission for the service.
type Config struct {
	OtelEndpoint        string
	Headers             string
	OTELMetricsExporter string
	ServiceName         string
	ServiceVersion      string
	Enabled             bool
}
