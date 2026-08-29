# Key results and OKR activity persistence

This guide explains where key-result code lives, which layer owns each
decision, and how the SQLC persistence boundary protects tenant data. It is
written for engineers who are new to Go, SQLC, or FortyOne.

## Where to start

| Concern                                             | Location                                         |
| --------------------------------------------------- | ------------------------------------------------ |
| Business values, validation, and typed commands     | `internal/modules/keyresults/domain`             |
| Actor binding and key-result use cases              | `internal/modules/keyresults/service`            |
| SQLC adapter and pgx transactions                   | `internal/modules/keyresults/repository`         |
| Handwritten key-result SQL                          | `internal/modules/keyresults/repository/queries` |
| Generated key-result code                           | `internal/modules/keyresults/repository/sqlc`    |
| HTTP decoding, response mapping, and status mapping | `internal/modules/keyresults/http`               |
| OKR audit values and list commands                  | `internal/modules/okractivities/domain`          |
| OKR audit actor binding                             | `internal/modules/okractivities/service`         |
| Handwritten OKR audit SQL and pgx adapter           | `internal/modules/okractivities/repository`      |

The `sqlc` directories are generated implementation details. Do not import
them from HTTP or service code and never edit their contents by hand. Change a
reviewed query under `queries`, run the supported SQLC generator, and review the
generated diff.

## Request and dependency flow

```text
HTTP, Maya, or email-reply caller
  -> key-results service (caller-owned repository/event ports)
    -> key-results domain (validation and typed intent)
      -> key-results repository (transaction and error mapping)
        -> generated SQLC queries
          -> PostgreSQL

objective/key-result use case
  -> OKR activities service
    -> OKR activities repository
      -> generated SQLC queries
```

The service owns the small interfaces it consumes. The concrete pgx
repositories are assembled only in bootstrap. This keeps SQLC types below the
repository boundary and lets another caller or integration provide a focused
adapter without changing business logic.

## Data relationships

```text
workspace
  -> team
    -> objective
      -> key_result
        -> key_result_contributors
        -> okr_activities
    -> team_key_result_sequences
```

A key result belongs to the same team as its objective. A lead or contributor
must be an active user who is currently a member of both the workspace and that
team. `team_key_result_sequences` allocates human-readable sequence numbers per
workspace and team.

## Authorization contract

Authorization is deliberately repeated at two layers:

1. The service reads the authenticated `platform/auth.Actor`, binds it to the
   requested workspace, checks the required objective read/write scope, and
   converts credential team restrictions into `domain.AccessScope`.
2. Every repository query proves current active-user, workspace-membership,
   resource-workspace, team-membership, and allowed-team predicates in SQL.

Middleware is useful request context, but it is not the database security
boundary. A user whose membership is removed or account is disabled stops
matching the SQL immediately. A restricted developer credential can only
narrow reads and writes to its explicit team set. Missing, inaccessible, and
cross-workspace mutation targets return the same not-found result where
revealing existence would leak information.

Callers must not accept a workspace or actor from a JSON payload. HTTP binds
both values from authenticated middleware. Compatibility methods that accept a
user ID verify that it is the same principal as the actor in context.

## Typed create and update intent

Create commands accept at most 20 key results. One batch must use one objective
and one authenticated creator. Domain normalization trims names, validates the
measurement enum, rejects NaN and infinity, normalizes dates to UTC calendar
days, requires `end_date >= start_date`, and de-duplicates contributors.

Updates use `domain.Patch`, not `map[string]any`. Every field carries a presence
bit:

```go
patch := keyresults.KeyResultPatch{
    CurrentValue: keyresults.SetField(0),
    Lead:         keyresults.ClearField[uuid.UUID](),
}
```

This distinguishes an omitted field from an intentional zero, empty list, or
nullable lead. Start and end dates cannot be cleared because the schema makes
them required. Unsupported fields cannot reach SQL because there is no patch
member or generated parameter for them.

`UpdateExternalUserActionIfUnchanged` adds compare-and-swap behavior for Maya
and other asynchronous user actions. It locks the row, compares the supplied
`updated_at`, and returns `ErrVersionConflict` if another writer won. A retry is
accepted as idempotent only when the requested patch is already the persisted
state. Contributor replacement is excluded from that compatibility path
because list equality cannot safely infer the intent of a stale external
action.

## Transaction boundaries

### Create

One native pgx transaction:

1. proves the actor can write the objective and resolves its team;
2. validates every lead and contributor before sequence allocation;
3. atomically allocates a contiguous sequence range using an upsert;
4. inserts each key result and its de-duplicated contributors; and
5. inserts the corresponding OKR create activity.

