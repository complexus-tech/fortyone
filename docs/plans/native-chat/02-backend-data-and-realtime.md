# Backend data and realtime architecture

## 1. Module ownership

Add `apps/server/internal/modules/chat` using the established `domain`, `service`, `repository`, `repository/queries`, `repository/sqlc`, and `http` structure. Wire the API and workers through bootstrap. Use Go 1.26.9, native pgx, SQLC, typed actors, existing request/error conventions, and the repository's unit-of-work/outbox contract.

Chat owns conversations, native messages, threads, reactions, pins, read state, conversation preferences, chat-resource links, history/replay, and chat-specific authorisation. Existing modules remain authoritative for team membership, users, workspaces, stories, attachments, notifications, billing, and personal Maya sessions.

The messaging module remains the shared provider-neutral assistance capability. Extend the existing `agentinvocations` engine with a trusted native-team-chat surface adapter; that engine owns durable acceptance, leases, lifecycle, proposals, continuations, execution/publication receipts and recovery. Chat owns source selection, conversation authorisation, audience-safe context and native message publication. Add evidence and per-item action records to that shared engine where required; do not create a second chat-only run lifecycle or unrestricted tool registry.

At inspected HEAD `0ee7c80448b722eeca817e60b76c630cfdf957f1`, migrations 000221–000223 and `internal/bootstrap/agentadapter` implement the story-comment integration. `agentinvocations.Accept` already joins a source unit of work, while `Runtime` separates provider preparation from fenced database apply/publication. Reuse these contracts, preserving the story-comment surface and its existing behaviour. Add explicit surface routing and claim compatibility before enabling chat: the current worker binds the engine to `CommentRuntime`, so changing an input surface string alone is insufficient. This is repository evidence, not confirmation that any environment has deployed those migrations. See [acceptance](../../../apps/server/internal/modules/agentinvocations/repository/acceptance.go), [runtime contract](../../../apps/server/internal/modules/agentinvocations/service/service.go), and [worker wiring](../../../apps/server/internal/bootstrap/worker/agent_invocations.go).

The current shared `Store.Execute` wraps one runtime `Apply` and one outcome in a fenced atomic unit. Independent batch-item commits and partial completion require an explicit engine extension: persist the immutable approved item plan, execute each item in its own fenced domain transaction with its receipt, checkpoint progress, and resume only unresolved items. Do not loop separate commits inside the existing atomic callback or create a chat-only execution engine. Preserve the existing single-action path; qualify mixed-version workers and cancellation between items before enabling batches.

Do not convert story comments into chat messages, personal AI JSON transcripts into multi-user rooms, or Yjs documents into chat storage. They have different access, ordering, and lifecycle rules. Keep the current Slack integration independent; native chat does not automatically mirror Slack or create a second provider-webhook runtime.

## 2. Data model

Names below are proposed domain/storage names. Select migration numbers only when implementing, after reconciling the active migration chain. All tenant relations must enforce ownership through composite foreign keys or equivalently strong scoped constraints. UUID validation alone is insufficient.

