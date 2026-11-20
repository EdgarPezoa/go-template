# Operations

This page is the production-oriented operational index. Detailed local
commands, endpoint contracts, configuration variables, and test prerequisites
are maintained in the linked references rather than duplicated here.

## Production procedure

1. Supply configuration through the deployment environment or secret manager;
   do not deploy `.env` or local emulator credentials.
   See [Configuration](configuration.md).
2. Provide managed Redis and Azure Service Bus. Both are mandatory API
   dependencies; the process fails startup if either cannot initialize.
3. Expose `/health` for process liveness and `/ready` for startup readiness.
   Readiness reflects successful initialization, not continuous dependency
   probing. See [HTTP API](http-api.md).
4. Deliver SIGTERM and allow the API's 10-second HTTP shutdown window and the
   queue/topic listener cancellation to complete. See [Architecture](architecture.md).
5. Review logs for configuration, initialization, listener, and shutdown
   failures before replacing an instance.

## Local operational reference

Use [Local development](local-development.md) for local modes, dependency
setup, addresses, and start/stop commands. For the complete local manual
runtime check, use [Manual Docker Compose verification](manual-docker-compose-testing.md).

## Runbooks and checks

- Configuration and secret rotation: [Configuration](configuration.md)
- Endpoint verification: [HTTP API](http-api.md)
- Dependency-backed verification: [Testing](testing.md)
- Startup, health, and shutdown behavior: [Architecture](architecture.md)
