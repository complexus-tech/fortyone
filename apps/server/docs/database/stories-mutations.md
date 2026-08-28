# Story mutation, relationship, and schedule persistence

Story write persistence is split by cohesive use case instead of one generic
repository file. The service owns actor policy and event intent, the domain
owns validated commands, and the repository owns SQLC parameter mapping,
authoritative database checks, row locks, and transaction boundaries. Generated
types never cross the repository boundary.

This document covers both the primary create/update/delete slice and the
secondary lifecycle and relationship slice. Story reads are documented in
[Stories read persistence](stories-read.md); comment writes are documented in
[Story comment persistence](comments.md).

## Source map

| Capability                                                       | Domain/service boundary                                                                                                                    | Handwritten repository                                          | Reviewed SQLC queries                                                                                                                 |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Create, typed patch, single soft delete                          | `mutation_intent.go`, `story_create.go`, `story_update.go`, `story_delete.go`                                                              | `mutations.go`, `mutation_mapping.go`, `mutation_helpers.go`    | `queries/mutations.sql`                                                                                                               |
| Durable story developer events                                   | `mutation_outbox.go`                                                                                                                       | `mutation_outbox.go`                                            | `queries/mutation_outbox.sql`                                                                                                         |
| Bulk soft delete, restore, archive, unarchive                    | `domain/secondary_mutations.go`, `service/secondary_mutations.go`, `service/story_secondary_lifecycle.go`                                  | `secondary_mutations.go`                                        | `queries/secondary_mutations.sql`                                                                                                     |
| Interactive hard delete and durable attachment-object retirement | `domain/secondary_mutations.go`, `service/story_secondary_lifecycle.go`, `http/hard_delete_media.go`                                       | `secondary_mutations.go`, `interactive_hard_delete.go`          | `queries/secondary_mutations.sql`, `queries/interactive_hard_delete.sql`, and the shared outbox statements in `queries/retention.sql` |
| Label and collaborator replacement                               | `domain/secondary_mutations.go`, `service/secondary_mutations.go`, `service/story_relationships.go`                                        | `secondary_mutations.go`                                        | `queries/secondary_mutations.sql` plus the shared authorized-label insert in `queries/mutations.sql`                                  |
| Watch/mute preference and notification audience                  | `service/story_relationships.go`                                                                                                           | `secondary_mutations.go`                                        | `queries/secondary_mutations.sql`                                                                                                     |
| Story association add/update/remove                              | `domain/associations.go`, `service/association_mutations.go`                                                                               | `association_mutations.go`                                      | `queries/association_mutations.sql`                                                                                                   |
| Story duplication and inline-media link copy                     | `domain/duplication.go`, `service/story_duplication.go`                                                                                    | `duplication.go`, `duplication_helpers.go`                      | `queries/duplication.sql`                                                                                                             |
| Explicit activity batches                                        | `domain/activities.go`, `service/activity_port.go`                                                                                         | `activity_writes.go`                                            | the shared activity query in `queries/mutations.sql`                                                                                  |
| Estimate, status, key-result, link, and activity support reads   | caller-owned ports in `service/support_ports.go`                                                                                           | `support_reads.go`                                              | `queries/support_reads.sql`                                                                                                           |
| Authorized auto-scheduling state                                 | typed port in `service/auto_scheduling.go`                                                                                                 | `automation_state.go`                                           | `queries/automation_state.sql`                                                                                                        |
| Comment roots, replies, and parent lookup                        | `domain/comments.go`, `service/comment_reads.go`, `service/story_comments.go`                                                              | `comment_reads.go`                                              | `queries/comment_reads.sql`                                                                                                           |
| Durable Maya schedule transitions                                | domain values in `domain/schedule_transition.go`; worker/service ports in `service/schedule_transition_outbox.go`                          | `schedule_transition_outbox.go`                                 | `queries/schedule_transition.sql`                                                                                                     |
| Story auto-archive, auto-close, and sprint-story migration       | `domain/automation.go`, `pkg/jobs/story_automation.go`, `internal/taskhandlers/story_automation.go`                                        | `story_automation.go`                                           | `queries/story_automation.sql`                                                                                                        |
| Deleted-story retention and attachment-object delivery           | `domain/retention.go`, `pkg/jobs/purge_stories.go`, `internal/taskhandlers/story_retention.go`                                             | `retention.go`                                                  | `queries/retention.sql`                                                                                                               |
| Sprint creation and inactivity shutdown                          | `internal/modules/teamsettings/domain/sprint_automation.go`, `pkg/jobs/sprint_automation.go`, `internal/taskhandlers/sprint_automation.go` | `internal/modules/teamsettings/repository/sprint_automation.go` | `internal/modules/teamsettings/repository/queries/sprint_automation.sql` and `sprint_schedule.sql`                                    |

