# API idempotency receipts

The shared receipt service gives explicitly adopted mutation endpoints one
consistent retry contract. It is implemented in
`internal/platform/idempotency`, backed by sqlc/pgx, and introduced by migration
`000156_api_idempotency_receipts`. The first adopted operation is
`POST /api/v1/workspaces/{workspaceId}/stories`, with the stable operation name
`stories.create`.

Adopting HTTP routes accept the exact `Idempotency-Key` request header. The
credentialed CORS policy already permits that header, but a route must not
advertise or act on it until its transaction/outbox review is complete.

The API composition root constructs the service from the native pgx pool and
passes it explicitly to the public API transport. That wiring does not make any
other endpoint idempotent: every additional route still needs a separate review
of its domain transaction and crash-recovery behavior before it may advertise
or act on `Idempotency-Key`.

## Contract at a glance

A receipt is unique within this complete scope:

```text
principal kind + principal ID + optional workspace ID
+ HTTP mutation method + stable route operation + SHA-256(key)
```

The request body is represented by the SHA-256 hash of its exact bytes. A raw
idempotency key or request body is never stored. The service has no logger and
its key value formats as `[REDACTED]` if it is accidentally passed to a
printf-style logger.

`Begin` returns exactly one state:

| State         | Meaning                                                              | Caller action                                                                           |
| ------------- | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `new`         | This caller owns a fresh or reclaimed lease.                         | Perform the reviewed domain mutation, then complete the receipt with the safe response. |
| `in_progress` | The same request hash already has a live lease.                      | Do not run the mutation. Retry no earlier than `RetryAt`.                               |
| `completed`   | The same request already completed.                                  | Replay only the returned status, body, and content type.                                |
| `conflict`    | The same scoped key is unexpired but covers different request bytes. | Do not run the mutation. Return the route's generic idempotency-conflict response.      |

An expired receipt starts a new lifecycle and may cover different request
bytes. A stale in-progress receipt can be reclaimed only after its lease
expires. Reclamation increments `lease_generation`; completion requires the
exact receipt ID and current generation and fails after lease expiry. A stale
handler therefore cannot overwrite the outcome of its replacement.

## Limits and defaults

| Value                           | Contract                                                                                                                                                      |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Idempotency key                 | 16–255 visible ASCII bytes. Prefer at least 16 cryptographically random bytes encoded with base64url. UUIDv4 provides 122 random bits and is also acceptable. |
| Route operation                 | 1–128 lowercase ASCII characters matching `[a-z][a-z0-9._:-]*`; for example `stories.create`. Never include a URL, slug, or resource ID.                      |
| HTTP method                     | `POST`, `PUT`, `PATCH`, or `DELETE`.                                                                                                                          |
| Request body covered by Phase 1 | At most 1 MiB. Larger or streaming mutations need a separately reviewed streaming/canonical hash design.                                                      |
| Replayed response status        | `200`–`599`. A route decides which terminal outcomes are safe to complete.                                                                                    |
| Replayed response body          | At most 64 KiB.                                                                                                                                               |
| Replayed content type           | Valid MIME media type, at most 128 bytes, with no CR/LF.                                                                                                      |
| Replayed headers                | None. `Set-Cookie`, authorization, cache, location, rate-limit, and arbitrary custom headers are not representable.                                           |
| Default lease                   | 2 minutes; configurable from 1 second through 15 minutes.                                                                                                     |
| Default retention               | 24 hours; configurable from 1 hour through 30 days and always longer than the lease.                                                                          |
| Purge batch                     | 1–1,000 expired receipts per call.                                                                                                                            |

Key length is a format check, not an entropy detector. Client SDK and API
documentation must require random, per-logical-mutation keys. A key is not an
authentication credential and never replaces actor authorization, current
workspace membership, or resource ownership checks.

## Exact request bytes

The caller must pass the same bounded bytes used by the handler. Hashing decoded
then re-encoded JSON is unsafe unless the route defines one canonical encoding:
whitespace, numeric representation, object member order, compression, and
trailing bytes can otherwise change the hash. A route adapter should read and
bound the body once, call `Begin` with those bytes, and then decode that same
buffer. The service does not log the body, key, or their digests.

Method and operation are part of the scope so one key may safely be reused for
different operations. Principal identity comes from `auth.Actor`. Optional
workspace identity is derived from the same actor rather than accepted as a
separate caller parameter, preventing accidental tenant mismatch.

## Response replay boundary

Construct replay data with `idempotency.NewResponse`. The body is copied on
input and output. Only these values cross the receipt boundary:

```text
status code
bounded response body
content type
```

