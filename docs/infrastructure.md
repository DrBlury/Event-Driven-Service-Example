# Infrastructure Guide

This document describes the infrastructure components, deployment options, and local development setup.

## Overview

The service can run with multiple messaging backends and includes observability, database, and development tooling. All infrastructure is defined using Docker Compose for local development and Terraform for cloud deployments.

## Local Development Stacks

### Available Compose Stacks

All stacks include base services (app, MongoDB, OpenObserve). Choose a messaging backend:

| Command | Backend | Additional Services | Use Case |
|---------|---------|---------------------|----------|
| `task up-kafka` | Kafka | Kafka, Kafdrop (UI) | High-throughput event streaming |
| `task up-rabbitmq` | RabbitMQ | RabbitMQ Management UI | Traditional message queuing |
| `task up-aws` | AWS SNS/SQS | LocalStack, Terraform | AWS service emulation |
| `task up-nats` | NATS | NATS server | Lightweight cloud-native messaging |
| `task up-http` | HTTP | MockServer | HTTP-based event delivery |
| `task up-io` | In-memory | - | Testing without external dependencies |

### Base Services (Included in All Stacks)

- **app**: The Go application
- **mongo**: MongoDB 8.0 database
- **mongo-express**: MongoDB web UI (port 8081)
- **openobserve**: Observability backend (port 5080)

### Stopping Stacks

```bash
task down-kafka       # Stop Kafka stack
task down-rabbitmq    # Stop RabbitMQ stack
task down-aws         # Stop AWS/LocalStack stack
# etc.
```

## Compose File Structure

The infrastructure uses a modular compose setup:

```text
infra/compose/
├── docker-compose.yml          # Base: app, MongoDB, OpenObserve
├── docker-compose.kafka.yml    # Kafka + Kafdrop overlay
├── docker-compose.rabbitmq.yml # RabbitMQ overlay
├── docker-compose.aws.yml      # LocalStack + Terraform overlay
├── docker-compose.nats.yml     # NATS overlay
├── docker-compose.http.yml     # MockServer overlay
├── docker-compose.io.yml       # IO mode (no external services)
└── docker-compose.debug.yml    # Development overlay (hot reload)
```

**Benefits:**

- No duplication across stack definitions
- Easy to add new messaging backends
- Clean separation of concerns

**Manual Usage:**

```bash
# Start Kafka stack manually
docker compose -f infra/compose/docker-compose.yml \
               -f infra/compose/docker-compose.kafka.yml up

# Start with debug mode
docker compose -f infra/compose/docker-compose.yml \
               -f infra/compose/docker-compose.kafka.yml \
               -f infra/compose/docker-compose.debug.yml up
```

## Service Details

### Kafka Stack

**Services:**

- `kafka`: Kafka broker (port 9092)
- `kafdrop`: Web UI for Kafka (port 9000)

**Access:**

- Kafka: `localhost:9092`
- Kafdrop UI: <http://localhost:9000>

**Configuration:** `infra/env/kafka.env`, `infra/env/kafdrop.env`

### RabbitMQ Stack

**Services:**

- `rabbitmq`: RabbitMQ broker with management plugin

**Access:**

- AMQP: `localhost:5672`
- Management UI: <http://localhost:15672> (guest/guest)

**Configuration:** `infra/env/rabbitmq.env`

### AWS/LocalStack Stack

**Services:**

- `localstack`: AWS service emulator (SNS, SQS, S3, etc.)
- `terraform`: Applies Terraform configuration to LocalStack

**Access:**

- LocalStack: `localhost:4566`
- Health: <http://localhost:4566/_localstack/health>

**Configuration:** `infra/env/localstack.env`

