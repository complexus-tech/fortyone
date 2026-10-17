# Native chat contract

This is the shared implementation contract for the first native chat release. All HTTP paths below are relative to `/workspaces/{workspaceSlug}/chat`. The authenticated workspace and actor come from middleware; request bodies never choose an author or tenant.

## Authorization and persistence

- Team conversations are provisioned for every team. Only current, active workspace members who also belong to that team can read or write its chat. Workspace administrators do not bypass chat membership. Credential team restrictions still apply.
- Direct conversations have exactly two distinct human participants in the same workspace. Both must be active workspace members when the conversation is created. Reading requires the current actor to be one of its participants and an active workspace member. Team-restricted credentials cannot access DMs.
- Maya uses the existing active system actor, resolved by bootstrap. Her displayed membership is implicit in team conversation member lists. System replies carry a separate requesting human whose current conversation access is checked again.
- All writes, sequence allocation, and `chat_outbox` events commit in one transaction. The outbox contains identifiers, not message bodies. SSE is an invalidation signal; authorized HTTP reads remain authoritative.
- Migration `000229_team_chat` owns conversations, messages, mentions, reactions, read cursors, attachment associations, and the outbox. Applied migrations remain immutable. SQL lives in the chat repository and is generated through SQLC.

## JSON shapes

All fields use camelCase. UUIDs and timestamps are strings. Timestamps are UTC RFC3339. `sequence` is a conversation-local monotonic integer; `version` starts at 1 and increments when the message is edited or deleted.

```ts
type ChatPerson = { id: string; fullName: string; avatarUrl: string | null; isSystem: boolean };
type Conversation = {
  id: string; kind: "team" | "dm"; teamId: string | null;
  name: string; description: string; color: string | null;
  members: ChatPerson[]; memberCount: number;
  lastMessageAt: string | null; lastMessageText: string;
  unreadCount: number; unreadRootCount: number; unreadReplyCount: number;
  lastReadSequence: number;
};
type Message = {
  id: string; conversationId: string; threadRootId: string | null;
  creationId: string; sequence: number; version: number;
  author: ChatPerson; text: string; richContent: object | null;
  createdAt: string; editedAt: string | null; deletedAt: string | null;
  mentionedUserIds: string[]; agentMentions: ("maya")[]; attachmentIds: string[];
  attachments: {id: string; fileName: string; mediaType: string; sizeBytes: number}[];
  reactions: { emoji: string; count: number; reacted: boolean }[];
  replyCount: number; unreadReplyCount: number;
  lastReplyAt: string | null; replyAuthors: ChatPerson[];
};
type MessagePage = {
  items: Message[]; olderCursor: string | null; newerCursor: string | null;
  hasOlder: boolean; hasNewer: boolean;
};
```

Conversation member payloads are capped at 50, including Maya first, while `memberCount` represents the full current membership.

Root messages include reply summaries and up to three recent distinct reply authors, never nested reply histories. Deleted leaf messages are omitted; a deleted root with visible replies is returned as a tombstone with empty content, mentions, attachments, and reactions. The database retains deletion records for retry and ordering correctness.

## HTTP endpoints

| Method and path | Request | Response |
| --- | --- | --- |
| `GET /unread` | — | `{totalUnread, unreadThreads, conversations: [{conversationId, unreadCount, unreadRootCount, unreadReplyCount}]}` across every accessible conversation; zero-count conversations omitted |
| `GET /threads/unread` | Optional `conversationId`, `cursor`, `limit` (default 50, max 100) | `{items: (Message & {conversationName, conversationKind, firstUnreadReplyId})[], nextCursor}` containing roots with unread replies, newest unread-reply activity first; `firstUnreadReplyId` is the earliest unread reply by sequence |
| `GET /conversations` | Optional `search`, `cursor`, `limit` (default 100, max 200) | `{items: Conversation[], nextCursor: string|null}` sorted by latest activity, then ID; empty team rooms follow active conversations |
| `POST /conversations/direct` | `{userId}` | Conversation; repeated or concurrent requests return the same participant pair |
| `GET /conversations/{id}` | — | Conversation |
| `GET /conversations/{id}/members` | `q`, `limit` default 15, max 50 | `{items: ChatPerson[], nextCursor: null}`; bounded suggestions search the full current membership |
| `GET /conversations/{id}/messages` | One of `before`, `after`, `around`; `limit` default 50, max 100 | MessagePage containing roots, always ascending by sequence |
| `GET /messages/{id}` | — | Message |
| `GET /messages/{rootId}/replies` | One of `before`, `after`, `around`; `cursor` aliases `before`; `limit` default 50, max 100 | MessagePage of replies, ascending by sequence; initial request starts at the oldest reply |
| `POST /conversations/{id}/messages` | `{creationId, threadRootId?, text, richContent?, mentionedUserIds?, agentMentions?, attachmentIds?}` | Persisted Message; 201 when created, 200 for identical idempotent replay |
| `PATCH /messages/{id}` | `{version, text, richContent?, mentionedUserIds?, agentMentions?, attachmentIds?}` | Updated Message; author only; 409 if version is stale |
| `DELETE /messages/{id}?version=N` | No body | `{id, conversationId, threadRootId, version, deletedAt}`; author only; repeated deletion succeeds |
| `PUT /messages/{id}/reactions/{emoji}` | No body | Updated Message; desired-state add, idempotent |
| `DELETE /messages/{id}/reactions/{emoji}` | No body | Updated Message; desired-state remove, idempotent |
| `POST /conversations/{id}/read` | `{throughMessageId, threadRootId?}` | `{conversationId, threadRootId, lastReadSequence, unreadCount}`; monotonic and scoped to the observed message |
| `POST /conversations/{id}/typing` | `{isTyping: boolean, threadRootId?: string|null}` | 204, including throttled repeats; current human membership and selected root scope are required |
| `GET /conversations/{id}/search` | `q`, `cursor`, `limit` default 25, max 100 | `{items: Message[], nextCursor: string|null}`; newest first, includes roots and replies, excludes deleted content |
| `GET /conversations/{id}/resources` | `kind=all|files|media|links`, `sort=newest|oldest` (default `newest`), `q`, `cursor`, `limit` | `{items: Resource[], nextCursor: string|null}`; authorized, nondeleted message resources |