The production story adapter is pgx-only. Its constructor accepts one
`*pgxpool.Pool`; every query is reviewed SQLC, and generated types remain inside
the repository package. Compatibility interfaces in the service exist only for
focused test doubles and do not create a second production persistence path.

```text
HTTP, Maya, or integration caller
  -> stories service: authenticated actor, typed intent, event payload
  -> stories domain: bounded command and cross-field validation
  -> stories repository: live authorization and one pgx transaction
  -> SQLC query set bound to pgx.Tx
  -> story state + durable event commit together
  -> worker claims event with a fenced lease
  -> bootstrap adapter publishes to subscribed outbound endpoints
```

## Authorization and tenant isolation

Every migrated write carries the complete platform actor and exact workspace.
The service requires `stories:write`, rejects an actor bound to another
workspace, and prevents a caller-supplied actor ID from replacing the
authenticated principal. Credential team restrictions can only narrow product
access.

Before changing any target, `AuthorizeSecondaryStoryTargets` locks targets in
stable UUID order and rechecks mutable authority inside the transaction:

- target IDs all belong to the supplied workspace;
- human and OAuth actors are active users with current workspace and story-team
  membership;
- personal tokens still belong to their subject user, remain unrevoked and
  unexpired, retain `stories:write`, and allow the target team;
- service-account principals and keys remain active, unrevoked, unexpired,
  scoped for `stories:write`, and allowed for the target team;
- system actors are active system users; and
- the in-memory credential restriction also allows every target team.

A missing or cross-workspace target produces `domain.ErrNotFound`. A visible
target for which live actor authority has been revoked produces
`domain.ErrMutationForbidden`. The repository never authorizes from an HTTP
role flag or a stale cached membership.

Deletion has the additional ownership policy used by the existing API: a
user-backed actor must be the reporter or a current workspace admin. Service
accounts cannot soft-delete or hard-delete stories, even with
`stories:write`; they may perform non-destructive writes allowed by their scope
and team restriction. A verified system actor may delete for internal recovery
workflows. Hard delete never trusts a client-provided attachment list.

Label replacement accepts only labels in the story workspace and either the
story team or the workspace-wide label set. Collaborators must be active,
non-system members of the story team and cannot duplicate the assignee. Watch
preferences are user-local: the actor ID must be the authenticated user-backed
principal, with current workspace and team membership. Machine actors cannot
create a human watch preference. Notification-audience reads also rejoin every
assignee, collaborator, and watcher to an active account plus current workspace
and story-team membership, so membership revocation takes effect before the
next delivery is assembled.

Comment reads intentionally use the established user visibility model. Each
query proves the actor is active, remains a workspace and story-team member,
binds the story to the workspace, excludes deleted stories, and applies the
credential's allowed-team set. Parent lookup also binds the comment to the
supplied story, preventing cross-story notification lookups.

## Transaction and event intent

One lifecycle command accepts at most 500 non-zero story IDs. IDs are
deduplicated before persistence. The repository locks the complete authorized
target set before changing state; a mixed valid and invalid batch changes
nothing.