| Record                              | Fields and constraints                                                                                                                                                                                                                                                                                                        |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `chat_conversations`                | `id`, `workspace_id`, closed `kind` enum, optional `team_id`, topic, creator, archive state, `access_generation`, `last_event_seq`, `last_root_seq`, timestamps. One canonical team conversation per team.                                                                                                                    |
| `chat_direct_members`               | Explicit human members of direct conversations; exactly two for the initial human DM type. Immutable participant set initially. A unique canonical pair key prevents duplicate one-to-one DMs within a workspace.                                                                                                             |
| `chat_user_state`                   | Conversation/user/workspace, `last_read_root_seq`, separate manual-unread marker, mute/snooze/notification policy, optional pinned/saved position. Team access is derived from authoritative team membership, not this row.                                                                                                   |
| `chat_agent_bindings`               | Optional explicit availability/configuration record for native agents, linked to the existing system identity. It grants no human membership or universal resource access. A projected implicit Maya binding is sufficient for the first version.                                                                             |
| `chat_messages`                     | UUID, workspace/conversation, author ID/kind, client message ID, thread root ID, optional immutable `reply_to_message_id`, root/reply sequence, creation event sequence, body schema version/document/plain text, body revision, aggregate `state_event_seq`, created/edited/deleted timestamps, optional originating run ID. |
| Root thread state                   | `last_reply_seq`, reply count, latest-reply metadata on the root, or a dedicated `chat_threads` row if query/locking evidence warrants it.                                                                                                                                                                                    |
| `chat_thread_user_state`            | Root/user, `last_read_reply_seq`, follow mode, mute override. Create for participants/followers/readers, not every possible user-thread pair.                                                                                                                                                                                 |
| `chat_catchup_checkpoints`          | Opaque ID, account/workspace/user/conversation, immutable pre-visit root and included-thread read vector, thread-scope policy/version, captured watermark, created/expiry timestamps. Created during visible conversation bootstrap before new read acknowledgements; never grants content access.                            |
| `chat_message_revisions`            | Revision metadata and any permitted retained content; author, time, reason. Deletion/retention must remove prohibited plaintext here as well as in the main message.                                                                                                                                                          |
| `chat_message_mentions`             | Message/revision, principal ID and mention kind, with scoped uniqueness and eligible-target validation.                                                                                                                                                                                                                       |
| `chat_message_reactions`            | Message/user/canonical emoji unique key. Set/unset operations, never a retry-sensitive toggle.                                                                                                                                                                                                                                |
| `chat_message_attachments`          | Message/attachment/workspace/conversation relation, display order, original-name metadata, uploader/reference provenance.                                                                                                                                                                                                     |
| `chat_resource_links`               | Message or run to story/document/objective, relationship type, source-selection ID and provenance. Store IDs; hydrate protected details on read.                                                                                                                                                                              |
| `chat_pins` / `chat_saved_messages` | Shared pins with author/time versus private user saves. Keep their access semantics separate.                                                                                                                                                                                                                                 |
| `chat_events`                       | UUID event ID, conversation/workspace, transactional `event_seq`, schema version, event type, entity ID/revision, actor, time, bounded metadata. Unique conversation/event sequence.                                                                                                                                          |
| `chat_outbox`                       | Event/consumer or destination, state, attempt count, next attempt, claim token/generation, lease expiry, terminal failure.                                                                                                                                                                                                    |
| Consumer receipts                   | Unique consumer/event identity for notification, indexing, and run-trigger side effects.                                                                                                                                                                                                                                      |
| Shared agent records and extensions | Reuse `agent_invocations`, proposals, outcomes, usage and recovery receipts. Add surface-scoped evidence manifests and stable per-item action results where required by the Maya document; no independent chat run ledger.                                                                                                    |
| Upload intent                       | Uploader, destination conversation, byte/type limits, checksum, expiry, attachment ID, scan/processing state.                                                                                                                                                                                                                 |

Use partial unique indexes for root order `(conversation_id, root_seq)` and reply order `(thread_root_id, reply_seq)`, plus tenant-scoped message identity and client-send uniqueness. Index event replay by `(conversation_id, event_seq)`, user state by `(workspace_id, user_id, conversation_id)`, replies by root/sequence, and queue claims by state/next-attempt. Add real query-plan tests for history, unread, search and cleanup with representative data.