Message cursors are opaque values representing a sequence. An initial root request returns the latest window. `before` returns the preceding window; `after` returns the following window. `around` is a message UUID and returns a balanced window containing its root (a reply resolves to its root). On the replies endpoint, `around` selects a reply within the specified root and returns its surrounding replies. Cursors and `hasOlder`/`hasNewer` describe the actual returned window, independent of direction. Invalid combinations return 400; a source outside the conversation returns 404.

`creationId` is a client-generated UUID retained across retries. Uniqueness is scoped to author and conversation. Reusing it with different content, thread, mentions, or attachment IDs returns 409. The server accepts up to 4,000 text characters, bounded rich-text JSON, and the existing finite reaction catalog. Mentions must identify visible participants; only `maya` is accepted as an agent mention and only in team rooms. Edits preserve the original creation identity and timestamp.

The conversation read cursor tracks root messages from other authors. Replies have separate per-thread read cursors, so opening a conversation does not silently mark unopened threads read. A read request must reference a message from its specified root/thread scope; it cannot advance based on a client clock. Its `unreadCount` remains scoped to that timeline or thread. Conversation projections and the workspace unread summary include both roots and replies, each counted once, excluding the viewer's own messages and deleted content. A missing cursor starts at zero; existing read cursors are preserved when unread counts are introduced. `unreadThreads` counts roots with at least one unread reply, and `unreadReplyCount` on a root is independent from its lifetime `replyCount`. Unread-thread pagination can discover old roots outside loaded conversation history, including deleted roots with visible unread replies; cursors are scoped to the viewer, workspace and conversation filter.

Resources use `{id, conversationId, messageId, kind:"file"|"link", title, description, author:ChatPerson, createdAt, attachmentId?, fileName?, mediaType?, sizeBytes?, url?}`. Attachment upload and signed access URL routes are supplied by the existing storage lifecycle integration; the core does not return storage keys or manufacture downloads. Uploads are pending and owned by their uploader in a specific conversation. Creation atomically claims them; published file access rechecks message visibility and current conversation access. Link resources are derived from validated rich-text links without remote preview fetching.

The `files` filter includes every attachment; `media` narrows attachments to image and video MIME types. Resource sorting applies to the full conversation, with resource ID as the deterministic tie-breaker for equal message dates. Resource cursors are scoped to the conversation, normalized filter, sort, and trimmed search query; clients must start a new cursor chain when any of these change.

## Backend integration ports

The core constructors are `chatrepository.New(log, pool)` and `chatservice.New(repo, options...)`; HTTP registration is `chathttp.Routes(Config{Log, SecretKey, BrowserSessions, WorkspaceResolver, Service}, app)`. Bootstrap supplies Maya's resolved UUID with a service option. Generated database types remain repository-private.

The repository exposes `ClaimEvents(ctx, limit)`, `CompleteEvent(ctx, eventID, leaseToken)`, `ReleaseEvent(ctx, eventID, leaseToken)`, `RenewEventLease(ctx, eventID, leaseToken)`, and `RecipientIDs(ctx, workspaceID, conversationID)` for the dispatcher. Every claimed event includes its lease token, event ID, type, workspace/conversation/message IDs, message sequence/version, actor ID, and creation time. The dispatcher claims only its next event, renews the active 60-second lease every 20 seconds, and cancels delivery on renewal failure. Redis hints have a concurrency limit of 16; notification writes remain sequential. `ReadRecipientIDs(ctx, event)` authorizes only the reader, so read acknowledgements never enumerate all conversation members. Unchanged read cursors create no event. `NotificationRecipients(ctx, event)` returns creation-time recipient snapshots, rechecking live message access before delivery. Edits do not create new mention notifications. Recipient lookup rechecks current memberships; the SSE event envelope can coalesce to `chat.updated` by conversation ID.

