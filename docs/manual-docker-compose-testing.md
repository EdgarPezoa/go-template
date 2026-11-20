# Manual Docker Compose verification

This guide is an executable smoke test for the containerized template. It
checks the API process, startup readiness, Redis and MongoDB operations,
and the Service Bus emulator and workers. It is a manual runtime
verification guide, not a replacement for automated tests or production
deployment checks.

## Prerequisites and boundaries

- Docker Engine and the Docker Compose plugin are installed and available as
  `docker compose`.
- Run every command from the repository root.
- Compose publishes host ports `3000`, `5300`, and `5672`; Redis `6379` and
  MongoDB `27017` are not published by `compose.yml`. Host-side Redis commands are applicable only
  when Redis is started separately.
- The local Service Bus emulator and SQL Server images are default
  infrastructure. They are
  local-only and does not verify managed Azure Service Bus behavior.
- The commands below use `curl --fail-with-body` so an unexpected HTTP status
  stops a shell block. The response checks make the expected status explicit.

## Default API, Redis, MongoDB, Service Bus emulator, and SQL Server infrastructure

Start the API, Redis, MongoDB, Service Bus emulator, and SQL Server:

```sh
docker compose up --build -d
```

Wait for API process health and startup readiness before checking MongoDB:

```sh
until curl --fail-with-body --silent http://localhost:3000/health; do sleep 1; done
until curl --fail-with-body --silent http://localhost:3000/ready; do sleep 1; done
```

Then create and retrieve a MongoDB example:

```sh
curl --fail-with-body --silent --show-error \
  -X POST http://localhost:3000/database/examples \
  -H 'Content-Type: application/json' \
  --data '{"value":"manual database verification"}'
# Copy the returned id, then:
curl --fail-with-body --silent --show-error \
  http://localhost:3000/database/examples/<id>
```

The POST returns `201`; the GET returns `200`. A malformed ObjectID returns
`400`, and a valid missing ObjectID returns `404`. MongoDB data remains after
`docker compose down`; use `docker compose down -v` to remove the named volume.

Expected responses are `200 {"status":"ok"}` from `/health` and
`200 {"status":"ready"}` from `/ready`.

Verify a Redis string write, read, overwrite, and missing-key response:

```sh
curl --fail-with-body --silent --show-error \
  -X PUT http://localhost:3000/cache/manual-string \
  -H 'Content-Type: application/json' \
  --data '{"value":"first"}'
curl --fail-with-body --silent --show-error \
  http://localhost:3000/cache/manual-string
curl --fail-with-body --silent --show-error \
  -X PUT http://localhost:3000/cache/manual-string \
  -H 'Content-Type: application/json' \
  --data '{"value":"overwritten"}'
curl --fail-with-body --silent --show-error \
  http://localhost:3000/cache/manual-string
curl --include --silent --show-error \
  http://localhost:3000/cache/manual-string-missing
```

The two `PUT` requests must return `201`; reads must return `200`, and the
overwritten read must contain `"value":"overwritten"`. The missing-key request
must return `404`.

Verify the equivalent hash operations and an invalid request:

```sh
curl --fail-with-body --silent --show-error \
  -X PUT http://localhost:3000/cache/hash/manual-hash \
  -H 'Content-Type: application/json' \
  --data '{"value":"first"}'
curl --fail-with-body --silent --show-error \
  http://localhost:3000/cache/hash/manual-hash
curl --fail-with-body --silent --show-error \
  -X PUT http://localhost:3000/cache/hash/manual-hash \
  -H 'Content-Type: application/json' \
  --data '{"value":"overwritten"}'
curl --fail-with-body --silent --show-error \
  http://localhost:3000/cache/hash/manual-hash
curl --include --silent --show-error \
  http://localhost:3000/cache/hash/manual-hash-missing
curl --include --silent --show-error \
  -X PUT http://localhost:3000/cache/hash/manual-hash \
  -H 'Content-Type: application/json' \
  --data '{"unexpected":"field"}'
```

Expect `201` for writes, `200` for reads, `404` for the missing hash, and
`400` for the invalid request. Stop this mode with:

```sh
docker compose down
```

## Long-running worker profile

The default stack already provides the API, Redis, MongoDB, Service Bus emulator,
and SQL Server.
Start both long-running workers with the `broker` profile:

```sh
docker compose --profile broker up --build -d
```

The emulator is dependent on SQL Server and may take longer than the API.
Before publishing, wait for the emulator logs to contain the exact readiness
text `Successfully Up!`:

```sh
until docker compose --profile broker logs servicebus-emulator 2>&1 | grep -Fq 'Successfully Up!'; do
  sleep 2
done
```

Then verify API health and readiness:

```sh
curl --fail-with-body --silent http://localhost:3000/health
curl --fail-with-body --silent http://localhost:3000/ready
```

Both requests must return `200`; the bodies must report `ok` and `ready`
respectively. Publish one message to each static configured queue/topic entity.
Use unique message content when repeating the check. Both responses must be
`202 {"accepted":true}`:

```sh
curl --fail-with-body --silent --show-error \
  -X POST http://localhost:3000/service-bus/queue/messages \
  -H 'Content-Type: application/json' \
  --data '{"message":"manual queue verification"}'
curl --fail-with-body --silent --show-error \
  -X POST http://localhost:3000/service-bus/topic/messages \
  -H 'Content-Type: application/json' \
  --data '{"message":"manual topic verification"}'
```

Inspect each worker separately and require evidence of the worker's
`Worker working` log after publishing:

```sh
docker compose --profile broker logs --tail=100 worker-queue
docker compose --profile broker logs --tail=100 worker-topic
```

The logs must show `Worker working` for both `worker-queue` and
`worker-topic`. A `202` only confirms acceptance by the API; it does not prove
that a worker consumed the message. Stop broker mode with:

```sh
docker compose --profile broker down
```

## One-shot worker example

Run the non-broker example worker with its dedicated `worker` profile:

```sh
docker compose --profile worker run --rm worker-example
```

The command must print `Worker working` and exit with status `0`. Clean up any
remaining profile resources with:

```sh
docker compose --profile worker down
```

## Interpreting startup races and failures

- Startup initializes mandatory Redis, Service Bus, and MongoDB dependencies in
  order; a dependency failure stops the API and prevents successful readiness.
  Keep polling `/ready` during startup, before exercising MongoDB endpoints.
- Do not publish to Service Bus until the emulator logs contain
  `Successfully Up!`. Starting the containers successfully is not sufficient.
- A broker `/ready` failure usually means the emulator, SQL Server, or
  connection configuration has not initialized. Inspect the API and emulator
  logs before retrying.
- A `202` without `Worker working` means acceptance succeeded but consumption
  was not demonstrated. Check the profile-aware worker logs and confirm the
  relevant worker is running.
- A `404` for a missing cache key and `400` for the deliberately invalid JSON
  shape are expected results, not startup failures.
- If a command fails because a prior run left resources behind, run the exact
  cleanup command for that mode, then start again.
- `docker compose down` stops the stack while retaining the MongoDB named
  volume; use `docker compose down -v` only when intentionally removing
  persisted MongoDB data. Resource cleanup follows the dependency lifecycle
  rather than deleting the volume by default.
