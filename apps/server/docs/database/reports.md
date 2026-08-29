# Reports persistence, authorization, and query contract

The reports module uses module-owned SQLC queries and the shared native
`pgx/v5` pool. It has no handwritten SQL builder, no `fmt.Sprintf` query
assembly, and no SQLx dependency. Generated SQLC code is checked in under the
repository because its parameter and result types are part of the reviewed
database contract.

This guide is the source of truth for navigating reports, understanding each
query's business grain, preserving tenant and private-team boundaries, and
reviewing query plans as data volume grows.

## Module navigation

```text
internal/modules/reports/
├── domain/                         dependency-neutral report types and errors
│   ├── filters.go                  common filter and analytics-event inputs
│   ├── statistics.go               personal and grouped statistics
│   ├── workload.go                 member/team workload results
│   ├── pulse.go                    pulse health and risk types
│   └── <capability>.go             one model file per report capability
├── http/
│   ├── routes.go                   authenticated workspace routes
│   ├── handlers.go                 handler construction and shared helpers
│   ├── filters.go                  fail-closed query parameter parsing
│   ├── <capability>_handler.go     request orchestration by capability
│   └── <capability>_models.go      transport DTOs and domain conversion
├── service/
│   ├── reports.go                  public service contract and use cases
│   ├── filters.go                  actor binding, deduplication, and bounds
│   ├── command_center.go           concurrent composite report assembly
│   ├── pulse.go                    pulse assembly and risk derivation
│   └── models.go                   compatibility aliases to domain types
└── repository/
    ├── queries/
    │   ├── authorization.sql        actor role and effective team scope
    │   ├── statistics.sql           personal/status/priority statistics
    │   ├── workload.sql             workload aggregates
    │   ├── pulse.sql                health aggregates
    │   └── <capability>.sql          reviewed static SQL by capability
    ├── sqlc/                         generated code; never edit by hand
    ├── repository.go                 pgx/SQLC construction and actor access
    ├── filters.go                    effective team-scope resolution
    └── <capability>.go               generated-row to domain mapping
```

Dependencies point inward:

```text
HTTP -> service -> domain <- repository -> repository/sqlc
```

The repository must never import `reports/service`, and generated SQLC types
must never escape the repository. HTTP DTOs stay at the transport boundary.
This lets internal callers such as Slack, Maya, workers, and future integration
adapters use the same service without depending on HTTP or database shapes.

## Request and authorization flow

Every request follows this sequence:

1. authentication middleware resolves the user;
2. workspace middleware resolves the workspace from its slug;
3. HTTP parsing rejects malformed UUID or date filters;
4. the service binds `ActorID` from the trusted authentication context;
5. the service deduplicates identifiers and enforces input bounds;
6. the repository reloads the actor's current role from PostgreSQL;
7. the repository resolves the actor's effective team scope;
8. static SQLC queries bind typed parameters;
9. handwritten adapters map generated rows into domain models; and
10. required projections fail closed if a database invariant is broken.

Middleware is not the final authorization boundary. The repository repeats the
current actor, workspace, role, account-state, workspace-state, and team-scope
checks so a worker or integration caller cannot bypass the policy by wiring the
service incorrectly.

## Workspace role policy

Workspace analytics are available only to current `member` and `admin`
memberships.

| Actor state                        | Result                              |
| ---------------------------------- | ----------------------------------- |
| Active admin in a live workspace   | Allowed across the workspace        |
| Active member in a live workspace  | Allowed within effective team scope |
| Guest                              | Denied                              |
| System user                        | Denied                              |
| Inactive user                      | Denied                              |
| Missing/cross-workspace membership | Denied                              |
| Deleted workspace                  | Denied                              |

The role comes from `workspace_members`; it is never trusted from an access
token claim or request parameter. Unknown roles fail closed.

### Effective team scope

An admin can see every team in the workspace. A member can see:

- every public team in the workspace; and
- a private team only when a matching `team_members` row exists.

For an explicit `teamIds` filter, every requested team must be visible to the
actor and belong to the current workspace. If even one requested team is
cross-tenant or is an unauthorized private team, the whole report returns
`ErrReportsAccessDenied`; the query is not broadened or partially executed.

An empty team filter is role-sensitive:

- for admins, the empty SQL array intentionally means all workspace teams;
- for members, the repository replaces it with the explicit list of visible
  public and joined-private teams; and
- if a member has no visible teams, the repository uses a deliberate
  non-matching UUID scope because an empty SQL array would otherwise mean all
  teams.

