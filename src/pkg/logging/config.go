package logging

type Config struct {
	// Open Telemetry
	OtelEndpoint   string
	Headers        string
	ServiceName    string
	ServiceVersion string

	// Logging basic configuration
	LogLevel string
	Logger   string
}
