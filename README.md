# Event-Driven Service Example

This repository is a from-scratch reference implementation of a production-style event-driven service. It exposes an HTTP API powered by **APIWeaver**, produces and consumes events via **Protoflow**, persists data in MongoDB, and stitches everything together with a modern tooling stack (Task, Docker, Terraform, Buf, oapi-codegen, act, OTEL, and more). Use it to learn, prototype, or as a baseline for your own services.

## Overview

- **Purpose**: Demonstrate how to combine synchronous APIs and asynchronous processing in a cohesive Go codebase.
- **HTTP surface**: API contracts live in `api/api.yml` (OpenAPI 3.1). APIWeaver and oapi-codegen generate request handlers that translate HTTP traffic into domain calls.
- **Event surface**: Protoflow wires Kafka, RabbitMQ, or AWS SNS/SQS pipelines, handling middleware, retries, tracing, and poison queues for you.
- **Foundation**: Configuration is centralized with Viper, logging uses `slog`, instrumentation flows through OpenTelemetry (OTEL), and protobuf models represent the domain.

## Key Technologies

| Layer | Tools & Libraries | Role |
| --- | --- | --- |
| HTTP & routing | **APIWeaver**, **OpenAPI**, **oapi-codegen** | Declarative API-first workflow with generated routers and request objects. |
| Event pipeline | **Protoflow**, Kafka/RabbitMQ/AWS | Typed JSON/Protobuf handlers with Watermill under the hood. |
| Data & contracts | **Protobuf**, **Buf** | Strongly typed domain messages shared across API and events. |
| Configuration & logging | **Viper**, **slog** | Environment-driven config loading plus structured logging. |
| Observability | **OTEL**, OpenObserve | Traces, metrics, and logs emitted via OpenTelemetry bridges. |
| Automation | **Task**, **Docker**, **Terraform**, **act** | Reproducible local dev (`task`), container stacks, IaC, and local CI emulation. |

## Architecture Highlights

- **API edge**: `src/internal/server` is generated from OpenAPI definitions. APIWeaver routes requests into use cases located in `src/internal/usecase`.
- **Domain models**: `proto/` definitions are compiled with Buf into Go types inside `src/internal/domain`.
- **Event orchestration**: `src/internal/events` registers Protoflow middleware, validators, and handlers. The same service publishes events via Protoflow producers.
- **Observability**: Logging bridges convert `slog` output into Protoflow-compatible logs, while OTEL exporters ship traces/metrics/logs to whatever backend you configure.
- **Infrastructure**: `infra/compose` holds Docker Compose stacks for Kafka, RabbitMQ, and LocalStack. `infra/terraform` demonstrates how to provision cloud resources with Terraform modules.

## Getting Started

### Prerequisites

- Go 1.25.4+
- Docker + Docker Compose
- [Task](https://taskfile.dev/) CLI (`brew install go-task/tap/go-task` on macOS)
- Optional: Terraform, act, Redocly CLI, Buf (these run via containers but installing locally speeds things up)

### Bootstrap the workspace

```bash
task gen-env-files          # copy example env vars into infra/env
task gen-buf                # compile protobufs into Go models
task gen-api                # lint OpenAPI + regenerate oapi-codegen stubs
task build-go               # compile the service
```

### Start a local stack

Pick the pub/sub backend you want to explore:

- `task up-kafka`
- `task up-rabbitmq`
- `task up-aws` (spins LocalStack + OpenObserve)
- `PUBSUB_SYSTEM=kafka task debug` to run the app with live code reloading against a compose stack.

Once the containers are healthy, the API is available at the address configured by `APP_SERVER_PORT` (default `:80`). Health probes live at `/healthz` and `/readyz`, while `/info/status` shows build metadata.

## Development Workflow

1. **Design or update the API**: edit `api/api.yml`, then run `task gen-api` to lint with Redocly and regenerate servers with oapi-codegen + APIWeaver bindings.
2. **Evolve events/domain**: change protobuf files under `proto/`, then run `task gen-buf`. Protoflow immediately sees new message types.
3. **Code business logic**: implement handlers in `src/internal/server/handler/*` and `src/internal/events`.
4. **Run locally**: use the compose tasks above or run only the Go binary with `go run ./src` while relying on external infra.
5. **Validate CI locally**: `task ci` executes every GitHub Actions job via `act`, matching the remote workflow.

## Observability & Operations

- **Logging**: Structured through `slog`, mirrored into Protoflow’s Watermill adapters.
- **Tracing & Metrics**: Exported with OTEL (`go.opentelemetry.io/otel` plus auto instrumentation). Configure OTLP endpoints via env vars (`OTEL_EXPORTER_OTLP_*`).
- **Poison queues & retries**: Protoflow middlewares provide correlation IDs, validation, retries, and poison queue routing. Tune values via `PROTOFLOW_*` env vars (loaded with Viper).
- **Monitoring**: When running the AWS/LocalStack stack, OpenObserve becomes available for quick dashboards.

## Infrastructure & Deployment

- **Docker & Compose**: Everything needed for local experimentation lives under `infra/compose`. Images follow the configs in `infra/build/dockerfiles/`.
- **Terraform**: Use `infra/terraform` to study how the service could be provisioned in real environments. Modules and environment definitions live under `infra/terraform/environments` and `infra/terraform/modules`.
- **Pipelines**: GitHub Actions workflows exercise linting, tests, and container builds. `act` mirrors those runs locally.

## Documentation Map

- `docs/setup.md` – deeper setup walkthroughs (IDE tooling, credentials, etc.).
- `docs/infrastructure.md` – Terraform module structure and deployment topology.
- `docs/protobuf.md` – rationale for protobuf-first domain modeling.
- `docs/pubsub.md` – how Kafka, RabbitMQ, and AWS are wired via Protoflow.
- `docs/watermill.md` – Watermill and OTEL integration stories.
- `docs/message_processing.md` – handler flowcharts, middleware ordering, and poison queue handling.

## Helpful Utilities

- `task git:web` – open the default Git remote; override with `REMOTE=<name>`.
- `scripts/git-web` – helper backing the task; add `scripts/` to your `PATH` for `git web`.
- `go run ./scripts/update-schema-index.go` – refresh `api/schemas/_index.yml` so new schema fragments are available to oapi-codegen and Redocly.

Happy hacking! Experiment with APIWeaver + Protoflow together, plug in new transports, or fork the infra to match your cloud of choice.
