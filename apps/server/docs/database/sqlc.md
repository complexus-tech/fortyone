# Typed database access with sqlc

FortyOne's production persistence boundary is native pgx/v5 with
sqlc-generated, compile-time-typed queries for static application SQL. The SQLx
runtime and its compatibility wiring have been removed; SQLx is now a prohibited
production dependency. Each module owns a complete, tested persistence slice
behind narrow domain or service ports.

Representative typed slices include `internal/modules/links/repository`,
`internal/modules/comments/repository`, `internal/modules/invitations/repository`,
`internal/modules/stories/repository`,
`internal/modules/teams/repository`, `internal/modules/teamsettings/repository`,
`internal/modules/users/repository`, and `internal/modules/workspaces/repository`.
Maya realtime session quota and tool-call idempotency also use this boundary;
their workspace lock, tenant binding, and replay semantics are documented in
[Maya realtime persistence](maya-realtime.md).
Story statuses, objective statuses, labels, and story activities now use the
same boundary; their authorization, locking, pagination, and unsupported-epics
contracts are documented in
[Reference and workflow persistence](reference-workflows.md).
Sprints use the same boundary for actor-scoped reads, tri-state compare-and-swap
updates, atomic audit events, and typed analytics inputs. Their directory map,
burndown CTE guide, indexes, and PostgreSQL 18 verification are documented in
[Sprint persistence and analytics](sprints.md).
Key results and OKR activities use the same boundary for typed create/update
intent, active actor/workspace/team scope, concurrent per-team sequences,
atomic audit activities, and deterministic pagination. Their directory map,
security contract, transaction behavior, integration seam, and PostgreSQL 18
verification are documented in
[Key results and OKR activity persistence](key-results.md).
Notifications use the same boundary for finite types and channel patches,
idempotent preference-aware creation, live inbox/portal/delivery authorization,
and deterministic pagination. Their query ownership, durable delivery intent,
concurrency behavior, and PostgreSQL 18 verification are documented in
[Notification persistence and delivery](notifications.md).
Feedback uses the same boundary for explicit account, verified-guest, external,
and anonymous identities; current actor/workspace/team fences; atomic public
identity and moderation mutations; durable publication/merge/delivery state
machines; and typed digest and retention work. Its directory map, identity
matrix, transaction contracts, and PostgreSQL verification are documented in
[Feedback persistence and delivery](feedback.md).
Links establishes tenant-scoped CRUD; comments proves author ownership and a
service-owned pgx transaction; invitations proves versioned bearer digests,
row-locked single-use acceptance, and state-plus-outbox atomicity; teams proves
active-membership authorization, public/private team policy, typed ordering,
and a cross-module workspace bootstrap bound to one native pgx transaction;
users proves active-account visibility, tenant-scoped private data, atomic OAuth
identity resolution, and purpose-bound one-time verification codes. Workspaces
proves an authoritative cache-free membership boundary, typed access metadata,
atomic member cleanup, and a callback-scoped cross-module unit of work that
cannot expose a raw transaction. The shared
`internal/platform/idempotency/repository` slice proves canonical platform
persistence ownership, an absent-row advisory-lock invariant, stale lease
takeover, and generation-fenced completion without leaking generated types.

## Ownership and directory layout

Each persisted module owns its SQL, generated package, and handwritten adapter:

```text
internal/modules/comments/
├── service/                 domain models, errors, and repository port
└── repository/
    ├── queries/comments.sql reviewed PostgreSQL and sqlc annotations
    ├── sqlc/                generated-only Go package; never hand-edit
    ├── comments.go          adapter and transaction construction
    ├── commands.go          generated-query calls and domain error mapping
    └── models.go            generated-row to domain-model mapping
```

HTTP handlers and services must not import a generated package. Generated types
are an implementation detail below the repository boundary. This keeps schema
changes from becoming API contracts and keeps module dependencies understandable.

The root `sqlc.yaml` contains one explicit generation block per persisted module.
Global overrides make UUID, timestamp, date, and nullable shapes consistent. A
repository-owned validator enforces the complete pgx/v5 generation profile; a
new block cannot silently change null handling, empty-slice behavior, parameter
structs, interfaces, naming, or database-aware rules. Unsupported generators,
duplicate ownership, and generated output directories not declared in the
configuration are rejected.

Shared persistence capabilities use the same boundary rather than hiding SQL in
a service package:

