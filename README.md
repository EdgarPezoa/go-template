# Go microservice template

A Go 1.25 service template with hexagonal boundaries, an HTTP API, mandatory
Redis cache, MongoDB database, and Azure Service Bus API dependencies, and queue/topic workers.

## Runtime flow

Each command loads configuration, builds its composition root, initializes only
the configured integrations, and then starts its runtime:

- `cmd/api` exposes health and enabled feature routes.
- `cmd/workers/queue` listens to the configured Service Bus queue,
  `cmd/workers/topic` listens to the configured topic subscription, and
  `cmd/workers/example` runs the standalone example worker.

See [`docs/README.md`](docs/README.md) for the complete documentation map.

## First five minutes

```sh
cp .env.example .env
go run ./cmd/api
curl http://localhost:3000/health
```

The API requires reachable Redis, MongoDB, and Service Bus dependencies before it starts.
The queue and topic contracts are `POST /service-bus/queue/messages` and
`POST /service-bus/topic/messages`, each accepting `{"message":"..."}` and
returning `202 {"accepted":true}`. Cache responses are
`{"key":"...","value":"..."}` and the API listens on port `3000` by default.
Database examples use `POST /database/examples` with `{"value":"..."}` and
`GET /database/examples/{id}`; invalid MongoDB ObjectIDs return `400` and
missing documents return `404`.

For local environment setup, dependency options, and health checks, see
[`docs/local-development.md`](docs/local-development.md).

Compose publishes API port `3000`; Redis and Service Bus emulator ports are
published only where declared in `compose.yml`. The broker integration tests
use static configured queue/topic entities and unique message content per test.

## Common direct-host commands

```sh
go run ./cmd/api
go run ./cmd/workers/queue
go run ./cmd/workers/topic
go run ./cmd/workers/example
```

## Adding a variable

Add the typed field and validation/default tags in `internal/config/config.go`,
document it in `.env.example` and [`docs/configuration.md`](docs/configuration.md),
then update the relevant environment configuration and consumers.
