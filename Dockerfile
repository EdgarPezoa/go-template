FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -o /out/worker-queue ./cmd/workers/queue && \
     CGO_ENABLED=0 go build -o /out/worker-topic ./cmd/workers/topic && \
     CGO_ENABLED=0 go build -o /out/worker-example ./cmd/workers/example

FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install --no-install-recommends --yes ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --uid 10001 app

USER app
WORKDIR /app

FROM runtime AS api

COPY --from=builder --chown=app:app /out/api /app/api
EXPOSE 3000
ENTRYPOINT ["/app/api"]

FROM runtime AS worker-queue
COPY --from=builder --chown=app:app /out/worker-queue /app/worker-queue
ENTRYPOINT ["/app/worker-queue"]

FROM runtime AS worker-topic
COPY --from=builder --chown=app:app /out/worker-topic /app/worker-topic
ENTRYPOINT ["/app/worker-topic"]

FROM runtime AS worker-example
COPY --from=builder --chown=app:app /out/worker-example /app/worker-example
ENTRYPOINT ["/app/worker-example"]
