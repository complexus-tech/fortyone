# Inbound webhook gateway

This document is the implementation guide for the provider-neutral inbound
webhook control plane in `internal/platform/webhooks`. Read it with
[ADR 0008](../architecture/decisions/0008-integration-capabilities.md),
[ADR 0009](../architecture/decisions/0009-webhook-delivery.md), and the
[provider registry](providers.md).

The gateway is internal infrastructure for compiled first-party adapters. It is
not a public plugin ABI, and it does not make provider payloads or semantics
universal. Slack still owns Slack signatures, event shapes, challenge handling,
and message rules; GitHub still owns GitHub signatures, delivery headers, and
event normalization. The gateway owns only the lifecycle those adapters share.

## Where each concern lives

| Concern                                           | Owner                                                        |
| ------------------------------------------------- | ------------------------------------------------------------ |
| Raw HTTP byte limit                               | HTTP adapter through `pkg/web.ReadBoundedBody`               |
| Signature, supported algorithm, and replay window | provider `WebhookVerifier`                                   |
| Provider payload decoding and event normalization | provider adapter                                             |
| Provider account to FortyOne installation lookup  | provider adapter                                             |
| Capability/version selection                      | `internal/platform/integrations` registry                    |
| Canonical envelope validation                     | `internal/platform/webhooks` gateway                         |
| Exact-body encryption                             | provider `PayloadProtector` implementation                   |
| Deduplication, leases, and retention              | SQLC inbox repository                                        |
| Queue handoff                                     | provider `Dispatcher`, receiving only provider plus inbox ID |
| Product authorization and business behavior       | consuming module service                                     |

Provider SDK request/response types must not cross into the gateway, service
domain, or SQLC repository.

## Receive path

```text
provider request
      |
      v
bounded exact body ---- overflow ----> reject before verification
      |
      v
provider verifier ----- invalid/stale -> safe classified rejection
      |
      v
installation + generation resolution
      |
      v
context-bound payload encryption
      |
      v
unique SQLC inbox receipt
      |
      +---- terminal duplicate -------> acknowledge, no redispatch
      |
      v
queue { provider, inbox_id }
      |
      v
record queue handoff and acknowledge
```

The exact signed body is cloned before the verifier runs. A verifier therefore
cannot accidentally mutate the bytes later encrypted for replay. Verification
errors are reduced to stable classifications; provider errors, signed bodies,
tokens, and account data are not copied into client errors or logs.

Authentication and availability are different contracts. A missing or invalid
signature, replay failure, or unauthorized signed identity maps to 401. If the
verifier cannot decide because its installation/grant repository is unavailable,
it returns `ErrVerificationUnavailable`, which the common ingress adapter maps
to retryable 503. The stable wrapper preserves that classification while
discarding the provider/database cause from the public response. Treating an
outage as 401 would cause providers to stop retrying an otherwise valid event.

Persistence always precedes dispatch. If dispatch fails, the durable pending
receipt remains recoverable and the provider receives a retryable response. If
dispatch succeeds but recording the handoff fails, duplicate tasks are allowed:
workers must be idempotent and acquire the inbox processing lease before doing
work. This is deliberate at-least-once delivery, not exactly-once delivery.

## Canonical envelope

Migration `000157_webhook_inbox_envelopes` extends the existing
`messaging_inbound_events` inbox without renaming or rewriting it during a
rolling deployment. A receipt contains:

| Field                     | Invariant                                                                   |
| ------------------------- | --------------------------------------------------------------------------- |
| `envelope_version`        | positive, currently `1`; independent of provider payload version            |
| `provider`                | stable key from the compile-time provider registry                          |
| `external_event_id`       | stable provider delivery identity used for deduplication                    |
| `external_workspace_id`   | provider account/tenant/workspace identity                                  |
| `event_type`              | normalized provider event name, not a Go type                               |
| `workspace_id`            | resolved FortyOne tenant; never accepted from an untrusted request directly |
| `installation_id`         | opaque internal installation identity                                       |
| `installation_generation` | lifecycle fence against reinstall, disconnect, or credential replacement    |
| `trace_id`                | optional safe correlation identifier, never a bearer value                  |
| `payload_encrypted`       | exact retained provider body; never queued or logged                        |
| `payload_expires_at`      | content-retention deadline                                                  |

The unique key remains `(provider, external_workspace_id, external_event_id)`.
When a duplicate uses that key but disagrees on workspace, installation,
generation, event type, or envelope version, the gateway fails closed with a
delivery conflict. It does not silently rewrite canonical identity.

The migration is additive for old API/worker compatibility. It backfills the
Slack installation ID where an existing inbox row can be mapped safely. The
new columns remain compatible with receipts created by an older binary during
the rolling window.

## State machine