```text
internal/platform/idempotency/
├── service.go
├── values.go
└── repository/
    ├── queries/receipts.sql
    ├── sqlc/
    └── repository.go
```

The architecture and sqlc-config gates allow this exact
`internal/platform/<capability>/repository/{queries,sqlc}` shape. They continue
to reject direct SQL elsewhere under `internal/platform`.

The first-party actor resolver uses this platform capability shape as well. Its
live active-system-user lookup and no-cache startup policy are documented in
[First-party system actor resolution](../security/system-actors.md).

The `internal/tools/sqlccontract` fixture compiles the type shapes on which
repository adapters rely, including nullable enums, numeric, JSONB, UUID arrays,
stable parameter structs, and non-nil empty results for `:many` queries.

## Pinned tool and supported commands

`tools/sqlc.lock` is the single source of truth for the sqlc version and official
Darwin/Linux amd64/arm64 archive checksums. The supported workflow never installs
`@latest` and never relies on whichever `sqlc` happens to be on `PATH`.

Run these commands from `apps/server`:

```bash
# Networked, explicit one-time installation into the ignored .tools directory.
make sqlc-bootstrap

# Intentionally update checked-in generated code.
make sqlc-generate

# Offline SQL compilation, clean temporary generation, drift comparison, and
# compile-only Go test across the API and worker.
make sqlc-check

# Database-aware query planning against a disposable DB at migration head.
SQLC_DATABASE_URL='postgresql://postgres:postgres@127.0.0.1:5432/fortyone_sqlc?sslmode=disable' \
  make sqlc-vet
```

`sqlc-generate` first generates and validates a complete temporary tree. It then
stages every output beside its destination and replaces the checked-in packages
with same-filesystem renames, retaining rollback copies until every replacement
succeeds. A generator error or interruption therefore leaves the last valid
checked-in output intact.

`sqlc-check` does not download tools or modify the working tree. The config
validator discovers every declared output, rejects duplicate ownership and
orphaned generated packages, checks every path component for symlinks, and
refuses to clean a directory containing a handwritten or unrecognized file.

`sqlc-vet` deliberately does not load `.env`. Its database URL must be provided
explicitly. Before sqlc plans any query, the wrapper verifies that
`schema_migrations` is clean and exactly matches the highest checked-in migration.
Use a disposable validation database, not a developer or production database.

## Writing a query

Use a stable, intention-revealing name and an annotation that reflects result
semantics:

```sql
-- name: UpdateLinkForWorkspace :execrows
UPDATE story_links AS link
SET title = COALESCE(CAST(sqlc.narg(title) AS text), link.title)
FROM stories AS story
WHERE link.link_id = sqlc.arg(link_id)
  AND story.id = link.story_id
  AND story.workspace_id = sqlc.arg(workspace_id);
```

Required query rules:

- List columns explicitly; do not use `SELECT *` or `RETURNING *`.
- Pass tenant identity explicitly and scope every workspace-owned read or write in
  SQL. Middleware checks are not a substitute for query scoping.
- Use `sqlc.arg(name)` for named parameters and `sqlc.narg(name)` only when SQL
  null has a deliberate meaning.
- Use `CAST(value AS type)` in application queries.
- Do not cast user input to `varchar(n)`: PostgreSQL truncates overlong values
  during an explicit constrained cast. Cast to `text` when sqlc needs an
  inference hint and let the destination column reject invalid lengths.
- Use `:execrows` when an affected-row count proves ownership, not-found, or an
  optimistic concurrency condition.
- Bound user-controlled arrays and page sizes before calling the repository.
- Convert Go `int`, `int64`, and `uint32` values to SQLC's PostgreSQL `int16`
  or `int32` types with `internal/platform/safecast`; never use an unchecked
  narrowing cast. Guard pagination multiplication before it can overflow, then
  reject out-of-range input instead of wrapping it into a different query.
- Keep dynamic identifiers behind typed allowlists; values always remain bound
  parameters.
- Do not record URLs, tokens, query payloads, or secret-bearing values in logs or
  tracing attributes.

sqlc establishes database-level types; it does not replace domain validation or
authorization. The adapter must translate `pgx.ErrNoRows`, affected-row results,
constraint errors, and generated values into service-owned outcomes.

