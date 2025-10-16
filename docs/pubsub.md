# Pub/Sub Systems

This project supports multiple pub/sub systems, including Kafka and RabbitMQ.

## Kafka

- Kafka is used for high-throughput, distributed messaging.
- Configuration file: `env/kafka.env`
- Docker Compose file: `compose/docker-compose.kafka.yml`

## RabbitMQ

- RabbitMQ is used for lightweight, reliable messaging.
- Configuration file: `env/rabbitmq.env`
- Docker Compose file: `compose/docker-compose.rabbitmq.yml`

## AWS (LocalStack)

- AWS SNS/SQS can be used for pub/sub messaging.
- Configuration file: `env/localstack.env`
- Docker Compose file: `compose/docker-compose.aws.yml`

## Switching Between Systems

- Use the appropriate Docker Compose file to switch between Kafka and RabbitMQ.
- Ensure the environment variables are correctly set in the `.env` files.
