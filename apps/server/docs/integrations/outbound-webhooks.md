# Outbound webhook delivery

This runbook describes the customer-facing webhook system implemented by
`internal/modules/outboundwebhooks`. It is separate from the provider-facing
[inbound webhook gateway](webhook-gateway.md): inbound adapters authenticate
GitHub, Slack, and other providers; outbound webhooks deliver FortyOne product
events to customer-controlled HTTPS endpoints.

Read this with [ADR 0004](../architecture/decisions/0004-actors-authorization-and-revocation.md),
[ADR 0006](../architecture/decisions/0006-cursor-pagination.md),
[ADR 0007](../architecture/decisions/0007-machine-credentials.md), and
[ADR 0009](../architecture/decisions/0009-webhook-delivery.md).

## Ownership boundaries

| Concern                                                                 | Owner                                                |
| ----------------------------------------------------------------------- | ---------------------------------------------------- |
| Endpoint lifecycle and authorization                                    | outbound webhook service                             |
| Endpoint, subscription, event, delivery, attempt, and audit persistence | outbound webhook SQLC repository                     |
| Signing-secret encryption and rotation                                  | shared credential vault plus outbound secret manager |
| URL and network safety                                                  | `internal/platform/safehttp`                         |
| Exact event envelope and event catalog                                  | outbound webhook domain                              |
| Delivery claims, signatures, retries, and outcome classification        | outbound dispatcher                                  |
| Scheduled recovery and bounded draining                                 | worker bootstrap                                     |
| Product event timing and stable event ID                                | calling product service                              |
| HTTP/OpenAPI models                                                     | versioned HTTP adapter                               |

Product services publish domain events through the narrow `Publisher` port.
They do not insert delivery rows, open signing secrets, make network calls, or
import generated SQLC types. The HTTP adapter does not import a repository.

## Endpoint management authorization

Endpoint management requires all of the following at the moment of the call:

- a `human_user` or `personal_token` actor;
- the actor tenant equal to the requested workspace;
- a current workspace `admin` role;
- the `webhooks:manage` scope (`first_party:*` satisfies this for browser sessions).

Service accounts cannot manage webhooks. The HTTP adapter performs its cheap
boundary checks and the service repeats the full policy. A human owner is
stored as a principal-registry identity, not as an unscoped user ID. Browser
sessions may provision that principal; a PAT may only resolve the human
principal that already backs the token. Every endpoint read or write includes
the workspace ID.

Before fan-out and again before a delivery claim, SQL verifies that the owner
principal is active, the subject user is active, and current workspace
membership exists. Removing membership, disabling the user, or disabling the
principal therefore stops both new fan-out and queued dispatch without waiting
for a cache to expire.

## Endpoint URL policy

An endpoint is accepted only when all of these conditions hold:

- scheme is exactly HTTPS;
- port is absent or exactly `443`;
- hostname is a valid multi-label ASCII DNS name;
- URL contains no user information or fragment;
- hostname is not an IP literal;
- every resolved address is globally routable.

Loopback, private, link-local, metadata, carrier-grade NAT, documentation,
reserved, multicast, and other non-public address classes are rejected. The
creation-time check gives fast feedback, but it is not trusted permanently.
Every attempt resolves DNS again, rejects the endpoint if any returned address
is unsafe, pins the TCP connection to a validated address, and still verifies
TLS against the original hostname. Redirects, environment proxies, and
connection reuse are disabled. This closes common DNS-rebinding and proxy-based
SSRF paths.

The client enforces a ten-second request timeout, five-second TLS/header
timeouts, bounded response headers, and a 64 KiB response-body limit. Response
bodies are hashed and discarded; their content is never persisted or logged.

## Event contract

The initial catalog is deliberately small:

- `story.created`
- `story.updated`
- `story.deleted`
- `comment.created`
- `comment.updated`
- `comment.deleted`

Adding an event is a contract change. Update the domain catalog, database
constraints, OpenAPI source, examples, and contract tests together. Do not emit
an undocumented string dynamically.

Every request body is a versioned JSON object:

```json
{
  "id": "8b879ad7-cc02-41bc-9d36-d24d8ef72bc8",
  "type": "story.updated",
  "payload_version": 1,
  "occurred_at": "2026-08-28T11:30:00Z",
  "data": {
    "storyId": "07a7d33f-aead-4c0c-8c12-75980f337b4e",
    "workspaceId": "60fe61ca-a12c-49a1-9581-0f66bf2ac29a",
    "changes": {
      "archived_at": null
    }
  }
}
```

`id` is supplied by the calling product use case and is the durable idempotency
identity. Retries of the publish operation must reuse it. Reusing an ID with a
different workspace, event type, subject, actor, timestamp, or semantic JSON
payload is a conflict. PostgreSQL `jsonb` equality is used for semantic payload
comparison so harmless key-order or whitespace differences remain idempotent.

The exact encoded body stored on the delivery is the exact body signed and sent
on every attempt. No worker re-marshalling is allowed.

Story version-1 data is use-case specific and uses the existing camel-case API
field convention:

- `story.created` contains `storyId`, `workspaceId`, `teamId`, `title`,
  `assigneeId`, and `reporterId`;
- `story.updated` contains `storyId`, `workspaceId`, and a `changes` object; and
- `story.deleted` contains only `storyId` and `workspaceId`.

Typed updates include only fields that actually committed. Restore, archive,
and unarchive use `deleted_at` or `archived_at`; label and collaborator
replacement use the literal `"changed"` marker instead of exposing identifier
lists. A no-op lifecycle or relationship retry creates no event. Creation title
and ordinary changed story fields are tenant product data in the documented
contract, so receivers must protect webhook bodies as workspace data. Story
events never include credentials, integration tokens, provider secrets, or
signing material.

