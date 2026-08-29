# API process lifecycle

The API is one supervised process. HTTP, server-sent events (SSE), the Redis
Streams consumer, tracing, and process-owned dependencies share a single root
context and one explicit shutdown path.

## Why this exists

Previously, the API started the SSE hub and stream consumer in untracked
goroutines. A startup error could be logged while HTTP continued serving, the
readiness route checked only PostgreSQL, and graceful shutdown waited only for
the HTTP server. That made a partially working deployment look healthy.

The supervisor in `internal/bootstrap/api/lifecycle.go` now answers four
questions for every long-lived component:

1. who starts it;
2. which context cancels it;
3. who waits for it;
4. where its terminal error is returned.

## Startup order

The API starts in this order:

1. parse and validate configuration and security keys;
2. open and ping the shared PostgreSQL pool;
3. open and ping Redis;
4. construct task, provider, route, and tracing dependencies;
5. initialize supervised components;
   - the Redis Streams consumer creates or verifies its consumer group;
   - the SSE hub validates its lifecycle dependencies;
6. bind the HTTP listener;
7. start the consumer, SSE hub, and HTTP server;
8. mark the process `ready`.

Consumer-group creation and listener binding happen before readiness. A failure
there is fatal and the listener is never advertised as ready.

## Health contract

`GET /liveness` preserves the existing diagnostics contract. The exact host and
runtime values vary by deployment:

```json
{
  "data": {
    "status": "ok",
    "hostname": "api-host",
    "GOMAXPROCS": 2
  }
}
```

`GET /readiness` combines process phase with bounded PostgreSQL and Redis pings:

```json
{
  "data": {
    "status": "ok",
    "phase": "ready",
    "checks": {
      "postgres": "ready",
      "redis": "ready"
    }
  }
}
```

The phases are:

| Phase      | Meaning                                                             | HTTP status |
| ---------- | ------------------------------------------------------------------- | ----------- |
| `starting` | Startup validation or component initialization is incomplete.       | 503         |
| `ready`    | The supervisor is accepting traffic; dependency checks decide.      | 200 or 503  |
| `draining` | Shutdown began and the instance must receive no new traffic.        | 503         |
| `failed`   | A startup or supervised runtime capability terminated unexpectedly. | 503         |

Dependency error strings are logged internally and never returned by the
health endpoint. Liveness retains its existing hostname, Kubernetes metadata,
and `GOMAXPROCS` fields for compatibility; none of those values participate in
the readiness decision.

## Shutdown order

On root-context cancellation or a terminal component error, the supervisor:

1. marks readiness `draining` for a normal signal, or `failed` for an error;
2. cancels the shared runtime context;
3. gracefully shuts down HTTP within `APP_API_SHUTDOWN_TIMEOUT`;
4. waits for HTTP, SSE, and the stream consumer within the same budget;
5. flushes OpenTelemetry within `APP_API_TELEMETRY_SHUTDOWN_TIMEOUT`;
6. closes the tasks client, Redis client, and shared PostgreSQL pool once.

SSE uses normal `net/http` streaming instead of hijacking sockets. The root
context therefore reaches active stream handlers, while the hub cancels and
waits for all per-client Redis Pub/Sub listeners.

## Configuration

| Variable                             | Default | Purpose                                              |
| ------------------------------------ | ------- | ---------------------------------------------------- |
| `APP_API_READ_HEADER_TIMEOUT`        | `10s`   | Bounds request-header reads against slow clients.    |
| `APP_API_READ_TIMEOUT`               | `5m`    | Bounds ordinary request reads.                       |
| `APP_API_WRITE_TIMEOUT`              | `5m`    | Bounds ordinary response writes; SSE clears its own. |
| `APP_API_IDLE_TIMEOUT`               | `60s`   | Bounds idle keep-alive connections.                  |
| `APP_API_SHUTDOWN_TIMEOUT`           | `30s`   | Shared HTTP/component drain budget.                  |
| `APP_API_READINESS_CHECK_TIMEOUT`    | `2s`    | Shared deadline for required dependency probes.      |
| `APP_API_TELEMETRY_SHUTDOWN_TIMEOUT` | `5s`    | Trace flush deadline after traffic stops.            |

The deployment platform's termination grace period must be longer than the API
shutdown and telemetry budgets combined.

Release builds pass the immutable commit identifier through the shared Docker
build argument `BUILD_VERSION`. Both API and worker binaries receive it via Go
linker flags; the API records it in startup logs and OpenTelemetry's service
version, and the worker records it in startup logs.

## Failure and recovery policy

- Redis consumer-group creation failure is fatal at startup.
- A missing consumer group at runtime is terminal; it cannot spin in a fast
  `NOGROUP` loop.
- Redis read errors other than `NOGROUP` retry with a context-aware delay.
  Readiness fails concurrently when the Redis ping is also unavailable.
- A terminal HTTP, SSE hub, or consumer return marks readiness failed, cancels
  sibling components, and propagates an error to the process entrypoint.
- Per-client Pub/Sub subscription failures disconnect that stream. They do not
  currently mark the whole SSE hub failed when Redis itself still answers ping.
- The Redis client does not classify every stream error as transient or
  permanent. A persistent stream-specific error while Redis ping remains
  healthy is retried and is not yet represented by a dedicated readiness check
  or consumer-lag metric.
- OpenTelemetry exporter delivery is managed by the SDK. Provider startup and
  bounded shutdown are supervised; asynchronous export diagnostics remain
  telemetry signals rather than process-fatal errors.

## Verification

Focused lifecycle tests cover startup failure, runtime component failure,
signal cancellation, active HTTP request cancellation, readiness transitions,
single-close ownership, dependency timeouts, repeated cancellation races, and
SSE hub cancellation. Run them with:

```bash
go test -race ./cmd/api ./internal/bootstrap/api ./internal/eventconsumer ./internal/platform/health ./internal/modules/health/http ./internal/sse
go vet ./cmd/api ./internal/bootstrap/api ./internal/eventconsumer ./internal/platform/health ./internal/modules/health/http ./internal/sse
```
