# Feedback persistence and delivery

The feedback module owns public feedback intake, contributor identity,
moderation, roadmap updates, merge operations, reviewer digests, and contributor
email delivery. This document explains the persistence boundary for someone
who is new to either FortyOne or Go.

## Directory map

```text
internal/modules/feedback/
├── http/
│   ├── feedback.go               shared handler dependencies and errors
│   ├── public_portal_handlers.go public portal and contributor reads
│   ├── internal_item_handlers.go authenticated item/read-state queries
│   ├── portal_board_handlers.go  portal and board administration
│   └── item_mutation_handlers.go item, comment, vote, merge, and story writes
├── domain/               transport-neutral models, errors, access, and worker ports
├── service/              validation, actor policy, orchestration, and caller ports
├── storyadapter/         maps the feedback-owned story port to stories
└── repository/
    ├── queries/          reviewed PostgreSQL consumed by sqlc
    ├── sqlc/             generated code; never edit by hand
    ├── repository.go     pgx pool and transaction construction
    ├── items.go          item reads, status, trash, restore, and read state
    ├── community.go      comments, votes, contributor, and reviewer projections
    ├── portals_boards.go portal and board persistence
    ├── updates.go        roadmap publication plus its durable outbox
    ├── merges.go         atomic merge plus its durable outbox
    ├── deliveries.go     contributor-delivery claims and retry state
    └── maintenance.go    digest and retention persistence
```

HTTP, service, task-handler, and scheduled-job packages do not execute SQL and
do not import generated sqlc types. They depend on feedback-owned interfaces and
finite domain values. The repository is the only translation point between
those values and PostgreSQL. Repository code imports `feedback/domain`, never
the concrete service package, so persistence remains independently testable.

## Identity is not a boolean

Feedback keeps participation identity explicit because authentication,
contactability, and public attribution are different decisions.

| Contributor kind | Global user | Email delivery                                         | Public name                      |
| ---------------- | ----------- | ------------------------------------------------------ | -------------------------------- |
| `account`        | required    | normal account notification rules                      | account profile                  |
| `verified_guest` | absent      | allowed after email verification                       | shown or masked by portal policy |
| `external`       | absent      | allowed when the integration supplied a valid identity | shown or masked by portal policy |
| `anonymous`      | absent      | never                                                  | always anonymous                 |

`feedback_items`, `feedback_comments`, and `feedback_votes` therefore retain a
`contributor_id` even when no `user_id` exists. Public serializers never expose
the contributor email. The administrator-only private-author projection is a
separate type so private contact data cannot accidentally be added to a public
response.

## Authorization

Public queries start from a public portal and bind every board, item, comment,
vote, follow, update, and contributor to that portal. A supplied contributor
session must be unexpired, unrevoked, attached to the same portal, and not
blocked. Anonymous contributors cannot comment, vote, follow, or receive mail.

Authenticated workspace operations carry an explicit actor/workspace scope to
the repository. SQL rechecks all of the following against current rows:

- the workspace is active;
- the human actor is an active workspace member;
- the target board or item belongs to the supplied workspace;
- the actor is a current member of the owning team, unless the operation's
  documented administrator policy allows workspace-wide access;
- credential team restrictions can only narrow product membership.

The current feedback HTTP surface is first-party human only. Personal tokens,
OAuth actors, service accounts, applications, and system actors fail closed
until a dedicated feedback scope and public API contract are introduced. Public
contributors use the explicit contributor-session boundary instead of being
misattributed as workspace users.

Middleware is only an early rejection and context-binding layer. It is not a
substitute for these query fences.

The production repository implements `ScopedCoreStore` and
`ScopedNextPhaseStore`. The service derives these finite scopes from the actor
context and passes them explicitly; repositories never import HTTP middleware
or read ambient authentication. Legacy-shaped ports remain only to keep small
service fakes readable and fail closed in the production adapter where an
actor-aware operation is required.

