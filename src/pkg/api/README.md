# API Package

## Overview

The `api` package bundles reusable HTTP utilities for services that expose an
OpenAPI driven REST interface. It focuses on three concerns:

- Consistent JSON error envelopes and structured logging via `Responder`.
- Lightweight health, version, and documentation endpoints powered by
  `InfoHandler`.
- Tooling integration (`go:generate`) that keeps the generated API handlers in
  sync with the specification.

The package embraces the **functional options** pattern so callers can opt in to
optional collaborators without juggling long parameter lists.

## Key Components

### Responder

`Responder` wraps common HTTP tasks such as rendering JSON, decoding request
bodies, and emitting structured error payloads. Errors are enriched with a ULID,
category, timestamp, and log metadata so they remain traceable across systems.
`WithStatusMetadata` lets you override the logging level or error labels for
individual HTTP status codes.

### InfoHandler

`InfoHandler` combines the generated handlers with convenience endpoints that
expose build metadata (`GET /version`), health information (`GET /status`), and
an HTML viewer for your OpenAPI document. These collaborators are injected via
`InfoOption` values so the handler can be assembled with only the bits you need.

### Generation Workflow

The `generate.go` file wires `go generate` to a Dockerised instance of
`oapi-codegen`. Run `task gen-api` (as defined in `taskfile.yml`) whenever the
OpenAPI specification changes to refresh the generated server stubs. The
embedded HTML assets under `embedded/` are served by the info handler.

## Usage Example

```go
responder := api.NewResponder(
    api.WithLogger(logger),
)

infoHandler := api.NewInfoHandler(
    api.WithInfoResponder(responder),
    api.WithBaseURL("https://api.example.com"),
    api.WithInfoProvider(func() any {
        return map[string]string{
            "version": version,
            "commit":  commit,
        }
    }),
    api.WithSwaggerProvider(func() ([]byte, error) {
        return embeddedSpec, nil
    }),
)

mux := http.NewServeMux()
mux.HandleFunc("/status", infoHandler.GetStatus)
mux.HandleFunc("/version", infoHandler.GetVersion)
mux.HandleFunc("/docs", infoHandler.GetOpenAPIHTML)
mux.HandleFunc("/openapi.json", infoHandler.GetOpenAPIJSON)
```

## Integration Notes

- The responder is transport agnostic and can be shared across handlers to keep
  error semantics consistent.
- Pair the info handler with the router package to get request validation and
  logging out of the box.
- When running behind a reverse proxy, set `WithBaseURL` so the HTML viewer
  fetches the specification from the correct origin.
