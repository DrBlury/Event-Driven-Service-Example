# Router Package

## Overview

The `router` package produces an `http.ServeMux` preconfigured with validation,
logging, CORS, and timeout middleware. It is designed to sit in front of the
auto-generated handlers from `oapi-codegen` and reuse the configuration stored
in `router.Config`.

## Middleware Stack

Requests flow through the following middleware chain in order:

1. OpenAPI validation: requests are validated against the generated schema using
  `oapi-codegen` middleware.
2. CORS headers: driven by `CORSConfig` to make cross-origin calls predictable.
3. Timeout enforcement: ensures slow handlers do not occupy server resources
  indefinitely.
4. Structured logging: dumps method, path, headers (with optional redaction),
  and body size unless the route is listed in `QuietdownRoutes`.

## Configuration

- `Timeout`: per-request deadline applied by the HTTP timeout handler.
- `CORS`: lists allowed origins, methods, headers, and whether credentials are
  permitted.
- `QuietdownRoutes`: paths that should skip verbose request logging (useful for
  noisy health checks).
- `HideHeaders`: case-insensitive header keys that will be redacted before
  logging.

## Usage Example

```go
swagger, err := openapi3.NewLoader().LoadFromFile("./internal/server/_gen/openapi.json")
if err != nil {
    log.Fatal(err)
}

mux := router.New(
    generatedHandler,
    &router.Config{
        Timeout: 5 * time.Second,
        CORS: router.CORSConfig{
            Origins: []string{"https://app.example.com"},
            Methods: []string{"GET", "POST"},
            Headers: []string{"Content-Type", "Authorization"},
            AllowCredentials: true,
        },
        QuietdownRoutes: []string{"/status"},
        HideHeaders:     []string{"Authorization"},
    },
    logger,
    swagger,
)

http.ListenAndServe(":8080", mux)
```

Pair the router with the `api.InfoHandler` to expose health and documentation
endpoints on the same multiplexer.