This distinction is security-critical. Do not remove the non-matching scope or
pass an unvalidated caller slice directly to a report query.

Personal story, contribution, and user statistics are tied to the authenticated
actor. Status and priority statistics additionally intersect the selected team
with the actor's team memberships. All of these entry points still apply the
same guest, inactive-account, and workspace checks.

### Analytics event writes

`CreateWorkspaceAnalyticsEvent` uses an `actor_access` CTE and an
`authorized_input` CTE. The insert produces one row only when:

- the actor is an active, non-system member or admin in a live workspace;
- every supplied team/story/objective/sprint/key-result belongs to that
  workspace; and
- every team-backed entity is public, belongs to the member, or is being
  referenced by an admin.

The repository checks affected rows and maps zero rows to access denied. Event
properties are JSON-encoded before SQL and limited to 64 KiB. Event names are
trimmed and limited to 120 bytes; surfaces are limited to 80 bytes. No token,
secret, raw query, or properties payload is logged.

## Input bounds and wire behavior

| Input                                 | Rule                             |
| ------------------------------------- | -------------------------------- |
| Team, assignee, sprint, objective IDs | At most 100 per dimension        |
| Duplicate IDs                         | Removed before repository access |
| Nil UUID                              | Rejected                         |
| Malformed UUID text                   | HTTP 400; never silently ignored |
| Required report range                 | Both start and end must exist    |
| Optional workload/pulse range         | Both absent, or both present     |
| Maximum range                         | 366 days                         |
| Reversed/zero range                   | Rejected                         |
| Analytics properties                  | Valid JSON, at most 64 KiB       |

`parseReportFilters` defaults date-bound workspace reports to the prior 60 days.
Workload and pulse allow both dates to be absent so they can describe current
open work. The service remains authoritative: internal callers receive the same
identifier and date checks even when HTTP parsing is not involved.

## Query inventory and business grain

All query text lives in `repository/queries`. Values use `sqlc.arg` or
`sqlc.narg`; filters use typed UUID arrays and timestamps. A zero-length UUID
array has the deliberate meaning documented in the effective-team section.

| Capability           | SQLC operations                                                                                                                        | Output grain and time basis                                      |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| Authorization        | `GetReportsActorAccess`, `ListReportsVisibleTeamIDs`                                                                                   | One actor role; then one visible team per row                    |
| Personal statistics  | `GetStoryStats`, `GetContributionStats`, `GetUserStats`                                                                                | One aggregate, one calendar day, one aggregate; actor-owned data |
| Status/priority      | `GetStatusStats`, `GetPriorityStats`                                                                                                   | One status or priority bucket; story `created_at` range          |
| Workspace overview   | `GetWorkspaceMetrics`, `ListWorkspaceCompletionTrend`, `ListWorkspaceVelocityTrend`                                                    | One aggregate, one created week, one completed/updated week      |
| Story analytics      | `ListStoryStatusBreakdown`, `ListStoryPriorityDistribution`, `ListStoryCompletionByTeam`, `ListStoryBurndown`                          | One status/team, priority, team, or completion date              |
| Objective progress   | `ListObjectiveHealthDistribution`, `ListObjectiveStatusBreakdown`, `ListKeyResultProgress`, `ListObjectiveProgressByTeam`              | One health/status/objective/team; objective `created_at` range   |
| Team performance     | `ListTeamWorkload`, `ListMemberContributions`, `ListTeamVelocity`, `ListWorkloadTrend`                                                 | One team, member/team, team, or created date                     |
| Workload             | `GetWorkloadSummary`, `ListMemberWorkload`, `ListTeamWorkloadSummary`, `GetUnassignedWorkload`                                         | One aggregate, member, team, one aggregate                       |
| Pulse                | `GetPulseStoryHealth`, `GetPulseSprintHealth`, `GetPulseObjectiveHealth`, `GetPulseRequestHealth`                                      | One aggregate for each health domain                             |
| Sprint analytics     | `ListSprintProgress`, `ListCombinedSprintBurndown`, `ListSprintTeamAllocation`, `ListSprintHealth`                                     | One sprint, date, team, or status                                |
| Timeline             | `ListStoryCompletionTimeline`, `ListObjectiveProgressTimeline`, `ListTeamVelocityTimeline`, `ListKeyMetricsTimeline`                   | One day, with team added for velocity                            |
| Integration requests | `ListRequestSourcePerformance`                                                                                                         | One provider                                                     |
| Engagement           | `GetWorkspaceEngagementTotals`, `ListWorkspaceEngagementByName`, `ListWorkspaceEngagementBySurface`, `ListWorkspaceEngagementTopUsers` | One aggregate, event name, surface, or active user               |
| Event write          | `CreateWorkspaceAnalyticsEvent`                                                                                                        | Exactly zero or one inserted event                               |

