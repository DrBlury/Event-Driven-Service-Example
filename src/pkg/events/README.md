# Events Package

## Overview

The `events` package assembles the infrastructure required to build event-driven
workloads with [Watermill](https://watermill.io/). It abstracts away provider
specific plumbing (Kafka, RabbitMQ, SNS/SQS) and exposes a single `Service`
object that owns the router, publisher, subscriber, and middleware chain.

Key goals:

- Pluggable transports selected at runtime via configuration.
- Composable middleware stitched together using declarative registrations.
- Schema-aware processing powered by protobuf validation and an optional outbox
  store.

## Architectural Pattern

The service acts as the central coordinator for message handling: application
code registers handlers and the service wires them to the configured
publisher/subscriber infrastructure. Middleware registrations execute in the
order they are registered, letting cross-cutting concerns wrap handlers without
invasive changes.

## Core Types

- `Service`: wraps the Watermill router and exposes helper methods for registering
  handlers, middlewares, and protobuf schemas.
- `Config`: declares queue names, retry policies, and transport specific
  parameters. Only the relevant fields are read for each backend.
- `ServiceDependencies`: optional collaborators such as outbox storage or a
  protobuf validator. Use `DisableDefaultMiddlewares` to replace the built-in
  chain entirely.
- `HandlerRegistration`: describes how a single handler should be wired,
  including overrides for publisher/subscriber instances and message prototypes.

## Default Middleware Chain

`DefaultMiddlewares()` returns a stack that prioritises observability and data
quality:

1. Correlation ID injection.
2. Structured message logging.
3. Protobuf validation (optional when a validator is provided).
4. Outbox persistence.
5. OpenTelemetry tracing spans.
6. Exponential backoff retry.
7. Poison queue forwarding for unrecoverable payloads.
8. Panic recovery to keep the router healthy.

Use `RegisterMiddleware` if you need full control over the ordering.

## Basic Usage

```go
ctx := context.Background()
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

svc := events.NewService(
    &events.Config{
        PubSubSystem:      "kafka",
        KafkaBrokers:      []string{"localhost:29092"},
        KafkaConsumerGroup: "signup-service",
        PoisonQueue:       "signup-poison",
    },
    logger,
    ctx,
    events.ServiceDependencies{
        Validator: myProtoValidator,
        Outbox:    myOutboxStore,
    },
)

err := svc.RegisterHandler(events.HandlerRegistration{
    Name:             "signup_processed",
    ConsumeQueue:     "signup-input",
    PublishQueue:     "signup-output",
    MessagePrototype: &domain.Signup{},
    Handler: func(msg *message.Message) ([]*message.Message, error) {
        // implement domain logic
        return nil, nil
    },
})
if err != nil {
    log.Fatal(err)
}

if err := svc.Start(ctx); err != nil {
    log.Fatal(err)
}
```

The example assumes the generated protobuf messages live in the `internal/domain`
package and are imported as `domain`.

## Extending the Pipeline

- Implement `MiddlewareBuilder` to lazily construct middlewares that require
  access to the `Service` or its dependencies.
- Use `OutboxMiddleware()` in conjunction with `ServiceDependencies.Outbox` to
  implement the transactional outbox pattern.
- Register standalone protobuf messages with `RegisterProtoMessage` so the
  validator recognises events produced elsewhere.

## Testing Helpers

The package ships with unit tests that demonstrate how to stub factories (Kafka,
RabbitMQ, AWS) and verify middleware behaviour. Override the `*_factory`
variables in tests to swap in fakes without touching production code.