**Terraform Setup:**
The stack automatically runs `terraform init` and `terraform apply` to provision SNS topics and SQS queues. See [Terraform section](#terraform).

### NATS Stack

**Services:**

- `nats`: NATS server

**Access:**

- NATS: `localhost:4222`
- Monitoring: `localhost:8222`

### HTTP Stack

**Services:**

- `mockserver`: HTTP mock server for event delivery

**Access:**

- MockServer: `localhost:1080`
- Dashboard: <http://localhost:1080/mockserver/dashboard>

### MongoDB

**Configuration:**

- Database: `serviceflow`
- User: `root` / Password: `example`
- Connection: `mongodb://root:example@mongo:27017/serviceflow?authSource=admin`

**Access:**

- MongoDB: `localhost:27017`
- Mongo Express UI: <http://localhost:8081>

**Configuration:** `infra/env/mongo.env`, `infra/env/mongo-express.env`

### OpenObserve

Observability backend for logs, traces, and metrics.

**Access:**

- UI: <http://localhost:5080>
- OTLP gRPC: `localhost:5081`
- Credentials: `root@example.com` / `Complexpass#123`

**Configuration:** `infra/env/openobserve.env`

**Integration:**
Set these environment variables in `app.env`:

```bash
OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="http://openobserve:5081"
OTEL_EXPORTER_OTLP_LOGS_HEADERS="Authorization=Basic <base64>,stream-name=service-name,organization=default"
```

## Data Persistence

All persistent data is stored in `_volume_data/` at the repository root:

```text
_volume_data/
├── mongo/          # MongoDB data
├── kafka/          # Kafka logs and data
├── rabbitmq/       # RabbitMQ data
├── localstack/     # LocalStack state
└── openobserve/    # OpenObserve data
```

**Note:** `_volume_data/` is in `.gitignore`. To reset all data:

```bash
docker compose down -v
rm -rf _volume_data/
```

## Development Mode

### Hot Reload with Air

Start the service with automatic rebuilds on code changes:

```bash
task debug              # Uses default or $SYSTEM env var
SYSTEM=kafka task debug # Explicitly select Kafka
```

**Configuration:** `.air.toml`

- Watches: `src/` directory
- Excludes: tests, vendor, generated code
- Rebuild delay: 1000ms

### Protoflow Web UI

Enable the Protoflow metadata API to inspect registered handlers:

```bash
# In infra/env/app.env
PROTOFLOW_WEBUI_ENABLED=true
PROTOFLOW_WEBUI_PORT=8085
```text

Access: <http://localhost:8085/api/handlers>

### Debug Without Hot Reload

For IDE-based debugging without Air:

```bash
task dev-no-reload
```

## Terraform

### Overview

Terraform modules provision cloud resources. The LocalStack environment demonstrates the patterns.

**Structure:**

```text
infra/terraform/
├── modules/
│   ├── iam/        # IAM roles and policies
│   ├── sns/        # SNS topics
│   └── sqs/        # SQS queues
└── environments/
    └── localstack/ # LocalStack configuration
```

### LocalStack Deployment

The Terraform configuration creates:

- SNS topics for event publishing
- SQS queues for event consumption
- Queue-to-topic subscriptions

**Automated (via Docker Compose):**

```bash
task up-aws
```

The `terraform` service automatically runs:

```bash
terraform init
terraform apply -auto-approve
```

**Manual Execution:**

```bash
cd infra/terraform/environments/localstack
terraform init
terraform plan
terraform apply
```

**Prerequisites:**

```bash
# Install Terraform v1.5+
brew install terraform  # macOS
# or download from https://www.terraform.io/downloads

# Set credentials (LocalStack accepts any values)
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
```

### Terraform Modules

#### IAM Module

Creates roles and policies for networking components.

**Usage:**

```hcl
module "iam" {
  source = "../../modules/iam"
  # ... module inputs
}
```

#### SNS Module

Creates SNS topics and access policies.

**Usage:**

```hcl
module "sns" {
  source      = "../../modules/sns"
  topic_name  = "example-events"
  environment = "dev"
}
```

#### SQS Module

Creates SQS queues with dead-letter queue support.

**Usage:**

```hcl
module "sqs" {
  source     = "../../modules/sqs"
  queue_name = "example-queue"
  # ... additional configuration
}
```

### Adapting for Production

For production deployment:

1. **Move modules to a shared repository:**

   ```hcl
   module "sns" {
     source = "git::https://github.com/org/terraform-modules//sns?ref=v1.0.0"
   }
   ```

2. **Create environment-specific configurations:**

   ```
   infra/terraform/environments/
   ├── dev/
   ├── staging/
   └── production/
   ```

3. **Use remote state:**

   ```hcl
   terraform {
     backend "s3" {
       bucket = "terraform-state"
       key    = "service/dev/terraform.tfstate"
       region = "us-east-1"
     }
   }
   ```

4. **Manage secrets externally:**
   - AWS Secrets Manager
   - HashiCorp Vault
   - Parameter Store

## Docker Images

### Application Image

**Dockerfile:** `infra/build/dockerfiles/Dockerfile`

**Multi-stage build:**

1. **Base**: Go 1.23 with dependencies
2. **Builder**: Compiles the Go binary
3. **Runtime**: Minimal distroless image with binary

**Build:**

```bash
docker build -f infra/build/dockerfiles/Dockerfile -t event-service:latest .
```

### Helper Images

**oapi-codegen helper:** `infra/build/dockerfiles/Dockerfile.oapi-codegen`

Used by `task gen-api` to generate OpenAPI bindings.

## CI/CD Integration

### GitHub Actions

The repository includes workflows in `.github/workflows/`:

- `ci.yml`: Lint, test, build, security scans
- Security: Trivy, gosec, govulncheck, Trufflehog

**Local CI Emulation:**

```bash
task ci  # Runs all GitHub Actions locally via act
```

### Pre-commit Hooks

Automated quality checks before commits:

```bash
# Setup
pre-commit install
pre-commit install --hook-type commit-msg

# Run manually
pre-commit run --all-files
```

**Hooks include:**

- golangci-lint, gofmt, goimports
- buf-lint (protobuf)
- terraform fmt/validate
- hadolint (Dockerfile)
- actionlint (GitHub Actions)
- Secret scanning (gitleaks, trufflehog)

## Security Scanning

### Trivy

Scan for vulnerabilities in dependencies and IaC:

```bash
task scan-security
```

### Gosec

Go security scanner:

```bash
task gosec
```

### Grype

Vulnerability scanner with SBOM support:

```bash
task grype              # Scan codebase
task grype-image IMAGE=event-service:latest  # Scan Docker image
```

### SBOM Generation

Generate Software Bill of Materials:

```bash
task sbom       # Generate SBOM with Syft
task sbom-scan  # Scan SBOM for vulnerabilities
```

## Networking

### Container Network

All services communicate via Docker's default bridge network. Service names act as hostnames:

- `mongo`: MongoDB
- `kafka`: Kafka broker
- `rabbitmq`: RabbitMQ
- `localstack`: AWS services
- `openobserve`: Observability backend

### Port Mapping

**Exposed Ports:**

- 8080: Application HTTP API
- 8085: Protoflow metadata API (if enabled)
- 8081: Mongo Express
- 9000: Kafdrop (Kafka UI)
- 15672: RabbitMQ Management
- 5080: OpenObserve UI
- 5081: OpenObserve OTLP endpoint

**Internal Ports:**

- 27017: MongoDB
- 9092: Kafka
- 5672: RabbitMQ AMQP
- 4566: LocalStack
- 4222: NATS

## Resource Requirements

### Minimum Requirements

- **CPU**: 4 cores
- **RAM**: 8 GB
- **Disk**: 10 GB

### Recommended for Development

- **CPU**: 8 cores
- **RAM**: 16 GB
- **Disk**: 20 GB (with room for logs and data)

### Stack-Specific Memory Usage

- **Kafka**: ~2 GB
- **RabbitMQ**: ~512 MB
- **LocalStack**: ~1 GB
- **MongoDB**: ~512 MB
- **OpenObserve**: ~512 MB

## Troubleshooting

### Services Not Starting

1. Check Docker daemon: `docker ps`
2. Review logs: `docker compose logs <service>`
3. Verify ports are not in use: `lsof -i :<port>`
4. Ensure sufficient resources: `docker stats`

### Network Issues

1. Inspect network: `docker network ls`
2. Check service connectivity: `docker exec <container> ping <service>`
3. Verify DNS resolution: `docker exec <container> nslookup <service>`

### Volume Permissions

If you encounter permission errors:

```bash
# Fix volume permissions
sudo chown -R $USER:$USER _volume_data/

# Or reset volumes
docker compose down -v
rm -rf _volume_data/
task up-<stack>
```

### LocalStack Not Applying Terraform

1. Check Terraform service logs: `docker compose logs terraform`
2. Verify LocalStack health: `curl <http://localhost:4566/_localstack/health`>
3. Manually apply: `docker compose exec terraform terraform apply`

## Best Practices

### Local Development

1. Use `task debug` for hot reload during development
2. Enable Protoflow Web UI for event handler inspection
3. Use `LOGGER=pretty` for readable logs
4. Monitor resource usage with `docker stats`

### Infrastructure as Code

1. Keep Terraform modules reusable and well-documented
2. Use variables for environment-specific values
3. Store state remotely in production
4. Version infrastructure changes alongside code

### Data Management

1. Regularly back up `_volume_data/` if preserving data
2. Use `docker compose down -v` to clean slate
3. Don't commit `_volume_data/` to version control

### Security

1. Never expose development ports to the internet
2. Use strong credentials (even locally)
3. Scan images before deployment: `task grype-image`
4. Keep base images updated

## Related Documentation

- [Configuration Guide](configuration.md) - Environment variables and settings
- [README.md](../README.md) - Quick start and overview
- [CONTRIBUTING.md](../.github/CONTRIBUTING.md) - Development workflow
