# Workspaces persistence and unit-of-work contract

The workspaces slice is the tenant-entry boundary for the API. It owns typed
workspace, membership, access-metadata, and workspace-settings persistence and
coordinates the one transaction that creates the initial workspace graph.

All production SQL for this slice is SQLC-owned. Neither workspace HTTP code,
workspace service code, nor workspace middleware receives a database handle.

## Directory and capability ownership

```text
internal/modules/workspaces/
├── domain/                         persistence-neutral models and errors
├── http/                           request/response mapping and route policy
├── service/                        use cases split by business capability
├── repository/
│   ├── queries/
│   │   ├── workspaces.sql          workspace reads and lifecycle mutations
│   │   ├── memberships.sql         membership, role, and access metadata
│   │   └── settings.sql            workspace work-schedule settings
│   ├── sqlc/                       generated pgx/v5 code; never hand-edit
│   └── *.go                        generated-row mapping and error translation
└── uow/                            transaction-bound workspace bootstrap scope
```

The files in `service` and `http` are named by responsibility—creation,
lifecycle, memberships, settings, and logos. Adding an integration or workspace
behavior should not require navigating a single multi-thousand-line file.

| Capability                                                  | SQL source                | Adapter                  |
| ----------------------------------------------------------- | ------------------------- | ------------------------ |
| list/member/public workspace reads                          | `queries/workspaces.sql`  | `workspace_reads.go`     |
| create/update/delete/restore and default objective statuses | `queries/workspaces.sql`  | `workspace_mutations.go` |
| add/remove/demote and administrator email lookup            | `queries/memberships.sql` | `memberships.go`         |
| authoritative request membership and access timestamps      | `queries/memberships.sql` | `memberships.go`         |
| get, initialize, get-or-create, and update settings         | `queries/settings.sql`    | `settings.go`            |
| generated/domain conversion                                 | generated row types       | `mappers.go`             |

Generated SQLC types never cross the repository boundary. The service consumes
the persistence-neutral models in `workspaces/domain`, and HTTP maps those
models to the existing JSON DTOs.

## Current-membership authorization boundary

`mid.Workspace` depends on the narrow `mid.WorkspaceResolver` capability, not a
pool, SQLX handle, Redis client, or concrete repository. The bootstrap adapter
uses the workspace service to implement exactly two operations:

1. resolve the user's current workspace membership by slug;
2. record non-critical access metadata after authorization succeeds.

The resolve query joins the workspace membership to an active account every
time. It returns the current role, so a membership deletion or role demotion is
observed on the next request without a cache flush or TTL delay. Missing,
revoked, inactive, cross-tenant, and unknown-workspace cases all become the same
non-enumerating not-found response.

The middleware allows only user-shaped principals currently supported by the
compatibility authentication boundary. Service accounts, OAuth applications,
system users, and external principals are denied until their routes opt into an
explicit actor policy. An administrator role does not by itself make a machine
principal valid.

Soft deletion intentionally retains memberships during the recovery window.
This preserves scheduled deletion and restore behavior; hard purge removes the
graph through foreign-key cascades. Public workspace lookup and the unscoped
logo lookup exclude soft-deleted rows.

### Access metadata failure semantics

After current membership is resolved, one SQL statement updates both the active
user's `last_login_at`/warning state and the workspace's
`last_accessed_at`/warning state. A membership CTE gates both updates; a revoked
user cannot touch either record.

This is telemetry-like metadata, not an authorization condition. Middleware
gives it a maximum 250 ms child context, logs a safe error, and continues the
authorized request if it fails. The request never waits without a bound and
never turns a transient timestamp-write failure into an application outage.

## Tenant and mutation invariants

- Member workspace reads bind both workspace identity and user identity and
  require the account to remain active.
- List ordering is deterministic: `created_at`, then `workspace_id`.
- The fallback active workspace uses the same deterministic ordering.
- Every select and returning clause lists columns explicitly.
- Slug uniqueness is a database constraint. Unique violations map to
  `workspacedomain.ErrSlugTaken` without parsing PostgreSQL messages.
- Restricted product slugs are a service invariant and apply to HTTP callers
  and future integrations equally.
- Roles are validated against the typed `guest`, `member`, and `admin` set
  before persistence. Unknown roles fail closed.