## Atomic mutations and durable effects

Multi-row invariants use one native pgx transaction and sqlc queries bound with
`WithTx`.

- Creating a board and its initial reviewer subscription succeeds or rolls
  back together.
- Deleting a board first locks the matching anonymous contributor rows through
  a tenant-scoped `EXISTS` query, then deletes the board and only contributors
  left without any retained item in the same transaction. Lock the contributor
  rows directly: PostgreSQL does not allow `FOR UPDATE` on a `DISTINCT`
  projection.
- Verification consumption, contributor upsert, preferences, and session
  creation are one single-use transaction.
- Creating contributor feedback plus its initial follow, merging items and
  their relationships, replacing update links, and rotating widget secrets are
  atomic.
- Publishing an update commits the published state, immutable audience
  snapshot, versioned payload, and publication outbox row together.
- Merging commits the canonical relationship changes and versioned merge
  outbox row together.

Publication and merge consumers claim rows with `FOR UPDATE SKIP LOCKED`, a
fresh claim token, and a stale-lease cutoff. Completion and retry updates must
match the event ID, claim token, and `processing` status. A repeated publish or
merge returns the already-recorded result instead of producing a second event.

Contributor mail is also replay safe. The durable delivery row has a unique
dedupe key; the queue payload contains only `deliveryId`. Claiming rechecks that
the contributor is still eligible and that the unsubscribe digest remains
valid. A send failure records only a bounded generic reason, never the provider
error, email body, destination query, token, or credential. Scheduled recovery
re-enqueues due or stale deliveries by ID.

The task handler receives `ContributorDeliveryStore` from worker bootstrap.
There is no SQL fallback in the handler: missing persistence is a configuration
error. The same composition root injects `MaintenanceStore` and `DigestStore`
into scheduled jobs, leaving those jobs responsible only for timing, bounded
batch orchestration, email rendering, and aggregate logging.

Reviewer digests use one row per workspace, recipient, and local date. Failed
or stale `processing` rows may be reclaimed. Advancing every included board
cursor and marking the delivery `sent` or `skipped` is one transaction, so a
partial cursor advance cannot lose feedback from the next digest.

## Retention

The generic scheduler invokes feedback-owned maintenance ports; it does not
know table names. Feedback permanently deletes items only after the 30-day
recovery window and removes an anonymous contributor only when no retained item
still references it. Expired verification challenges, contributor sessions,
unsubscribe tokens, widget assertion nonces, and expired secret-rotation grace
rows are deleted through typed feedback queries in one maintenance operation.

## Query and pagination rules

- All static SQL lives in `repository/queries` and is compiled by the feedback
  block in `sqlc.yaml`.
- UUID, timestamp, nullable, JSON, and array values use the repository-wide
  sqlc overrides documented in [Typed database access](sqlc.md).
- Every list has a service-level maximum. Integer narrowing uses
  `internal/platform/safecast`; pagination multiplication is checked before the
  final cast.
- Sort and filter choices are finite values selected in Go. User values remain
  bound parameters; they never become identifiers or SQL fragments.
- Public list projections deliberately omit private email and token columns.

## Verification

The focused feedback gate includes:

```bash
go test -race ./internal/modules/feedback/... ./internal/taskhandlers ./pkg/jobs
make architecture-check
make security-check
make sqlc-check
```

The `integration`-tagged repository test uses `internal/testkit.NewPostgres`
and the full migration chain. It requires PostgreSQL 18 and covers tenant
isolation, concurrent digest claims, rollback after an intermediate board
write, tenant-safe board deletion and anonymous-contributor cleanup, and a
representative feedback-list `EXPLAIN` plan. Database-backed
`make sqlc-vet` runs only against an explicitly supplied disposable
migration-head database.

See [Feedback contributor delivery security](../security/feedback-deliveries.md)
for key separation, token reconstruction, rotation, and incident-response
rules.
