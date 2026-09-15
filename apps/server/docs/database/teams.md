# Teams persistence contract

The teams module owns all production team SQL under
`internal/modules/teams/repository/queries` and generates its pgx/v5 bindings
under `internal/modules/teams/repository/sqlc`. Only the handwritten repository
adapter imports the generated package. Services and HTTP handlers continue to
use the existing `teams.Core*` types and domain errors.

## Capabilities and file ownership

| Capability                                  | Reviewed SQL                   | Handwritten adapter |
| ------------------------------------------- | ------------------------------ | ------------------- |
| scoped list, public list, get               | `queries/teams.sql`            | `reads.go`          |
| create with defaults, update, delete        | `queries/teams.sql`            | `team_commands.go`  |
| add, join, leave, remove, member AI context | `queries/memberships.sql`      | `memberships.go`    |
| replace a user's team order                 | `queries/orderings.sql`        | `orderings.go`      |
| generated pgx implementation                | generated from all query files | `repository/sqlc`   |

Permanent deletion is composed by `internal/modules/teams/uow`. Team queries
remain in the team adapter; story deletion/events and attachment cleanup use
their owning repository capabilities on the same transaction.

Generated rows never become service or JSON contracts. `models.go` performs the
explicit conversion to `teams.CoreTeam`, including member counts and sprint
settings.

## Authorization and tenant rules

Team identifiers are globally unique, but that does not make an ID-only query
safe. Every externally scoped read or mutation binds the supplied workspace and
team in SQL. Actor-aware reads also prove that the actor is still active and is
still a workspace member at query time. This protects the database operation if
a cached middleware decision is stale.

- A team member may list and get teams they have joined.
- A workspace admin may list and get any team in that workspace unless the
  caller requests joined-only results.
- A public-team list is available only to an active member of that workspace and
  excludes teams the actor has already joined.
- Public self-join requires an active workspace membership and a non-private
  team in the same workspace. Ineligible, private, missing, and cross-workspace
  targets deliberately share `teams.ErrTeamNotFound`.
- Explicit member add requires the target user to be active and still belong to
  the team's workspace. A duplicate maps to `teams.ErrTeamMemberExists`.
- Self-leave accepts only the authenticated actor identity. Explicit removal is
  workspace-scoped and returns `teams.ErrTeamMemberNotFound` when no scoped
  membership was deleted.
- Member AI context can be changed only while the target remains an active
  member of both the team and its workspace.

These predicates are defense in depth. Route policies still decide whether the
actor may perform an administrative operation; repository scope prevents a
valid action in one workspace from mutating another.

## Transaction contracts

Creating a team is atomic. The repository inserts the team, its story automation
settings, and every default story status in one native pgx transaction. Any
missing default row or database error rolls the entire operation back. A unique
workspace/code violation maps at the adapter boundary to
`teams.ErrTeamCodeExists`.

Workspace bootstrap owns a wider transaction. It passes one `pgx.Tx` through
the transaction-specific workspace, teams, and users operations so the new
workspace, creator membership, default team, creator team membership, default
objective/team statuses, automation settings, user convenience pointer, and
workspace settings share one commit decision. The user convenience pointer is
best-effort, matching existing behavior; every required entity remains atomic.

Replacing custom ordering is also atomic. The repository first verifies the
actor's active workspace membership, deletes the old order, and inserts the new
positions. Every team must belong to that workspace and be visible to the actor
as a team member or workspace admin. A duplicate, inaccessible, or
cross-workspace team fails the operation and restores the previous order through
transaction rollback.

## Permanent team deletion

The authenticated settings route delegates to the configured deletion manager.
It accepts a first-party human actor and rechecks active workspace-admin
membership in SQL. Team, administrator, objective, feedback, story and candidate
attachment locks keep authorization, ownership and file references stable until
commit. Cross-workspace or inaccessible targets return the existing not-found
response without changing data.

Within one transaction the manager:

1. Locks the team and its dependent records, including archived and soft-deleted
   stories, and captures story, inline and feedback attachment IDs.
2. Removes document relationships and entity notifications while their target
   identifiers can still be resolved. Shared documents remain intact.
3. Deletes stories in batches and writes one durable `story.deleted` event per
   deleted story using the existing integration payload and authenticated actor.
4. Deletes the team and cascades its remaining owned records.
5. Retires only attachments with no remaining story, inline, document or
   feedback reference, and records physical object deletion in the existing
   durable outbox. Provider calls happen after commit through existing workers.

Any query, constraint, event or outbox failure rolls back the complete operation.
Missing stories in already-queued GitHub synchronization work complete without
retrying a permanently absent resource. The client retains failed confirmations,
blocks duplicate submissions while pending, and refreshes affected workspace
and credential caches after success.

Migration `000189` adds objective/label cascades, repairs historical cross-team
status references and enforces matching team/status ownership. It also preserves
surviving feedback merge sources as closed records when their target is deleted,
retains original durable merge history, and routes calendar cleanup to the
actual provider. Migration `000190` revokes and audits a machine credential when
its final team restriction disappears; removing a restriction cannot broaden
credential access. Other team restrictions and previously unrestricted
credentials retain their existing behavior.

Deploy the new migrations with the API/worker changes. Historical status repairs
and credential revocations are not reversed by rolling back schema definitions;
see `migration-operations.md` for rollout and recovery details.

## Error and empty-result semantics

| Database outcome                                | Repository outcome                            |
| ----------------------------------------------- | --------------------------------------------- |
| hidden, missing, or cross-workspace team row    | `teams.ErrTeamNotFound`                       |
| duplicate workspace/team code                   | `teams.ErrTeamCodeExists`                     |
| duplicate team membership                       | `teams.ErrTeamMemberExists`                   |
| missing scoped team membership                  | `teams.ErrTeamMemberNotFound`                 |
| zero-row list                                   | non-nil empty slice                           |
| infrastructure or unexpected constraint failure | wrapped error preserving the PostgreSQL cause |

Search text and pagination values are bound parameters. Positive limits and
offsets are checked before conversion to PostgreSQL `integer`; a non-positive
limit retains the existing unpaged API behavior.

## Verification

The ordinary repository tests use the generated `Querier` contract to verify
parameter mapping, generated-row conversion, domain error mapping, transaction
orchestration, and unsafe pagination rejection. SQL source contract tests retain
the security-critical workspace, actor, active-user, and membership predicates.

Integration tests use `internal/testkit.NewPostgres` with the `integration`
build tag. They create two real workspaces and prove cross-workspace reads,
member additions, public joins, removal, AI context changes, and ordering are
rejected without modifying data. Separate transaction tests prove required team
defaults commit together, caller rollback removes all team state, and a forced
default-status failure leaves no partial team, setting, or status rows.

Run from `apps/server`:

```bash
go test -race ./internal/modules/teams/repository
TEST_DATABASE_URL='postgresql://...' \
  go test -tags=integration -count=1 ./internal/modules/teams/repository
make sqlc-check
SQLC_DATABASE_URL='postgresql://...' make sqlc-vet
```
