# Backend data and realtime architecture

## 1. Module ownership

Add `apps/server/internal/modules/chat` using the established `domain`, `service`, `repository`, `repository/queries`, `repository/sqlc`, and `http` structure. Wire the API and workers through bootstrap. Use Go 1.26.9, native pgx, SQLC, typed actors, existing request/error conventions, and the repository's unit-of-work/outbox contract.

Chat owns conversations, native messages, threads, reactions, pins, read state, conversation preferences, chat-resource links, history/replay, and chat-specific authorisation. Existing modules remain authoritative for team membership, users, workspaces, stories, attachments, notifications, billing, and personal Maya sessions.

The messaging module remains the shared provider-neutral assistance capability. Native chat supplies a new trusted surface adapter. A chat-specific orchestration package owns durable run lifecycle; it does not duplicate story rules or introduce another unrestricted tool registry. Promote a common primitive only after its meaning is stable across native and provider surfaces.

Do not convert story comments into chat messages, personal AI JSON transcripts into multi-user rooms, or Yjs documents into chat storage. They have different access, ordering, and lifecycle rules. Keep the current Slack integration independent; native chat does not automatically mirror Slack or create a second provider-webhook runtime.

## 2. Data model

Names below are proposed domain/storage names. Select migration numbers only when implementing, after reconciling the active migration chain. All tenant relations must enforce ownership through composite foreign keys or equivalently strong scoped constraints. UUID validation alone is insufficient.

| Record                              | Fields and constraints                                                                                                                                                                                                                                                                                                           |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `chat_conversations`                | `id`, `workspace_id`, closed `kind` enum, optional `team_id`, topic, creator, archive state, `access_generation`, `last_event_seq`, `last_root_seq`, timestamps. One canonical team conversation per team.                                                                                                                       |
| `chat_direct_members`               | Explicit human members of direct conversations; exactly two for the initial human DM type. Immutable participant set initially. A unique canonical pair key prevents duplicate one-to-one DMs within a workspace.                                                                                                                |
| `chat_user_state`                   | Conversation/user/workspace, `last_read_root_seq`, separate manual-unread marker, mute/snooze/notification policy, optional pinned/saved position. Team access is derived from authoritative team membership, not this row.                                                                                                      |
| `chat_agent_bindings`               | Optional explicit availability/configuration record for native agents, linked to the existing system identity. It grants no human membership or universal resource access. A projected implicit Maya binding is sufficient for the first version.                                                                                |
| `chat_messages`                     | UUID, workspace/conversation, author ID and actor kind, client message ID, root ID if reply, root or reply sequence, creation event sequence, body schema version, validated body, plain-text search representation, body revision, aggregate `state_event_seq`, created/edited/deleted timestamps, optional originating run ID. |
| Root thread state                   | `last_reply_seq`, reply count, latest-reply metadata on the root, or a dedicated `chat_threads` row if query/locking evidence warrants it.                                                                                                                                                                                       |
| `chat_thread_user_state`            | Root/user, `last_read_reply_seq`, follow mode, mute override. Create for participants/followers/readers, not every possible user-thread pair.                                                                                                                                                                                    |
| `chat_catchup_checkpoints`          | Opaque ID, account/workspace/user/conversation, immutable pre-visit root and included-thread read vector, thread-scope policy/version, captured watermark, created/expiry timestamps. Created during visible conversation bootstrap before new read acknowledgements; never grants content access.                               |
| `chat_message_revisions`            | Revision metadata and any permitted retained content; author, time, reason. Deletion/retention must remove prohibited plaintext here as well as in the main message.                                                                                                                                                             |
| `chat_message_mentions`             | Message/revision, principal ID and mention kind, with scoped uniqueness and eligible-target validation.                                                                                                                                                                                                                          |
| `chat_message_reactions`            | Message/user/canonical emoji unique key. Set/unset operations, never a retry-sensitive toggle.                                                                                                                                                                                                                                   |
| `chat_message_attachments`          | Message/attachment/workspace/conversation relation, display order, original-name metadata, uploader/reference provenance.                                                                                                                                                                                                        |
| `chat_resource_links`               | Message or run to story/document/objective, relationship type, source-selection ID and provenance. Store IDs; hydrate protected details on read.                                                                                                                                                                                 |
| `chat_pins` / `chat_saved_messages` | Shared pins with author/time versus private user saves. Keep their access semantics separate.                                                                                                                                                                                                                                    |
| `chat_events`                       | UUID event ID, conversation/workspace, transactional `event_seq`, schema version, event type, entity ID/revision, actor, time, bounded metadata. Unique conversation/event sequence.                                                                                                                                             |
| `chat_outbox`                       | Event/consumer or destination, state, attempt count, next attempt, claim token/generation, lease expiry, terminal failure.                                                                                                                                                                                                       |
| Consumer receipts                   | Unique consumer/event identity for notification, indexing, and run-trigger side effects.                                                                                                                                                                                                                                         |
| Chat AI records                     | Runs, evidence manifests, steps, proposals/actions, per-item results, publication state, usage. Defined in the Maya document.                                                                                                                                                                                                    |
| Upload intent                       | Uploader, destination conversation, byte/type limits, checksum, expiry, attachment ID, scan/processing state.                                                                                                                                                                                                                    |

