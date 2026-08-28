# Reference and workflow persistence

This document defines the persistence and authorization contract for story
statuses, objective statuses, labels, story activities, and the currently
unsupported epics capability. It is the review companion for the SQL in each
module; generated sqlc files are implementation output and are never edited by
hand.

No schema migration was required for this slice. The existing `statuses`,
`objective_statuses`, `labels`, and `story_activities` tables already express
the durable data model. There is no `epics` table and no story-to-epic foreign
key.

## Module layout

Each supported module uses the same boundary:

```text
internal/modules/<module>/
├── domain/                 storage-independent models and typed errors
├── http/                   transport parsing and error/status mapping
├── service/                application-facing repository port
└── repository/
    ├── queries/*.sql       reviewed, static, tenant-scoped SQL
    ├── sqlc/               generated pgx/v5 code; never hand-edit
    └── repository.go       domain mapping and transaction orchestration
```

HTTP and service packages do not import generated types. Repositories accept a
native `*pgxpool.Pool`; these modules contain no production SQLx dependency and
no handwritten SQL string in Go.

All Go `int` values that cross into PostgreSQL `integer` or `smallint` columns
use `internal/platform/safecast`. An out-of-range page, limit, or order value is
rejected before a generated query runs, rather than wrapping during conversion.

## Authorization matrix

Authorization is repeated in SQL. Route middleware establishes the request
identity and workspace, but it is not treated as proof that a later database
operation remains authorized.

| Capability                            | Required live database state                                                        | Cross-tenant behavior           |
| ------------------------------------- | ----------------------------------------------------------------------------------- | ------------------------------- |
| List story statuses                   | active workspace membership and team membership                                     | empty list                      |
| Create/update/delete story status     | active workspace membership; target team/status must belong to the workspace        | generic not found               |
| List objective statuses over HTTP     | active workspace membership                                                         | empty list                      |
| Create/update/delete objective status | active workspace administrator                                                      | generic not found               |
| Label CRUD/list                       | active workspace membership; optional team must belong to the same workspace        | generic not found or empty list |
| Append story activity                 | active actor membership; live story and actor must belong to the supplied workspace | generic scope mismatch          |
| List story activities                 | active account and current membership in the supplied workspace                     | empty list                      |

Generic not-found outcomes deliberately do not reveal whether a foreign
workspace, inaccessible team, inactive account, or missing record exists.

Two read methods are intentionally trusted internal capabilities:

- `states.Service.Get` and `states.Service.TeamList` are used by
  already-authorized notification, agent-readiness, workspace bootstrap, seed,
  and Maya workflows. The public status-list handler uses `TeamListForMember`,
  never the trusted list method.
- `objectivestatus.Service.List` supports existing internal readiness workflows.
  HTTP uses `ListForMember`, which verifies current membership.

Keep those methods out of new request handlers. A new transport must accept an
actor and use the member-scoped operation.

## Status concurrency invariants

Status creation and deletion are multi-statement invariants and therefore run
inside native pgx transactions.

Story status creation:

1. Rechecks that the actor is active and the team belongs to the workspace.
2. Takes a transaction-scoped advisory lock for the workspace/category ordering
   sequence.
3. Takes a separate per-team default-status lock.
4. Clears the previous default when required, allocates the next order, and
   inserts the row in the same transaction.

Objective status creation follows the same pattern, with administrator
authorization and a workspace-wide default lock.

Deletion locks the target row and then the applicable team/category or
workspace/category key. It counts live attached stories/objectives and the
remaining statuses while holding that lock. This prevents two concurrent
deletes from both observing two rows and removing the last status in a category.
An attached record returns a typed conflict; deleting the final category status
also returns a typed conflict.

Status list order is deterministic: `order_index ASC`, then the status UUID.

## Labels

Label SQL is static. Search and team filtering use named sqlc arguments; no
column name or predicate is assembled from request input. A team filter includes
workspace-global labels (`team_id IS NULL`) and labels for the requested team,
matching the existing product behavior.

Offset pagination is retained for this internal endpoint and is bounded by the
shared HTTP pagination policy. SQL adds `label_id DESC` after `created_at DESC`
so equal timestamps cannot duplicate or omit rows solely because PostgreSQL
chooses a different physical order. New public API list endpoints should use the
cursor contract instead of copying this legacy offset contract.

## Story activities

Activities are immutable audit entries. Creation is one `INSERT ... SELECT`
whose predicates verify the live story, supplied workspace, active actor, and
current workspace membership atomically. Deleted stories and mismatched tenants
affect zero rows and return `ErrScopeMismatch`; callers cannot accidentally
append a valid-looking audit row for a foreign or deleted story.

Reads retain the current product contract of returning the signed-in user's own
activities. They use explicit start/end timestamps, a limit from 1 to 100, and
deterministic `created_at DESC, activity_id DESC` ordering. Inactive users and
removed memberships receive an empty list.

## Epics: explicit unsupported capability

Epics do not currently have a durable aggregate, story foreign key, or active
list consumer. The previous repository returned randomly generated placeholder
rows, which made non-existent data appear real and could not be tenant-scoped.
That behavior has been removed.

The authenticated internal route
`GET /workspaces/{workspaceSlug}/epics` now returns HTTP `501 Not Implemented`
through `web.RespondError`. Consequently it uses the normal internal error
envelope and does not expose the domain error text:

```json
{
  "data": null,
  "error": {
    "code": "internal_error",
    "message": "internal server error",
    "hint": "Retry only when the operation is safe to repeat.",
    "request_id": "<request correlation id>"
  }
}
```

The developer `/api/v1` surface does not expose an epics resource. Story reads
also reject an `epicId` filter because the database cannot enforce it. Adding
epics later requires an explicit aggregate and tenant/authorization design, a
new forward-only migration, SQLC queries, tests, and public API documentation;
do not reintroduce synthetic rows as a compatibility shortcut.

## Verification

Run from `apps/server`:

```bash
make sqlc-check

TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration -count=1 \
  ./internal/modules/states/repository \
  ./internal/modules/objectivestatus/repository \
  ./internal/modules/labels/repository \
  ./internal/modules/activities/repository

go test -race \
  ./internal/modules/states/... \
  ./internal/modules/objectivestatus/... \
  ./internal/modules/labels/... \
  ./internal/modules/activities/... \
  ./internal/modules/epics/...
```

The PostgreSQL tests apply the complete migration chain to an isolated database
and cover inactive/cross-workspace access, role checks, concurrent default/order
allocation, concurrent last-status deletion, concurrent immutable activity
appends, and deterministic same-timestamp label pagination.