All three comment event types use the same version-1 data shape:

```json
{
  "comment_id": "48bde60f-d1e4-46f5-bd75-f32a193e85cc",
  "story_id": "406741c5-5cff-4673-a3bb-4036cc6c78c9",
  "parent_id": null
}
```

Comment content, mentions, names, email addresses, credentials, provider
metadata, and secrets are deliberately excluded. An integration that needs the
current representation must fetch it through an authorized API read. A parent
delete invalidates its reply subtree because replies use database cascade
semantics; only the explicitly addressed parent has a `comment.deleted` event.

## Signing and rotation

An endpoint receives a `whsec_...` secret exactly once when it is created or
rotated. FortyOne stores only its credential-vault envelope, authenticated to
the workspace, endpoint ID, credential type, and generation. String, GoString,
JSON, and structured-log representations are redacted by default.

Each attempt sends:

```text
Webhook-Id: <stable delivery UUID>
Webhook-Timestamp: <Unix seconds>
Webhook-Signature: v1,<base64 HMAC-SHA256>
```

The exact JSON body is non-empty and bounded to 256 KiB by the domain,
persistence, signing, and dispatch boundaries.

The signed bytes are:

```text
<Webhook-Id>.<Webhook-Timestamp>.<exact request body bytes>
```

Receivers must verify the signature against the unmodified raw body before
JSON decoding, compare in constant time, enforce their own timestamp replay
window, and deduplicate `Webhook-Id` before applying business effects.

Rotation uses compare-and-swap on the expected secret generation. For 24 hours,
FortyOne signs with the new secret first and the previous secret second:

```text
Webhook-Signature: v1,<current> v1,<previous>
```

Receivers should accept any valid `v1` value during the overlap. A stale
rotation cannot overwrite a newer generation. If the previous key is
unavailable, the current signature still delivers; if the current envelope
cannot authenticate, the endpoint is disabled because its local credential
state is corrupt.

## Delivery lifecycle

```text
publish event
    -> transactionally store event and endpoint fan-out
    -> pending
    -> delivering (30-second lease)
       -> succeeded
       -> retry_scheduled
       -> failed
       -> cancelled
```

The worker schedules a recovery/drain task every five seconds on the
`integrations` queue. A task first recovers expired leases and then handles at
most 16 deliveries. PostgreSQL `FOR UPDATE SKIP LOCKED` distributes work across
workers. An endpoint with a `delivering` row is excluded from new claims, so
requests to one destination stay serial while unrelated endpoints progress.

Every completion transaction writes the immutable attempt before changing the
leased delivery. The delivery update requires the exact lease token and attempt
number. A stale worker therefore cannot complete a lease reclaimed by another
worker.

## Outcome and retry policy

| Result                                              | Behavior                                           |
| --------------------------------------------------- | -------------------------------------------------- |
| HTTP `2xx`                                          | success; reset endpoint consecutive failures       |
| HTTP `408`, `425`, `429`                            | retry; respect a safe `Retry-After` up to 24 hours |
| HTTP `5xx`                                          | retry                                              |
| network/timeout failure                             | retry                                              |
| HTTP `410`                                          | fail and disable endpoint                          |
| unsafe DNS/URL or redirect                          | fail and disable endpoint                          |
| other HTTP `3xx`/`4xx`                              | terminal delivery failure                          |
| vault not configured or key temporarily unavailable | retry without penalizing endpoint health           |
| unauthentic/malformed current secret                | fail and disable endpoint                          |

The default retry delays are 1 minute, 5 minutes, 30 minutes, 2 hours, 8 hours,
and 24 hours with stable plus-or-minus ten-percent jitter derived from delivery
ID and attempt number. The maximum is eight attempts. Destination failures
increment endpoint health; 20 consecutive failures disable the endpoint.
Internal vault availability failures do not increment that counter.

## Data minimization and logs

Never log or place in queue payloads:

- signing secrets or vault envelopes;
- exact webhook request bodies;
- receiver response bodies;
- bearer credentials or URLs containing credentials;
- arbitrary receiver error text.

Safe persisted attempt facts are delivery/attempt IDs, timestamps, outcome,
resolved public IP, HTTP status, response byte count, SHA-256 response digest,
bounded machine error code, and duration. Audit metadata contains stable counts
and generations, not secrets or bodies.

## Operator checks

Investigate oldest eligible delivery, oldest delivering lease, retry depth,
terminal failures, disabled endpoints, and repeated lease recovery. A high
`secret_vault_unavailable` rate is an internal key/configuration incident; do
not ask the endpoint owner to rotate until vault availability is restored.

Do not manually mutate attempt or audit rows: PostgreSQL rejects updates and
deletes. Repair a stuck delivery through an explicit audited operation or a new
migration/tool, never an ad hoc status edit.

## Verification

Fast checks:

```bash
go test -race ./internal/platform/safehttp ./internal/modules/outboundwebhooks/... ./internal/bootstrap/worker
.tools/bin/sqlc compile -f sqlc.yaml
./scripts/check-sqlc.sh
```

PostgreSQL contract coverage requires the disposable PostgreSQL 18 control
database described in [integration infrastructure](../testing/integration-infrastructure.md):

```bash
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/outboundwebhooks/repository
```

The integration suite proves tenant fencing, owner revocation, semantic
idempotency, secret-generation CAS, endpoint-serial claims, retry state,
immutable attempts, and immutable audit history.
