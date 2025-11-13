package logging

// Format represents the console log output style.
type Format string

const (
	// FormatText uses the slog text handler.
	FormatText Format = "text"
	// FormatJSON uses the slog JSON handler.
	FormatJSON Format = "json"
	// FormatPretty renders coloured, pretty printed JSON to the console.
	FormatPretty Format = "pretty"
)

// Config captures the high level logger configuration that can be derived
// from environment variables or configuration files.
type Config struct {
	Level          string
	Format         Format
	AddSource      bool
	ConsoleEnabled bool
	SetAsDefault   bool
	OTel           OTelConfig
}

// OTelConfig holds OpenTelemetry related settings.
type OTelConfig struct {
	Enabled         bool
	MirrorToConsole bool
	Endpoint        string
	Headers         string
	ServiceName     string
	ServiceVersion  string
}
