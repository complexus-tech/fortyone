# Maya realtime persistence contract

Maya realtime voice sessions and tool-call idempotency use SQLC with the
process-owned native pgx pool. The HTTP adapter never receives a database
handle. Its dependency path is:

```text
internal/modules/maya/http
  -> internal/modules/maya/service
    -> RealtimeRepository port
      -> internal/modules/maya/repository
        -> internal/modules/maya/repository/queries/realtime.sql
          -> internal/modules/maya/repository/sqlc (generated)
```

The handwritten service owns product policy and request canonicalization. The
repository owns transaction, locking, SQL error, and affected-row semantics.
Generated SQLC params and rows never leave the repository package.

## Session reservation and quota

`Service.BeginRealtimeVoiceSession` applies the domain limits: ten minutes per
workspace per calendar month and five minutes per session. The repository then:

1. opens a native pgx transaction;
2. locks the current, non-deleted workspace row with `FOR UPDATE`;
3. rechecks the workspace's trial or paid-subscription access inside that same
   transaction, closing the gap between an earlier HTTP authorization check and
   persistence;
4. counts ended sessions by their bounded actual duration and open sessions by
   their full five-minute reservation;
5. rejects exhausted quota or inserts the new session; and
6. commits before the API calls the external realtime provider.

The workspace lock serializes concurrent reservations for one tenant. Two
simultaneous five-minute reservations can consume the ten-minute allowance; a
third cannot observe the same pre-reservation total and oversubscribe it.
Provider network calls never run inside the transaction. If provider session
creation fails after reservation, the HTTP use case immediately ends the local
session so its actual elapsed time, rather than a permanent open reservation,
is counted.

Session validation and termination bind all three identifiers:
`workspace_id`, `user_id`, and `session_id`. A session from another tenant or
user is never accepted merely because its UUID exists. Validation also requires
an open session whose bounded lifetime has not elapsed.

## Tool-call idempotency

Every realtime tool call is uniquely identified by `(session_id, call_id)`.
The service trims the tool and call identifiers and hashes the exact tool name,
a zero-byte separator, and the exact JSON argument bytes with SHA-256. The
repository atomically inserts a claim with `ON CONFLICT DO NOTHING`.

The outcomes are deliberate:

| Stored state                        | Incoming state            | Result                      |
| ----------------------------------- | ------------------------- | --------------------------- |
| no row                              | valid claim               | insert and execute once     |
| same hash, no response              | duplicate while executing | in-progress conflict        |
| same hash, stored response          | completed duplicate       | return stored JSON response |
| different hash for the same call ID | conflicting duplicate     | conflict; never execute     |

Completion updates only a claimed row whose response is still null and requires
exactly one affected row. A second completion is a typed conflict. PostgreSQL
validates the response as JSONB; the service rejects invalid JSON before it
reaches persistence.

## Query and change workflow

The reviewed SQL is
`internal/modules/maya/repository/queries/realtime.sql`. Generated files under
`internal/modules/maya/repository/sqlc` must never be edited. After changing the
query or schema contract, run:

```bash
make sqlc-generate
make sqlc-check
SQLC_DATABASE_URL='<disposable database at migration head>' make sqlc-vet
go test -race ./internal/modules/maya/...
TEST_DATABASE_URL='<disposable PostgreSQL control URL>' \
  go test -race -tags=integration -count=1 ./internal/modules/maya/repository
```

The integration suite proves concurrent quota serialization, current billing
access, cross-tenant session rejection, exact-request idempotency, replay, and
completion conflicts against the real migration chain.

Realtime is one capability within Maya's completed pgx/SQLC boundary. Six files
under `internal/modules/maya/repository/queries` own the module's 30 named
operations for entitlement, work plans, scheduling, realtime, work-focus, and
worker reads. The handwritten repository is Maya's only PostgreSQL adapter;
HTTP, services, task handlers, and jobs depend on Maya-owned domain types and
narrow ports rather than generated rows or database handles. API and worker
composition reuse their process pool and share one Maya repository instance
across the surfaces in that process. SQLx is a prohibited production dependency;
do not add raw SQL to HTTP, service, task-handler, or job packages.

## Native approvals and conversation history

Sessions request `client: "mobile"` to advertise the new deletion capability;
legacy web sessions retain their existing tool catalog. Native instructions
direct Maya to prepare a review card and wait for the app, without treating
spoken assent as approval. This capability flag is not an authorization secret.

The `delete_story` voice tool resolves a visible current story, checks the
current actor's admin/creator permission, and previews its title and reference.
Its confirmation signature binds the voice session, workspace, actor, story ID,
title, reference, and current update timestamp. An exact approved call uses the
ordinary story deletion service (soft deletion), including its transactional
authorization and lifecycle work. The existing `(session_id, call_id)` receipt
surrounds this operation just like the other realtime tools. Native clients
must hold confirmation tokens outside the model conversation and send them
only following an explicit approval in the UI.

The Projects application exposes authenticated `POST /api/chat/voice-history`:

```json
{
  "id": "16-character-id!",
  "workspace": { "slug": "example" },
  "messages": [
    { "id": "voice-user-1", "role": "user", "text": "Show my tasks" },
    { "id": "voice-assistant-1", "role": "assistant", "text": "Here are your tasks." }
  ]
}
```

Only finalized text transcripts are accepted; tool parts, approvals, tokens,
and system messages are rejected. The endpoint reads the authenticated user's
canonical history and uses chat message-write begin/finalize generations to
append without replacing existing content or bypassing pending approvals.
Successful retries with the same message IDs and text are no-ops. A delayed
assistant reply must include its original user transcript; it cannot attach to
a newer text request. Initial assistant-only transcripts remain buffered until
a user turn is available. Clients must retain failed batches for explicit retry
with the original IDs. HTTP 409 signals a stale/conflicting write, and 200
returns `{ "messages": [...] }` with canonical UI messages. Native clients must
serialize text sends and transcript saves for one conversation, and keep all
assistant followups for a user turn in one saved batch.

## Text admission boundary

`POST /api/chat` accepts `client: "mobile"` and an explicit IANA `timezone`.
Legacy requests without a timezone use UTC. Workspace identity and role,
terminology, user memories, identity, billing, and current monthly usage come
from authenticated APIs, not the corresponding client hints. Native models
receive the task-focused tool subset; workflow/label catalogs are read-only,
while supported task mutations retain the canonical approval ledger.

New text model calls check the current user/workspace message allowance before
reserving a generation or invoking the provider. Approval responses take the
existing approval-only path and do not consume another model turn. Failed
billing/usage reads fail closed. This is an admission check, **not an atomic
billing meter**: concurrent requests can observe the same allowance;
regeneration is not separately counted; and the existing count query attributes
user messages to the chat session's creation month and excludes deleted chats.
Voice transcript user messages also enter that existing count. Correct
per-generation, cross-month accounting needs a separate durable reservation
policy rather than a client counter or a read-before-write assertion. The
transactional voice-minute quota described above is unchanged.