| Mutation                      | Database transition                                                                                                          | Durable developer event                         | Version-1 `data.changes` intent                                                  |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- | -------------------------------------------------------------------------------- |
| create                        | insert story, sequence, references, labels, and activity                                                                     | `story.created`                                 | not applicable; creation data is documented in the webhook contract              |
| typed update                  | compare-and-swap story patch, activities, media reconciliation                                                               | `story.updated`                                 | exact normalized changed fields                                                  |
| soft delete                   | set `deleted_at` and `updated_at` only when not deleted                                                                      | `story.deleted`                                 | no changes map                                                                   |
| hard delete                   | collect attachment candidates, delete stories, retire only unreferenced attachment metadata, and enqueue each retired object | `story.deleted`                                 | no changes map                                                                   |
| restore                       | clear `deleted_at` only when deleted                                                                                         | `story.updated`                                 | `{"deleted_at": null}`                                                           |
| archive                       | set `archived_at` only for active, unarchived rows                                                                           | `story.updated`                                 | `{"archived_at": "<UTC timestamp>"}`                                             |
| unarchive                     | clear `archived_at` only for active, archived rows                                                                           | `story.updated`                                 | `{"archived_at": null}`                                                          |
| labels                        | replace the exact normalized set                                                                                             | `story.updated` only when the set changes       | `{"label_ids": "changed"}`                                                       |
| collaborators                 | replace the exact normalized set                                                                                             | `story.updated` only when the set changes       | `{"collaborator_ids": "changed"}`                                                |
| association add/update/remove | lock the association and every old/new endpoint, then insert/CAS-update/delete                                               | one `story.updated` for every affected endpoint | IDs, action, association type, and previous type only; story titles are excluded |
| duplicate                     | CAS-lock the source version, allocate a team sequence, copy the story and inline-media links                                 | `story.created` for the caller-chosen target ID | the existing bounded story-created contract                                      |
| watch or mute                 | change the acting user's delivery preference                                                                                 | none                                            | user-local preference is not shared story state                                  |

State, the corresponding field activity, and the row in
`story_mutation_events` use the same `pgx.Tx`. Interactive hard delete also
retires only attachment metadata proven unreferenced and appends its
`attachment_object_deletion_outbox` rows in that transaction. The label and
collaborator activity values are derived from the sets actually read and
replaced under the story lock; callers cannot forge the before value. An event-ID
conflict, invalid relationship reference, failed activity, failed attachment
outbox insert, or failed media reconciliation rolls back the story change.
Hard-deleting a story does not delete its durable developer event because that
outbox deliberately retains an immutable snapshot for delivery and bounded
retention.

Idempotent lifecycle retries return the normalized requested receipt but append
events only for rows whose state changed. Two concurrent archive attempts are
serialized on the story row: both can return successfully, but exactly one
changes the row and creates `story.updated`. Label and collaborator no-ops do
not create an event.

Association updates use an expected snapshot. If another request changes the
same association after preparation, the later compare-and-swap returns
`domain.ErrMutationConflict`; it never overwrites the newer endpoints or type.
When endpoints change, the event and activity set covers the union of old and
new stories. The repository rechecks all of them under row locks, so moving an
association cannot be used to write across a tenant or a team restriction.

Duplication is also compare-and-swap safe. The service chooses the target story
ID, activity ID, and event ID before persistence, and passes the source
`updated_at` version. The repository locks and reauthorizes the source, rejects
a stale version, allocates the sequence, rewrites story-scoped media URLs,
copies authorized inline-media links, writes the create activity, and appends
the durable `story.created` row in one serializable transaction. A retry reuses
the same identities. Duplication requires a user-backed or verified system
actor because `reporter_id` and media-link `created_by` are user-attributed;
service accounts cannot invent that human attribution.

The worker claims at most 100 story events with `FOR UPDATE SKIP LOCKED`.
Processing rows carry a unique claim token; completion and retry require the
same token. A stale processing lease can be reclaimed with a new token. Retry
publication preserves the original event ID, actor, timestamp, and JSON payload
and never replays the business mutation.

## Bounded story and sprint automation

