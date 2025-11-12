# Project Structure and Tooling

This document provides a comprehensive overview of the Event-Driven Service Example project structure, organization, and all tooling used throughout the development lifecycle.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Directory Structure](#directory-structure)
3. [Build and Development Tools](#build-and-development-tools)
4. [Code Generation Tools](#code-generation-tools)
5. [Linting and Code Quality](#linting-and-code-quality)
6. [Testing](#testing)
7. [Container and Deployment Tools](#container-and-deployment-tools)
8. [Configuration Management](#configuration-management)
9. [CI/CD](#cicd)
10. [Development Workflow](#development-workflow)

---

## Project Overview

This is an event-driven microservice built in Go (1.25.1) that demonstrates various pub/sub patterns using:
- **Apache Kafka** for high-throughput message streaming
- **RabbitMQ** for traditional message queuing
- **AWS SNS/SQS** (via LocalStack) for cloud-native messaging

The service features OpenAPI-based REST API, Protocol Buffers for domain models, and comprehensive observability through OpenTelemetry.

---

## Directory Structure

```
Event-Driven-Service-Example/
├── .actrc                      # Act (GitHub Actions local runner) configuration
├── .editorconfig               # Editor configuration for consistent coding styles
├── .gitignore                  # Git ignore patterns
├── README.md                   # Project README with documentation links
├── taskfile.yml                # Task automation definitions (Task runner)
├── go.work                     # Go workspace file (Go 1.25.1)
├── go.work.sum                 # Go workspace checksums
│
├── api/                        # OpenAPI specification (API-first design)
│   ├── .gitignore
│   ├── api.yml                 # Main OpenAPI spec file
│   ├── resources/              # API endpoint definitions
│   │   ├── _index.yml
│   │   ├── info/               # Info endpoints (status, version, OpenAPI docs)
│   │   └── signup/             # Signup endpoint definitions
│   └── schemas/                # API schema definitions
│       ├── _index.yml
│       ├── enum/               # Enumeration types
│       ├── types/              # Data type definitions
│       ├── requests/           # Request body schemas
│       └── errors/             # Error response schemas
│
├── proto/                      # Protocol Buffer definitions
│   └── domain/                 # Domain model protobuf files
│       ├── Info.proto
│       ├── address.proto
│       ├── customer.proto
│       ├── date.proto
│       ├── paymentMean.proto
│       ├── signup.proto
│       └── signupInfo.proto
│
├── buf.yaml                    # Buf (protobuf tool) configuration
├── buf.gen.yaml                # Buf code generation configuration
├── buf.lock                    # Buf dependency lock file
│
├── src/                        # Go source code (main application)
│   ├── go.mod                  # Go module dependencies
│   ├── go.sum                  # Go module checksums
│   ├── .golangci.yml           # GolangCI-Lint configuration
│   ├── main.go                 # Application entry point
│   │
│   ├── internal/               # Private application code
│   │   ├── app/                # Application initialization and config
│   │   ├── database/           # Database layer (MongoDB)
│   │   ├── domain/             # Generated domain models (from protobuf)
│   │   ├── events/             # Event handling and pub/sub logic
│   │   ├── server/             # HTTP server implementation
│   │   │   ├── generated/      # Generated API code (from OpenAPI)
│   │   │   ├── handler/        # Request handlers
│   │   │   │   └── apihandler/ # API handler implementations
│   │   │   │       └── embedded/ # Embedded resources (OpenAPI UI)
│   │   │   └── transform/      # Data transformation utilities
│   │   └── usecase/            # Business logic layer
│   │
│   └── pkg/                    # Reusable packages (could be extracted)
│       ├── logging/            # Logging utilities with OpenTelemetry support
│       ├── metrics/            # Metrics and monitoring
│       ├── router/             # HTTP router configuration
│       └── tracing/            # Distributed tracing setup
│
├── build/                      # Build artifacts and Dockerfiles
│   └── dockerfiles/
│       ├── Dockerfile.app      # Production application Dockerfile
│       ├── Dockerfile.debug    # Debug configuration with Delve
│       └── Dockerfile.protobuf # Protobuf code generation container
│
├── deploy/                     # Deployment configurations
│   ├── compose/                # Docker Compose files
│   │   ├── README.md
│   │   ├── docker-compose.yml         # Base compose (app + mongo + mongo-express)
│   │   ├── docker-compose.kafka.yml   # Kafka + Kafdrop setup
│   │   ├── docker-compose.rabbitmq.yml # RabbitMQ setup
│   │   ├── docker-compose.aws.yml     # LocalStack (AWS SNS/SQS) setup
│   │   └── docker-compose.debug.yml   # Debug mode with Delve debugger
│   │
│   ├── env/                    # Environment configuration
│   │   ├── .gitignore
│   │   └── example/            # Example environment files
│   │       ├── app.env.example
│   │       ├── kafka.env.example
│   │       ├── kafdrop.env.example
│   │       ├── rabbitmq.env.example
│   │       ├── mongo.env.example
│   │       ├── mongo-express.env.example
│   │       ├── localstack.env.example
│   │       └── openobserve.env.example
│   │
│   └── infra/                  # Infrastructure as Code
│       ├── rabbitmq.conf       # RabbitMQ configuration
│       └── terraform/          # Terraform IaC for AWS (LocalStack)
│           ├── .gitignore
│           ├── README.md
│           ├── environments/   # Environment-specific configs
│           │   └── localstack/ # LocalStack environment
│           │       ├── .terraform.lock.hcl
│           │       ├── main.tf
│           │       └── variables.tf
│           └── modules/        # Terraform modules
│               ├── readme.md
│               ├── iam/        # IAM resources
│               ├── sns/        # SNS topic module
│               └── sqs/        # SQS queue module
│
└── docs/                       # Project documentation
    ├── infrastructure.md       # Infrastructure and deployment guide
    ├── message_processing.md   # Event processing flow documentation
    ├── protobuf.md             # Protocol Buffers usage and rationale
    ├── pubsub.md               # Pub/Sub systems overview
    ├── setup.md                # Setup and installation guide
    ├── watermill.md            # Watermill library and OpenTelemetry
    └── project-structure-and-tooling.md  # This file
```

---

## Build and Development Tools

### Go Workspace (go.work)
- **Version**: Go 1.25.1
- **Purpose**: Manages the Go module in `src/` directory
- **Usage**: Automatically used by Go tools when working in the repository

### Task (taskfile.yml)
Task is the primary automation tool, replacing Make with a simpler YAML syntax.

**Key Tasks**:
```bash
# Environment setup
task gen-env-files              # Generate .env files from examples

# Code generation
task gen-buf                    # Generate Go code from protobuf definitions
task gen-api                    # Generate API interface from OpenAPI spec
task gen-api-std                # Generate standard library HTTP server stubs

# Linting
task lint                       # Lint Go code using golangci-lint (in Docker)
task lint-api                   # Lint OpenAPI spec using Redocly

# Running services (different pub/sub backends)
task up-kafka                   # Start with Kafka
task up-rabbitmq                # Start with RabbitMQ
task up-aws                     # Start with AWS (LocalStack)
task down-kafka                 # Stop Kafka setup
task down-rabbitmq              # Stop RabbitMQ setup
task down-aws                   # Stop AWS setup

# Debugging
task debug                      # Start in debug mode with Delve (reads PUBSUB_SYSTEM from .env)

# CI/CD
task ci                         # Run GitHub Actions locally using act
task act-test                   # Run specific GitHub Action job locally

# Utilities
task install-tools              # Install Go tools globally (oapi-codegen, protoc-gen-go, prism-cli)
task scc                        # Show code statistics using scc
```

**Task Configuration**:
- Loads environment from `./env/app.env`
- All tasks use Docker containers to ensure consistent environments
- Silent mode enabled for cleaner output

---

## Code Generation Tools

### 1. Protocol Buffers (Buf)

**Configuration Files**:
- `buf.yaml`: Main Buf configuration
  - Uses v2 configuration format
  - Module path: `proto/`
  - Linting: DEFAULT ruleset
  - Breaking change detection: FILE level
  - Dependencies: `buf.build/bufbuild/protovalidate` for validation

- `buf.gen.yaml`: Code generation config
  - Plugin: `protoc-gen-go` (local)
  - Output: `src/internal/` directory
  - Option: `paths=source_relative`

**Tools**:
- **Buf CLI**: Modern protobuf toolchain
- **protoc-gen-go**: Official Go protobuf compiler plugin
- **protovalidate**: Validation rules for protobuf messages

**Custom Docker Image**: `drblury/protobuf-gen-go`
- Built from: `build/dockerfiles/Dockerfile.protobuf`
- Contains: Buf, protoc-gen-go, and all necessary tools

**Generated Output**: `src/internal/domain/*.pb.go`
- Domain models with validation
- Type-safe enumerations
- Timestamp handling
- All models are Go structs with protobuf tags

### 2. OpenAPI Code Generation (oapi-codegen)

**Tools**:
- **Redocly CLI**: OpenAPI linting and bundling
- **oapi-codegen**: Generates Go server interfaces from OpenAPI spec

**Configuration**:
- `src/internal/server/handler/apihandler/server-std.cfg.yml`: oapi-codegen config
- Generates: Standard library HTTP server interfaces

**Process**:
1. Lint OpenAPI spec with Redocly
2. Bundle modular OpenAPI files into single `bundle.yml`
3. Generate Go interfaces from bundled spec
4. Output: `src/internal/server/generated/api.gen.go`

**Generated Code Includes**:
- Request/response type definitions
- Server interface definitions
- Request validation
- Parameter binding

---

## Linting and Code Quality

### Go Linting (GolangCI-Lint)

**Configuration**: `src/.golangci.yml`

**Enabled Linters**:
- `bodyclose`: Checks HTTP response body is closed
- `funlen`: Function length checker (max 60 lines, 40 statements)
- `gochecknoinits`: Checks for init functions
- `gocyclo`: Cyclomatic complexity checker (max 15)
- `gosec`: Security vulnerability scanner

**Security Rules (gosec)**:
Comprehensive security checks including:
- G101-G110: General security issues
- G201-G204: SQL injection risks
- G301-G307: File permission issues
- G401-G505: Cryptographic vulnerabilities
- G602: Slice bound checks

**Formatters**:
- `gofmt`: Standard Go formatting with simplification
- `goimports`: Import organization

**Exclusions**:
- Generated code (lax rules)
- Test files (goconst disabled)
- Third-party and example code

**Execution**:
```bash
# Via Task (recommended - uses Docker)
task lint

# Direct execution
docker run --rm -v "./src:/code" golangci/golangci-lint:latest \
  /bin/sh -c "cd /code && golangci-lint run"
```

### OpenAPI Linting (Redocly)

**Tool**: Redocly CLI

**Execution**:
```bash
task lint-api

# Or directly:
docker run --rm -v ./api/:/spec redocly/cli lint api.yml
```

**Features**:
- OpenAPI 3.0 validation
- Schema validation
- Best practices enforcement
- Reference resolution

### EditorConfig

**File**: `.editorconfig`

**Global Settings**:
- End of line: LF (Unix-style)
- Insert final newline: true
- Charset: UTF-8
- Trim trailing whitespace: true
- Indent style: spaces

**File-Specific Indentation**:
- YAML/YML: 2 spaces
- JSON: 2 spaces
- JavaScript/TypeScript: 2 spaces
- Markdown: 4 spaces (trailing whitespace preserved)
- Dockerfile: 2 spaces

---

## Testing

### Current State
The project currently has **39 Go source files** but **no test files** (`*_test.go`) are present yet.

### Testing Infrastructure Available

**Go Testing**:
- Standard Go testing framework available via `go test`
- Test files should follow convention: `*_test.go`

**Potential Testing Tools** (based on dependencies):
- Standard library `testing` package
- Could add: testify, gomock, or other testing libraries as needed

**Note**: Testing infrastructure exists but tests are not yet implemented. This is an area for future development.

---

## Container and Deployment Tools

### Docker

**Dockerfiles** (in `build/dockerfiles/`):

#### 1. `Dockerfile.app` (Production)
Multi-stage build:
- **Builder stage**: `golang:alpine`
  - Installs build dependencies (gcc, git, musl-dev, ca-certificates)
  - Supports private Go modules via SSH
  - Sets build metadata via ARGs (VERSION, COMMIT, COMMIT_DATE, BRANCH, USER, NOW)
  - Compiles static binary with CGO disabled
  - Embeds version info in binary via `-ldflags`

- **Runtime stage**: `alpine:3`
  - Minimal runtime image
  - Runs as unprivileged `nobody` user
  - Includes: libxml2, tzdata, libc6-compat, curl
  - Timezone: Europe/Berlin
  - Single binary execution

**Build Arguments**:
- `GOPROXY`: Go module proxy (default: https://proxy.golang.org,direct)
- `VERSION`: Application version (default: dev)
- `COMMIT`: Git commit hash
- `COMMIT_DATE`: Commit timestamp
- `BRANCH`: Git branch name
- `USER`: Build user
- `NOW`: Build timestamp

#### 2. `Dockerfile.debug` (Development)
Adds Delve debugger for remote debugging:
- Based on production build
- Includes Delve Go debugger
- Exposes debug port (typically 2345)
- Allows live debugging in containers

#### 3. `Dockerfile.protobuf` (Code Generation)
Custom image for protobuf code generation:
- Contains Buf CLI
- Contains protoc-gen-go
- Used by `task gen-buf`

### Docker Compose

**Base File**: `deploy/compose/docker-compose.yml`
Never used directly; provides base configuration for:
- **app**: Go application container
  - Uses `golang:1.25.1-alpine` image
  - Mounts source code for live reload
  - Port 8080 exposed
  - Health check on `/info/status`
  - Depends on MongoDB

- **mongo**: MongoDB database
  - Official mongo image
  - Data volume: `_volume_data/mongo`
  - Health check via mongosh
  - Quiet logging to `/dev/null`

- **mongo-express**: MongoDB web UI
  - Port 8081 exposed
  - For database inspection and management

**Override Files** (extend base config):

#### `docker-compose.kafka.yml`
Adds Kafka infrastructure:
- **kafka**: Apache Kafka (latest)
  - Port 9092 exposed
  - Kraft mode (no Zookeeper needed)
  - Data volume: `_volume_data/kafka`
  - Health check on Kafka process

- **kafdrop**: Kafka web UI
  - Port 9000 exposed
  - Topic inspection and message browsing
  - Version: 4.2.0

**App environment**: `PUBSUB_SYSTEM=kafka`

#### `docker-compose.rabbitmq.yml`
Adds RabbitMQ infrastructure:
- **rabbitmq**: RabbitMQ with management plugin
  - Ports: 5672 (AMQP), 15672 (management UI)
  - Custom config: `deploy/infra/rabbitmq.conf`
  - Data volume: `_volume_data/rabbitmq`

**App environment**: `PUBSUB_SYSTEM=rabbitmq`

#### `docker-compose.aws.yml`
Adds LocalStack for AWS services:
- **localstack**: AWS service emulation
  - SNS and SQS services enabled
  - Port 4566 exposed (unified endpoint)
  - Data volume: `_volume_data/localstack`
  - Terraform applies infrastructure on startup

- **openobserve**: Observability platform
  - Ports: 5080 (HTTP), 5081 (gRPC)
  - Logs, traces, and metrics collection
  - Web UI for observability data

**App environment**: `PUBSUB_SYSTEM=aws`

#### `docker-compose.debug.yml`
Debugging configuration:
- Replaces `go run` with Delve debugger
- Exposes debug port for IDE attachment
- Works with any pub/sub backend
- Command: `dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient`

**Usage Pattern**:
```bash
# Always combine base + backend + optional debug
docker compose -f deploy/compose/docker-compose.yml \
               -f deploy/compose/docker-compose.kafka.yml \
               up --build

# With debugging:
docker compose -f deploy/compose/docker-compose.yml \
               -f deploy/compose/docker-compose.kafka.yml \
               -f deploy/compose/docker-compose.debug.yml \
               up --build
```

### Infrastructure as Code (Terraform)

**Location**: `deploy/infra/terraform/`

**Purpose**: Provisions AWS resources in LocalStack for local development

**Structure**:
- `environments/localstack/`: Environment configuration
  - `main.tf`: Main infrastructure definition
  - `variables.tf`: Input variables
  - `.terraform.lock.hcl`: Provider version lock

- `modules/`: Reusable Terraform modules
  - `iam/`: IAM roles and policies
  - `sns/`: SNS topic creation
  - `sqs/`: SQS queue creation

**Resources Created**:
- SNS topics for event publishing
- SQS queues for event consumption
- Dead letter queues (DLQ)
- IAM roles and policies for least-privilege access

**Execution**:
Automatically applied when using `docker-compose.aws.yml` via init script in LocalStack container.

---

## Configuration Management

### Environment Files

**Location**: `deploy/env/example/` (templates)

**Files**:

#### `app.env.example` (Main Application)
Key configurations:
- **Server**: Port (8080), timeout (5m), base URL
- **Logging**: Level (debug), format (prettyjson/otel-and-console)
- **MongoDB**: Connection URL, database name, credentials
- **Pub/Sub Queues**: 
  - `QUEUE`: Main message queue
  - `QUEUE_PROCESSED`: Processed messages queue
  - `QUEUE_SIGNUP`: Signup events queue
  - `QUEUE_SIGNUP_PROCESSABLE`: Processable signup queue
  - `POISON_QUEUE`: Dead letter queue
- **CORS**: Origins, methods, headers, credentials
- **OpenTelemetry**:
  - Protocol: gRPC
  - Exporters: OTLP for logs, traces, metrics
  - Endpoints: OpenObserve (http://openobserve:5081)
  - Headers: Basic auth and stream configuration
  - Metrics: Prometheus producer enabled

**Security Features**:
- `APP_SERVER_QUIETDOWN_ROUTES`: Suppress logs for health checks
- `APP_SERVER_HIDE_HEADERS`: Redact sensitive headers (Authorization)

#### `kafka.env.example` (Kafka)
- Kafka server configuration
- Listener settings
- Log directory

#### `kafdrop.env.example` (Kafka UI)
- Kafka broker connection
- UI port configuration

#### `rabbitmq.env.example` (RabbitMQ)
- Username and password
- Virtual host configuration
- Management plugin settings

#### `mongo.env.example` (MongoDB)
- Root username and password
- Database initialization

#### `mongo-express.env.example` (MongoDB UI)
- Admin credentials
- MongoDB connection details

#### `localstack.env.example` (LocalStack)
- AWS region (us-east-1)
- Services to enable (sns, sqs)
- Edge port (4566)

#### `openobserve.env.example` (OpenObserve)
- Root user credentials
- Organization settings
- Data storage path

**Generation**:
```bash
task gen-env-files
```
Copies all `.example` files to working `.env` files in `deploy/env/`.

### Go Module Configuration

#### `go.work` (Workspace)
- Go version: 1.25.1
- Single module: `./src`
- Enables workspace mode for potential future multi-module setup

#### `src/go.mod` (Module)
**Module Name**: `drblury/event-driven-service`

**Key Dependencies**:

**Pub/Sub Libraries**:
- `github.com/ThreeDotsLabs/watermill`: v1.5.1 - Event streaming library
- `github.com/ThreeDotsLabs/watermill-amqp/v3`: v3.0.2 - RabbitMQ adapter
- `github.com/ThreeDotsLabs/watermill-kafka/v3`: v3.1.2 - Kafka adapter
- `github.com/ThreeDotsLabs/watermill-aws`: v1.0.1 - AWS SNS/SQS adapter

**AWS SDK**:
- `github.com/aws/aws-sdk-go-v2`: v1.38.1
- `github.com/aws/aws-sdk-go-v2/service/sns`: v1.37.2
- `github.com/aws/aws-sdk-go-v2/service/sqs`: v1.42.1

**HTTP and API**:
- `github.com/gorilla/mux`: v1.8.1 - HTTP router
- `github.com/getkin/kin-openapi`: v0.133.0 - OpenAPI handling
- `github.com/oapi-codegen/oapi-codegen/v2`: v2.5.0 - API code generator
- `github.com/oapi-codegen/nethttp-middleware`: v1.1.2 - HTTP middleware

**Protocol Buffers**:
- `google.golang.org/protobuf`: v1.36.9 - Official protobuf runtime
- `buf.build/go/protovalidate`: v1.0.0 - Protobuf validation
- `buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go`: Generated validation types

**Database**:
- `go.mongodb.org/mongo-driver`: v1.17.4 - MongoDB driver

**Observability (OpenTelemetry)**:
- `go.opentelemetry.io/otel`: v1.38.0 - Core API
- `go.opentelemetry.io/otel/sdk`: v1.38.0 - SDK
- `go.opentelemetry.io/otel/sdk/log`: v0.14.0 - Logging
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`: v0.63.0 - HTTP instrumentation
- `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc`: v0.14.0 - Log exporter
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`: v1.38.0 - Trace exporter
- `go.opentelemetry.io/contrib/bridges/otelslog`: v0.13.0 - slog integration

**Logging and Configuration**:
- `github.com/samber/lo`: v1.51.0 - Utility functions
- `github.com/samber/slog-multi`: v1.5.0 - Multi-handler slog
- `github.com/spf13/viper`: v1.21.0 - Configuration management
- `github.com/google/uuid`: v1.6.0 - UUID generation

**Kafka Client**:
- `github.com/IBM/sarama`: v1.43.3 - Kafka client library
- `github.com/dnwe/otelsarama`: v0.0.0-20240308230250-9388d9d40bc0 - Kafka tracing

**RabbitMQ Client**:
- `github.com/rabbitmq/amqp091-go`: v1.10.0 - AMQP 0.9.1 client

**Tools** (declared in go.mod):
- `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen`: API code generator

---

## CI/CD

### GitHub Actions (Local Execution with Act)

**Configuration**: `.actrc`
```
--container-architecture linux/amd64
-P ubuntu-latest=catthehacker/ubuntu:act-latest
```

**Purpose**: Allows running GitHub Actions workflows locally using [nektos/act](https://github.com/nektos/act)

**Usage**:
```bash
# Run all workflows
task ci

# Run specific job
JOB=<job-name> task act-test
```

**Benefits**:
- Test CI pipelines locally before pushing
- Faster feedback loop
- No GitHub Actions minutes consumed during testing

**Note**: The `.github/` directory is not present in this clone, but the infrastructure is ready for CI/CD workflows.

---

## Development Workflow

### Initial Setup

1. **Clone Repository**
   ```bash
   git clone <repository-url>
   cd Event-Driven-Service-Example
   ```

2. **Generate Environment Files**
   ```bash
   task gen-env-files
   ```
   Edit `deploy/env/*.env` files as needed.

3. **Choose Your Pub/Sub System**
   Edit `deploy/env/app.env` and set:
   ```
   PUBSUB_SYSTEM=kafka    # or rabbitmq, or aws
   ```

4. **Generate Code**
   ```bash
   # Generate domain models from protobuf
   task gen-buf
   
   # Generate API interfaces from OpenAPI
   task gen-api
   ```

### Daily Development

1. **Start Infrastructure**
   ```bash
   # Choose one based on PUBSUB_SYSTEM in app.env
   task up-kafka
   # or
   task up-rabbitmq
   # or
   task up-aws
   ```

2. **Development with Hot Reload**
   The Docker setup mounts source code as a volume, so changes are reflected immediately:
   - Edit code in `src/`
   - Container restarts on crashes (restart: on-failure)
   - View logs: `docker compose logs -f app`

3. **Debugging**
   ```bash
   task debug
   ```
   Then attach your IDE debugger to `localhost:2345`:
   - **VS Code**: Use launch configuration with `dlv` attach mode
   - **GoLand**: Remote debug configuration

4. **Linting**
   ```bash
   # Lint Go code
   task lint
   
   # Lint OpenAPI spec
   task lint-api
   ```

5. **Regenerate Code After Schema Changes**
   ```bash
   # After editing .proto files
   task gen-buf
   
   # After editing api.yml
   task gen-api
   ```

6. **Stop Infrastructure**
   ```bash
   task down-kafka
   # or
   task down-rabbitmq
   # or
   task down-aws
   ```

### Code Statistics

Use `scc` (Source Code Counter) to view project metrics:
```bash
task scc
```

### Accessing Services

**Application**:
- REST API: http://localhost:8080
- Status: http://localhost:8080/info/status
- Version: http://localhost:8080/info/version
- OpenAPI JSON: http://localhost:8080/info/openapi.json
- OpenAPI UI: http://localhost:8080/info/openapi.html

**Database UIs**:
- Mongo Express: http://localhost:8081

**Pub/Sub UIs**:
- Kafdrop (Kafka): http://localhost:9000
- RabbitMQ Management: http://localhost:15672
- LocalStack: http://localhost:4566

**Observability**:
- OpenObserve: http://localhost:5080

### Best Practices

1. **Use Task Commands**: Always use `task` commands instead of running Docker/tools directly
2. **Keep Environment Files Updated**: After pulling, regenerate environment files if examples changed
3. **Lint Before Committing**: Run `task lint` and `task lint-api` before committing
4. **Regenerate After Schema Changes**: Always regenerate code after modifying protobuf or OpenAPI schemas
5. **Use Correct Pub/Sub Backend**: Ensure `PUBSUB_SYSTEM` in `.env` matches the Docker Compose stack you're running
6. **Volume Data**: The `_volume_data/` directory contains persistent data for MongoDB, Kafka, etc. Add to `.gitignore`
7. **Health Checks**: Wait for health checks to pass before testing (check `docker compose ps`)

---

## Dependencies and Third-Party Tools

### Runtime Dependencies (via Docker)
All development tools run in Docker containers, no local installation required (except Docker and Task):

- **golang:1.25.1-alpine**: Go runtime
- **alpine:3**: Production container base
- **golangci/golangci-lint:latest**: Go linting
- **redocly/cli**: OpenAPI tooling
- **apache/kafka:latest**: Kafka broker
- **obsidiandynamics/kafdrop:4.2.0**: Kafka UI
- **rabbitmq**: RabbitMQ with management
- **mongo**: MongoDB
- **mongo-express**: MongoDB UI
- **localstack/localstack**: AWS services emulation
- **openobserve**: Observability platform
- **ghcr.io/lhoupert/scc:master**: Code statistics

### Optional Local Tools
Can be installed globally for convenience (via `task install-tools`):
- **oapi-codegen**: OpenAPI code generator
- **protoc-gen-go**: Protobuf compiler
- **prism-cli**: OpenAPI mock server

### Build Dependencies (in Docker)
- **gcc**: C compiler for CGO
- **musl-dev**: C standard library for Alpine
- **git**: Source control
- **ca-certificates**: SSL/TLS certificates
- **openssh-client**: SSH for private Go modules

---

## Summary

This project demonstrates a well-structured, modern Go microservice with:

✅ **Clean Architecture**: Separation of concerns (internal/pkg, domain/usecase/handler)  
✅ **API-First Design**: OpenAPI specification drives API development  
✅ **Type-Safe Domain Models**: Protocol Buffers with validation  
✅ **Multiple Pub/Sub Backends**: Kafka, RabbitMQ, AWS (pluggable via Watermill)  
✅ **Containerized Everything**: Consistent environments via Docker  
✅ **Comprehensive Observability**: OpenTelemetry for logs, traces, metrics  
✅ **Infrastructure as Code**: Terraform for AWS resources  
✅ **Automated Workflows**: Task runner for all common operations  
✅ **Code Quality**: Linting and security scanning built-in  
✅ **Developer Experience**: Hot reload, debugging, UI tools for all services  

The tooling supports the entire development lifecycle from code generation to deployment, with strong emphasis on consistency, automation, and developer productivity.
