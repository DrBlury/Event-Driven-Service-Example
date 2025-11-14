# Tracing Package

## Overview

The `tracing` package initialises the global OpenTelemetry tracer provider. It
leans on `autoexport` to detect the correct OTLP exporter or falls back to a
console exporter for local debugging. The goal is to make distributed tracing an
opt-in feature that requires only configuration, not bespoke setup code.

## Configuration

`Config` mirrors common OTLP environment variables:

- `OTELTracesExporter`: choose between `console`, `otlp`, or leave empty to
  disable tracing (a no-op exporter is used).
- `OtelEndpoint` and `Headers`: forwarded to the underlying OTLP exporter when
  relevant.
- `ServiceName` and `ServiceVersion`: attached to traces as resource attributes.
- `Enabled`: optional flag callers can evaluate before invoking
  `NewOtelTracer`.

## Pattern

`NewOtelTracer` composes exporters, resources, and the tracer provider before
installing it with `otel.SetTracerProvider`.

## Usage Example

```go
if tracingCfg.Enabled {
    if err := tracing.NewOtelTracer(ctx, logger, tracingCfg); err != nil {
        logger.With("component", "tracing").Error("failed to initialise tracing", slog.Any("error", err))
    }
}
```

Once configured you can instrument code using `otel.Tracer("component")` and
leverage the spans exported by your chosen backend.
