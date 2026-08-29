# Sprint persistence and analytics

The sprint module uses SQLC-generated PostgreSQL queries and the native pgx/v5
pool. Its persistence boundary is intentionally small and navigable:

```text
internal/modules/sprints/
├── domain/                         sprint values, commands, validation, analytics math
├── service/                        use cases and the repository port
├── http/                           strict JSON DTOs and response mapping
└── repository/
    ├── queries/
    │   ├── reads.sql               actor-scoped list, single, and running reads
    │   ├── mutations.sql           authorization, CRUD, CAS, and audit persistence
    │   └── analytics.sql           breakdown, workweek, allocation, and burndown inputs
    ├── sqlc/                        generated code; never edit by hand
    ├── reads.go                    generated-row mapping and bounded pagination
    ├── mutations.go                pgx transactions and domain error mapping
    ├── analytics.go                concurrent independent reads and pure calculations
    └── mapping.go                  checked SQL-to-domain conversions
```

HTTP and service packages never import `repository/sqlc`. Generated parameter
and row types stop at the handwritten repository adapter. The service depends on
a narrow repository interface expressed entirely in sprint-domain types.

## Typed query surface

Every query has a static name, explicit columns, named bound parameters, and a
bounded result. Sprint filters are a finite `domain.ListFilter`; the legacy
in-process map is converted immediately and rejects unknown names. It can never
become a SQL column, operator, fragment, or order expression.

Sprint lists are ordered by `(end_date DESC, sprint_id DESC)`. The UUID tie-break
makes results deterministic when several sprints end on the same day. HTTP
offset pagination requests one extra bounded row to calculate `hasMore`; the
repository rejects negative offsets and limits outside `1..500`. Story summary
counts exclude archived and soft-deleted stories.

`Running` receives `today` from the service clock. PostgreSQL does not decide
the application date with `NOW()`, so tests and callers have one deterministic
time source. A sprint is returned as running only when its team's automatic
sprint setting is enabled and the actor can still access the team.

## Partial updates and concurrency

The JSON update DTO preserves three different states for every property:

| JSON state      | Domain meaning           | Example                         |
| --------------- | ------------------------ | ------------------------------- |
| omitted         | keep the stored value    | `{}`                            |
| concrete value  | replace the stored value | `{"goal":"Improve activation"}` |
| explicit `null` | clear a nullable value   | `{"goal":null}`                 |

`name`, `startDate`, and `endDate` cannot be cleared. `goal` and `objectiveId`
can. Dates are normalized to UTC calendar dates, the resulting interval is
validated after combining stored and patched values, and an objective must
belong to the sprint's exact workspace and team.

Clients that read `updatedAt` can send it as `expectedUpdatedAt`. The update
locks the sprint, compares the version, repeats the comparison in the `UPDATE`,
and returns a typed conflict when another writer won. Two concurrent updates
with the same version therefore produce exactly one success and one conflict.
Changing a date explicitly disables `schedule_managed_by_automation`; unrelated
updates preserve it.

## Transaction and audit contract

Create, update, and delete each use one native pgx transaction:

```text
authorize current actor and references
  -> lock the mutation target when present
  -> apply one typed mutation
  -> insert the sprint audit event
  -> commit
```

The audit row is not a best-effort side effect. If audit insertion fails, the
sprint mutation rolls back. No provider call, queue operation, or network call
runs inside this transaction. External developer events should be added through
the shared durable outbox, in the same transaction, rather than published
directly from this repository.

## Analytics query explained

The service first performs an actor-scoped sprint read, then executes four
independent typed reads concurrently:

1. workspace working days;
2. active-story status breakdown;
3. sprint scope and completion changes; and
4. active team-member allocation.

Every statement repeats live active-user, workspace-role, and sprint-team
membership checks. An authorization precondition is not treated as a durable
grant if membership changes while a request is running.

The burndown query is deliberately SQL-heavy because PostgreSQL is the source
of historical story state. Its named CTEs have these responsibilities:

| CTE                                     | Responsibility                                            |
| --------------------------------------- | --------------------------------------------------------- |
| `params`                                | one typed tenant, sprint, date-range, and actor input row |
| `authorized_sprint`                     | prove live sprint visibility before producing dates       |
| `date_series`                           | generate one calendar row for every sprint day            |
| `initial_stories_list`                  | reconstruct stories in scope at sprint start              |
| `daily_scope_changes`                   | count stories entering and leaving the sprint each day    |
| `stories_ever_in_sprint`                | bound later status history to relevant stories            |
| `daily_completion_changes`              | count transitions into and out of completed status        |
| `initial_completed_list`                | reconstruct completed stories at sprint start             |
| `initial_scope` / `initial_completions` | reduce the reconstructed sets to counts                   |

The query returns typed daily changes, not presentation calculations. Pure
functions in `domain/analytics.go` calculate remaining work, ideal burn,
completion percentage, elapsed workdays, remaining workdays, and health. This
separation keeps complex SQL reviewable and the product rules fast to unit test.

Legacy activity rows store historical IDs as JSON scalars or text. The SQL
supports both established representations and applies UUID casts only to values
expected by that historical contract. A future activity-schema migration should
replace this compatibility logic with typed columns rather than expanding it.

## Indexes and query plans

Migration `000167_add_sprint_query_indexes` adds:

- `idx_sprints_workspace_end_id` for tenant lists and stable ordering;
- `idx_sprints_workspace_team_end_id` for team-filtered navigation; and
- `idx_stories_workspace_sprint_status_active` for active sprint summaries.

The migration is reversible and does not change data. PostgreSQL 18 integration
tests force index-eligible plans and assert the list and summary paths use the
intended indexes. Do not edit migration `000167` after it has been applied;
create a new migration for later index changes.

## Verification

Run from `apps/server`:

```bash
make sqlc-generate
make sqlc-check
SQLC_DATABASE_URL='<disposable database at migration head>' make sqlc-vet
go test -race ./internal/modules/sprints/...
TEST_DATABASE_URL='<PostgreSQL 18 control URL>' \
  go test -race -tags=integration -count=1 ./internal/modules/sprints/repository
```

The integration suite uses the real migration chain in an isolated database and
proves tenant isolation, inactive/guest/outsider rejection, live revocation,
reference isolation, tri-state patches, concurrent CAS, atomic audit rollback,
analytics correctness, active-member allocation, and query-plan indexes.

See [Sprint authorization](../security/sprint-authorization.md) for the complete
access matrix and [Typed database access](sqlc.md) for repository-wide rules.
