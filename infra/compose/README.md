# Docker Compose Infrastructure

This directory contains modular Docker Compose configurations for local development.

## Quick Start

```bash
task up-kafka       # Start with Kafka
task up-rabbitmq    # Start with RabbitMQ
task up-aws         # Start with LocalStack (AWS)
task up-nats        # Start with NATS
task up-http        # Start with HTTP/MockServer
task up-io          # Start with in-memory queues
```text

## Architecture

**Base file** (`docker-compose.yml`): Common services (app, MongoDB, OpenObserve)

**Overlay files**: Add messaging backend and related tools

- `docker-compose.kafka.yml` - Kafka + Kafdrop UI
- `docker-compose.rabbitmq.yml` - RabbitMQ + Management UI
- `docker-compose.aws.yml` - LocalStack + Terraform provisioner
- `docker-compose.nats.yml` - NATS server
- `docker-compose.http.yml` - MockServer
- `docker-compose.io.yml` - In-memory mode (no external messaging)
- `docker-compose.debug.yml` - Development mode with hot reload

## Manual Usage

```bash
# Kafka stack
docker compose -f docker-compose.yml -f docker-compose.kafka.yml up

# With debug mode (hot reload)
docker compose -f docker-compose.yml -f docker-compose.kafka.yml -f docker-compose.debug.yml up
```

## Data Persistence

All persistent data is stored in `../../_volume_data/`:

```text
_volume_data/
├── mongo/          # MongoDB data
├── kafka/          # Kafka logs
├── rabbitmq/       # RabbitMQ data
├── localstack/     # LocalStack state
└── openobserve/    # Observability data
```

**Reset all data:**

```bash
docker compose down -v
rm -rf ../../_volume_data/
```

## Service Ports

| Service | Port | Description |
| ------- | ---- | ----------- |
| app | 8080 | HTTP API |
| app | 8085 | Protoflow metadata API (if enabled) |
| mongo-express | 8081 | MongoDB web UI |
| kafdrop | 9000 | Kafka web UI |
| rabbitmq | 15672 | RabbitMQ management UI |
| openobserve | 5080 | Observability UI |
| localstack | 4566 | AWS service emulator |

## Environment Configuration

Environment files are stored in `../env/`:

```bash
# Generate from examples
task gen-env-files

# Edit as needed
vim ../env/app.env
vim ../env/kafka.env
# etc.
```text

## Detailed Documentation

See [Infrastructure Guide](../../docs/infrastructure.md) for comprehensive information on:

- Service details and configuration
- Networking and ports
- Terraform integration
- Troubleshooting
- Best practices


