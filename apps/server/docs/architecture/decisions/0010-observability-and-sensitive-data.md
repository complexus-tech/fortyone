# ADR 0010: Observability, SLOs, and sensitive data

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering and operations

## Context

Logs and health endpoints are useful only when they describe the real lifecycle
and can be correlated without leaking secrets. Unbounded labels and raw payloads
create both operational and security risk. Signals without owners or response
procedures do not make the service operable.

## Decision

Use structured logs, OpenTelemetry traces, and bounded-cardinality metrics with
the shared fields appropriate to the operation: request/trace ID, deployment
version, route template, operation, actor kind, workspace ID when authorized,
provider, job/task kind, outcome, duration, and stable safe error code.

Never record bearer tokens, cookies, authorization headers, OAuth codes,
verification/invitation tokens, encryption material, raw webhook bodies, email
bodies, or unrestricted user/provider payloads. Database statements and
arguments containing customer data are not exported. Presence, version,
fingerprint, length, and sanitized provider request IDs may be logged when useful.

Liveness means the process/event loop is alive. Readiness means the instance can
accept its assigned work and required dependencies pass bounded checks. During
shutdown readiness turns false before listeners/consumers drain. Build/version
identity is exposed in deployment telemetry, not secrets.

Every critical journey receives an SLI, target, owner, alert, and runbook. Initial
journeys are API availability/latency, authentication, database pool saturation,
webhook acceptance/processing, queue age/failure, and provider credential refresh.
Alerts use user-impacting symptoms plus actionable saturation/error signals.

## Enforcement and adoption

- Lifecycle tests prove truthful readiness and bounded drain.
- Log/trace tests and security scans assert sensitive-value absence.
- Metric review rejects raw IDs or attacker-controlled labels.
- Release verification observes version, readiness, error rate, and worker/queue health.

## Consequences

Diagnostics become correlatable and safer, but engineers must choose fields and
cardinality deliberately. SLOs create explicit operational ownership rather than
an assumption that dashboards alone are sufficient.

## Revisit when

Telemetry cost, privacy requirements, or incident evidence justifies a different
retention/sampling policy; sensitive-data exclusions remain fail-closed.
