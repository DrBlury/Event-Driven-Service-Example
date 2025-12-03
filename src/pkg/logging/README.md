# Logging Package

Configurable structured logging façade over Go's `slog` with OpenTelemetry integration.

## Features

- Functional options for clean configuration
- Multiple output formats (text, JSON, pretty)
- OpenTelemetry log export via OTLP
- Development-friendly console rendering
- Third-party library adapters (Resty, etc.)

## Quick Start

```go
logger := logging.SetLogger(
    context.Background(),
    logging.WithLevel(slog.LevelDebug),
    logging.WithPrettyFormat(),
)

logger.Info("service started", "port", 8080)
```

## Configuration Options

### Log Level
```go
logging.WithLevel(slog.LevelDebug)
logging.WithLevelString("info")  // From config/env var
```

Levels: `debug`, `info`, `warn`, `error`

### Output Format

**Text** (human-readable):
```go
logging.WithTextFormat()
```

**JSON** (structured):
```go
logging.WithJSONFormat()
```

**Pretty** (colorized, indented - for development):
```go
logging.WithPrettyFormat()
```

### OpenTelemetry Export

**OTLP only** (no console):
```go
logging.WithOTel(
    "service-name",
    "v1.0.0",
    logging.WithOTelEndpoint("http://localhost:5081"),
)
```

**OTLP + console mirror**:
```go
logging.WithOTel(
    "service-name",
    "v1.0.0",
    logging.WithOTelEndpoint("http://localhost:5081"),
    logging.WithOTelConsoleMirror(),  // Also print to console
)
```

### Environment-Based Config

```go
// Load from Viper/env vars
logging.WithConfig(cfg.Logger)
```

## Environment Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `LOGGER` | `text`, `json`, `pretty`, `otel`, `otel-and-console` | Output format |
| `LOGGER_LEVEL` | `debug`, `info`, `warn`, `error` | Minimum log level |

See [Configuration Guide](../../../docs/configuration.md) for OTEL-specific variables.

## Third-Party Adapters

### Resty HTTP Client

```go
restyClient := resty.New()
restyClient.SetLogger(logging.NewRestyLogger(logger))
```

## Advanced Usage

### Custom Attributes

```go
logger = logger.With(
    "service", "api",
    "environment", "production",
)

logger.Info("request handled", "duration_ms", 42)
```

### Pretty Console Handler

The pretty handler renders colorized, indented JSON for local debugging:

```go
handler := logging.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
logger := slog.New(handler)
```

**Features:**
- Color-coded levels
- Indented JSON attributes
- Timestamp formatting
- Removes duplicate fields

## Best Practices

**Development:**
```go
logging.WithPrettyFormat()
logging.WithLevel(slog.LevelDebug)
```

**Production:**
```go
logging.WithJSONFormat()
logging.WithLevel(slog.LevelInfo)
logging.WithOTel(serviceName, version, ...)
```

**Testing:**
```go
logging.WithJSONFormat()
logging.WithLevel(slog.LevelError)  // Reduce noise
```

## Related Packages

- [metrics/](metrics/) - OpenTelemetry metrics configuration
- [tracing/](tracing/) - OpenTelemetry tracing setup

## Documentation

- [Configuration Guide](../../../docs/configuration.md) - Complete configuration reference
- [Go slog package](https://pkg.go.dev/log/slog) - Standard library documentation
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/) - OTEL integration guide

