# Documentation

This directory is the canonical reference for operating and extending the
template. The root [`README.md`](../README.md) is intentionally limited to
newcomer onboarding and the shortest path to a running service; these documents
contain the detailed contracts.

- [Architecture](architecture.md): boundaries, composition roots, lifecycle,
  readiness, and shutdown.
- [Learn Go](learn-go.md): a TypeScript-to-Go guide grounded in this template,
  including boundaries, flows, testing, and feature development.
- [Configuration](configuration.md): variables, precedence, defaults, and
  secret boundaries.
- [Local development](local-development.md): local modes, commands, health
   checks, and dependency addresses.
- [Manual Docker Compose verification](manual-docker-compose-testing.md):
   executable local API, Redis, MongoDB, broker, and worker smoke checks.
- [HTTP API](http-api.md): routes, payloads, status codes, and JSON behavior.
- [Testing](testing.md): test layers, integration prerequisites, and test
  commands.
- [Operations](operations.md): production-oriented deployment and operational
  procedures.