| State        | Meaning                                        | Valid next state                                                |
| ------------ | ---------------------------------------------- | --------------------------------------------------------------- |
| `pending`    | durably accepted, not yet leased               | `processing`                                                    |
| `processing` | worker owns a bounded lease                    | `completed`, `ignored`, `failed`, `cancelled`, or lease reclaim |
| `failed`     | retryable processing attempt ended             | `processing`                                                    |
| `completed`  | product effect completed                       | terminal                                                        |
| `ignored`    | valid delivery required no product effect      | terminal                                                        |
| `cancelled`  | installation/grant lifecycle made work invalid | terminal                                                        |

Only a `processing` row can be completed. A second worker receives
`ErrLeaseBusy` until the lease expires. A terminal duplicate is acknowledged
without another queue handoff.

Completion stores only a bounded machine outcome code such as
`provider.rate_limited`; arbitrary provider/database errors and raw payload
fragments are rejected at this boundary.

## Recovery and fairness

`Gateway.Recover` claims rows with `FOR UPDATE SKIP LOCKED`, increments a
recovery generation, and applies bounded exponential backoff. The default
policy is:

- claim at most 100 rows per provider and scan;
- stop after 20 processing attempts;
- recover pending handoffs after 30 seconds;
- recover failed work after five minutes;
- reclaim a processing lease after ten minutes;
- start recovery backoff at 30 seconds, capped by a validated shift.

Claims are provider-scoped so one provider cannot consume every recovery batch.
Production scheduling must also enforce per-tenant/provider concurrency and
observe queue age before increasing global parallelism.

If one dispatch fails, recovery releases that exact recovery generation and
continues with the rest of the batch. A stale worker cannot clear a newer
claim. Operators should alert on oldest pending age, repeated recovery,
attempt exhaustion, and terminal failure—not on raw error strings.

## Payload retention

The default encrypted payload retention is 30 days and is configurable at
gateway construction between one hour and 90 days. `ExpirePayloads` clears only
encrypted content in bounded batches. Delivery ID, provider, installation
generation, timestamps, status, and attempt counters remain as safe audit and
deduplication facts.

Changing retention requires a documented provider/legal need and an operations
review. Never extend retention by leaving `payload_expires_at` unset for new
receipts.

## Adding a provider runtime

1. Declare `control.webhook_verification` major version 1 in the provider
   descriptor.
2. Implement a small provider verifier. Verify authenticity and replay before
   trusting event fields, then resolve an active installation and its current
   generation.
3. Use `pkg/web.ReadBoundedBody` in the HTTP adapter with the provider's
   documented maximum. The gateway's in-memory check is defense in depth, not
   permission to allocate an unbounded body first.
4. Implement payload protection. Bind provider, workspace, installation,
   generation, and delivery ID in authenticated context.
5. Implement a dispatcher that serializes only `{provider, inbox_id}`.
6. Register the runtime explicitly in bootstrap. Duplicate, unknown, missing,
   or capability-incompatible registrations must fail startup.
7. In the worker, load the inbox record by ID, acquire its lease, recheck the
   active installation/grant generation, decrypt only at the provider parsing
   boundary, execute idempotently, and complete with a safe outcome code.
8. Add sanitized provider fixtures and run the common contract plus the
   provider-specific signature/replay suite.

Challenge requests or single-use provider triggers with deadlines shorter than
durable dispatch may stay in the provider adapter. They must still be bounded
and signature-verified, and the exception must be tested and documented. Do not
persist an expiring one-use trigger merely to satisfy a generic abstraction.

## Required tests

Each adapter must prove:

- exact body preservation and provider body limit;
- valid, invalid, missing, and algorithm-confused signatures;
- current and stale signed timestamp behavior where available;
- stable delivery-ID deduplication where timestamp replay protection is absent;
- unknown/revoked installation and stale generation;
- terminal duplicate quick acknowledgement;
- persistence before dispatch and dispatch failure recovery;
- concurrent lease and recovery claims;
- queue payload contains no raw body or credential;
- payload expiry retains safe audit identity;
- provider/database error text is absent from public errors and logs;
- compatibility with tasks already queued by the previous API/worker version.

Fast package checks:

```bash
go test -race ./internal/platform/webhooks/...
.tools/bin/sqlc compile -f sqlc.yaml
```

Repository contract tests require the disposable PostgreSQL 18 control database
described in [integration infrastructure](../testing/integration-infrastructure.md):

```bash
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/platform/webhooks/repository
```

## Rollout and rollback

Deploy the additive migration before binaries that write envelope metadata.
During a normal rolling deployment, old tasks and old receipts must remain
readable by the new worker, and the old worker must tolerate the added columns.
Deploy API before worker when introducing the new queue envelope only after the
worker can read both legacy and new tasks.

Rollback of the binary is safe while the additive columns remain. Do not run
the down migration while any new binary can write or read the new envelope,
while new-format tasks remain queued, or while operator recovery depends on the
new fields. Prefer rolling the binary forward; schema rollback is an explicit
maintenance operation.
