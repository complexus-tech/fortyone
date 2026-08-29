# Team settings persistence and automation contract

Team settings are a tenant-scoped aggregate made of sprint automation, story
automation, and estimation settings. The module uses native pgx/v5 and checked-in
sqlc output exclusively. No HTTP model, service method, or worker receives a
generated sqlc type, and no team-settings repository file contains handwritten Go
SQL or SQLx access.

## Module map

```text
internal/modules/teamsettings/
├── domain/                 transport-neutral values and domain errors
├── service/                authorization, validation, ports, and use cases
├── http/                   request/response mapping and route policy
└── repository/
    ├── queries/            reviewed static SQL owned by this module
    ├── sqlc/               generated code; never edit by hand
    ├── repository.go       native pool and shared transaction runner
    ├── reads.go            scoped default initialization and reads
    ├── *_settings.go       typed mutations and immutable audit writes
    └── sprint_schedule.go  transactional sprint schedule reconciliation
```

The dependency direction is HTTP -> service -> domain. The repository imports
only the domain and generated query packages. Bootstrap supplies the concrete
repository and post-commit task scheduler. This prevents persistence or Asynq
details from becoming business or transport contracts.

## API and authorization policy

The API exposes these workspace routes:

| Method | Path suffix                                 | Minimum workspace role | Principal policy                                                      |
| ------ | ------------------------------------------- | ---------------------- | --------------------------------------------------------------------- |
| `GET`  | `/teams/{teamId}/settings`                  | member                 | human user, personal token, or delegated OAuth user with `teams:read` |
| `PUT`  | `/teams/{teamId}/settings/sprints`          | admin                  | human user with `teams:read`                                          |
| `PUT`  | `/teams/{teamId}/settings/story-automation` | admin                  | human user with `teams:read`                                          |
| `PUT`  | `/teams/{teamId}/settings/estimation`       | admin                  | human user with `teams:read`                                          |

The workspace middleware performs the first membership lookup, but the service
remains authoritative. It verifies the actor kind, selected workspace, scope,
credential-level team restriction, and current workspace role. A non-admin read
also performs a current database check that the actor is active and belongs to
the requested team. Workspace administrators may read any team in their current
workspace. Settings writes are intentionally human-admin-only until a separate,
reviewed integration mutation scope is introduced.

The repository independently binds every read and write to both `team_id` and
`workspace_id`, and joins the settings row back to the authoritative `teams`
row. A team ID paired with a second tenant is indistinguishable from a missing
resource and cannot create default settings in that tenant.

## Typed patch semantics

Update DTO fields are pointers because JSON must distinguish omission from a
literal `false`, `0`, or empty value. The HTTP mapper converts each pointer into
`PatchField[T]`, which carries a value and an explicit `Present` bit. Repository
code maps those bits to fixed sqlc parameters. SQL uses static `CASE` clauses;
it never constructs column names, fragments, or argument lists at runtime.

An entirely omitted patch is rejected. Domain validation constrains sprint
weekdays, durations, working-day uniqueness, upcoming counts, next generated
sprint number, story inactivity windows, and the supported `points`/`tshirt`
estimation schemes. PostgreSQL check violations are mapped to the same stable
domain errors as application validation.

## Transaction and audit boundary

Each settings mutation uses the repository's shared serializable pgx transaction
runner:

1. Initialize defaults from an authoritative, same-workspace team when absent.
2. Apply the typed update through a tenant-and-team-scoped sqlc query.
3. Reconcile managed sprint dates when an enabled cadence changes.
4. Insert an immutable `audit_events` row containing only safe changed-field or
   schedule metadata.
5. Commit all state together, or roll back settings, sprint dates, and audit data
   together.

PostgreSQL serialization failures and deadlocks become `ErrConcurrentUpdate`,
which the API reports as a retryable conflict. A custom future sprint that would
overlap an automation-managed schedule becomes `ErrSprintScheduleConflict`; the
entire settings transaction rolls back.

Sprint reconciliation locks the canonical settings row, the managed future
sprints it may update, and reads custom future sprint boundaries in the same
transaction. It retains an active sprint as the schedule anchor and only moves
sprints marked `schedule_managed_by_automation`. User-managed sprint dates are
never silently rewritten.

The schedule date deliberately comes from PostgreSQL `CURRENT_DATE`, once per
transaction. This preserves the existing database-calendar semantics and makes
API and worker instances agree even when process clocks or time zones differ.
Tests derive their fixture dates from that same database value instead of the
wall clock.

## Post-commit automation wake-up

Task dispatch is outside the repository transaction and occurs only after the
repository returns successfully. The service owns a narrow
`AutomationScheduler` port; bootstrap adapts the existing task service to it.
Updating settings never calls Asynq before commit.

The sprint creation, story auto-close, and story auto-archive jobs already scan
durable settings state on periodic schedules. An immediate enqueue is therefore
a coalescing wake-up, not the durable source of truth. If dispatch fails after a
commit, the service records a safe error and returns the committed settings
successfully; the periodic scan provides recovery. Reporting a rollback to the
client in that case would invite a duplicate mutation while the database had in
fact committed.

Sprint and story automation workers receive narrow SQLC-backed store
capabilities composed at worker startup. They do not share or bridge a
team-settings transaction; the committed settings row remains the durable
source of truth for their next bounded scan.

## Verification

Hermetic tests cover validation, literal zero/false patch presence, role and
principal policy, team restrictions, second-tenant denial, active membership,
SQLSTATE mapping, SQL query security clauses, and post-commit scheduling order.

PostgreSQL 18 integration tests use `internal/testkit.NewPostgres`, which creates
an isolated database at the complete migration head. They verify:

- default initialization and tenant/team scoping;
- active, inactive, other-team, and second-tenant membership results;
- typed zero/false updates plus immutable audit events;
- rollback of the setting, sprint date, and audit event on schedule conflict;
- deterministic concurrent serializable updates with one commit and one mapped
  retryable conflict.

Run the focused gates from `apps/server`:

```bash
go test -race ./internal/modules/teamsettings/...
TEST_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/test_control?sslmode=disable' \
  go test -tags=integration -race -count=1 ./internal/modules/teamsettings/repository
make sqlc-check
SQLC_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/sqlc_validation?sslmode=disable' \
  make sqlc-vet
```

`TEST_DATABASE_URL` is a disposable control database whose role can create and
drop isolated test databases. `SQLC_DATABASE_URL` is a separately migrated,
disposable planning database. Never point either command at production.