`thread_root_id` groups the thread; `reply_to_message_id` identifies the exact message being answered. Persist and validate the latter against the same conversation and thread. Native Maya continuation resolves only that explicit pending question/proposal and its originating run; never select the last Maya message or infer an approval from thread membership. Bind the immutable target, source revision and expected proposal/run version into the send/continuation identity and retry hash. A conflicting or stale target requires a visible conflict/clarification; a plain thread reply without an eligible target does not silently continue a pending run.

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
2. Validate the request body, root ownership, optional exact reply target, attachment upload finalisation/reference eligibility, mention eligibility, archive state and rate limit. Resolve any continuation against its immutable target/source/proposal version. A valid processing placeholder is distinct from a scan-clean file permitted for download or extraction.
3. Look up the scoped `client_message_id`. An identical retry returns the existing canonical result; reuse with different request bytes/content is a conflict.
4. In the transaction, recheck membership against concurrent removal, allocate order, and insert message, mentions, attachment relations, resource links, event, and outbox intent.
5. If the validated structured mention or authorised continuation invokes Maya, call the existing `agentinvocations.Accept` through the native adapter in the same unit of work. Bind a stable surface/source idempotency key, initiating human, source revision/hash and conversation scope. Acceptance must commit or roll back with the message; a second after-send acceptance transaction would leave a lost-invocation window.
6. Commit, then return the canonical ID, revision, sequences, timestamps and run identity.
7. Relay live projections/hints, indexing, notifications, and AI work after commit. Reuse after-commit invocation dispatch and the engine's durable recovery when enqueue fails; never acknowledge an invocation that exists only in Redis.

The service must prove the concurrent membership-removal boundary using locking or an equivalent tested transaction design. Middleware-only checks and a stale membership cache are insufficient.

The shared idempotency service provides useful hashes, scopes and leases, but does not automatically make an unrelated domain transaction safe. Couple mutation and receipt, enforce permanent logical-send uniqueness in the domain, or reconcile from a committed result record. A 24-hour generic receipt expiry must not permit a queued old send to duplicate its original message.

### Outbox execution

Use bounded `FOR UPDATE SKIP LOCKED` claims with fencing tokens, retry classification, jitter, attempt limits, poison-message diagnostics and reconciliation. An outbox publish may happen twice if a process dies before marking completion. Consumers must therefore apply an event once by durable identity.

