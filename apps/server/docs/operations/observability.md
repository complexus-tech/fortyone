# API observability and safe correlation

Every request receives a server-generated UUID in the `X-Request-ID` response header. Error envelopes repeat the same value as `error.request_id`, allowing support and operators to locate the request without asking for a URL containing private identifiers or tokens.

The structured logger automatically adds:

| Field        | Source                            | Use                                             |
| ------------ | --------------------------------- | ----------------------------------------------- |
| `request_id` | Server-generated per HTTP request | Correlate all application logs for one request. |
| `trace_id`   | Active OpenTelemetry span         | Correlate logs with a distributed trace.        |
| `span_id`    | Active OpenTelemetry span         | Locate the exact span that emitted a record.    |
| `service`    | Logger construction               | Distinguish API, worker, and command records.   |

`request_id` is diagnostic metadata only. It is not an authentication credential, idempotency key, CSRF token, or stable user identifier. The server does not trust a caller-supplied value as its canonical request ID.

## Safe logging rules

- Record matched route patterns such as `/invitations/{token}`, never raw paths or request URIs. Raw paths can contain invitation codes and other bearer values.
- Record principal, workspace, credential, installation, resource, and delivery IDs only when they are needed for the event and the log access class permits them.
- Never record authorization headers, cookies, OAuth codes/state, provider tokens, vault envelopes, HMAC inputs/keys, database URLs, full webhook bodies, or environment values.
- Record safe fingerprints only when correlating an opaque value is necessary. A fingerprint must be one-way, purpose-bound, short, and insufficient to authenticate.
- Provider errors are classified at the adapter boundary. Log the safe class/status/request ID; do not copy an arbitrary upstream response body.
- The final HTTP transport boundary logs only that an unexpected handler failure occurred, its status, and the request correlation fields. It deliberately omits `err.Error()` because an unclassified error can contain SQL, credentials, or provider payloads.
- The logger is a defense-in-depth privacy boundary: known token, secret, cookie, body, payload, raw-message, error-message, and email attribute keys are replaced with `[REDACTED]`. Any attribute whose value implements `error` becomes a structured diagnostic containing a bounded Go type chain, never arbitrary `Error()` text. Callers still must omit sensitive attributes; redaction is not permission to collect them.
- Operational boundaries that need an actionable production message must wrap the cause with a package-level `logger.MustDefineError` definition. This adds a stable `error.code` and reviewed `error.safe_message` while preserving `errors.Is`/`errors.As`. Safe messages are literal, bounded copy; they must never include runtime values, `err.Error()`, SQL, provider responses, URLs, credentials, or user input.
- HTTP logs use `r.Pattern`, not `r.URL.Path`, so high-cardinality resource IDs do not become metric or trace dimensions.
- HTTP request logs omit raw client addresses. Edge/network telemetry owns source-address evidence under its separate access and retention policy.

## Incident lookup

1. Ask for the `X-Request-ID` or `error.request_id` and approximate time.
2. Locate the API record with that `request_id` and confirm the matched route, status, deployment version, and safe actor/workspace identifiers.
3. Follow `trace_id` into the trace backend for dependency timing and the failing span.
4. For queued work, follow the durable task/delivery ID recorded by the enqueue event rather than assuming the original HTTP span remains open.
5. If an error is still unclassified, add an explicitly reviewed error definition, safe field, or metric at the boundary that understands the failure; never temporarily log the secret-bearing payload.

Process-fatal API and worker records always include `version`, `phase`, and an
`error` object. Query `error.code` first; startup classifications distinguish
configuration, PostgreSQL, Redis, HTTP, scheduler, queue, migration, and Slack
failures. For example, a missing Slack signing secret is reported as
`worker.slack.signing_secret_missing` with the reviewed summary
`SLACK_SIGNING_SECRET is not configured`, without writing the secret value or
the underlying provider/database text.

Readiness and liveness expose stable dependency states without returning underlying error strings. The API changes readiness to draining before it stops accepting work; the worker reports its queue/scheduler state independently. See `docs/architecture/api-lifecycle.md` for lifecycle semantics.

## Review checklist

- New log fields have bounded cardinality or are excluded from metrics.
- Error responses contain a safe stable message and request ID, never raw database/provider errors.
- Background work has its own durable correlation identifier.
- Sensitive routes use matched patterns in logs and spans.
- Tests assert secrets and raw bearer paths are absent from logs.
- A new external dependency has latency, error-class, retry, and saturation signals before it becomes release-critical.
