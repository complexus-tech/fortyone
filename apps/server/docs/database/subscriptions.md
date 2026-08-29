# Subscriptions and Stripe state

The subscriptions module owns workspace billing snapshots, invoice history, and the durable Stripe webhook inbox. PostgreSQL is the authority for processed-delivery state; Stripe remains the authority for the current provider subscription.

## Where to find the code

| Concern                           | Location                                                                   |
| --------------------------------- | -------------------------------------------------------------------------- |
| Domain values and provider cursor | `internal/modules/subscriptions/domain/`                                   |
| Billing use cases                 | `internal/modules/subscriptions/service/`                                  |
| Handwritten SQL                   | `internal/modules/subscriptions/repository/queries/`                       |
| SQLC adapter and mappings         | `internal/modules/subscriptions/repository/`                               |
| Generated SQLC code               | `internal/modules/subscriptions/repository/sqlc/`                          |
| Webhook lease schema              | `internal/migrations/000153_stripe_webhook_processing_leases.*.sql`        |
| Event-order schema                | `internal/migrations/000164_subscription_event_ordering.*.sql`             |
| PostgreSQL 18 contracts           | `internal/modules/subscriptions/repository/repository_integration_test.go` |

Generated files are adapter details and must not be imported by HTTP or service packages. Change the SQL under `repository/queries`, run `make sqlc-generate`, and review the generated signatures.

## Tenant boundaries

First-party reads and operator-triggered synchronization always include `workspace_id`. A Stripe webhook has no authenticated workspace route, so its trusted boundary is different: after exact-body signature verification, a globally unique Stripe subscription or invoice ID is resolved to its immutable workspace binding.

- subscription creation binds `(stripe_subscription_id, workspace_id)` and refuses reassignment;
- subscription snapshot/deletion writes resolve the existing provider ID and return its workspace;
- snapshot and manual-sync writes cannot adopt a Stripe customer already bound to another workspace;
- invoice writes require both the workspace and the already-bound Stripe customer ID;
- an existing invoice ID cannot be reassigned to another workspace;
- terminal webhook rows record the resolved workspace for safe operations and audit queries.

Do not add a provider-ID-only mutation unless the provider identity is globally unique, the request is cryptographically authenticated, and the adapter returns the bound workspace.

## Delivery state

`stripe_webhook_events` is a leased inbox:

```text
received/claim -> processing -> processed (handled or ignored)
                         \----> failed -> retry claim
                expired lease -> replacement claim
```

The event ID is the idempotency key. A lease token fences completion and failure, so a crashed or slow owner cannot overwrite a replacement owner. A successful terminal event cannot be reclaimed. A repeated event ID with a different event type is rejected as identity confusion. Raw Stripe payloads are deliberately not retained.

Domain effects are replay-safe: subscription snapshots are monotonic upserts, invoices are workspace-fenced upserts, trial notifications have stable task IDs, and seat/plan/cancellation provider writes carry intent-derived idempotency keys. If the process crashes after a domain write but before inbox completion, Stripe can retry without duplicating billing state.

## Out-of-order events

Stripe delivery order is not guaranteed. Migration `000164` adds a cursor to each subscription:

```text
(event created_at, event priority, event_id)
```

Writes advance only monotonically. Same-second semantic priority is `subscription.created < current-provider snapshot < subscription.deleted`, so a delayed creation cannot overwrite an update and a snapshot cannot resurrect a deleted subscription. Creation, update, and checkout handlers read the current Stripe subscription before applying their cursor instead of trusting a potentially old delivery body.

Stripe event IDs do not encode chronology. They are used only as a stable deterministic tie-breaker between otherwise equivalent cursors; correctness for those ties comes from applying a current-provider snapshot. Exact-event retries remain idempotently applicable.

The cursor is persistence metadata, not an authorization credential or a public API field.

## Query and index expectations

- workspace subscription reads use `workspace_id`, deterministic `created_at`/subscription-ID ordering, and a five-row invoice cap;
- provider mutations use the primary key on `stripe_subscription_id`;
- customer-binding mutations take a transaction advisory lock before testing cross-workspace ownership;
- invoice upserts use the unique Stripe invoice index from migration `000153`;
- webhook claim recovery uses `idx_stripe_webhook_events_retryable`;
- every query is static, parameterized, and compiled by SQLC.

## Verification

```sh
go test ./internal/modules/subscriptions/...
TEST_DATABASE_URL='postgres://.../postgres?sslmode=disable' \
  go test -tags=integration -count=1 -race ./internal/modules/subscriptions/repository
make sqlc-check
make sqlc-vet
make architecture-check
```

The integration role must target disposable PostgreSQL 18 and have `CREATEDB`. The test kit applies the complete migration chain to a random database and removes it after the run.