Use partial unique indexes for root order `(conversation_id, root_seq)` and reply order `(thread_root_id, reply_seq)`, plus tenant-scoped message identity and client-send uniqueness. Index event replay by `(conversation_id, event_seq)`, user state by `(workspace_id, user_id, conversation_id)`, replies by root/sequence, and queue claims by state/next-attempt. Add real query-plan tests for history, unread, search and cleanup with representative data.

Every mutable projection needs a monotonic version. Keep message `body_revision` for edit conflicts and exact AI evidence; use `state_event_seq` for chat-owned aggregate changes such as reactions, deletion state and root reply summaries. Update affected aggregate stamps in the same chat transaction. Shared attachments and stories carry their own domain revisions; a scan or work change emits an authorised invalidation without pretending it edited the message body. Clients merge each independently versioned projection, so a late HTTP response cannot restore an old reaction count or scanning state.

### Identity and body format

Persist a bounded, versioned rich-text document rather than arbitrary HTML. The server validates an allowlist of node/mark types, link schemes, nesting depth, text size, mention targets, and attachment references. Derive searchable plain text and mentions from the validated body. Clients render the same semantic contract with platform-specific components. Model output uses typed presentation blocks and resource references, not executable HTML.

A proposed initial limit is 10,000 Unicode characters, 64 KiB encoded body, 20 direct mentions, and 10 attached files per message, with lower tenant quotas where appropriate. These are configurable launch limits, not existing system facts. Reject over-limit input predictably before committing; do not silently truncate.

## 3. Ordering and transaction boundaries

The order users see must not depend on device clocks, UUID lexical order, Redis delivery order, or an independently allocated global sequence.

Within a short mutation transaction, lock the conversation counter and allocate the next event sequence. Root sends also allocate a root sequence; replies allocate the next reply sequence on their root. Always acquire locks in the same order: conversation, root/message, dependent relation rows. Independent conversations remain concurrent.

Edits, deletions, reactions, and durable conversation lifecycle changes advance event order without changing the original message position. Read updates, typing, presence, worker heartbeats and token previews do not acquire this room lock. A final Maya reply uses the normal durable message transaction; every token does not.

