Docker Compose layout and usage
================================

This folder uses a base Compose file with small, mode-specific override
files to avoid duplication and follow the Docker multiple-compose-files
recommendation.

Files
-----

- docker-compose.yml
  - Base file containing the shared services: app, mongo, mongo-express,
    openobserve.

- docker-compose.kafka.yml
  - Kafka mode: adds `kafka` and `kafdrop`, and makes `app` use
    PUBSUB_SYSTEM=kafka.

- docker-compose.rabbitmq.yml
  - RabbitMQ mode: adds `rabbitmq` and makes `app` use
    PUBSUB_SYSTEM=rabbitmq.

- docker-compose.aws.yml
  - AWS/localstack mode: adds `localstack` and `terraform`, and sets
    PUBSUB_SYSTEM=aws for `app`.

- docker-compose.nats.yml
  - NATS mode: adds `nats` and sets PUBSUB_SYSTEM=nats for `app`.

- docker-compose.http.yml
  - HTTP mode: adds `mockserver` and sets PUBSUB_SYSTEM=http for `app`.

- docker-compose.io.yml
  - IO mode: sets PUBSUB_SYSTEM=io for `app`.

Examples
--------

Start Kafka mode (base + kafka override):

```bash
docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.kafka.yml up
```

Start RabbitMQ mode (base + rabbitmq override):

```bash
docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.rabbitmq.yml up
```

Start AWS/localstack mode (base + aws override):

```bash
docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.aws.yml up
```

Start NATS mode:

```bash
docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.nats.yml up
```

Start HTTP mode:

```bash
docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.http.yml up
```

Start IO mode:

```bash
docker compose -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.io.yml up
```

Notes
-----

- If you need to override settings for development (ports, volumes, env) you
  can add a `docker-compose.override.yml` or pass an additional `-f` file.
- The `app` service's PUBSUB_SYSTEM is set by the mode-specific files so the
  code can switch between Kafka, RabbitMQ or AWS (localstack) without
  duplicating the full `app` service definition.

Volume & infra changes
-----------------------

- Persistent data for services is now consolidated under the repository root
  `_volume_data/` folder. Subfolders include `mongo`, `kafka`, `rabbitmq`,
  `localstack`, and `openobserve`. Compose files mount from
  these new paths (e.g. `../../_volume_data/mongo`).
