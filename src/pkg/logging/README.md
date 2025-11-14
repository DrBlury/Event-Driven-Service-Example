# Logging Package

## Overview

The `logging` package exposes a configurable façade over Go's structured slog
API. It helps boot services with consistent logging defaults, optional
OpenTelemetry export, and development friendly console rendering.

Design highlights:

- **Functional options** configure the logger without long parameter lists.
- Output formats (text, JSON, pretty) and exporters (console, OTLP) are selected
  at runtime by composing slog handlers.
- Adapters such as `RestyLogger` allow third-party libraries to reuse the
  configured slog instance.

## Primary Entry Point

`SetLogger(ctx, opts...)` builds a slog logger driven by the supplied `Option`
set. When `WithoutGlobal` is not used the logger also replaces the process-wide
slog default so standard library integrations automatically inherit the
configuration.

Common options include:

- `WithConfig` to hydrate settings from environment/config files.
- `WithLevel`, `WithLevelString` to specify the minimum log level.
- `WithJSONFormat`, `WithPrettyFormat`, or `WithTextFormat` to select console
  rendering.
- `WithOTel` to emit records through the OpenTelemetry bridge.

## Pretty Console Output

`NewPrettyHandler` wraps the JSON handler and renders colourised, indented logs
for local debugging sessions. It intentionally drops duplicate level, message,
and timestamp fields so the output stays compact yet readable.

## OpenTelemetry Integration

`WithOTel` enables the OTLP exporter via `autoexport`. When
`WithOTelConsoleMirror` is specified, the wrapped handler is mirrored through
`MyWrapperHandler`, which flattens JSON attributes before they are exported and
also keeps them readable in console output.

## Usage Example

```go
logger := logging.SetLogger(
    context.Background(),
    logging.WithLevel(slog.LevelDebug),
    logging.WithPrettyFormat(),
    logging.WithOTel(
        "signup-service",
        version,
        logging.WithOTelEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
        logging.WithOTelConsoleMirror(),
    ),
)
restyClient.SetLogger(logging.NewRestyLogger(logger))
```

The returned `*slog.Logger` can be decorated with `WithAttrs` to stamp service
identifiers or request metadata.