`internal/platform/safecast` is the one shared integer-conversion boundary. Its
functions return a typed range error and have table plus fuzz tests at numeric
limits. Domain validation should normally keep values far inside those limits;
the repository conversion is still required as a final defense when a service
is bypassed, a configuration changes, or a platform-sized `int` reaches a
PostgreSQL-sized column.

## Adding or extending a persisted module

1. Identify a complete repository slice and its transaction boundary. Do not
   split one business invariant across transactions or persistence adapters;
   bind every statement to one `pgx.Tx`.
2. Add or update `internal/modules/<module>/repository/queries/*.sql`.
3. For a new persisted module, add one explicit, alphabetically placed block to
   `sqlc.yaml`, generating to `internal/modules/<module>/repository/sqlc` with a
   module-specific package.
4. Run `make sqlc-generate` and review generated signatures and nullability.
5. Wrap the generated methods in the module repository. Keep generated rows and
   params out of the service and HTTP packages.
6. Add adapter unit tests for parameter mapping, domain error mapping, and
   zero-row behavior.
7. Add PostgreSQL integration tests for success, cross-workspace rejection,
   transaction behavior, and nullable/edge values. Integration tests use the
   `integration` build tag and `internal/testkit.NewPostgres`. The testkit uses
   `TEST_DATABASE_URL` as a control connection, creates one isolated database,
   applies the real embedded migration chain, and registers deletion with
   `t.Cleanup`. The configured role therefore needs `CREATEDB` on disposable,
   non-production PostgreSQL. Tagged tests fail rather than silently skip when
   the variable or database is unavailable.
8. Run `make sqlc-check`, database-backed `make sqlc-vet`, focused tests,
   `make test`, and `go vet ./...`.

When a query is genuinely dynamic, first prefer a finite set of typed query
variants. Retain a small query builder only when the combinations are truly
unbounded and document why sqlc cannot express the operation safely. Do not use
SQLx as an escape hatch.

## Runtime connection ownership

The API, worker, and seed bootstraps each own one native `pgxpool.Pool` for their
process. Repositories receive that pool, and multi-statement invariants derive a
`pgx.Tx` from it through the shared transaction runner. Generated query sets bind
to the same transaction with `Queries.WithTx`; there is no SQLx compatibility
view, second application pool, or alternate production transaction API.

The native pool is bounded by `APP_DB_MAX_OPEN_CONNS`. Its minimum size, startup
timeout, idle lifetime, maximum lifetime, and health-check cadence use the
`APP_DB_MIN_CONNS`, `APP_DB_CONNECT_TIMEOUT`, `APP_DB_MAX_CONN_IDLE_TIME`,
`APP_DB_MAX_CONN_LIFETIME`, and `APP_DB_HEALTH_CHECK_PERIOD` settings. Process
shutdown drains and closes this one native pool after request or worker work has
stopped.

The only production connection opened through the `database/sql` API is the
short-lived handle returned by `internal/platform/database.OpenMigrationConnection`.
It exists solely because the golang-migrate PostgreSQL driver requires
`*sql.DB`, closes when migration startup finishes, and is never injected into an
application repository. The integration testkit uses a separate standard-library
control connection only to create a disposable database and apply the real
migration chain before opening the repository's native pgx pool.

Story comments are the reference state-plus-outbox transaction example. Create,
update, mention replacement, delete, developer-event persistence, and endpoint
fan-out use module-local generated queries bound to one `pgx.Tx`. The author
must still be an active workspace and story-team member, and credential team
restrictions can only narrow that access. If any requested user is not an active
member of the workspace, the repository returns a generic invalid-mention
outcome and the transaction preserves both the old content and old mention set.
Missing, cross-workspace, inaccessible, and non-author mutation targets share
the same not-found outcome. The full contract is documented in
[Story comment persistence and event contract](comments.md).

Stories use one module-owned generated package for statically typed reads,
primary mutations, secondary lifecycle and relationship mutations, comment-tree
reads, bounded story/sprint automation, retention, attachment-object deletion,
and the schedule-transition outbox. Services and jobs depend on stories-owned
ports and domain values; bootstrap owns cross-module adapters such as comment
creation. Automation runs capture one application UTC `as_of`, page with stable
ordering and hard limits, and execute each batch under a transaction-scoped
advisory lock. Auto-close activities and sprint-story migration activities plus
audit events must match the exact transition count before commit.

