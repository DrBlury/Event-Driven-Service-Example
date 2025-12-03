# Tracing Package

OpenTelemetry distributed tracing initialization.

## Overview

Configures and registers a global `trace.TracerProvider` for the application.

## Quick Start

```go
if cfg.Tracing.Enabled {
    if err := tracing.NewOtelTracer(ctx, logger, cfg.Tracing); err != nil {
        logger.Error("failed to initialize tracing", "error", err)
    }
}

// Use anywhere in the application
tracer := otel.Tracer("component-name")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()
```

## Configuration

```go
type Config struct {
    Enabled            bool
    OTELTracesExporter string  // "console", "otlp", or empty (disabled)
    OtelEndpoint       string  // OTLP endpoint
    Headers            string  // Custom OTLP headers
    ServiceName        string  // Resource attribute
    ServiceVersion     string  // Resource attribute
}
```

## Exporters

### Console
Pretty-printed trace output for development:
```go
cfg.OTELTracesExporter = "console"
```

### OTLP
Send traces to OpenTelemetry collector:
```go
cfg.OTELTracesExporter = "otlp"
cfg.OtelEndpoint = "http://localhost:5081"
```

### Disabled
Use no-op provider (no overhead):
```go
cfg.Enabled = false
// or
cfg.OTELTracesExporter = ""
```

## Environment Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `TRACING_ENABLED` | `true`, `false` | Enable distributed tracing |
| `OTEL_TRACES_EXPORTER` | `console`, `otlp` | Exporter backend |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | URL | OTLP endpoint for traces |
| `OTEL_EXPORTER_OTLP_TRACES_HEADERS` | Headers | Custom headers (key=value pairs) |

## Usage Example

```go
// Initialize (typically in main)
tracingCfg := &tracing.Config{
    Enabled:            true,
    OTELTracesExporter: "otlp",
    OtelEndpoint:       "http://openobserve:5081",
    ServiceName:        "my-service",
    ServiceVersion:     "v1.0.0",
}

if err := tracing.NewOtelTracer(ctx, logger, tracingCfg); err != nil {
    log.Fatal(err)
}

// Use in application code
tracer := otel.Tracer("http-server")

func HandleRequest(ctx context.Context, req *http.Request) {
    ctx, span := tracer.Start(ctx, "HandleRequest",
        trace.WithAttributes(
            attribute.String("http.method", req.Method),
            attribute.String("http.url", req.URL.String()),
        ),
    )
    defer span.End()

    // Add events
    span.AddEvent("processing request")

    // Nested spans
    processData(ctx)

    // Record errors
    if err := validateRequest(req); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "validation failed")
    }
}

func processData(ctx context.Context) {
    _, span := tracer.Start(ctx, "processData")
    defer span.End()
    // ... work ...
}
```

## Automatic Instrumentation

For HTTP servers and clients, use instrumentation libraries:

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// Wrap HTTP handler
handler := otelhttp.NewHandler(myHandler, "my-operation")

// Wrap HTTP client
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}
```

## Best Practices

1. **Use semantic conventions** for span names and attributes
2. **Propagate context** through the call chain
3. **Keep span names low-cardinality** (e.g., "GET /users/:id", not "GET /users/12345")
4. **Add meaningful attributes** for filtering and debugging
5. **Record errors** with `span.RecordError(err)`
6. **Set span status** to indicate success/failure
7. **Add events** for significant points in execution

## Span Lifecycle

```go
ctx, span := tracer.Start(ctx, "operation")
defer span.End()  // Always defer End()

// Add attributes
span.SetAttributes(
    attribute.String("key", "value"),
    attribute.Int("count", 42),
)

// Add events
span.AddEvent("checkpoint reached", trace.WithAttributes(
    attribute.String("detail", "processed 100 items"),
))

// Record errors
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

## Related Packages

- [metrics/](../metrics/) - OpenTelemetry metrics setup
- [logging/](../) - Structured logging with OTEL integration

## Documentation

- [Configuration Guide](../../../../docs/configuration.md) - Complete configuration reference
- [OpenTelemetry Tracing](https://opentelemetry.io/docs/specs/otel/trace/) - Specification
- [Go Tracing API](https://pkg.go.dev/go.opentelemetry.io/otel/trace) - API reference
- [Instrumentation Libraries](https://opentelemetry.io/ecosystem/registry/?language=go&component=instrumentation) - Pre-built integrations

