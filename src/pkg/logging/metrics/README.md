# Metrics Package

OpenTelemetry metrics initialization with minimal configuration.

## Overview

Configures and registers a global `metric.MeterProvider` for the application.

## Quick Start

```go
if cfg.Metrics.Enabled {
    if err := metrics.NewOtelMetrics(ctx, cfg.Metrics, logger); err != nil {
        logger.Error("failed to initialize metrics", "error", err)
    }
}

// Use anywhere in the application
meter := otel.Meter("component-name")
counter := meter.Int64Counter("requests_total")
counter.Add(ctx, 1)

```

## Configuration

```go

type Config struct {
    Enabled             bool
    OTELMetricsExporter string  // "console", "otlp", "prometheus"
    OtelEndpoint        string  // OTLP endpoint
    Headers             string  // Custom OTLP headers
    ServiceName         string  // Resource attribute
    ServiceVersion      string  // Resource attribute
}

```

## Exporters

### Console

Pretty-printed JSON output for development:

```go

cfg.OTELMetricsExporter = "console"

```

### OTLP

Send metrics to OpenTelemetry collector:

```go

cfg.OTELMetricsExporter = "otlp"
cfg.OtelEndpoint = "http://localhost:5081"

```

### Prometheus

Expose Prometheus-compatible metrics endpoint:

```go

cfg.OTELMetricsExporter = "prometheus"

```

## Environment Variables

| Variable | Values | Description |
| ---------- | -------- | ------------- |
| `OTEL_METRICS_EXPORTER` | `console`, `otlp`, `prometheus` | Exporter backend |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | URL | OTLP endpoint for metrics |
| `OTEL_EXPORTER_OTLP_METRICS_HEADERS` | Headers | Custom headers (key=value pairs) |
| `OTEL_METRICS_PRODUCERS` | `prometheus` | Additional metric producers |

## Usage Example

```go

// Initialize (typically in main)
metricsCfg := &metrics.Config{
    Enabled:             true,
    OTELMetricsExporter: "otlp",
    OtelEndpoint:        "http://openobserve:5081",
    ServiceName:         "my-service",
    ServiceVersion:      "v1.0.0",
}

if err := metrics.NewOtelMetrics(ctx, metricsCfg, logger); err != nil {
    log.Fatal(err)
}

// Use in application code
meter := otel.Meter("http-server")

requestCounter, _ := meter.Int64Counter(
    "http.server.requests",
    metric.WithDescription("Total HTTP requests"),
)

requestDuration, _ := meter.Float64Histogram(
    "http.server.duration",
    metric.WithDescription("HTTP request duration"),
    metric.WithUnit("ms"),
)

// Record metrics
requestCounter.Add(ctx, 1, metric.WithAttributes(
    attribute.String("method", "GET"),
    attribute.String("path", "/api/users"),
))

requestDuration.Record(ctx, 42.5, metric.WithAttributes(
    attribute.String("method", "GET"),
))

```

## Best Practices

1. **Use semantic conventions** for metric names and attributes
2. **Add units** to histogram metrics
3. **Use appropriate instruments:**

   - Counter: Monotonically increasing values (requests, errors)
   - Histogram: Distribution of values (latency, request size)
   - Gauge: Current value (CPU usage, queue depth)

4. **Keep cardinality low** on attribute values
5. **Use resource attributes** for service-level metadata

## Related Packages

- [tracing/](../tracing/) - OpenTelemetry tracing setup
- [logging/](../) - Structured logging with OTEL integration

## Documentation

- [Configuration Guide](../../../../docs/configuration.md) - Complete configuration reference
- [OpenTelemetry Metrics](https://opentelemetry.io/docs/specs/otel/metrics/) - Specification
- [Go Metrics API](https://pkg.go.dev/go.opentelemetry.io/otel/metric) - API reference