Every automation entry point captures one application-owned UTC `as_of` value
and reuses it for the complete run. SQL does not call the database clock, so
eligibility, transition timestamps, activities, and audit events agree even
when a run crosses midnight. Jobs check cancellation, validate that every page
is within its declared limit, require cursor progress where a cursor is used,
and return an explicit backlog error instead of looping without a bound.

Story auto-archive and auto-close process at most 1,000 stories per transaction;
sprint-story migration processes at most 500. Each job stops after 100 full
batches. Every batch takes a transaction-scoped advisory lock, selects a stable
`updated_at/id` or sprint/story order, applies the configured workspace/team
settings and exact status guard, and locks only its bounded candidates with
`FOR UPDATE SKIP LOCKED`:

- auto-archive changes only completed or cancelled, non-deleted, unarchived
  stories older than the team's configured interval;
- auto-close changes only inactive unstarted or started stories and chooses one
  deterministic cancelled status from the same workspace and team. The status
  transition and one system activity per changed story must commit together;
  and
- sprint-story migration moves only backlog, unstarted, or started work from a
  sprint that ended during the previous UTC day into the first later sprint
  starting within 30 days. The transition, one activity, and one audit event per
  story must all match before commit.

The guarded transitions remove rows from their own eligibility set. A retried
transaction therefore cannot append a duplicate activity or audit event for a
transition that already committed.

Sprint auto-creation is owned by the team-settings repository. Its worker
keyset-pages by `(workspace_id, team_id)` with 100 teams per page and at most 100
pages, using one UTC `as_of` for the run. Each team runs in its own
advisory-locked transaction: current settings are locked, managed dates are
reconciled, missing sprints and audit events are inserted, and the sequence
counter advances with an expected-value guard. A concurrent counter conflict is
retried at most three times. The inactivity path uses the same bounded keyset
scan and rechecks the 90-day human-activity threshold plus the 30-day settings
grace period inside the disabling transaction before recording its audit event.

## Attachment retirement and object deletion

Interactive hard delete accepts the shared lifecycle limit of at most 500
distinct story IDs. Before deleting anything, its transaction locks and
authorizes the complete workspace-scoped story set, gathers attachment
candidates with a 25,001-row look-ahead, and rejects a request that would exceed
25,000 candidates. It then deletes the exact locked stories and retires only
attachment metadata no longer referenced by any story, inline-story link, or
document. Shared attachments remain intact.

Each retired attachment produces one durable outbox row containing only its
workspace, provider, container, and blob routing metadata; credentials and
provider responses are never stored. The story delete, unreferenced metadata
retirement, object-deletion rows, and immutable `story.deleted` developer events
commit or roll back together. The production repository sets
`AttachmentObjectDeletionDeferred`, so the HTTP path does not call object
storage after the repository transaction; the outbox worker owns delivery.

The daily retention job applies the same retirement transaction to stories that
have remained soft-deleted for 30 days. It walks a stable `(deleted_at, id)`
keyset in batches of 100, stops after 100 batches, and rejects a batch with more
than 5,000 attachment candidates before deleting its stories.

A separate minute-level job claims at most 100 due or expired-lease object rows
per batch and stops after 100 batches. Claims use `FOR UPDATE SKIP LOCKED`, a
five-minute lease, and a unique claim token. Its task timeout is four minutes,
so a single invocation cannot hold a claim for the full lease. Completion and
failure updates require the same token, which prevents a stale worker from
finalizing a reclaimed row. Failures store only the safe message
`object storage deletion failed` and retry exponentially from one minute to a
24-hour cap. Completed rows are retained for 30 days, then purged in bounded
500-row pages, also capped at 100 pages per run.

## Comment tree reads

Comment pagination applies to root comments only. Root order is
`created_at DESC, comment_id DESC`, with a maximum page size of 100 and one
look-ahead row for `hasMore`. Replies for the returned roots are fetched in one
bounded SQLC query and ordered oldest first per parent; this avoids an N+1 query
loop while preserving the existing response tree.