Keep interactive chat dispatch separate from slower maintenance queues. No network call, model call, upload, email, or Redis publish occurs inside the message transaction. Do not use `SKIP LOCKED` for normal history reads; it is a work-claim mechanism. [PostgreSQL SELECT locking](https://www.postgresql.org/docs/18/sql-select.html)

## 4. Transport decision

Use **HTTP commands/history plus one multiplexed authenticated SSE stream per active client/workspace** as the chosen transport. Send full authorised, versioned projections for actively viewed conversations and coalesced hints for sidebar/unloaded conversations. This reuses the existing application pattern and avoids an extra fetch for each visible message. Qualify it at the initial 1,000-user workload and subsequent growth checkpoints; user count alone is not a reason to replace SSE with WebSocket. No audited native-client requirement currently forces WebSocket.

SSE carries live state, but PostgreSQL remains authoritative for committed state and recovery. Redis Pub/Sub can lose events while a subscriber is disconnected; its relay is an acceleration path, never the only durable record. [Redis Pub/Sub](https://redis.io/docs/latest/develop/use-cases/pub-sub/)

For active conversations, emit a bounded event envelope containing `schemaVersion`, `eventId`, `conversationId`, `eventSeq`, `accessGeneration`, the active `subscriptionGeneration`, event type, and the authorised changed projections/tombstones. Each projection carries its entity version, including message `bodyRevision`/`stateEventSeq` where applicable. Emit only committed state. Do not include unrelated private tool state, unrestricted resource details, or reusable signed download URLs. Hydrate resource/file details under their own access and version contracts. Reuse a projection across equivalent authorised audiences where safe; avoid a database hydration query for every recipient.

For unloaded conversations, coalesce recent changes into a hint such as:

```json
{
  "type": "chat.conversation.changed",
  "schemaVersion": 1,
  "conversationId": "uuid",
  "latestEventSeq": 481,
  "accessGeneration": 7
}
```

An active recipient can render the canonical projection immediately; the sender also reconciles its optimistic message from the HTTP response. `/chat/sync` remains the authoritative recovery path for gaps, reconnects, bootstrap and event expiry. A live event advances a conversation's contiguous cursor only when every preceding event has been processed. Receiving event 481 after 479 may update the visible entity by version, but must retain cursor 479 and request event 480 onward; neither `latestEventSeq` nor a newer entity version proves complete event coverage. Keep cursor bookkeeping separate from projection merging.

Sidebar hints trigger a debounced, coalesced summary refresh for eligible conversations, not a history fetch for every room. Share in-flight requests and fetch room deltas when that room becomes active. A lost inactive-room hint is repaired on next Chat/sidebar open, foreground or reconnect; the 35-second silent-loss objective applies only to the active conversation. Add an explicit account-summary watermark mechanism before promising that deadline for inactive badges. Benchmark full-projection fanout, hint coalescing, payload sizes and recovery fetch amplification under identical workloads. WebSocket remains an ADR only if a demonstrated client or bidirectional-traffic requirement justifies it; both transports still need this ordering, authorisation and recovery design.

### Active subscriptions and the history race

An authenticated HTTP command associates a bounded active-conversation set with the caller's current SSE connection and returns an acknowledgement/generation. Start with one visible conversation, allowing only a small explicit limit for supported split views; this limit is a configuration decision. Validate connection ownership, workspace and current conversation access server-side. Client-supplied IDs express interest, never a subscription capability. Hidden/unmounted views release active subscriptions; connection cleanup removes them. Sidebar hints are still filtered by current access.

Route that command to the process owning the stream, or use a tested shared control path; a load-balanced HTTP request may reach another replica. Acknowledge only after the owning gateway installs the authorised set. Use versioned updates so a delayed older command cannot restore a previous view's subscriptions. An opaque connection ID is scoped to its authenticated account/session, not sufficient authority by itself.

Register the authorised subscription before acknowledging it. The client buffers bounded events for that acknowledged generation while loading a consistent history snapshot and baseline `W`; then it merges versioned projections, discards buffered events already covered by `W`, and processes contiguous events above `W`. It calls sync to fill any gaps and reach a captured watermark. A disconnect or unknown acknowledgement restarts this handshake; overflow requests resync instead of silently discarding required state. Workspace/conversation switches invalidate stale responses and generations. A single atomic subscribe-and-bootstrap endpoint is also valid if it proves the same no-gap boundary. Never assume that fetching history and subsequently opening a stream leaves no race.

Bootstrap establishes a new cache baseline: replace the returned history window and metadata, and invalidate older cached pages/projections not certified by that snapshot before treating `W` as covered. A one-page history response must not silently certify stale offscreen pages. Older content remains available through freshly authorised pagination.

Typing and presence use short-lived Redis state and throttled authenticated HTTP updates. Start with conversation-scoped typing, not workspace-wide presence. Use approximately one typing update every three seconds, an expiry around six seconds, and stop signals as best effort. These are tunable UX defaults. Maya previews use batched ephemeral frames keyed by run and offset; canonical snapshots recover them. Release safe progress immediately and validate source/audience eligibility before broadcasting answer blocks; a final-only check cannot retract leaked tokens.

### Required gateway changes

- Add server-authorised conversation subscriptions and bounded active sets, with membership-generation invalidation before fanout. Do not publish private conversation IDs or content to the existing unrestricted workspace-update lane. Recheck queued payload eligibility at emission; losing an invalidation must not preserve an indefinitely authorised subscription.
- Share Redis topic subscriptions per gateway process and fan out to locally eligible connections. Do not multiply Redis listeners by every tab and every conversation unnecessarily.
- Set bounded queue bytes and event counts. Drop stale typing first; when durable state is behind, request resync and close/drain the slow connection rather than silently claiming it is current.
- Maintain origin checks, session expiry, write deadlines, request limits, heartbeats, graceful drain, connection limits and subscriber cleanup.
- Validate the chosen SSE implementation's HTTP/2, buffering/compression behaviour, proxy timeouts, native stream support and full-projection delivery in the deployment benchmark. Heartbeats must contain data and fit the actual proxy policy. [MDN SSE guidance](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events), [AWS load balancer attributes](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/edit-load-balancer-attributes.html)
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

The client advances only through a contiguous event prefix it successfully processed, whether supplied by SSE, sync or their overlap. It must not advance to the advertised high watermark before accounting for every intervening event. Dedupe by conversation/event sequence and stable event identity. A deleted entity still produces a no-content event or tombstone that safely advances the cursor. Projection versions prevent stale rendering; they do not authorise skipping event gaps.

Each page is a new short snapshot; pagination does not hold a transaction open. Entity versions prevent a later replay page from overwriting newer visible state. Sync means convergence to current state, not disclosure of historical deleted plaintext.

Initial history load returns its event baseline consistently with the loaded page. Use the acknowledged subscription/buffer handshake above, then sync after that baseline. Reconcile on stream reconnect, network restoration, foreground, workspace change, opening a conversation, ambiguous send, and gateway subscription recovery. Also compare the visible conversation's watermark on a jittered schedule with a maximum healthy interval of 30 seconds, for example 25–30 seconds: this detects a lost final live projection or hint even when no later message arrives.

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

| Operation                           | Proposed route/contract                                                                                                                                                                                                                |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Conversation navigation             | `GET /chat/conversations` with cursor/filter and authoritative unread/attention summaries.                                                                                                                                             |
| Metadata and participant projection | `GET /chat/conversations/{id}` and `/members`; system/human kinds remain internal metadata.                                                                                                                                            |
| Create/find one-to-one DM           | `POST /chat/direct-conversations` with one eligible recipient; canonical pair uniqueness.                                                                                                                                              |
| Message page/bootstrap              | `GET /chat/conversations/{id}/messages` with `before`, `after` or `around`, bounded page size. Explicit visit bootstrap returns a consistent event baseline and server-issued catch-up checkpoint.                                     |
| Send                                | `POST /chat/conversations/{id}/messages`, client identity and versioned body.                                                                                                                                                          |
| Edit/delete                         | `PATCH`/`DELETE /chat/messages/{id}`, expected revision and logical operation ID.                                                                                                                                                      |
| Thread                              | `GET /chat/messages/{rootId}/replies`; send accepts thread root ID plus optional persisted `replyToMessageId`. Continuations additionally bind source/proposal versions to the exact target, with same-conversation/thread validation. |
| Reaction                            | `PUT`/`DELETE /chat/messages/{id}/reactions/{emoji}`.                                                                                                                                                                                  |
| Shared pins/private saves           | Separate idempotent pin/unpin and save/unsave capabilities; pins require conversation permission, saves belong only to the current user.                                                                                               |
| Read/follow                         | `PUT /chat/conversations/{id}/read`, thread-specific read/follow routes.                                                                                                                                                               |
| Preferences                         | `PATCH /chat/conversations/{id}/preferences`, scoped to current user.                                                                                                                                                                  |
| Sync                                | `POST /chat/sync` with bounded conversation cursor vector.                                                                                                                                                                             |
| Active subscriptions                | Authenticated connection-scoped command replacing a bounded active-conversation set; server validates connection ownership and current access, then acknowledges a subscription generation.                                            |
| Typing                              | `PUT /chat/conversations/{id}/typing`, optional root ID and short expiry.                                                                                                                                                              |
| Upload                              | Conversation-bound intent, finalize, and authorised attachment download endpoints.                                                                                                                                                     |
| Search                              | `GET /chat/search` with query and authorised conversation/author/date filters.                                                                                                                                                         |
| Maya                                | Extend existing agent-invocation capabilities for run read, cancel, clarify, proposal review, execute/retry-remaining; surface/destination guards and no client-controlled authority fields.                                           |
| Task claim                          | Existing story capability extended with conditional `expectedUnassigned`/version semantics if necessary.                                                                                                                               |

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