- Adding, removing, and changing a member are administrator routes. Affected-row
  results distinguish a missing scoped membership.
- Removing a workspace member and that user's team memberships in this
  workspace is one repository transaction. A cleanup failure restores the
  workspace membership automatically.
- Settings use an atomic `INSERT ... ON CONFLICT ... RETURNING` get-or-create
  statement. Concurrent first reads cannot create duplicate state.
- Work-schedule day and minute ranges are validated before the typed update.

## Workspace creation unit of work

`workspaces/uow.Manager` owns the complete initial graph transaction:

```text
one pgx transaction
  ├── workspace
  ├── default objective statuses
  ├── creator workspace membership
  ├── default team
  ├── default team automation settings
  ├── default team story statuses
  ├── creator team membership
  ├── creator last-used-workspace pointer
  └── workspace settings
```

The manager begins one native `pgx.Tx`, then asks the workspace, teams, and users
repositories for their transaction-specific capability. Every repository binds
its generated query set with `Queries.WithTx(tx)`. The callback receives only a
`workspaces.Transaction`; it cannot commit, roll back, obtain the raw tx, or
reach a pool-backed sibling repository.

The capability scope is fresh for each callback. It serializes operations
because pgx transactions do not support concurrent query use. The scope closes
before commit and rejects every retained or late call with
`workspaceuow.ErrTransactionScopeClosed`. An accidentally retained reference
therefore fails deterministically instead of escaping the transaction.

Required graph operations preserve their domain error and roll back the entire
graph. The last-used-workspace pointer remains a documented convenience: its
failure is logged and does not invalidate an otherwise complete graph. Seed
stories and the trial-start task run after commit because service or queue calls
must not hold the database transaction open.

Those post-commit effects are not yet outbox-backed. A process crash after
commit can omit seed content or the trial notification. That is explicit
delivery debt under ADR 0003; network or queue work must not be moved into the
transaction as a workaround.

## Error mapping

| Database/result condition                                     | Domain outcome                                     |
| ------------------------------------------------------------- | -------------------------------------------------- |
| member-scoped row absent, inactive user, or cross-tenant slug | `workspacedomain.ErrNotFound`                      |
| duplicate workspace slug                                      | `workspacedomain.ErrSlugTaken`                     |
| duplicate workspace membership                                | `workspacedomain.ErrAlreadyWorkspaceMember`        |
| membership update/removal affects zero rows                   | `workspacedomain.ErrMemberNotFound`                |
| invalid role                                                  | `authorization.ErrInvalidWorkspaceRole` before SQL |
| retained unit-of-work capability                              | `workspaceuow.ErrTransactionScopeClosed`           |
| unexpected PostgreSQL/infrastructure failure                  | wrapped error retaining the cause                  |

HTTP error handling uses `errors.Is`; it does not compare message strings or
expose PostgreSQL details.

## Verification

Ordinary tests cover role/principal failure matrices, bounded access-write
failure behavior, SQLC adapter validation, constructor guards, and middleware
response contracts. PostgreSQL 18 integration tests prove:

- active same-tenant membership succeeds;
- inactive, unknown, revoked, and cross-tenant membership fails closed;
- a role demotion appears on the next resolve;
- access metadata changes only for a current member;
- member removal cannot delete another tenant's team membership;
- a forced team-cleanup error rolls back membership removal;
- a forced unit-of-work error removes every row in the initial graph;
- a successful unit of work commits every required row together;
- a retained transaction capability is unusable after the callback;
- concurrent creation of one slug yields one graph and one typed conflict.

Run from `apps/server` against a disposable PostgreSQL control database whose
role has `CREATEDB`:

```bash
go test -race ./internal/modules/workspaces/... ./internal/platform/http/middleware

TEST_DATABASE_URL='postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable' \
  go test -race -tags=integration -count=1 \
  ./internal/modules/workspaces/repository \
  ./internal/modules/workspaces/uow \
  ./internal/platform/http/middleware

make sqlc-check
SQLC_DATABASE_URL='postgresql://postgres:postgres@127.0.0.1:5432/fortyone_sqlc?sslmode=disable' \
  make sqlc-vet
```

The SQLC vet database must be disposable, fully migrated, clean, and exactly at
the checked-in migration head. Never point a test or vet run at production.