### Status and time semantics

Story status behavior comes from `statuses.category`, not display names.
`completed` and `cancelled` are terminal for open-work calculations. `started`
and `paused` feed their corresponding health counters. A missing category is
treated as open, matching the legacy behavior.

Most range-bound inventory queries use entity `created_at`. Completion velocity
uses story `updated_at`, because the schema does not currently expose a distinct
completed timestamp. That means reopening or editing a completed story can move
its apparent completion period. This is an intentional compatibility constraint
and must be revisited explicitly if a completion-event source becomes canonical.

Workload counts current non-deleted, non-archived, non-draft stories. Dates,
when supplied, constrain story creation. Overdue means an open story whose
`end_date` is before the database's current date. Estimate totals use the
stored `estimate_unit`; absent estimates are counted separately.

Team velocity retains the existing compatibility shape: three placeholder week
fields and an average over the reviewed three-week window. Do not reinterpret
those fields without a versioned API decision and frontend coordination.

## CTE review notes

CTEs are used where they make grain and authorization explicit:

- `GetContributionStats`: `dates` creates the requested daily series and
  `activity_counts` aggregates actor activity before the left join. This keeps
  zero-contribution days in the response.
- Workload queries: `workload_stories` establishes one filtered story per row
  before member, team, or summary aggregation. Do not join one-to-many tables
  inside it without restoring story uniqueness.
- `GetPulseSprintHealth`: `sprint_scope` establishes one row per sprint and
  computes open/overdue/unestimated story counts before the final health
  aggregation.
- `CreateWorkspaceAnalyticsEvent`: `actor_access` proves role and account state;
  `authorized_input` proves every optional resource before the insert.

When adding a CTE, document its row grain in the SQL review. PostgreSQL may
inline a non-materialized CTE; use `MATERIALIZED` only with plan evidence.

## Nullability and fail-closed mapping

SQLC exposes database nullability instead of hiding it in a handwritten scan.
The contract is:

- aggregate counts use `CAST` and `COALESCE` so domain counts are concrete;
- optional dates and identifiers remain pointers;
- optional display fields use `COALESCE` only where the API historically
  returns an empty string;
- empty list results are allocated as empty slices, not `nil` JSON values; and
- required projected UUIDs are checked before domain construction.

Sprint IDs, team IDs, engagement user IDs, and other required identity
projections return `ErrInvalidProjection` when zero. The mapper never invents a
UUID or silently drops a malformed row. A new nullable generated field must be
classified as required or optional during review; do not dereference it merely
to satisfy the compiler.

## Composite report consistency

The command-center report runs independent report sections concurrently to
control latency. It is a best-effort operational snapshot, not a
transactionally consistent database snapshot. Each section may observe a
slightly different committed state.

Non-context section failures are returned in `sectionErrors` and the failed
section keeps an explicit empty shape. Client-visible section messages are
sanitized; the detailed cause remains in server logs and tracing. An access
denial is never treated as partial success: it fails the entire command-center
request. Context cancellation also aborts the whole report. This partial-result
behavior is intentional for ordinary section failures; single report endpoints
still return their error normally.

If a future consumer requires one exact point-in-time snapshot, introduce a
read-only repeatable-read unit of work as a separate contract rather than
quietly changing the current latency and failure behavior.

## Index coverage and watch list

The migration chain currently supplies the main access-path indexes:

- `workspace_members (workspace_id, user_id)` and its primary key support
  actor-role checks;
- `team_members (team_id, user_id)` and user indexes support private-team scope;
- teams have a workspace/created index and a unique workspace/code index;
- stories have workspace, workspace/team, partial live workspace/team,
  assignee/sprint, status, objective, and completed-update indexes;
- objectives have `(workspace_id, team_id)`;
- integration requests have `(workspace_id, team_id, status, created_at DESC)`
  and `(workspace_id, status, created_at DESC)`; and
- analytics events have workspace/time, event/time, surface/time, user/time,
  and JSONB properties indexes.

Known plan watch points are range scans combining workspace/team with
`created_at`, sprint workspace/team filtering, objective date filtering, and
large `ANY(uuid[])` scopes. Sprints currently rely heavily on their primary key
and filtering. Do not add speculative indexes in this module: capture a
production-shaped `EXPLAIN (ANALYZE, BUFFERS, WAL, FORMAT JSON)` first, then add
a forward-only migration with write-amplification and storage costs documented.

