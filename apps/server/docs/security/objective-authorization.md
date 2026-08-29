# Objective authorization

Objective authorization is enforced twice: the service validates the caller's
authenticated actor and scopes, then SQL checks current database membership and
resource visibility. Route middleware is useful, but it is not the final
security boundary.

## Actor checks in the service

For normal HTTP and integration calls, the actor must:

1. be bound to the requested workspace;
2. represent a user principal;
3. hold `objectives:read` or `objectives:write`; and
4. have unrestricted credential-level team access on the legacy routes.

Restricted personal/OAuth credentials currently fail closed because the
legacy objective routes do not yet carry their permitted team set into the SQL
query. Do not “temporarily” ignore that restriction. A future versioned API can
add a typed allowed-team set and enforce it alongside product team membership.

Service accounts and OAuth applications are not accepted as objective activity
actors today. `okr_activities.user_id` requires a user row, so allowing a
machine principal without an explicit attribution design would either break the
transaction or falsely attribute a change to a human.

## Database authorization matrix

| Operation               | Workspace member | Supported role | Objective team member | Active user | Scope checked by service |
| ----------------------- | ---------------: | -------------: | --------------------: | ----------: | ------------------------ |
| list/get objective      |              yes |   member/admin |                   yes |         yes | `objectives:read`        |
| objective analytics     |              yes |   member/admin |                   yes |         yes | `objectives:read`        |
| create objective        |              yes |   member/admin |           target team |         yes | `objectives:write`       |
| update objective        |              yes |   member/admin |        objective team |         yes | `objectives:write`       |
| delete objective        |              yes |   member/admin |        objective team |         yes | `objectives:write`       |
| read strategy shell     |              yes |   member/admin |          not required |         yes | `objectives:read`        |
| see strategy alignment  |              yes |   member/admin |        objective team |         yes | `objectives:read`        |
| change strategy/pillars |              yes |   member/admin |          not required |         yes | `objectives:write`       |
| align/unalign objective |              yes |   member/admin |        objective team |         yes | `objectives:write`       |

The SQL returns not-found for a hidden objective mutation. This avoids telling
an attacker whether an identifier exists in another workspace or team.

## Reference validation

Creation and updates validate referenced rows inside the transaction:

- team belongs to the requested workspace;
- objective status belongs to the requested workspace;
- objective and key-result leads are active workspace and team members; and
- every key-result contributor is an active workspace and team member.

Contributor IDs are de-duplicated before insertion. A missing, foreign, or
inactive assignee fails the entire aggregate rather than silently dropping that
person. The actual objective, key-result, contributor, and update statements
repeat these checks, closing the gap between a precondition read and a
concurrent membership or reference change.

The objective-owned HTTP endpoints that list or create key results first call
the objective read boundary. Child-resource persistence is therefore unable to
bypass objective team visibility even if its own lookup is only workspace
scoped.

## Revocation behavior

Authorization uses database state at mutation time, not role data embedded in
a session token. Update and delete SQL repeat the live workspace membership,
role, team membership, and active-user predicates. Analytics and strategy reads
also repeat live checks in each statement so a precondition result is not a
long-lived authorization grant.

Repository integration tests cover:

- cross-workspace reads, writes, deletion, and strategy alignment;
- same-workspace users who are not members of the objective team;
- inactive users whose membership rows still exist;
- guest users who still have a team-membership row;
- invalid foreign statuses and assignees;
- rollback when activity persistence fails; and
- two concurrent compare-and-swap updates where exactly one succeeds.

## Objective privacy field

The schema contains `objectives.is_private`, but it has no objective ACL or
participant table. The current enforceable resource boundary is therefore
workspace plus objective-team membership for every objective, whether the flag
is true or false. Treating the boolean alone as authorization would be unsafe.

If product requirements later define private objectives as lead-only or
participant-only, first design an explicit ACL relation and migration, then
update every list/get/analytics/strategy/mutation predicate and its tenant
isolation tests together.

## Trusted internal read compatibility

Some background jobs historically call `Service.Get` without an actor. That
path is intentionally narrow:

- it requires both `workspace_id` and `objective_id`;
- it is represented by an explicit `Internal` query flag;
- it is not reachable from HTTP; and
- it is read-only.

New background consumers should prefer an explicit system-actor policy rather
than expanding this compatibility path.

## Events and audit durability

Objective and key-result activities are written in the same transaction as the
domain mutation, so a successful mutation cannot lose its database audit row.
The existing Redis `objective.updated` notification is published after commit;
publisher failure never rolls back a durable business change. That preserves
current behavior but does not guarantee delivery. Durable external webhooks for
objective lifecycle events require an objective outbox written by the same
repository transaction.