`GetVisibleComment` is the parent-author lookup used by reply notification
workflows. It returns `domain.ErrNotFound` for a wrong story, wrong workspace,
inaccessible team, deleted story, inactive actor, or revoked membership. Those
states are deliberately indistinguishable to prevent enumeration.

## Schedule-transition outbox

Maya scheduling decisions have a separate durable internal event snapshot
because their existing notification payload includes the schedule transition,
audience, and original timestamp. The repository locks the story, verifies the
expected story version, updates auto-scheduling state, allocates a monotonically
increasing per-story transition sequence, and inserts the immutable outbox row
in one transaction.

Repeating the current state with the latest semantic fingerprint acknowledges
the already-durable transition without inserting a duplicate. The same state
may still be emitted again after an intervening transition. Pending, retrying,
and stale-processing rows are claimed with `FOR UPDATE SKIP LOCKED`; every
claim receives a new UUID token. Completion, retry, and terminal malformed
payload failure require that token, so a stale worker cannot finalize a newer
lease. Provider error text is trimmed, normalized, and capped at 4,000
characters by both the adapter and database constraint.

Schedule payloads are internal product events, not the developer webhook
envelope. They must not contain credentials or provider secrets. The
application stores the exact immutable JSON snapshot and publishes that same
snapshot on every retry.

## Production persistence boundary

There is no alternate production persistence adapter for stories or comments.
The former raw command, query, and media-reconciliation paths were deleted after
their callers moved to typed ports. SQLx is prohibited: do not recreate a generic
command map or raw query file, or add a fallback database dependency. Add a
finite domain command, a caller-owned service or job port, a cohesive repository
file, and a statically named SQLC query instead.

## Schema and verification

Migration `000166_story_comment_tree_indexes` adds two reversible, additive
partial indexes. `idx_story_comments_roots_page` matches the root page's
`story_id, created_at DESC, comment_id DESC` ordering, while
`idx_story_comments_replies_page` matches batched reply hydration by
`story_id, parent_id, created_at, comment_id`. The story mutation,
schedule-transition, and attachment-object deletion outbox tables provide their
respective atomicity, identity, lease, retry, and retention invariants; migration
000166 does not change those outboxes or any persisted business data.

Fast checks from `apps/server`:

```bash
make sqlc-check
go test -count=1 ./internal/modules/stories/... ./internal/modules/comments/... ./internal/modules/teamsettings/... ./pkg/jobs/...
go test -count=1 -race ./internal/modules/stories/... ./internal/modules/comments/... ./internal/modules/teamsettings/... ./pkg/jobs/...
go vet ./internal/modules/stories/... ./internal/modules/comments/... ./internal/modules/teamsettings/... ./pkg/jobs/...
```

PostgreSQL verification must target a disposable PostgreSQL 18 control server
whose role can create databases. The testkit creates an isolated database and
applies the complete migration chain for each test:

```bash
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -count=1 -race -parallel=1 -tags=integration \
  ./internal/migrations ./internal/modules/stories/repository
```

The focused suite proves mixed-tenant rollback, live membership revocation,
machine deletion restrictions, concurrent lifecycle idempotency, hard-delete
event survival, attachment reference safety, outbox rollback and claim fencing,
relationship replacement rollback, user-only watch preferences,
association CAS and event rollback, concurrent sequence-safe duplication,
duplication source CAS, duplication lookup index selection, comment tree
visibility, bounded story/sprint automation transaction rollback, schedule event
rollback, disjoint claims, claim-token fencing, and PostgreSQL major version 18.
The migration test applies the full
chain through the current repository head, moves the isolated database to
000166, removes and reapplies exactly migration 000166, restores the head, and
checks both index identities after each transition. The comment integration
test disables sequential scans only inside a test-local transaction and asserts
that `EXPLAIN` selects the root and reply indexes for their corresponding static
access patterns. Representative staging `EXPLAIN (ANALYZE, BUFFERS, SETTINGS,
FORMAT JSON)` remains required before production rollout.
