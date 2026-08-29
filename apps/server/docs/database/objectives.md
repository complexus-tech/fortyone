# Objectives database guide

This guide explains how objective data is stored and changed. It is intended
for engineers who are new to Go, SQLC, or the FortyOne codebase.

## Where the code lives

| Concern                                | Location                                      |
| -------------------------------------- | --------------------------------------------- |
| Business types and invariants          | `internal/modules/objectives/domain`          |
| Use cases and actor/scope checks       | `internal/modules/objectives/service`         |
| SQLC queries and transaction ownership | `internal/modules/objectives/repository`      |
| Generated query code                   | `internal/modules/objectives/repository/sqlc` |
| JSON requests and responses            | `internal/modules/objectives/http`            |

Only the `queries` directory is handwritten SQL. Never edit files in `sqlc`;
run `./scripts/generate-sqlc.sh` after changing a query.

## Main tables

```text
workspaces
  ├─ teams
  │   ├─ objectives
  │   │   ├─ key_results
  │   │   │   └─ key_result_contributors
  │   │   ├─ okr_activities
  │   │   └─ strategy_objective_alignments
  │   ├─ team_objective_sequences
  │   └─ team_key_result_sequences
  ├─ objective_statuses
  ├─ workspace_strategies
  └─ strategic_pillars
```

Every application query supplies `workspace_id`, even when a foreign key
already implies the workspace. That repetition is deliberate: it makes tenant
scope visible at the call site and keeps a query safe if it is reused later.

## Reads

`ListObjectives` and `GetObjective` return the objective, story counts, key
result count, and schedule forecast in one typed result. The visibility CTE
first proves all of the following:

1. the objective belongs to the requested workspace;
2. the actor is a current, active workspace member;
3. the actor has the supported workspace role; and
4. the actor is a current member of the objective's team.

The remaining CTEs only join rows from that visible objective set. This avoids
calculating another tenant's statistics and then filtering at the end.

`LIMIT 0` means “no explicit limit” for the compatibility list endpoint. New
paginated callers request `pageSize + 1`; the extra row tells HTTP whether a
next page exists without a separate count query. Search is bounded to 200
characters and result size to 100 rows plus the look-ahead row.

Analytics queries repeat the tenant and current-membership predicate in every
statement. The initial `CanReadObjective` check provides a clear not-found
result, while the per-query checks prevent a membership revocation between the
check and a later analytics query from widening access.

## Create transaction

`Repository.Create` owns one database transaction for the full aggregate:

1. validate the typed command;
2. re-check active actor, workspace role, team membership, status, lead, and
   key-result assignees;
3. allocate one team-scoped objective sequence;
4. insert the objective;
5. allocate a contiguous key-result sequence range;
6. insert key results and de-duplicated contributors;
7. insert objective and key-result creation activities; and
8. commit.

Any error rolls back every step, including sequence allocation and activities.
The insert also repeats the actor authorization predicate so stale authorization
cannot be reused accidentally by a future caller.

The `team_id` and `sequence_id` columns on key results are required by the
current schema. They must always be populated together; SQLC makes omitting one
a compile-time error.

## Typed updates and nulls

HTTP update fields have three states:

| JSON             | Meaning                  |
| ---------------- | ------------------------ |
| field omitted    | keep the stored value    |
| `"field": value` | replace the stored value |
| `"field": null`  | clear a nullable field   |

A plain Go pointer cannot distinguish omission from explicit `null`, so HTTP
uses `PatchField[T]` and converts it to `domain.Field[T]`. The repository maps
each allowed field to a static SQLC parameter. It never interpolates a JSON key
or column name into SQL.

`name`, `statusId`, `isPrivate`, and `color` cannot be null. Description, short
summary, lead, dates, priority, and health may be cleared. Domain validation
also checks date order, color syntax, health values, and the database-backed
length limits for objective names, priority, short summaries, key-result names,
strategy goals, and pillar names.

An update transaction locks the objective row, compares `expectedUpdatedAt`
when the caller supplied it, validates referenced status/lead rows, applies the
static patch, and writes activities before commit. The update statement repeats
the actor, role, team, status, and lead checks so a concurrent revocation cannot
turn an earlier precondition into permission. `updated_at` always advances by
at least one microsecond, which makes compare-and-swap reliable even for two
updates in the same database clock tick.

## Strategy map

Strategy reads and writes also carry workspace and actor IDs. The strategy
shell is visible to active members and admins, not guests; objective alignments
are filtered again through objective team membership. Aligning an objective
requires the objective and pillar to be in the same workspace.

## Error mapping

The repository maps stable PostgreSQL outcomes to domain errors:

| Outcome                        | Domain error                         | HTTP status |
| ------------------------------ | ------------------------------------ | ----------- |
| invalid input/reference        | `ErrInvalid` / `ErrInvalidReference` | 400         |
| current actor is not allowed   | `ErrForbidden`                       | 403         |
| scoped row is absent or hidden | `ErrNotFound`                        | 404         |
| duplicate objective name       | `ErrNameExists`                      | 409         |
| stale `expectedUpdatedAt`      | `ErrVersionConflict`                 | 409         |

Unknown database failures remain internal errors; callers must not infer row
existence from a cross-workspace lookup.

## Verification

Run the focused checks from `apps/server`:

```sh
./scripts/generate-sqlc.sh
./scripts/check-sqlc.sh
go test ./internal/modules/objectives/...
go test -race -count=1 ./internal/modules/objectives/...
TEST_DATABASE_URL='postgres://…' go test -race -tags=integration -count=1 ./internal/modules/objectives/repository
```

Integration tests require a disposable PostgreSQL 18 control database and run
the complete migration chain before exercising transactions, tenant isolation,
authorization revocation, sequence concurrency, and stale-write conflicts.

## Intentional compatibility boundaries

- Old in-process callers may still pass a finite allowlisted map into the
  service. The service immediately converts it to a typed command; maps never
  reach persistence.
- Trusted background jobs may read one objective by workspace and ID without a
  request actor. The repository marks that path explicitly as internal and
  never exposes it through HTTP.
- Legacy Redis `objective.updated` publication happens after commit and is
  best-effort. Database activities are atomic, but Redis delivery is not a
  durable outbox. A future objective event migration should add a repository-
  owned outbox rather than publishing inside the transaction.
