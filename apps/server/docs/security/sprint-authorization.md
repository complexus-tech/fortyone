# Sprint authorization and tenant isolation

A sprint UUID is an untrusted identifier. Every product-facing operation must
prove the current actor, workspace, and sprint team from live database state.
Route middleware establishes request context, but SQL is the resource-security
boundary.

## Required actor state

An allowed human actor must satisfy all of the following at query time:

1. the user exists and is active;
2. the user is a current member of the requested workspace;
3. the workspace role is `member` or `admin`; and
4. the user is a current member of the sprint's team.

Guests, inactive users, workspace members outside the team, revoked team
members, and users from another workspace cannot read or mutate the sprint.
List operations return an empty result. Direct reads, analytics, updates, and
deletes return the same not-found outcome used for a nonexistent UUID, avoiding
cross-tenant existence disclosure. Create returns a generic forbidden outcome
when actor or target-team authorization fails.

## Operation matrix

| Operation            | Active user | Workspace member/admin | Sprint-team member | Exact workspace |         Live SQL check |
| -------------------- | ----------: | ---------------------: | -----------------: | --------------: | ---------------------: |
| list sprints         |         yes |                    yes |                yes |             yes |                    yes |
| list running sprints |         yes |                    yes |                yes |             yes |                    yes |
| get sprint           |         yes |                    yes |                yes |             yes |                    yes |
| get analytics        |         yes |                    yes |                yes |             yes | repeated per statement |
| create sprint        |         yes |                    yes |        target team |             yes |      transaction start |
| update sprint        |         yes |                    yes |        sprint team |             yes |     locked transaction |
| delete sprint        |         yes |                    yes |        sprint team |             yes |     locked transaction |

Credential-level scopes and allowed-team restrictions remain service concerns
above this repository. They can narrow product membership but must never widen
it. A future versioned external route must pass a typed actor policy into the
service; it must not add an HTTP-only authorization shortcut around these SQL
predicates.

## Reference isolation

An optional objective is valid only when its `workspace_id` and `team_id` equal
the sprint's. Create first proves actor access to the target team, then validates
the objective inside the same transaction. This distinction allows an
authorized caller to receive a generic invalid-reference response without
conflating it with actor authorization. Updates validate the resulting
objective under the locked sprint's tenant and team.

The database foreign key is useful integrity defence but is not sufficient: it
proves only that an objective UUID exists, not that it belongs to the same
tenant or team.

## Revocation and request races

Authorization is based on current database rows, not workspace or team roles
cached in a browser session. Mutation targets and membership rows are locked for
the transaction. Analytics statements repeat actor predicates instead of
assuming the first sprint lookup remains authoritative for the rest of the
request. Removing a team membership therefore immediately blocks subsequent
list, single-read, analytics, update, and delete calls.

## Data returned by analytics

Story counts include only stories whose workspace and sprint match and whose
`deleted_at` and `archived_at` values are null. Team allocation includes only
active users who remain member/admin workspace members and sprint-team members.
An inactive or guest row is not exposed merely because a stale team-membership
row remains.

Historical activity reads are bounded to stories in the requested workspace
and sprint. Burndown output is produced only after the `authorized_sprint` CTE
revalidates the actor.

## Mutation durability

Sprint state and its audit event commit in one pgx transaction. A failed audit
insert rolls back the state change. The audit metadata contains safe change
facts, IDs, names, and dates; it must never include bearer credentials, provider
tokens, signed URLs, raw headers, or request bodies.

## Review checklist

Before merging a sprint change, verify:

- every product read receives an actor ID as well as workspace and sprint IDs;
- workspace role is restricted to member/admin in SQL;
- active-user and current team-membership checks are present;
- hidden and cross-tenant direct IDs return not found;
- objective references match both workspace and team;
- nullable fields preserve omitted/value/null semantics;
- multi-table mutations and audit persistence share one pgx transaction;
- list inputs remain bounded and cannot become SQL identifiers;
- analytics exclude archived/deleted stories and stale members;
- no secret-bearing value enters logs, traces, events, or audit metadata; and
- PostgreSQL 18 integration tests cover allowed, denied, revoked, rollback, and
  concurrency paths.

Implementation details and the burndown CTE guide are in
[Sprint persistence and analytics](../database/sprints.md). Repository-wide
rules are in [Typed database access](../database/sqlc.md) and
[Authorization](authorization.md).
