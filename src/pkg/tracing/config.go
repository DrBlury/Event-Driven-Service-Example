package tracing

type Config struct {
	OtelEndpoint          string
	OtelAuthorization     string
	OtelUseBasicAuth      bool
	OtelBasicAuthUsername string
	OtelBasicAuthPassword string
	OTELTracesExporter    string
	ServiceName           string
	ServiceVersion        string
	Enabled               bool
}