Interactive hard delete and the daily deleted-story retention job remove
unreferenced attachment metadata and enqueue a credential-free deletion request
in `attachment_object_deletion_outbox` in the same transaction as the story
delete. A separate bounded worker claims object deletions with `FOR UPDATE SKIP
LOCKED`, a lease, and a claim token; external storage calls occur only after the
database transaction commits. The full read and mutation boundaries are
documented in
[Stories read persistence](stories-read.md) and
[Story mutation persistence](stories-mutations.md).

Sprint auto-creation uses the same rules in the team-settings repository. The
job keyset-pages at most 100 teams per page and 100 pages against one UTC
`as_of`; each team runs in its own advisory-locked pgx transaction so schedule
reconciliation, new sprints, counters, and audit events commit or roll back
together. Concurrent counter updates receive a bounded retry rather than
silently losing work.

Teams are the reference membership and cross-module unit-of-work example. Team
reads require the actor to remain an active workspace member at query time, then
apply team-member or workspace-admin policy in the same statement. Public-team
self-join, explicit member add/remove, self-leave, and member AI context updates
all bind the team to the supplied workspace in SQL. User-defined ordering is
replaced transactionally: every referenced team must belong to the workspace and
be visible to the actor, otherwise the delete and every preceding insert roll
back together.

Workspace creation starts one pgx transaction and passes it only through the
transaction-specific workspace, teams, and users operations. The workspace,
creator membership, default team, creator team membership, objective/team
statuses, automation settings, and workspace settings therefore commit or roll
back as one unit. SQLx must not be reintroduced as a parallel transaction or
repository path.

The cross-module creation boundary is documented in
[Workspaces persistence and unit-of-work contract](workspaces.md). Team-specific
authorization and ordering are documented in
[Teams persistence contract](teams.md).

Team settings are the reference typed-patch and post-commit wake-up example.
Literal zero and false values carry explicit presence bits into static sqlc
updates. Settings, managed sprint dates, and immutable audit rows share one
serializable pgx transaction; Asynq scheduling occurs only after commit and is
recoverable through periodic durable-state scans. The complete contract is in
[Team settings persistence and automation contract](team-settings.md).

## Database TLS contract

Use `APP_DB_SSL_MODE=disable` only for a trusted local PostgreSQL instance. An
external database should use `APP_DB_SSL_MODE=verify-full`, which validates both
the certificate chain and database hostname. `APP_DB_SSL_ROOT_CERT` is optional:
leave it empty to use Go's operating-system trust store, set it to `system` to
request the same system-root behavior across supported pgx versions, or provide a
CA file path visible inside the API and worker runtime. Production private CAs
must be mounted through the managed deployment secret mechanism and exposed to
both processes at the configured path. Do not copy CA material into an image or
commit it to the repository.

The supported modes are `disable`, `require`, `verify-ca`, and `verify-full`.
Explicit `require` is rejected unless a root certificate is supplied, because
plain `require` encrypts traffic without authenticating the server. The legacy
`APP_DB_DISABLE_TLS` setting is read only when `APP_DB_SSL_MODE` is empty: `true`
maps to `disable`, while `false` now maps to authenticated `verify-full`.
The direct `make migrate-*` commands resolve `APP_DB_*` through the same Go
connection-string builder, including credential escaping, SSL-mode validation,
and private root certificates. The helper loads `.env` strictly and aborts on a
malformed or unreadable file instead of falling back to local defaults. `DB_URL`
remains an explicit expert override.

When `APP_ENVIRONMENT=production`, API, worker, and migration startup require
`APP_DB_SSL_MODE=verify-full`. The API and worker also reject disabled Redis
TLS. Redis certificate and hostname verification cannot be disabled. These
checks use `APP_ENVIRONMENT`, never an integration- or email-specific
environment setting.

## Test gates

`make test` runs the complete hermetic server test suite. Time-sensitive Maya
scheduling decisions use an injected clock, so the default suite does not
exclude tests or depend on the current wall-clock date.

The CI database gate uses PostgreSQL 18, applies every checked-in migration to
the sqlc validation database, runs `sqlc vet`, and then runs `go test
-tags=integration -count=1 ./...`. Every PostgreSQL integration test receives a
separate migrated database from the shared testkit, so parallel packages cannot
share mutable tenant state. The gate also runs generation drift checks, the
repository-boundary architecture test, the complete hermetic server suite, and
`go vet ./...`.