Any error rolls back the rows, contributors, activities, and sequence update.
Concurrent creators receive unique sequence numbers.

### Update

One native pgx transaction locks the authorized key result with `FOR UPDATE`,
checks an optional expected version, validates effective dates and changed
assignees, applies the static typed update, replaces contributors when present,
and writes one OKR activity per changed field. A no-op patch commits no update
and no activity. An audit failure rolls back the data change.

### Delete

One native pgx transaction locks the authorized key result, deletes it, and
writes a delete activity against the surviving objective. The activity stores
the deleted name because the key-result foreign key is necessarily absent. An
audit failure restores the key result.

The repository owns all transaction control. Services never expose `pgx.Tx`,
and SQLx is not used or bridged inside these modules.

## Reads and pagination

Objective lists preserve the existing newest-first behavior and add the ID as a
deterministic tie-breaker. The filterable endpoint accepts only these sort keys:

- `name`
- `created_at`
- `updated_at`
- `objective_name`

Direction is limited to `asc` or `desc`; invalid values fall back to the
documented default. SQL uses finite `CASE` branches rather than interpolating
identifiers. Every branch ends with the key-result ID as a stable tie-breaker.

Page size is bounded to 100. Offset multiplication is checked before narrowing
to PostgreSQL `int32`, and database counts are checked before conversion to Go
`int`. HTTP rejects malformed UUID/date filters with 400 instead of silently
dropping them and widening a query.

OKR activity lists use `pageSize + 1` look-ahead pagination and order by
`created_at DESC, activity_id DESC`. Objective and key-result activity reads
repeat active actor, workspace membership, team membership, and resource scope
inside SQL. The legacy activity API cannot yet express a restricted credential
team set, so the service fails those credentials closed instead of broadening
access.

## Events and integrations

Key-result lead, contributor, target, and end-date changes can notify the
existing internal publisher through the service-owned `EventPublisher`
interface. Publication runs only after the database transaction commits, so a
rolled-back mutation never emits an event.

This publisher call is currently best effort. A publish failure is logged and
does not roll back committed data. It is **not** a durable integration or
webhook guarantee. Any future external integration that requires at-least-once
delivery must write a module event or outbox row in the same transaction, then
deliver it asynchronously through the platform integration boundary. Do not
hide provider-specific Slack, GitHub, GitLab, or custom-integration logic in
the key-result service.

## Error behavior

| Domain error          | Meaning                                              | HTTP behavior |
| --------------------- | ---------------------------------------------------- | ------------- |
| `ErrInvalid`          | Malformed value, patch, filter, date, or pagination  | 400           |
| `ErrForbidden`        | Invalid actor/scope or inaccessible create objective | 403           |
| `ErrNotFound`         | Missing or inaccessible existing key result          | 404           |
| `ErrInvalidReference` | Lead/contributor/resource relationship is invalid    | 400           |
| `ErrVersionConflict`  | Compare-and-swap version is stale                    | 409           |

Unexpected database or infrastructure errors remain internal errors. SQL and
repository errors must never echo query text, tokens, or credentials to an API
response.

## Verification

Run from `apps/server`:

```bash
go test -race ./internal/modules/keyresults/... ./internal/modules/okractivities/...

TEST_DATABASE_URL='postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable' \
  go test -race -tags=integration \
  ./internal/modules/keyresults/repository \
  ./internal/modules/okractivities/repository

make sqlc-check

SQLC_DATABASE_URL='postgresql://postgres:postgres@127.0.0.1:5432/fortyone_sqlc?sslmode=disable' \
  make sqlc-vet
```

The integration tests require disposable PostgreSQL 18 and create their own
migrated database through `internal/testkit.NewPostgres`. They cover successful
CRUD, stable pagination, tenant/team/member denial, inactive actors, mismatched
objective/key-result relationships, transaction rollback, concurrent sequence
allocation, and optimistic concurrency.

## Adding or changing a field

1. Decide whether the field is required, nullable, or clearable in the domain.
2. Add a typed domain member and validation; never introduce a generic CRUD map.
3. Update explicit SQL projections and static update parameters.
4. Regenerate SQLC and map generated values only inside the repository.
5. Add domain, service, query-contract, HTTP, and PostgreSQL integration tests
   for zero/null intent, tenant negatives, rollback, and ordering as relevant.
6. Update this guide when the invariant, event behavior, or transaction boundary
   changes.
