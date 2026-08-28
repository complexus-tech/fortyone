# Story comment persistence and event contract

Story-comment writes are owned by `internal/modules/comments`. The package is a
complete mutation slice: HTTP and compatibility callers construct typed
commands, the service validates identity and bounded input, and the repository
uses module-local SQLC queries over native pgx. Generated SQLC types stay below
the repository boundary.

The legacy stories workflow retains a stories-owned narrow `CommentCreator`
port because first-party and provider callers still use its existing
notification workflow. A bootstrap adapter maps that caller-owned command and
result to the concrete comments service; neither feature service imports the
other. The stories module no longer inserts comments or writes mentions. All
comment create, update, and delete persistence passes through the comments
service and repository.

## Authorization invariants

Every write carries the comment or story ID, workspace ID, and the complete
authenticated actor. The service requires a user-backed actor
(`human_user`, `personal_token`, or `oauth_user`) with `comments:write` and an
actor workspace that exactly matches the command workspace.

The SQL statement then rechecks current authoritative state:

- the user is active and remains a workspace member;
- the user remains a member of the story's team;
- the credential's team restriction includes the story team, unless the
  credential is unrestricted;
- the story belongs to the supplied workspace and is not deleted;
- updates and deletes target a comment authored by the actor; and
- a reply parent belongs to the same story.

A missing target, wrong workspace, inaccessible team, revoked membership,
non-author mutation, and deleted story all produce the same not-found outcome.
This prevents resource enumeration. Mentioned users must be active members of
the same workspace; one invalid mention rejects the complete write.

## Transaction boundary

Create, update, and delete each own one pgx transaction. Create and update also
replace the normalized mention set inside that transaction. The same
transaction appends the developer event and fans out pending deliveries to
currently authorized webhook endpoints. A mention failure, event-ID conflict,
or delivery fan-out failure rolls back both comment state and the outbox write.
There is no network call inside the transaction.

Comment content is limited to 10,000 Unicode code points and a comment may
mention at most 100 users. Duplicate mention IDs are removed before persistence.
Concurrent updates lock through PostgreSQL's row update and produce one durable
event per committed update. `updated_at` advances monotonically even when two
writes occur within the database clock's timestamp resolution.

Deleting a parent uses the schema's existing `ON DELETE CASCADE` behavior for
replies. `comment.deleted` identifies the explicitly requested root mutation;
consumers must treat deletion of a parent as invalidation of its reply subtree.
Cascade-deleted replies do not currently produce separate events.

## Developer event contract

The stable event names are `comment.created`, `comment.updated`, and
`comment.deleted`, all at payload version `1`. The exact request body is stored
once and reused for signing and every delivery attempt:

```json
{
  "id": "1f128146-a75d-4c16-a5ee-c87f994e99d7",
  "type": "comment.updated",
  "payload_version": 1,
  "occurred_at": "2026-08-28T11:30:00Z",
  "data": {
    "comment_id": "48bde60f-d1e4-46f5-bd75-f32a193e85cc",
    "story_id": "406741c5-5cff-4673-a3bb-4036cc6c78c9",
    "parent_id": null
  }
}
```

The payload intentionally contains identifiers only. Comment content, mention
lists, names, emails, credentials, provider metadata, tokens, and secrets are
not part of the external contract. Receivers use the event ID for idempotency
and fetch any currently authorized resource representation separately.

## Verification

Fast checks from `apps/server`:

```bash
go test -race ./internal/modules/comments/...
.tools/bin/sqlc compile -f sqlc.yaml
./scripts/check-sqlc.sh
```

The PostgreSQL 18 suite applies the real migration chain and verifies tenant
and team isolation, author ownership, live membership revocation, atomic mention
replacement, event/delivery atomicity, event-ID conflict rollback, payload data
minimization, and concurrent updates:

```bash
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/comments/repository
```

Paginated comment roots, batched replies, and parent-comment notification
lookup now use `internal/modules/stories/repository/queries/comment_reads.sql`
with SQLC and native pgx. They repeat the active-user, current workspace and
team membership, story/workspace, deleted-story, and credential-team fences.
Migration `000166_story_comment_tree_indexes` adds separate partial indexes for
root pagination and ordered reply hydration; the primary-key index remains the
single-comment lookup path. PostgreSQL 18 tests exercise the full migration
chain through the current repository head, isolate the safe 166 down/up
transition, restore the head, and prove deterministic `EXPLAIN` selection of
both tree indexes. No production story-comment read, mutation, or mention path
uses SQLx. The complete story-side read and lifecycle contract is documented in
[Story mutation persistence](stories-mutations.md).
