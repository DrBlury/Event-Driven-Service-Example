# GitHub Copilot Instructions for Event-Driven-Service-Example

# https://docs.github.com/en/copilot/customizing-copilot/adding-repository-custom-instructions-for-github-copilot

applyTo: "\*\*"

instructions: |

## Project Overview

This is an event-driven Go service using APIWeaver for HTTP APIs and Protoflow for message handling.

## Architecture

-   `src/internal/server/` - HTTP handlers generated from OpenAPI specs
-   `src/internal/events/` - Event handlers using Protoflow
-   `src/internal/usecase/` - Business logic layer
-   `src/internal/domain/` - Domain models generated from protobuf
-   `src/internal/database/` - Database operations (MongoDB)
-   `proto/` - Protobuf definitions compiled with Buf
-   `api/` - OpenAPI 3.1 specifications
-   `async-api/` - AsyncAPI 3.0 event contract documentation

## Code Style

-   Follow Go idioms and Effective Go guidelines
-   Use structured logging with `slog`
-   Handle errors explicitly, never ignore them
-   Prefer dependency injection over global state
-   Write table-driven tests
-   Use context.Context for cancellation and deadlines

## API Development

-   API contracts are defined in `api/api.yml` (OpenAPI 3.1)
-   Run `task gen-api` after modifying OpenAPI specs
-   Use oapi-codegen generated types, don't create duplicates

## Event Handling

-   Events use Protoflow with Watermill under the hood
-   Domain messages are defined in `proto/domain/v1/`
-   Event contracts documented in `async-api/asyncapi.yml` (AsyncAPI 3.0)
-   Supports inbound events (event → event processing), not just API → event
-   Run `task gen-buf` after modifying protobuf files
-   Run `task gen-asyncapi` to lint and bundle AsyncAPI spec
-   AsyncAPI JSON served at `/info/asyncapi.json`, docs at `/info/asyncapi.html`
-   Always include correlation IDs in events

## Testing

-   Write unit tests alongside implementation
-   Use testify/assert for assertions
-   Mock external dependencies
-   Run tests with race detector: `go test -race ./...`

## Common Tasks

-   `task gen-api` - Regenerate API code from OpenAPI
-   `task gen-asyncapi` - Lint and bundle AsyncAPI spec
-   `task gen-buf` - Regenerate protobuf Go code
-   `task gen-all` - Regenerate all code (protobuf + OpenAPI + AsyncAPI)
-   `task lint-go` - Run golangci-lint
-   `task lint-all` - Run all linters (Go, API, AsyncAPI, Proto, Docker, Actions)
-   `task test-go` - Run tests with coverage
-   `task up-kafka` / `task up-rabbitmq` - Start local infrastructure

## Dependencies

-   Configuration: Viper
-   HTTP routing: APIWeaver + net/http (stdlib)
-   Event handling: Protoflow (Watermill-based)
-   Database: MongoDB driver
-   Observability: OpenTelemetry (OTEL)
-   Logging: slog with OTEL bridge
