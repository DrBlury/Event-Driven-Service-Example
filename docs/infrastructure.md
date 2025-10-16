# Infrastructure

The local infrastructure for this project is managed using Terraform and Docker Compose.

## Terraform

- The `infra/terraform` folder contains the Terraform configuration files.
- Use `main.tf` to define the infrastructure resources required on AWS (LocalStack).
- The infrastructure will be automatically set up when you run the Docker Compose files with AWS configuration.

## Docker Compose

The `compose/` folder contains multiple Docker Compose files for different configurations, each file is designed to override the base Docker Compose configuration.

The `docker-compose.yml` is the base configuration. So do not use this directly.

- `docker-compose.kafka.yml`: For Kafka-based pub/sub.
- `docker-compose.rabbitmq.yml`: For RabbitMQ-based pub/sub.
- `docker-compose.aws.yml`: For AWS-based pub/sub.

Use the appropriate file to start the services:

```bash
docker-compose -f compose/<file-name> up
```