Do not reconstruct or infer omitted headers during replay. If an endpoint needs
to set a cookie, rotate a credential, create a one-time URL, or return another
per-delivery header, it is not directly eligible for this Phase 1 replay shape.
Design a resource lookup or another explicit recovery response instead.

## Transaction coordination is mandatory

The receipt repository uses one pgx transaction to serialize `Begin`, including
the absent-row race, stale takeover, and expiry restart. `Complete` is one fenced
update. Those transactions do not automatically include an unrelated module's
domain write.

The dangerous crash window is:

```text
domain transaction commits
process exits before receipt completion
lease expires
replacement handler runs the domain mutation again
```

Before adopting a route, prove at least one of these designs:

1. The domain table has a tenant-scoped uniqueness/idempotency invariant that
   turns the second execution into the same logical result.
2. Receipt state and the domain mutation participate in one reviewed shared
   pgx unit of work, with generated queries bound to the same `pgx.Tx`.
3. The domain transaction writes a durable outbox/result record from which a
   stale receipt can be completed without repeating the side effect.

An external API call, email, queue publication, or provider mutation needs an
outbox or provider-native idempotency contract. Holding a database transaction
open across network I/O is not an acceptable substitute.

For each adopting operation, document:

- which bytes are hashed and the maximum request size;
- the stable operation name and actor/workspace scope;
- which success and error responses are terminal and replay-safe;
- the domain uniqueness, shared transaction, or outbox invariant;
- behavior for a crash before the domain write, after the domain commit, and
  after receipt completion;
- metrics, retention, and an operator recovery query that expose no key or
  payload material.

## Expiry and purge

Receipt expiry is lazy: `Begin` atomically replaces an expired scoped row with a
new receipt ID, request hash, and generation one. `PurgeExpired` removes only
rows whose `expires_at` has passed, in a bounded ordered batch using
`FOR UPDATE SKIP LOCKED`. Multiple purge callers are safe. The worker schedules
`cleanup:api_idempotency_receipts` hourly at minute 7. Each database statement
deletes at most 1,000 rows; the handler continues until a short batch proves the
backlog is drained. Its Asynq execution has a ten-minute timeout and five
retries, so cancellation bounds one worker execution without widening a
database statement. Every 100 full batches emits one aggregate warning to make
sustained growth visible. The task never logs key, request, response, principal,
or workspace material.

Useful telemetry is limited to safe dimensions and aggregates: route operation,
state, reclaimed count, lease-lost count, purge count, age buckets, and latency.
Never include keys, bodies, response bodies, key digests, request hashes, raw
URLs, or database error details that reproduce retained values.

## Migration `000156` rollout and recovery

The migration is additive and schema-first:

1. Back up PostgreSQL and apply `000156` before deploying service code.
2. Verify the table is empty and its scope uniqueness, expiry, and stale-lease
   indexes exist.
3. Deploy the schema before the API and worker code. Old binaries ignore the
   table, so this stage is rolling-deploy compatible.
4. Adopt one route only after its coordination review. Before accepting keys on
   that route, remove or isolate old API instances that could still serve it
   without the receipt contract.
5. Deploy the worker cleanup handler and verify its aggregate purge telemetry.
6. Exercise new, in-progress, completed replay, conflict, stale takeover,
   completion fencing, expiry, and purge on PostgreSQL 18.

The migration is operationally reversible only while the table is empty. Its
down migration refuses to discard retained rows. Before rollback after route
adoption, disable every adopting route or remove its idempotency behavior, wait
for the configured retention window, purge expired receipts, and prove the
table is empty. If adoption cannot be paused or receipts must be retained, keep
schema `000156` and deploy a forward application fix.

The machine-readable rollout source of truth and generated operator procedure
are [`internal/migrations/manifest.json`](../../internal/migrations/manifest.json)
and [`docs/database/migration-operations.md`](../database/migration-operations.md).

## Verification

The default suite covers typed limits, redaction, immutable response bodies,
schema security contracts, exact SHA-256 behavior, and fuzzed key/request hash
inputs. Tagged tests use the shared isolated-database testkit on PostgreSQL 18
and cover same-body replay, different-body conflict, one concurrent winner,
stale takeover, completion fencing, expiry, purge, and principal/workspace scope.

Run from `apps/server`:

```bash
go test ./internal/platform/idempotency/...
go test -race ./internal/platform/idempotency/...
go test ./internal/taskhandlers ./internal/bootstrap/worker
TEST_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/test_control?sslmode=disable' \
  go test -tags=integration -count=1 ./internal/platform/idempotency
make sqlc-check
SQLC_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/sqlc_validation?sslmode=disable' \
  make sqlc-vet
make migration-check
```