### Plan budgets

Use these review targets with representative workspace cardinalities; they are
budgets, not claims about a tiny test fixture:

| Query class                       | Target execution time | Additional guardrail                            |
| --------------------------------- | --------------------: | ----------------------------------------------- |
| Actor access/effective team scope |                 25 ms | No unbounded sort or spill                      |
| One aggregate/list SQL statement  |                150 ms | No temp-file spill; bounded output              |
| Workload/pulse capability         |          500 ms total | Each child statement within single-query budget |
| Command-center report             |             1 s total | Slow sections observable independently          |
| Analytics event insert            |                 50 ms | Exactly one or zero affected rows               |

At review time inspect estimated versus actual rows, loops, shared reads, temp
blocks, sort method, and rows removed by filters. A sequential scan is not
automatically wrong for a small table; it becomes actionable when the measured
plan violates the budget or scales with irrelevant tenant rows. Recheck plans
after meaningful data growth or PostgreSQL upgrades.

## HTTP surface

All routes are authenticated and workspace-scoped:

```text
GET  /workspaces/{workspaceSlug}/analytics/summary
GET  /workspaces/{workspaceSlug}/analytics/contributions
GET  /workspaces/{workspaceSlug}/analytics/users
GET  /workspaces/{workspaceSlug}/analytics/status
GET  /workspaces/{workspaceSlug}/analytics/priority
GET  /workspaces/{workspaceSlug}/analytics/overview
GET  /workspaces/{workspaceSlug}/analytics/story-analytics
GET  /workspaces/{workspaceSlug}/analytics/objective-progress
GET  /workspaces/{workspaceSlug}/analytics/team-performance
GET  /workspaces/{workspaceSlug}/analytics/workload-analysis
GET  /workspaces/{workspaceSlug}/analytics/pulse
GET  /workspaces/{workspaceSlug}/analytics/sprint-analytics
GET  /workspaces/{workspaceSlug}/analytics/timeline-trends
GET  /workspaces/{workspaceSlug}/analytics/command-center
POST /workspaces/{workspaceSlug}/analytics/events
```

API documentation for external integrations should describe the transport DTO,
filter defaults, limits, role policy, and error shape. It must not expose SQLC
models or imply that a workspace token can bypass user/team authorization.

## Changing or adding a report

1. Put neutral input/result types in the matching `domain/<capability>.go`.
2. Add reviewed static SQL to `repository/queries/<capability>.sql`.
3. Use `sqlc.arg`/`sqlc.narg` with explicit casts; never concatenate SQL.
4. State the input and output grain, time basis, terminal-status semantics, and
   null policy in the query review.
5. Apply workspace predicates and effective team scope to every entity source.
6. Reuse `scopedQueryFilters`; never interpret an empty member scope as all.
7. Run SQLC generation and review parameter/result nullability.
8. Map rows in the capability repository file and fail closed for required IDs.
9. Keep service orchestration bounded and caller-independent.
10. Add unit tests plus PostgreSQL 18 positive, cross-tenant, guest, inactive,
    and private-team negative coverage.
11. Capture a production-shaped query plan when a new join or large aggregate
    is introduced.
12. Update this inventory and the external API documentation.

## Verification

From `apps/server`:

```bash
make sqlc-generate
make sqlc-check
SQLC_DATABASE_URL='postgresql://validation-role:password@127.0.0.1:5432/fortyone_validation?sslmode=disable' \
  make sqlc-vet
go test -count=1 -race ./internal/modules/reports/...
go vet ./internal/modules/reports/... ./internal/bootstrap/api ./internal/bootstrap/worker
go test -count=1 ./internal/bootstrap/architecture
TEST_DATABASE_URL='postgresql://test-role:password@127.0.0.1:5432/postgres?sslmode=disable' \
  go test -count=1 -race -parallel=1 -tags=integration \
  ./internal/modules/reports/repository
```

The SQLC validation database must be migrated to repository head. The
integration server must be disposable PostgreSQL 18 and satisfy
[`../testing/integration-infrastructure.md`](../testing/integration-infrastructure.md).
The report integration fixture proves positive member/admin access, explicit
cross-tenant denial, guest and inactive denial, unauthorized private-team
denial, private-team exclusion for an empty member filter, joined-private-team
access, and fail-closed analytics entity writes.
