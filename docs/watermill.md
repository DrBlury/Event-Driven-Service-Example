# Watermill and OpenTelemetry

## Watermill Library

- Watermill is used for building event-driven applications.
- It provides abstractions for pub/sub systems, making it easier to switch between Kafka, AWS (Localstack with SQS/SNS) and RabbitMQ.
- Handlers are defined to process incoming messages and perform business logic.
- It allows easy integration with middleware for logging, tracing, and error handling.

## OpenTelemetry

- OpenTelemetry is used for logging and tracing.
- Middlewares and loggers are integrated to provide detailed insights into the application's behavior.
- Logs are sent to a centralized logging system for analysis.
