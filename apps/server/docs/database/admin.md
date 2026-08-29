# Admin persistence

The admin repository owns all SQL for the internal administration surface. It
uses the native pgx pool and the generated `adminsql` query package; SQLx,
application SQL strings, named-parameter maps, and runtime SQL fragments are not
permitted.

## Tables read

- `users`, `workspaces`, `workspace_members`, `teams`, and `stories` provide
  account and workspace summaries.
- `workspace_subscriptions` supplies the most recently updated subscription
  projection for a workspace. Equal timestamps are resolved by
  `stripe_subscription_id`, so selection is deterministic.
- `slack_workspaces` and `github_installations` provide current integration
  counts and per-workspace installed flags.
- `admin_audit_logs` and `admin_notes` provide immutable operator history.

## Tables written

- `workspaces.trial_ends_on`, `deleted_at`, `deleted_by`, and `updated_at`;
- `users.is_active`, `is_internal`, and `updated_at`;
- append-only rows in `admin_notes` and `admin_audit_logs`.

There are no update or delete queries for admin audit rows or notes. New code
must preserve that append-only contract.

## Transaction ownership

Every repository method starts a short transaction and locks an active internal
actor row. A concurrent demotion or deactivation must wait for an already
authorized operation, while a request that begins after that change is denied.

Mutation ordering is deliberately consistent:

1. lock/recheck actor;
2. lock the target row;
3. validate the current stored state;
4. write the state change with an `updated_at` compare-and-swap predicate;
5. append one or more audit rows;
6. reload the response projection;
7. commit.

An error in steps 3–6 rolls the state change back. User-state operations lock
the actor and target user in UUID order, preventing cross-demotion deadlocks and
making the self-mutation decision atomic.

## Typed filters

The SQL never concatenates a `WHERE` clause. One static query receives a typed,
finite filter value:

- workspace status: `active`, `trialing`, `expired`, `expiring`, `paid`,
  `past_due`, or `deleted`;
- audit target: `workspace`, `user`, `subscription`, or `system`;
- audit action: the action constants in `internal/modules/admin/domain`;
- note target: `workspace` or `user`.

Search text remains a value parameter used by `ILIKE`; it never becomes SQL
syntax.

## Pagination and ordering

The existing first-party admin HTTP contract uses bounded offset pagination,
which is permitted for this internal surface. The shared pagination primitive
caps the page size at 100, and checked conversions reject offsets that cannot be
represented by PostgreSQL integer parameters.

Every list has a deterministic unique order:

- workspaces: `created_at DESC, workspace_id DESC`;
- users: `created_at DESC, user_id DESC`;
- audit: `created_at DESC, id DESC`;
- notes: `created_at DESC, id DESC`;
- memberships and members include their resource UUID as the final tie-breaker.

Migration `000168_admin_stable_pagination_indexes` adds matching composite
indexes for the four externally paged lists. The timestamp remains the business
sort key; the UUID is the unique tie-breaker that prevents equal timestamps
from moving records between pages:

| List       | Index                                                                  |
| ---------- | ---------------------------------------------------------------------- |
| users      | `idx_users_admin_created_id (created_at DESC, user_id DESC)`           |
| workspaces | `idx_workspaces_admin_created_id (created_at DESC, workspace_id DESC)` |
| audit logs | `idx_admin_audit_logs_created_id (created_at DESC, id DESC)`           |
| notes      | `idx_admin_notes_created_id (created_at DESC, id DESC)`                |

The PostgreSQL integration test disables sequential scans for its plan probe
and requires each representative ordering query to name the corresponding
index. This makes an accidental ordering/index drift a test failure instead of
an undocumented performance regression.

## JSON audit values

Audit old/new values and metadata are marshaled once through `encoding/json`.
SQLC passes the resulting bytes to PostgreSQL with `CAST(value AS jsonb)`. Reads
decode valid JSON into the legacy API's JSON value shape. Invalid stored JSON is
treated as a persistence error rather than silently returned as a different
type.

## Stripe boundary

Stripe synchronization cannot share the admin repository transaction. The
repository therefore records request and result audit facts in separate
transactions around the network-backed subscription capability. See
`docs/modules/admin.md` for the exact failure semantics.

## Verification

Repository integration tests use the fully migrated PostgreSQL 18 testkit and
cover authorization revocation, rollback after forced audit failure, concurrent
user updates, stable pagination, filters, integration projections, and
representative `EXPLAIN` output. SQLC compile and database-backed vet validate
the named queries against the real migration chain.

Run the focused gates from `apps/server`:

```sh
go test -race -count=1 ./internal/modules/admin/...
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration -count=1 ./internal/modules/admin/repository
SQLC_DATABASE_URL='postgresql://<read-only-schema-role>@<postgres-18-host>/<database>?sslmode=disable' \
  make sqlc-vet
```

Both database URLs must point to approved disposable or validation
infrastructure. Never aim these commands at production.
