# Event-Driven Service Example Documentation

Welcome to the documentation for the Event-Driven Service Example. This project demonstrates an event-driven architecture using various technologies and tools. Below is an overview of the documentation structure:

## Documentation Structure

- [Setup Guide](./docs/setup.md): Steps to set up the service, including generating protobuf files and configuring the environment.
- [Infrastructure](./docs/infrastructure.md): Details about the infrastructure setup, including Terraform and Docker Compose configurations.
- [Protobuf and Domain Models](./docs/protobuf.md): Explanation of why Protobuf is used for domain models and its benefits.
- [Pub/Sub Systems](./docs/pubsub.md): Overview of the different pub/sub systems supported by the project.
- [Watermill and OpenTelemetry](./docs/watermill.md): Information about the Watermill library and OpenTelemetry integration.
- [Message Processing Flow](./docs/message_processing.md): Detailed explanation of the message processing flow, including handlers and middleware.

Each section provides in-depth information about the respective topics.

Use the links above to navigate through the documentation.

## Developer Utilities

- `task git:web` opens the repository's default remote in your browser. Override with `task git:web REMOTE=upstream` to target another remote.
- The helper script lives in `scripts/git-web`. Add the repo's `scripts` directory to your `PATH` to invoke it as `git web` (Git picks up executables named `git-*` as custom subcommands).
