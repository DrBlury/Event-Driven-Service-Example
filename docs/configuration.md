# Configuration Guide

This document describes all configuration options available in the Event-Driven Service Example.

## Configuration Sources

Configuration is loaded using [Viper](https://github.com/spf13/viper) in the following order (highest priority first):

1. **Environment variables** (highest priority)
2. **Configuration files** (if specified)
3. **Default values** (defined in code)

## Quick Start

Generate environment files from examples:

```bash
task gen-env-files
```

This copies templates from `infra/env/example/` to `infra/env/` with default values.

## Application Configuration

### Server Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_NAME` | `example-service` | Service identifier used in logs and traces |
| `APP_SERVER_PORT` | `80` | HTTP server port |
| `APP_SERVER_BASE_URL` | `http://localhost:8080` | Base URL for the service |
| `APP_SERVER_TIMEOUT` | `60s` | Request timeout duration |
| `VERSION` | `dev-local` | Application version for telemetry |

### CORS Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_SERVER_CORS_ORIGINS` | `*` | Allowed origins (comma-separated) |
| `APP_SERVER_CORS_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| `APP_SERVER_CORS_HEADERS` | `*` | Allowed headers |
| `APP_SERVER_CORS_ALLOW_CREDENTIALS` | `false` | Allow credentials in CORS requests |

### Security & Privacy

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_SERVER_HIDE_HEADERS` | `Authorization` | Headers to redact from logs (comma-separated) |
| `APP_SERVER_QUIETDOWN_ROUTES` | `/info/version,/info/status,/info/openapi.json` | Routes excluded from verbose logging |

## Logging Configuration

### Basic Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `LOGGER` | `json` | Output format: `text`, `json`, `pretty`, `otel`, `otel-and-console` |
| `LOGGER_LEVEL` | `debug` | Minimum log level: `debug`, `info`, `warn`, `error` |

### Logger Format Options

- **`text`**: Human-readable text format (slog TextHandler)
- **`json`**: Structured JSON output (slog JSONHandler)
- **`pretty`**: Colorized, indented JSON for local development
- **`otel`**: Send logs via OpenTelemetry only (no console output)
- **`otel-and-console`**: Mirror logs to both console and OTEL

## OpenTelemetry (OTEL) Configuration

### Logs

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_LOGS_EXPORTER` | - | Exporter type: `otlp`, `console`, or empty to disable |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | - | OTLP endpoint for logs (e.g., `http://localhost:5081`) |
| `OTEL_EXPORTER_OTLP_LOGS_HEADERS` | - | Custom headers for OTLP logs (comma-separated key=value) |

### Traces

| Variable | Default | Description |
|----------|---------|-------------|
| `TRACING_ENABLED` | `false` | Enable distributed tracing |
| `OTEL_TRACES_EXPORTER` | - | Exporter type: `otlp`, `console`, or empty to disable |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | - | OTLP endpoint for traces |
| `OTEL_EXPORTER_OTLP_TRACES_HEADERS` | - | Custom headers for OTLP traces |

### Metrics

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_METRICS_EXPORTER` | - | Exporter type: `otlp`, `console`, `prometheus`, or empty to disable |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | - | OTLP endpoint for metrics |
| `OTEL_EXPORTER_OTLP_METRICS_HEADERS` | - | Custom headers for OTLP metrics |
| `OTEL_METRICS_PRODUCERS` | - | Additional metric producers: `prometheus` |

### OTEL Protocol

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` | Protocol: `grpc` or `http/protobuf` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | Common endpoint for all signals (if specific endpoints not set) |

### OpenObserve Integration

When using OpenObserve, configure authorization headers:

```bash
# Base64 encoded "email:password"
OTEL_EXPORTER_OTLP_LOGS_HEADERS="Authorization=Basic <base64>,stream-name=service-name,organization=default"
```

## Database Configuration

### MongoDB

| Variable | Default | Description |
|----------|---------|-------------|
| `MONGO_URL` | - | Full MongoDB connection string |
| `MONGO_DB` | `serviceflow` | Database name |
| `MONGO_USER` | `root` | MongoDB username |
| `MONGO_PASSWORD` | `example` | MongoDB password |

Example connection string:

```bash
MONGO_URL=mongodb://root:example@mongo:27017/serviceflow?authSource=admin
```

## Event Processing (Protoflow)

### Pub/Sub System Selection

| Variable | Default | Description |
|----------|---------|-------------|
| `PUBSUB_SYSTEM` | - | Backend: `kafka`, `rabbitmq`, `aws`, `nats`, `http`, `io` |

### Queue Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PROTOFLOW_POISON_QUEUE` | `messages-poison` | Dead letter queue for failed messages |
| `EVENTS_DEMO_CONSUME_QUEUE` | `messages` | Demo handler input queue |
| `EVENTS_DEMO_PUBLISH_QUEUE` | `messages-processed` | Demo handler output queue |
| `EVENTS_EXAMPLE_CONSUME_QUEUE` | `example-records` | Example record input queue |
| `EVENTS_EXAMPLE_PUBLISH_QUEUE` | `example-records-processed` | Example record output queue |

### Protoflow Web UI

Protoflow includes a metadata API for debugging registered handlers.

| Variable | Default | Description |
|----------|---------|-------------|
| `PROTOFLOW_WEBUI_ENABLED` | `false` | Enable metadata API server |
| `PROTOFLOW_WEBUI_PORT` | `8085` | Metadata API port |

When enabled, visit `http://localhost:8085/api/handlers` to see registered event handlers and their configuration.

### Kafka-Specific Configuration

Kafka configuration is typically provided via environment variables or Protoflow config. Refer to [Watermill Kafka](https://watermill.io/pubsubs/kafka/) documentation for advanced options.

### RabbitMQ-Specific Configuration

RabbitMQ connection is configured through Protoflow. See [Watermill AMQP](https://watermill.io/pubsubs/amqp/) for details.

### AWS (SNS/SQS) Configuration

When using LocalStack or AWS:

| Variable | Default | Description |
|----------|---------|-------------|
| `AWS_REGION` | `us-east-1` | AWS region |
| `AWS_ENDPOINT` | - | Override endpoint (for LocalStack) |
| `AWS_ACCESS_KEY_ID` | - | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | - | AWS secret key |

## Container-Specific Configuration

### Kafka

See `infra/env/example/kafka.env.example` for Kafka broker settings.

### MongoDB

See `infra/env/example/mongo.env.example` for MongoDB initialization variables.

### OpenObserve

See `infra/env/example/openobserve.env.example` for observability backend credentials.

## Development Configuration

### Hot Reload (Air)

The project uses [Air](https://github.com/air-verse/air) for live code reloading during development.

Configuration is in `.air.toml`:

- **Watch paths**: `src/`
- **Excludes**: `_test.go`, vendor, generated files
- **Rebuild delay**: 1000ms

Start with hot reload:

```bash
task debug              # With selected system
SYSTEM=kafka task debug # Explicitly select Kafka
```

### Debug Mode

| Variable | Default | Description |
|----------|---------|-------------|
| `HOT_RELOAD` | `true` | Enable Air hot reload in debug mode |
| `SYSTEM` | - | Override pub/sub system for debug mode |

## API Documentation

### OpenAPI

The service exposes OpenAPI documentation:

- **Spec**: `GET /info/openapi.json`
- **HTML Docs**: `GET /info/openapi.html`

### AsyncAPI

Event contract documentation:

- **Spec**: `GET /info/asyncapi.json`
- **HTML Docs**: `GET /info/asyncapi.html`

## Environment File Structure

```text
infra/env/
├── app.env                    # Application configuration
├── kafka.env                  # Kafka broker settings
├── kafdrop.env                # Kafka UI settings
├── mongo.env                  # MongoDB settings
├── mongo-express.env          # MongoDB UI settings
├── rabbitmq.env               # RabbitMQ settings
├── openobserve.env            # Observability backend
└── localstack.env             # LocalStack (AWS emulation)
```

## Configuration Best Practices

### Security

1. **Never commit secrets** to version control
2. Use `.env` files locally (listed in `.gitignore`)
3. Use secret management in production (AWS Secrets Manager, Vault, etc.)
4. Rotate credentials regularly

### Local Development

1. Run `task gen-env-files` to create initial configs
2. Customize `infra/env/*.env` for your local setup
3. Use `LOGGER=pretty` for readable local logs
4. Enable Protoflow Web UI with `PROTOFLOW_WEBUI_ENABLED=true`

### Production

1. Set `LOGGER=otel` or `LOGGER=json`
2. Configure proper OTEL endpoints
3. Use strict CORS policies
4. Enable TLS/HTTPS
5. Set appropriate timeouts
6. Hide sensitive headers with `APP_SERVER_HIDE_HEADERS`

## Troubleshooting

### Configuration Not Loading

1. Verify environment variable names are correct (case-sensitive)
2. Check `infra/env/` files exist (run `task gen-env-files`)
3. Ensure `taskfile.yml` includes correct `dotenv` paths
4. Review logs for configuration errors

### OTEL Not Exporting

1. Verify OTEL endpoint is reachable
2. Check `OTEL_*_EXPORTER` is set correctly
3. Ensure `LOGGER=otel-and-console` or `LOGGER=otel`
4. Review OTEL logs with `LOGGER_LEVEL=debug`

### Database Connection Issues

1. Verify MongoDB is running: `docker ps | grep mongo`
2. Check `MONGO_URL` connection string
3. Ensure network connectivity between containers
4. Review MongoDB logs: `docker logs <container>`

## Related Documentation

- [README.md](../README.md) - Project overview and quick start
- [CONTRIBUTING.md](../.github/CONTRIBUTING.md) - Development workflow
- [Infrastructure Guide](infrastructure.md) - Deployment and infrastructure
- [docs/code-metrics.md](code-metrics.md) - Code complexity tracking