Human typing uses `WithTypingBroadcaster(chatadapter.NewTypingBroadcaster(redis, publisher))`. Its ephemeral `chat.typing` SSE event is `{type:"chat.typing", workspaceId, userId, conversationId, actorId, fullName, threadRootId:null|string, isTyping, expiresAt}`. The server derives the actor identity and name, excludes the actor from recipients, and rechecks live membership before the throttle and before fanout. Redis uses one three-second key per workspace actor; repeated activity in the same scope is throttled, while start/stop and scope changes pass immediately. Both start and stop expiry timestamps are assigned atomically by Redis when accepted, six seconds into the future. Clients immediately remove stopped actors, ignore older events, and expire activity without receiving a stop. There are no database writes, durable events, or reconnect replay for typing.

The service provides `GetMessage`, `BoundedContext`, and `CreateAgentReply` for bootstrap-owned Maya adapters. The agent reply command separates `Requester` and `Author`, requires an idempotent creation UUID, and rechecks the requesting human's access in the same transaction as the reply. It uses the same message and outbox tables as human messages. Existing durable agent invocations own model runs, approvals, task proposals, and delivery recovery; chat does not create a second AI job system.

## Verification boundary

Core tests cover cross-workspace access, private team and DM isolation, inactive and removed membership, stale versions, creation replay/conflicts, reaction desired-state races, root/reply deletion behavior, scoped pagination/search, monotonic cursors, transactional rollback, and outbox leasing. Migration execution and sample seeding are local-only and separately owned by bootstrap work; this contract authorizes no production writes.

## Workspace rollout

Messages is enabled by default for existing and new workspaces. The persistent
switch is `workspace_settings.chat_enabled`, exposed as `chatEnabled` through the
existing workspace settings API and **Workspace Settings → General → Features**.
The authenticated settings update route remains admin-only. Chat uses the same
settings table and client hooks as Objectives; it does not require a new webhook.
The former `APP_CHAT_ENABLED` and `APP_CHAT_WORKSPACE_IDS` configuration variables
have been removed.

`GET /workspaces/{workspaceSlug}/chat/capabilities` requires authentication and
current workspace membership. Its normal API response contains
`{ "enabled": boolean, "attachmentsEnabled": boolean }`; it remains available
while Messages is disabled and is never cached. Attachments and voice recordings
are available when Chat is enabled, subject to the existing upload validation and
permissions. No scanner service or scanner environment variable is required.
See [Chat upload validation](chat-file-scanning.md) for file limits and persisted
attachment states.

Every chat endpoint checks the workspace setting after live membership resolution, including
attachment reads/uploads, thumbnails, typing and transcription. Service checks
also prevent Maya from reading or publishing in disabled workspaces. The shared
SSE endpoint continues to serve other application events but suppresses chat events.
Already queued push and email delivery filter disabled chat items before sending.

Outbox and transcription claims filter enabled workspaces before acquiring leases.
Disabling preserves message data and pending jobs; enabling again resumes those
jobs. Maya runs are reauthorized before processing and mutation; a run denied while
disabled may need an explicit retry after re-enabling. Already issued short-lived
attachment URLs cannot be revoked by this workspace setting.

Disable Chat in the existing workspace Features settings to stop new work for that
workspace. API and worker processes read the authoritative setting; no configuration
redeploy or restart is required. Re-enabling resumes eligible pending work without
deleting conversations. An external call already in progress cannot be recalled;
stop/drain workers when an immediate pause of those calls is required. Do not remove
migrations or delete chat rows as a feature rollback mechanism.

The client refreshes capabilities immediately after its own settings mutation and
also rechecks on focus, reconnect and a 30-second foreground polling interval.
Other open clients can retain their previous navigation briefly; backend access
checks still enforce the current setting.

### Activity tracking under realtime traffic

Every request continues to read current account, workspace membership, session
revocation, and policy state from authoritative storage. Authorization is not
cached. Browser session registry timestamps and user/workspace activity timestamps
are refreshed at most once per minute while activity continues; existing inactivity
warnings clear on the next authorized access. Workspace and user touches are
independent and do not hold both rows across the update. Fresh browser sessions
skip registry writes only after the live policy read has verified the owner,
issuance time, expiry, and stored metadata.

### Concurrent message writes

Chat mutations retain Serializable transactions and reuse the caller's creation
ID. PostgreSQL serialization and deadlock failures retry the complete unit with
at most eight attempts, with capped exponential jitter between attempts (at most
630 milliseconds of cumulative backoff). Request cancellation stops retries and
interrupts the wait. Each failed transaction rolls back before the next attempt;
receipts and outbox events commit together, and post-commit callbacks run only
for the successful transaction. Transport errors during COMMIT are not replayed
automatically because their outcome may be ambiguous.