PostgreSQL sequence allocation is not a commit-order guarantee. Use the room transaction to establish the replay boundary, with concurrency tests proving it. [PostgreSQL sequence functions](https://www.postgresql.org/docs/18/functions-sequence.html)

### Durable send

1. Authenticate and resolve the current actor, workspace, conversation, and membership.
2. Validate the request body, root ownership, attachment upload finalisation/reference eligibility, mention eligibility, archive state and rate limit. A valid processing placeholder is distinct from a scan-clean file permitted for download or extraction.
3. Look up the scoped `client_message_id`. An identical retry returns the existing canonical result; reuse with different request bytes/content is a conflict.
4. In the transaction, recheck membership against concurrent removal, allocate order, and insert message, mentions, attachment relations, resource links, event, and outbox intent.
5. If the message genuinely invokes Maya, insert the unique activation/run intent in the same transaction.
6. Commit, then return the canonical ID, revision, sequences, timestamps and run identity.
7. Relay live hints, indexing, notifications, and AI work after commit.

The service must prove the concurrent membership-removal boundary using locking or an equivalent tested transaction design. Middleware-only checks and a stale membership cache are insufficient.

The shared idempotency service provides useful hashes, scopes and leases, but does not automatically make an unrelated domain transaction safe. Couple mutation and receipt, enforce permanent logical-send uniqueness in the domain, or reconcile from a committed result record. A 24-hour generic receipt expiry must not permit a queued old send to duplicate its original message.

### Outbox execution

Use bounded `FOR UPDATE SKIP LOCKED` claims with fencing tokens, retry classification, jitter, attempt limits, poison-message diagnostics and reconciliation. An outbox publish may happen twice if a process dies before marking completion. Consumers must therefore apply an event once by durable identity.

Keep interactive chat dispatch separate from slower maintenance queues. No network call, model call, upload, email, or Redis publish occurs inside the message transaction. Do not use `SKIP LOCKED` for normal history reads; it is a work-claim mechanism. [PostgreSQL SELECT locking](https://www.postgresql.org/docs/18/sql-select.html)

## 4. Transport decision

Use **HTTP commands and history plus one multiplexed authenticated SSE stream per active client/workspace** initially. This reuses the current application pattern and is appropriate to validate at 1,000 concurrent users. No audited native-client requirement currently forces WebSocket.

SSE is the transport for change signals, not the database. Redis Pub/Sub can lose events while a subscriber is disconnected, so use it only for prompt wake-ups. PostgreSQL stores the recoverable state and event ledger. [Redis Pub/Sub](https://redis.io/docs/latest/develop/use-cases/pub-sub/)

The initial payload can be a coalesced versioned hint:

```json
{
  "type": "chat.conversation.changed",
  "schemaVersion": 1,
  "conversationId": "uuid",
  "latestEventSeq": 481,
  "accessGeneration": 7
}
```

Receivers request canonical deltas through `/chat/sync`; the sender already receives its own canonical message from the send response. Coalesce nearby hints and share an in-flight sync request. Measure the extra round trip under the load profile. If it prevents the delivery target, add authorised canonical projections to the same SSE envelopes while retaining the exact sync protocol. Do not add a second transport merely to avoid defining replay.

Typing and presence use short-lived Redis state and throttled authenticated HTTP updates. Start with conversation-scoped typing, not workspace-wide presence. Use approximately one typing update every three seconds, an expiry around six seconds, and stop signals as best effort. These are tunable UX defaults. Maya previews use batched ephemeral frames keyed by run and offset; canonical snapshots recover them. Release safe progress immediately and validate source/audience eligibility before broadcasting answer blocks; a final-only check cannot retract leaked tokens.

### Required gateway changes

- Add conversation-aware subscriptions/authorisation. Do not publish private conversation IDs or content to the existing unrestricted workspace-update lane.
- Share Redis topic subscriptions per gateway process and fan out to locally eligible connections. Do not multiply Redis listeners by every tab and every conversation unnecessarily.
- Set bounded queue bytes and event counts. Drop stale typing first; when durable state is behind, request resync and close/drain the slow connection rather than silently claiming it is current.
- Maintain origin checks, session expiry, write deadlines, request limits, heartbeats, graceful drain, connection limits and subscriber cleanup.
- Confirm HTTP/2, buffering/compression behaviour, proxy timeouts and native fetch-stream support in a deployment spike. Heartbeats must contain data and fit the actual proxy policy. [MDN SSE guidance](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events), [AWS load balancer attributes](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/edit-load-balancer-attributes.html)
- Authorisation caches are generation-bound and short-lived. Lost invalidation signals must not leave access valid indefinitely. Sensitive reads/writes still check authoritative membership.
- A stream health signal covers the Redis subscription and gateway, not merely the client TCP connection.

## 5. Replay and bootstrap protocol

Maintain a cursor per conversation. A scalar browser `Last-Event-ID` does not encode the independently ordered histories of every conversation in a workspace. Existing notification events may share the connection without sharing chat's replay semantics.

Proposed sync request:

```json
{
  "conversations": [{ "id": "uuid", "afterEventSeq": 450, "limit": 200 }]
}
```

For each conversation, the server:

1. Checks current access.
2. In a short consistent read snapshot, captures high watermark `W`.
3. Returns ordered events where `afterEventSeq < eventSeq <= W`, within the page limit.
4. Hydrates affected messages/reactions/thread summaries with current authorised projections, or tombstones.
5. Returns `throughEventSeq`, `highWatermark`, `hasMore`, and replay floor/reset metadata.

The client advances only through events it successfully applied. It must not advance to the advertised high watermark before fetching all intervening pages. A deleted entity still produces a no-content event or tombstone that safely advances the cursor.

Each page is a new short snapshot; pagination does not hold a transaction open. Entity versions prevent a later replay page from overwriting newer visible state. Sync means convergence to current state, not disclosure of historical deleted plaintext.

Initial history load returns its event baseline consistently with the loaded page. Subscribe first or buffer hints while loading, then sync after that baseline. Reconcile on stream reconnect, network restoration, foreground, workspace change, opening a conversation, ambiguous send, and gateway subscription recovery. Also compare the visible conversation's watermark on a jittered schedule with a maximum healthy interval of 30 seconds, for example 25–30 seconds: this detects a lost final hint even when no later message arrives.

At 1,000 simultaneously foreground users, a 30-second interval is about 33 requests/second; a uniform 25–30-second interval averages about 36 requests/second before batching. Do not poll every conversation at that frequency. Refresh sidebar summaries on relevant hints, foreground and reconnect.

Keep event replay for a proposed seven days initially. If a cursor is older, return `reset_required` and load a fresh canonical window/read summary while preserving drafts. Human message retention is independent of replay retention.

## 6. Reads, unread state and attention

Maintain separate concepts:

- Event cursor: what state changes the client has applied.
- Root read position: channel messages the person reached.
- Thread read position: replies in one thread the person reached.
- Attention records: mentions, DMs, followed-thread replies, requested results.
- Manual unread/save markers: deliberate user intent.

Read updates are monotonic per user and stream, bounded by server-known accessible sequence values. Use the highest continuously viewed sequence while focused. A jump to a search hit does not silently acknowledge the intervening gap; provide an explicit mark-through action when useful. Do not derive unread counts by subtracting two sequences: own messages, deleted messages, replies and non-message events make this incorrect.

At the initial target, use indexed unread existence/capped-count queries and thread state only for relevant users. Avoid a database row for every user-message combination. Materialise aggregate counters only after profiling, with a repairable projection and reconciliation job.

Conversation bootstrap also issues a user-scoped catch-up checkpoint from a consistent snapshot of the pre-visit read vector, before the client advances reads. Launch scope is channel roots plus the explicitly included followed threads; disclose that scope. A Maya catch-up request references this opaque checkpoint, not client-supplied cursor values or the mutable current read row. Capture the ending message boundary at invocation. A proposed checkpoint expiry is 24 hours; an expired checkpoint requires a new visible scope selection, not a silent boundary change. Reauthorise all source access when used. Periodic sync does not create a fresh visit checkpoint. Out-of-order notification views acknowledge their attention item only; the initial model does not persist sparse message-read exceptions, so catch-up is explicitly watermark-based and may include a previously viewed isolated target above a gap.

## 7. API contract outline

Use the established workspace routing/authentication conventions. The following relative routes describe capabilities rather than bypassing the existing public/internal API versioning rules.

| Operation                           | Proposed route/contract                                                                                                                                                                            |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Conversation navigation             | `GET /chat/conversations` with cursor/filter and authoritative unread/attention summaries.                                                                                                         |
| Metadata and participant projection | `GET /chat/conversations/{id}` and `/members`; system/human kinds remain internal metadata.                                                                                                        |
| Create/find one-to-one DM           | `POST /chat/direct-conversations` with one eligible recipient; canonical pair uniqueness.                                                                                                          |
| Message page/bootstrap              | `GET /chat/conversations/{id}/messages` with `before`, `after` or `around`, bounded page size. Explicit visit bootstrap returns a consistent event baseline and server-issued catch-up checkpoint. |
| Send                                | `POST /chat/conversations/{id}/messages`, client identity and versioned body.                                                                                                                      |
| Edit/delete                         | `PATCH`/`DELETE /chat/messages/{id}`, expected revision and logical operation ID.                                                                                                                  |
| Thread                              | `GET /chat/messages/{rootId}/replies`; send uses root ID in the validated message command.                                                                                                         |
| Reaction                            | `PUT`/`DELETE /chat/messages/{id}/reactions/{emoji}`.                                                                                                                                              |
| Shared pins/private saves           | Separate idempotent pin/unpin and save/unsave capabilities; pins require conversation permission, saves belong only to the current user.                                                           |
| Read/follow                         | `PUT /chat/conversations/{id}/read`, thread-specific read/follow routes.                                                                                                                           |
| Preferences                         | `PATCH /chat/conversations/{id}/preferences`, scoped to current user.                                                                                                                              |
| Sync                                | `POST /chat/sync` with bounded conversation cursor vector.                                                                                                                                         |
| Typing                              | `PUT /chat/conversations/{id}/typing`, optional root ID and short expiry.                                                                                                                          |
| Upload                              | Conversation-bound intent, finalize, and authorised attachment download endpoints.                                                                                                                 |
| Search                              | `GET /chat/search` with query and authorised conversation/author/date filters.                                                                                                                     |
| Maya                                | Run read, cancel, clarify, proposal review, execute/retry-remaining endpoints; no client-controlled authority fields.                                                                              |
| Task claim                          | Existing story capability extended with conditional `expectedUnassigned`/version semantics if necessary.                                                                                           |

Responses include stable error codes and request IDs. Define conflict, permission/no-access, archived, invalid-root, attachment-not-ready, rate-limit, stale-proposal, replay-reset and uncertain-action outcomes. Return retry timing where appropriate. Never expose SQL/provider error bodies.

Version event and body schemas explicitly. Maintain compatibility with older web sessions and installed native binaries; unknown optional event types trigger a safe refresh, not a crash. Required incompatible changes need capability gates and a documented minimum version.

## 8. Authorisation and disclosure

Current workspace/team membership is the authority for team rooms. Explicit DM participants are the authority for DMs. Workspace administration alone does not grant ordinary UI access to private conversations. Support/compliance access, if introduced, is a separately authorised and audited product feature.

Enforce this policy on message reads/writes, threads, members, replay, search/snippets/counts, attachments, previews, source links, Maya context, action execution, and notification rendering. Membership removal increments an access generation and invalidates queued work and active subscriptions. Current reads and writes fail closed; periodic checks cover lost live invalidations.

The practical revocation guarantee is no newly authorised read or write after the revocation boundary, plus prompt connection/cache teardown. Bytes already in flight or previously viewed cannot be recalled. Offline caches require explicit retention, device locking and next-contact cleanup rather than an impossible promise of instantaneous offline erasure.

Rate-limit by actor, conversation and workspace. Bound message sizes, connection counts, subscription counts, search complexity, replay pages, uploads, typing and mentions. A hot room must not starve unrelated tenants.

## 9. Attachments and link previews

Reuse object storage and attachment lifecycle, adding chat-specific access and reference ownership. Upload intents bind the actor and destination; a raw attachment UUID is not sufficient authority to attach or download a file.

Validate file length, actual MIME, checksum, image dimensions, archive expansion limits and quotas. Keep objects outside message bodies. The current attachment implementation does not include a deployed malware-scanner provider; `unscanned` compatibility behaviour must not be described as scanning.

For the wider chat rollout, add a real quarantine/scanning integration. Proposed policy: text can send with a processing attachment placeholder, but download, preview and Maya extraction remain blocked until clean. Reject infected files; distinguish transient scanner failure from malicious content; provide retry/support paths. If scanner work is deferred, restrict the file beta explicitly rather than silently opening broad uploads.

Signed download URLs are short-lived and issued only after live access checks. Active HTML/SVG and unsafe document rendering are sandboxed or forced to download. Public unfurls use a safe outbound fetcher with DNS/redirect/private-network protection, response size/time limits, and no forwarded credentials. Private resource previews use domain APIs, never a generic URL fetch.

Extend orphan cleanup, object-reference counting, workspace/team deletion, user deletion and retention jobs. A new relation table without corresponding cleanup changes can delete files still referenced by chat or retain files that should be removed.

## 10. Search and derived knowledge

Start with PostgreSQL full-text search and a GIN index on validated plain text. Apply tenant and conversation access in SQL before ranking, counts and snippets. Use stable keyset pagination and bounded queries. PostgreSQL recommends GIN for text-search indexing. [PostgreSQL text-search indexes](https://www.postgresql.org/docs/18/textsearch-indexes.html)

Edits update search; deletions remove content; membership and archive policy apply immediately to result access. Highlighting must escape content. Search analytics may contain timing and result counts, not private query/message text by default.

Semantic retrieval is a later optimisation with source ID/revision, access scope, deletion and retention metadata. Filter before model exposure, then reauthorise canonical IDs. Do not retrieve unrestricted top-K content and rely on post-generation filtering to remove secrets.

## 11. Retention and lifecycle

Recommended initial human-history default: retain until an authorised deletion or configured workspace policy. Expose the actual policy before launch; do not reuse external-provider messaging's 30-day retention accidentally. Design optional 30/90/365-day policies without forcing a product decision into the schema.

Separate retention classes: human message bodies; message revision history; operational replay events; AI evidence snapshots; AI traces; usage/audit metadata; pending notification excerpts; attachments and extracts; exports/backups. Proposed initial operational defaults are seven-day replay, 30-day restricted AI diagnostic evidence, and bounded metadata retention aligned with existing audit policy. These require owner approval before launch and must be configurable without changing historical migrations.

Deletion must fan out to search, extracts, embeddings, cached summaries, pending notifications, attachment relations and AI evidence. Keep an explicit deletion work ledger and completion age. Ordinary source edits may mark a summary stale; sensitive-content deletion requires redacting/removing affected derived text and evidence or regenerating from eligible sources. A stale badge alone is insufficient. Independently created work follows its own domain lifecycle, with coordinated workspace erasure and sensitive-content remediation as described in the Maya plan.

Backups follow documented restoration and deletion policy. A restore must reapply deletion/access tombstones before exposing recovered data. Exports must be scoped, audited, rate-limited and protect DMs; workspace-admin export does not implicitly mean unrestricted private-message access.

## 12. Failure isolation

| Failure                       | Required behaviour                                                                                                                                                                  |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PostgreSQL unavailable        | Do not acknowledge send; preserve pending content and retry identity.                                                                                                               |
| Redis unavailable             | Durable data remains authoritative; fall back to bounded sync where deployment readiness permits. Review current Redis-dependent readiness rather than assuming this already works. |
| Relay/gateway failure         | Outbox/replay repairs delivery; reconnect with jitter.                                                                                                                              |
| Slow client                   | Bounded memory; ephemeral drops or explicit resync; no silent permanent loss.                                                                                                       |
| Notification provider failure | Chat remains sent and readable; delivery retries separately.                                                                                                                        |
| Model provider failure        | Human messaging remains usable; Maya has a visible run failure and safe retry.                                                                                                      |
| Scanner failure               | Text/history work; affected file remains unavailable with clear processing failure.                                                                                                 |
| Search/index lag              | Canonical history remains usable; stale/deleted content never bypasses access checks.                                                                                               |
| Permission change             | Reauthorise queued actions and publication; stop subscriptions and pending disclosures.                                                                                             |

Expose independent switches for chat admission, live delivery, Maya reads, Maya writes, media processing and external notifications. Recovery must preserve message history and committed action receipts.
