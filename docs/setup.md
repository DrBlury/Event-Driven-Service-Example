# Setup Guide

This guide explains how to set up the Event-Driven Service Example.

## Prerequisites

- Docker and Docker Compose installed.
- Go installed.
- Taskfile installed for task automation.

## Steps

1. **Generate Protobuf based domain models**:
   This command generates the domain models using the protobuf files.

   ```bash
   task gen-buf
   ```

   The generated files will be located in the `src/internal/domain` folder and will contain the validation logic as per the protobuf definitions.

2. **Embedded API and API Interface**:
   - The `api` folder contains the OpenAPI specifications.
   - Use the `api.yml` file and the linked folders for resources and schemas to define the API interface.
   - Generate the bundled OpenAPI specification and API interface GO code using:

     ```bash
     task gen-api
     ```

3. **Configuration Files**:
   - `env/`: Contains environment configuration files for different services.
     - `app.env`: Main application configuration.
     - `kafka.env`: Kafka-specific configuration.
     - `rabbitmq.env`: RabbitMQ-specific configuration.
     - `localstack.env`: AWS-specific configuration.
     - `mongo.env`: MongoDB-specific configuration.
     - `mongo-express.env`: MongoExpress-specific configuration.
     - `openobserve.env`: OpenObserve-specific configuration.
   - `infra/terraform/main.tf`: Terraform setup code.
   - `infra/terraform/variables.tf`: Terraform variables for infrastructure setup.

4. **Run the Service**:
   Use the appropriate Docker Compose file based on the pub/sub system:
   - Kafka: `docker-compose.kafka.yml`
   - RabbitMQ: `docker-compose.rabbitmq.yml`
   - AWS: `docker-compose.aws.yml`
   These files will override the base docker compose.
   - Default: `docker-compose.yml`

  Check the taskfile for predefined tasks to start the services for different pub/sub systems.
  And have the application configured automatically to work with the selected pub/sub system.
