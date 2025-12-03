# Contributing to Event-Driven-Service-Example

First off, thanks for taking the time to contribute! 🎉

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Style Guidelines](#style-guidelines)
- [Commit Messages](#commit-messages)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- [Task](https://taskfile.dev/) (task runner)
- [Buf](https://buf.build/) (protobuf tooling)
- Node.js 20+ (for API tooling)

### Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/Event-Driven-Service-Example.git
   cd Event-Driven-Service-Example
   ```

2. **Install dependencies**
   ```bash
   # Go dependencies
   cd src && go mod download && cd ..

   # Node.js tooling
   npm install -g @redocly/cli
   ```

3. **Start infrastructure**
   ```bash
   docker compose -f infra/compose/docker-compose.yml up -d
   ```

4. **Run the service**
   ```bash
   task run
   ```

5. **Run tests**
   ```bash
   task test-go
   ```

## Development Workflow

### Branching Strategy

- `main` - Production-ready code
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates

### Common Tasks

```bash
# Run all linters
task lint

# Run Go linter
task lint-go

# Run tests with coverage
task test-go

# Generate protobuf stubs
task gen-buf

# Generate API code
task gen-api

# Run security scans
task scan-security
```

### Making Changes

1. Create a new branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Run linters and tests
   ```bash
   task lint
   task test-go
   ```

4. Commit your changes (see [Commit Messages](#commit-messages))

5. Push and create a Pull Request

## Pull Request Process

1. **Before submitting:**
   - [ ] Run `task lint` and fix any issues
   - [ ] Run `task test-go` and ensure all tests pass
   - [ ] Update documentation if needed
   - [ ] Add tests for new functionality

2. **PR Title:** Use a descriptive title following conventional commits
   - `feat: add new event handler`
   - `fix: resolve race condition in queue`
   - `docs: update API documentation`

3. **PR Description:** Fill out the PR template completely

4. **Review Process:**
   - At least one approval required
   - All CI checks must pass
   - Address review feedback promptly

5. **Merge:** Maintainers will merge using squash-and-merge

## Style Guidelines

### Go Code

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (handled by golangci-lint)
- Write meaningful comments for exported functions
- Keep functions focused and small
- Handle errors explicitly

### API Design

- Follow OpenAPI 3.1 specification
- Use meaningful HTTP status codes
- Document all endpoints and schemas
- Include examples in API definitions

### Protobuf

- Follow [Buf style guide](https://buf.build/docs/lint/rules)
- Use meaningful field names
- Document messages and fields

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes
- `deps`: Dependency updates

### Examples

```
feat(events): add support for NATS JetStream

fix(api): handle nil pointer in user handler

docs(readme): add deployment instructions

deps(go): bump go.mongodb.org/mongo-driver to v1.13.0
```

## Questions?

Feel free to open a [Discussion](https://github.com/DrBlury/Event-Driven-Service-Example/discussions) if you have questions!
