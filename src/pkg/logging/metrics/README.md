# Metrics Package

## Overview

The `metrics` package wires OpenTelemetry metric exporters with minimal setup.
It hides the boilerplate required to build a `metric.MeterProvider` and updates
the global provider used by the SDK once initialisation succeeds.

## Configuration

`Config` mirrors the most common knobs required by OTLP exporters:

- `OTELMetricsExporter`: selects the exporter backend. Supported values are
  `console` (pretty printed JSON) and `otlp` (auto-detected OTLP exporter).
- `OtelEndpoint` and `Headers`: additional OTLP transport parameters when using
  `autoexport`.
- `ServiceName` and `ServiceVersion`: propagated as resource attributes so they
  show up in dashboards.
- `Enabled`: conventional flag that callers can check before invoking
  `NewOtelMetrics`.

## Pattern

`NewOtelMetrics` assembles the required exporters and installs the resulting
meter provider. It returns an error when the configuration is unsupported or
misconfigured so callers can decide how to degrade gracefully.

## Usage Example

```go
if metricsCfg.Enabled {
    if err := metrics.NewOtelMetrics(ctx, metricsCfg, logger); err != nil {
        logger.With("component", "metrics").Error("failed to initialise metrics", slog.Any("error", err))
    }
}
```

Once the provider is registered you can instrument components using
`otel.Meter("component")` without additional wiring.
