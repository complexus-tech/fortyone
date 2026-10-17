# Clients and notifications

## 1. Implementation boundary

This is a proposed implementation contract for web, hosted desktop and Expo mobile at the initial target of **1,000 concurrent users**. Repository findings describe the committed implementation inspected on 16 October 2026, including Maya in story comments; they do not establish deployed behaviour or measured chat capacity. Read this with [product interactions](01-product-and-interactions.md), [backend and replay](02-backend-data-and-realtime.md), and [Maya orchestration](03-maya-orchestration.md).

The client must make accepted messages, pending operations, unread content and Maya's actual outcomes distinguishable. The server owns identity, membership, ordering, permissions and committed results. Client optimism improves responsiveness without inventing authority or success.

The existing comment collaboration is now an implemented foundation: typed Maya mentions, ordinary threaded answers, recoverable pending presentation, exact-proposal review, conversational continuation, stable creation IDs and independent thread read state. Reuse its contracts and extract neutral presentation primitives. Story comments remain their own storage and scoped surface; the new chat feature must not import story-private controllers throughout its implementation.

## 2. Module map and reusable foundations

| Location                                                                   | Implementation responsibility                                                                                                                    |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `apps/projects/src/modules/chat/public/`                                   | Chat page, navigation summary, event-reducer entry point and typed capabilities exposed to the shell.                                            |
| `modules/chat/components/`                                                 | Conversation navigation/header, timeline, message actions, thread, composer, search, files, Maya status and result blocks.                       |
| `modules/chat/hooks/`, `queries/`, `actions/`                              | Typed server boundary, history, commands, read tracking, drafts and operation reconciliation.                                                    |
| `modules/chat/cache/`                                                      | Query keys, message windows, version checks, pending operations and deterministic sync reducers.                                                 |
| `apps/projects/src/shell/chat/`                                            | Composition of Chat with existing story details, navigation and global attention.                                                                |
| `apps/projects/src/shell/realtime/`                                        | Shared transport lifecycle; chat owns decoding and applying its resource events.                                                                 |
| `apps/server/internal/modules/agentinvocations/` and a native-chat adapter | Existing durable invocation authority plus new conversation/message bindings; expose safe snapshots and ordinary message publication to clients. |
| `apps/projects/src/shell/desktop/` and `apps/desktop/src-tauri/`           | Capability negotiation, native destinations, cache policy, notifications and explicit outbox support.                                            |
| `apps/mobile/modules/chat/` and mobile routes                              | Native presentation, lifecycle, composer, timeline, deep links and local operation storage.                                                      |
| `packages/lib/` where appropriate                                          | Stable platform-independent DTOs, route/target parsing and pure reconciliation rules. No browser or native runtime dependencies.                 |

Reuse the existing [navigation](../../../apps/projects/src/components/shared/sidebar/navigation.tsx), [Maya avatar](../../../apps/projects/src/components/ui/maya-avatar.tsx), [people stack](../../../apps/projects/src/components/ui/story/people-avatars.tsx), [story results](../../../apps/projects/src/components/ui/chat/story-results.tsx), and [feature composition pattern](../../../apps/projects/src/shell/work-feature-slots.tsx). Extract presentation primitives rather than embedding the entire personal Maya controller.

The current [comment composer](../../../apps/projects/src/modules/story/components/comment-input.tsx) provides Tiptap, mentions, draft recovery, attachment progress and submission guards. Chat needs a neutral, versioned document contract shared with native rendering, rather than reliance on arbitrary comment HTML. Continue using existing work-item terminology, theme tokens, icons and public story capabilities.

The [implemented comments lifecycle](../../../apps/server/docs/maya-comments.md) supplies the following concrete starting points:

| Existing contract                                                                                                                                                                                          | Reuse in Chat                                                                                    | Chat-specific work still required                                                                                                                     |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Mention identity](../../../apps/projects/src/modules/story/hooks/use-comment-mentions.ts) and typed agent targets                                                                                         | Existing persisted Maya ID/avatar; separate human notifications from agent activation.           | Conversation-authorised member catalog and capability-aware picker.                                                                                   |
| [Pending reply projection](../../../apps/projects/src/modules/story/hooks/use-maya-pending-replies.ts) and [status controls](../../../apps/projects/src/modules/story/components/comment-maya-pending.tsx) | Temporary reply, reduced-motion feedback, delayed/offline state, Check again versus Retry.       | Bind to message/root IDs and shared invocation events; never store the placeholder or count it as a reply.                                            |
| [Exact change review](../../../apps/projects/src/modules/story/components/comment-maya-approval.tsx) and versioned [run controls](../../../apps/projects/src/modules/story/hooks/comment-maya.ts)          | Ordinary Maya answer with expandable review and requester-only approval.                         | Typed task proposals/results, expiry display, continuation binding and conditional task claim. Current comment reviews are text, not chat task cards. |
| [Draft/submission identity](../../../apps/projects/src/modules/story/hooks/use-comment-submission.ts)                                                                                                      | Preserve the submitted draft and logical creation ID across retries; do not clear newer content. | Give every chat send a stable ID, not only agent-targeted messages/replies; add chat persistence to native outboxes.                                  |
| [Root/thread read tracking](../../../apps/projects/src/modules/story/hooks/comment-read-tracking.ts)                                                                                                       | Focus/visibility checks and separation of collapsed replies from root reads.                     | Sequence-based contiguous reads, replay-gap handling, immutable catch-up checkpoints and account attention aggregation.                               |

Use one server `agentinvocations` lifecycle with a `native_team_chat` adapter, not a parallel chat run controller. Its existing comment adapter updates only the owning story and publishes ordinary comments; chat adds conversation/audience context, bounded same-team creation tools and message publication. Surface capabilities are explicit: existing comment update support does not implicitly enable chat updates, cross-team actions, personal-history retrieval or attachment extraction. The client renders the server's capability result and permission-checked run snapshot; it cannot widen the tool catalog by exposing a button.

## 3. Navigation, identity and private conversations

Add Chat alongside My work using the existing navigation's selected and collapsed states. Its secondary column contains All unread, Threads, Teams and Direct messages. Keep the selected row stable when other conversations become active. The numeric badge represents attention, not total team message volume.

Proposed typed routes are `/{workspaceSlug}/chat` and `/{workspaceSlug}/chat/{conversationId}`, with optional `thread={rootMessageId}` and `message={messageId}`. A `story={storyId}` parameter may select the contextual work panel. Validate IDs and relationship ownership; a reply must belong to the indicated conversation/root. One route helper must support development path prefixes and production workspace subdomains through existing workspace routing conventions.

Opening Chat restores the last accessible conversation. A direct link loads around its target, rather than downloading history from the end. Refresh, browser Back/Forward and copied links preserve meaningful context. A removed target displays a neutral unavailable state without disclosing inaccessible titles.

Use one side panel for thread or story details. Opening a ticket from a thread preserves the thread, scroll anchor and draft in navigation state; Back restores them. Narrow layouts and mobile use stacked screens. Do not keep hidden duplicate composers mounted and competing for drafts or focus.

Maya uses the existing `members/maya` endpoint, [member hook](../../../apps/projects/src/lib/hooks/members.ts) and persisted system user. Show **Maya** and her actual avatar in the participant stack and mention picker. No appended AI badge or “in every team” banner. A normal profile can explain that Maya is the workspace assistant and describe available help. Internally retain actor kind; do not report a human seat, human presence, last-seen time or read receipt for Maya.

One-to-one human DMs are first-class roadmap scope, with a separate privacy release gate after the team pilot. They have exactly two human participants, immutable initially, and a canonical pair per workspace. Maya is not automatically a participant and does not silently ingest these conversations. The private Maya entry should initially adapt or open the existing personal Maya route; do not copy private transcripts or personal memory into a team room. Group DMs follow later after participant-change and history rules are designed. Avoid displaying nonfunctional DM rows in the pilot as if the capability were released.

## 4. Composer and timeline state contracts

Persist drafts under account, workspace, conversation, root and edit-target identity. Draft restoration must wait for the correct session binding. A late acknowledgement must not clear a newer draft. Browser/native storage failure is visible and must not be mistaken for successful persistence.

The composer supports the approved rich-text nodes, structured member mentions and attachment references. Derive mention IDs from validated structure; quoted text, code blocks and a plain spelling of `@Maya` are not agent activation. A reply to a specific pending Maya question/review can continue that invocation without a mention, using its persisted reply identity and the server's continuation rules. Ordinary text elsewhere does not activate her. Respect IME composition, selection, pasted code and the Enter/Shift+Enter preference. The configured byte, character, depth and file limits must match the server.

Model each local send explicitly:

| State                | Behaviour                                                                                                      |
| -------------------- | -------------------------------------------------------------------------------------------------------------- |
| Draft                | Editable, recoverable content; no server operation exists.                                                     |
| Uploading/processing | Individual files show progress, cancel, failure or scan status.                                                |
| Sending              | Stable client message ID and visible pending row.                                                              |
| Accepted             | Canonical server ID, revision and sequences reconciled into that row.                                          |
| Uncertain            | Connection failed after submission; recover by logical send ID before offering a new send.                     |
| Retryable failure    | Preserve exact content/files and retry the same operation identity.                                            |
| Rejected             | Explain permission, archive or validation failure; preserve recoverable user text subject to retention policy. |

Processing files may appear as unavailable placeholders on an accepted text message under the backend attachment policy. No download, preview or Maya extraction is enabled before readiness. Cancelled local uploads must not create orphan message references. Do not store base64 images or temporary `blob:` URLs in durable message bodies.

Edits carry an expected revision. On conflict, retain the attempted edit and present current content rather than overwrite it. Deletion leaves a root tombstone when replies survive. Newly added human mentions can generate one relevant attention event; repeated saves cannot repeat it. Adding Maya through an edit requires the explicit Ask Maya behaviour specified in the product plan.

The timeline is chronological and cursor-paginated, with stable message keys and variable-height virtualisation. Preserve the first visible message and pixel offset when prepending history, images resolve or previews expand. Keep focused/selected rows mounted where required. Auto-follow new content only while at the live edge or after the user's explicit send; otherwise display a new-message control. Store scroll anchors separately from server read positions.

Root and reply feeds have independent windows, composers and read state. Loading a search hit must not mark intermediate history read. The thread root displays current reply count, participant avatars and unread-reply indication. Subscription/follow state is explicit and reversible; merely reading a thread need not subscribe someone permanently.

Maya uses an ordinary temporary reply under the triggering message, following the implemented comments interaction. It becomes a published reply when available and never increases stored message/reply counts while pending. Reuse “Thinking…”, “Still working on this…” after the current 30-second threshold, and static feedback with **Check again** after three minutes or connection loss. Check again refreshes state without executing a new action; an explicit Retry follows server eligibility and version checks. Additional labels such as “Creating 3 tickets” must reflect observable work, not hidden reasoning. Final replies replace progress rather than create a message per tool call. A channel request replies in its thread with a compact root preview.

A clear, bounded request can create one to five same-team tasks directly under the policy in the Maya plan, only when the whole batch qualifies. Ambiguity produces a normal question; suggestions or actions requiring review produce an exact proposal. Any currently authorised member may answer the specific pending question or review without another mention. The server binds that reply to the pending invocation and creates a linked invocation owned and authorised by the answering person. It must not inherit the previous requester's authority or guess a target from the most recent Maya message. An unqualified affirmative may accept an exact replacement proposal under the answerer's permissions while retaining values, source preconditions and expiry; “yes, but change the deadline” is a new instruction. Only the requester can click a proposal's approval button, bound to its exact version and expiry. Replies to completed answers and Maya-authored messages do not recursively activate the agent. These changes do not silently alter personal Maya behaviour.

Committed task cards contain actual IDs and permission-checked current data. Draft proposals are clearly different. Use existing story queries as authority, conditional assignment for **Assign to me**, and ordinary story detail composition. A partial result reports each committed/failed item; Retry remaining uses stable action identities. Cancellation reports already committed work and never silently erases it.

## 5. Cache, sync and optimistic reconciliation

Use account/workspace-scoped keys for conversation lists, metadata, message windows, thread windows, read summaries, preferences and run snapshots. Include before/after/around parameters in window identity. Keep pending operations separate from canonical entities so a refresh cannot accidentally erase an unacknowledged send. Reuse canonical story cache keys for linked work.

Three independent cursors must never be conflated:

1. `throughEventSeq`: the latest contiguous durable event successfully applied for a conversation.
2. `last_read_root_seq`: the user's acknowledged root-message boundary.
3. `last_read_reply_seq`: the user's acknowledged boundary within a particular thread.

`highWatermark` advertises what the server knows. It is not permission to advance any cursor. Applying edits, reactions or tombstones advances sync state without reading messages. Receiving history, push or SSE advances no read state.

Use **HTTP commands/history and one authenticated multiplexed SSE connection per active client/workspace**. Register the active conversation/thread before requesting its HTTP baseline and buffer incoming events during bootstrap. The history snapshot supplies a consistent event baseline; visible visit bootstrap also obtains the server-issued catch-up checkpoint before starting read acknowledgements. Preserve its ID for “since last here” Maya requests. Apply the snapshot, reconcile buffered events by version and sync beyond the baseline to repair races. Changing the active target updates the authorised subscription; stale responses cannot reactivate a previous workspace or room.

For the active conversation/thread, SSE carries bounded, authorised canonical projections for the affected messages, revisions, reactions, root summaries and publishable Maya state/results. Apply them immediately without a mandatory GET before displaying the change. “Full payload” means the complete projection for that event, not an entire history, private tool trace or unrestricted source snapshot. Referenced files and stories keep their independent access/version checks. Disclosure is checked at fanout as well as subscription; a permission change fences already queued deliveries and invalidates affected local projections when received.

Inactive conversations receive coalesced hints and safe navigation/attention summaries rather than every message body. Treat `chat.conversation.changed` and its advertised high watermark as a reason to reconcile the relevant summary or sync on opening, not a reason to fetch every inactive room immediately. Share one in-flight sync per conversation and bound background work. The sender's command response and the active SSE event converge on the same message identity.

Applying a newer projection is distinct from proving event continuity. A full event above a missing sequence may update its message immediately, but cannot advance `throughEventSeq` across that gap. A filtered thread delivery likewise cannot prove that omitted conversation events were applied. Only a complete contiguous event range, or the explicit coverage contract defined by the backend, advances that cursor; recover gaps through durable `/chat/sync` pages. Apply pages until their returned boundary is reached, preserving newer entity versions already seen live. Persist a cursor only alongside compatible cached state. A crash must not restore a new cursor with an older cache and permanently skip events.

The backend's seven-day proposed replay window can return `reset_required`. Preserve drafts and pending operation identities, replace canonical windows/read summaries from a fresh snapshot, and recover ambiguous sends by identity. A reconnect is a state reconciliation, not a replay of sounds or toasts. Unknown optional events trigger a bounded refresh; incompatible schemas need explicit capability handling.

Reducer invariants:

- HTTP acknowledgement and realtime sync can arrive in either order; match both by client operation ID and canonical ID.
- Ignore duplicate events and older projection versions. Use body revision for text, aggregate state sequence for reactions/thread summaries and domain revisions for shared files/stories. A newer hydrated entity may legitimately arrive before its older event page; the later page must not overwrite it.
- A tombstone prevents a late fetch from resurrecting deleted content. Access-generation changes invalidate affected projections.
- Never infer missing visible messages from sequence subtraction; event/root/reply sequences have different meanings and eligible unread content excludes own/deleted items.
- Roll back only the failed optimistic operation. Whole-cache snapshot restoration must not erase concurrent accepted events.
- Ignore late requests from disposed account/workspace scopes. Cancel requests and clear subscription state on session transitions.
- Do not invalidate all stories and all conversation histories per message or streamed token. Apply resource-specific changes and reconcile bounded windows.

The current [workspace SSE lifecycle](../../../apps/projects/src/shell/realtime/workspace-realtime-connection.tsx) invalidates existing caches on reconnect and has no chat event/replay contract. Extend it through chat-owned parsing/reducers and durable sync; browser `Last-Event-ID` alone cannot cover multiple conversation histories. The existing [comment run hook](../../../apps/projects/src/modules/story/hooks/comment-maya.ts) polls active runs every two seconds and inactive feeds every 30 seconds. Reuse its lifecycle concepts and controls, not its polling transport: two-second polling must not become primary chat delivery or the source of every participant's Maya progress.

Reconcile on foreground, reconnect, conversation open, ambiguous command and gateway recovery. Follow the backend's proposed visible-conversation watermark check with a maximum 30-second interval and bounded jitter, for example 25–30 seconds. That is approximately 33–40 requests/second at 1,000 foreground users before batching. Do not poll every conversation at that frequency. Measure hot-room rendering, memory and sync amplification under the agreed capacity profile.

## 6. Read state, attention and cross-device behaviour

Chat owns all messaging activity. The Chat navigation badge counts unread top-level messages plus unread replies, with each message counted once. Conversation badges show their share of that total, and a Threads entry lists roots with unread replies, including older roots outside the loaded timeline. Own and deleted messages do not count. Typing, edits, reactions and run progress never create message alerts.

Advance a read boundary only while the relevant surface is foreground, focused and actually rendered, through a continuously viewed range from its known boundary. Historical search jumps and virtualised gaps do not prove the intervening messages were seen. Provide an explicit mark-through/mark-all action when the person wants that behaviour. Bound acknowledgements by server-known sequences and merge monotonically; an old device cannot move normal read state backwards.

Manual mark-unread is a separate durable marker with explicit resolution rules. Capture the visit's initial unread boundary so new acknowledgements do not erase the divider during reading. Thread read state remains independent: viewing the channel or root does not consume unseen replies. This distinction is already implemented in the [comment reader](../../../apps/projects/src/modules/story/hooks/comment-read-tracking.ts), with persisted per-root reply cutoffs. Reuse its focus/expanded-viewport gates, but replace its timestamp/container boundary model with chat's sequence and gap contract; implementation of thread reads does not itself establish gap-safe chat catch-up.

Resolving a notification link does not itself mark the content read. After the target is rendered and viewed, acknowledge its attention version. Advance root/thread read position only if the viewed range is contiguous with its existing boundary; an out-of-order target must not clear an intervening unread gap. Versioned attention acknowledgements cannot clear later events. Catch-up remains watermark-based unless sparse viewed exceptions are deliberately added. Updates propagate to other devices; periodically reconcile badges from server aggregates. Already displayed OS notifications may not disappear immediately.

Chat activity never appears in the general Notifications list or its unread count. Internal notification receipts may still provide deduplication and compatibility for old notification links. Chat sends a separate recipient-scoped `chat.notification` event; the native app badge combines the work notification count and chat unread-message count. The current SSE subscription and badges cover the active workspace while the app is running. Account-wide aggregation and delivery after quitting remain separate work.

## 7. Notification policy and preferences

The approved default is alerts for all incoming messages in accessible, unmuted conversations, with sound. Desktop/browser permission remains explicit. Chat settings provide device-local banner and sound controls, conversation mute, quiet hours and sound preview. Mute and quiet hours suppress interruptions without changing unread counts. Mobile delivery retains its existing push preference contract; these device-local settings do not change another device's preferences.

| Event                          | Chat state       | Inbox attention                   | Desktop/mobile interruption              | Email default                 |
| ------------------------------ | ---------------- | --------------------------------- | ---------------------------------------- | ----------------------------- |
| Ordinary team root message     | Unread           | No                                | If unattended and allowed                | No                            |
| Direct human mention           | Unread           | No                                | If unattended and allowed                | Existing email preference     |
| Human DM                       | Unread           | No                                | If unattended and allowed                | Existing email preference     |
| Thread reply                   | Thread unread    | No                                | If unattended and allowed                | Existing email preference     |
| Maya published answer/result   | Thread unread    | No                                | Published message policy                 | Existing email preference     |
| Maya creates/assigns work      | Result card      | Existing affected-user work rules | Aggregate overlapping result/work alerts | Existing work cadence         |
| Reaction, edit, deletion       | Update content   | No                                | No repeat alert                          | No                            |
| Typing, presence, run progress | Ephemeral update | No                                | No                                       | No                            |

The initial conversation policy is all messages or muted. Muted unread remains in the Chat and conversation totals. Quiet hours suppress delivery during the selected device-local time window; exiting quiet hours does not replay old sounds. Mentions-only modes, per-thread overrides, snooze and synchronized account-level preferences remain follow-on capabilities.

Compute recipient eligibility from trusted membership, mention and thread-subscription records. Exclude the actor, inactive accounts and unavailable content. Apply preferences, quiet hours and suppression, then create durable delivery intent keyed by source event and recipient. Before every delayed send/retry, recheck current access, content version and preferences. Removing access or deleting content cancels undelivered disclosures.

The client groups live events in a short conversation burst before presentation. Suppress the banner for message IDs actually seen in a focused timeline, while allowing a subtle receive sound. Never infer visibility merely from the selected conversation. Read cursors are rechecked before backend delivery, and recent viewed-message evidence is shared across browser tabs. Native and browser presentation use one delivery branch, with transactional receipts and sound ownership to prevent duplicate interruptions.

Group bursts by conversation and recipient while preserving individual actionable targets in-app. Maya progress never notifies. A completed creation run and resulting assignment events use a shared correlation ID to present one coherent alert where audiences overlap. This is presentation/delivery aggregation, not deletion of audit events. Reconnection/replay rebuilds state without emitting old interruptions.

Reuse the comment adapter's deterministic reply publication and recipient delivery deduplication, extending them to chat message IDs and attention policy. A continuation reply belongs to the answering human's new invocation, so its result notification targets that person rather than inheriting the earlier requester's identity. Clarification/review messages are actionable ordinary replies; pending animation and background refresh produce no additional notification. Publish one durable message first, then recover its delivery independently without generating the answer or executing work again.

Quiet hours use the user's configured timezone and respect daylight-saving changes. At the end, send a bounded useful summary rather than one notification per accumulated message; recheck whether it was already read. Preserve existing weekly routine-email limits. Any separate missed-chat email cadence must be explicit, opt-in and deduplicated against the weekly summary. Do not echo native chat into Slack by default.

Preview preferences offer full content, sender/conversation only, and generic text. Delivery payloads contain typed navigation IDs and bounded safe text, never HTML, signed attachment URLs or unrestricted tool output. Opening fetches authorised current content. A lock-screen preview already delivered cannot be recalled reliably; restrictive preview modes must work before disclosure.

## 8. Delivery reliability and native integration

### Expo push

The inspected [Expo client](../../../apps/server/pkg/expopush/client.go) discards ticket IDs, and the [worker](../../../apps/server/internal/taskhandlers/notification_push.go) marks sent after initial acceptance. Extend the delivery model before treating push as a dependable chat attention channel.

Persist per-device attempts, payload version, provider ticket ID, retry classification and outcome. Poll receipts asynchronously, disable invalid devices, and retry failed items rather than an entire partly accepted batch. Expo distinguishes service acceptance, handoff receipts and actual device receipt; even a successful receipt does not prove the person received or read the message. Current documentation specifies receipt checks, 100 messages/request and 600 notifications/second/project. [Expo delivery guidance](https://docs.expo.dev/push-notifications/sending-notifications/)

A timeout after submission is uncertain, not definitely unsent. Preserve the attempt, apply bounded retries with grouping/dedupe, and tolerate possible duplicate OS delivery. Do not claim exactly-once push when the provider cannot prove it. Measure provider acceptance separately from application acknowledgement. Use finite TTL, appropriate priority, per-conversation grouping and bounded backlog recovery; grouping semantics differ between `collapseId`, Android `tag` and iOS `threadId`. [Expo payload fields](https://docs.expo.dev/push-notifications/sending-notifications/#message-request-format)

### Hosted desktop

The repository already contains native cache, drafts, offline navigation and an outbox. The [outbox operation enum](../../../apps/projects/src/shell/desktop/outbox-types.ts) supports story/comment/upload operations, and [cache policy](../../../apps/projects/src/shell/desktop/query-cache.ts) does not include chat. Add chat intentionally rather than assuming these generic-looking facilities already cover it.

Extend the web parser, JS bridge, Rust validation, native click destination and capability flag together. The existing [notification bridge](../../../apps/projects/src/shell/desktop/notification.ts) already accounts for older binaries rejecting unknown fields. Older installations need a safe supported destination or clear minimum-version gate; new payloads must not break their entire notification flow. Preserve origin, authenticated scope, receipt dedupe and logout checks in [native delivery](../../../apps/desktop/src-tauri/src/hosted/notifications.rs).

Distinguish live hosted notifications from closed-app/background delivery; the existing active WebView SSE does not imply either. Validate actual macOS and Windows binaries, permissions, badge behaviour, click activation and session changes. A browser fixture is not native notification proof.

### Mobile

The [native tab layout](<../../../apps/mobile/app/(tabs)/_layout.tsx>) already includes Home, My Work, Inbox, Search and Maya. Prototype Chat as a primary tab with Search as a header/command action and Maya reachable privately from Chat, preserving existing destinations until navigation approval. Do not add an unreviewed sixth tab.

Add typed conversation/thread routes for foreground, background and cold-start activation. Validate recipient and session before workspace switching; recheck after asynchronous navigation. Replace the current [unconditional foreground banner/sound](../../../apps/mobile/modules/notifications/hooks/use-push-notifications.ts) with active-view policy. Subscribe while active, recover from durable cursors on resume, and use push for background attention rather than assuming a permanent background socket.

Native timelines/composers need keyboard-safe layout, attachment picking, draft recovery and stable scroll anchors. Share data/reducer contracts, not browser components. Existing [session cache](../../../apps/mobile/lib/session-query-client.ts) isolates account/workspace and rejects offline mutations; [persistence](../../../apps/mobile/lib/query-provider.tsx) explicitly excludes mutations. Queued chat sends require new durable work.

## 9. Offline promise and local storage

The pilot may support cached reading and recovered drafts with **Reconnect to send**. Do not label a draft as queued or promise automatic sending before a persistent outbox passes its release gate. The complete offline-send experience remains roadmap scope.

A durable chat outbox must persist logical send identity, destination, versioned body, attachment dependencies, attempt state and acknowledgement mapping atomically. Preserve ordering within a destination, provide cancel/retry, recover ambiguous commits, and reauthorise before transmitting. Queued Maya requests activate only after server acceptance. Cross-account recovery is forbidden; logout and revocation cleanup must not race with late persistence writes.

Bound message/cache storage by size and age. Exclude temporary URLs and private data not permitted by workspace policy. Define encryption/device-lock expectations for native storage and the limitations of browser storage. Do not describe MMKV or SQLite usage alone as encrypted storage. A disconnected device cannot learn revocation instantly; purge on the next verified contact/session change, and document retained-cache policy honestly.

## 10. Accessibility and acceptance gates

Provide keyboard routes through navigation, message actions, mentions, threads, attachments and ticket controls. Restore focus on panel close, preserve focused rows during virtualisation, respect reduced motion and expose meaningful labels/timestamps. Test IME, right-to-left text, emoji, long code/URLs, high zoom, contrast and touch targets.

Use restrained polite announcements for new content and Maya completion. Do not announce every token or historical page as a new message. A chronological log and a virtualised feed have different semantics; choose and test their focus/loading behaviour with VoiceOver and NVDA. ARIA roles alone do not establish accessibility. [WAI log semantics](https://www.w3.org/TR/wai-aria-1.2/#log), [WAI feed pattern](https://www.w3.org/WAI/ARIA/apg/patterns/feed/)

Required acceptance cases include:

1. Lost acknowledgement, duplicate/reordered sync and app restart converge to one accepted message.
2. A fresh cursor cannot restore over an older cache; replay reset preserves pending sends and drafts.
3. History paging, image loading and Maya output preserve scroll/focus; new messages do not pull a reader away.
4. Thread reads remain independent; search jumps do not clear unseen gaps; a read acknowledgement cannot clear newer attention.
5. Account/workspace changes fence late fetches, uploads, outbox writes and native notification activation.
6. Team removal blocks history/search/files/previews, queued delivery and Maya actions; clients evict affected cached content when informed.
7. Two simultaneous ticket claims produce one owner, and all surfaces converge.
8. One event across web, phone, desktop and replay produces coherent attention and no application-generated duplicate sound.
9. Quiet hours, muting and read-during-grace suppress appropriately without permanently losing attention.
10. Partial push acceptance, uncertain submission, rejected receipt and provider outage recover without retrying successful batches wholesale.
11. Older desktop binaries and mobile cold launches reach a safe authorised destination.
12. Outbox-enabled builds survive offline restart and attachment dependency failures; draft-only builds make no queued-send claim.
13. Human chat remains usable with Maya unavailable; partial Maya creation shows actual receipts and retries only remaining actions.
14. The 1,000-user capacity profile includes large histories, a hot room, reconnect bursts and slow native devices, with bounded render/cache memory and measured sync traffic.
15. Active SSE projections render without a follow-up GET; inactive bursts coalesce without downloading bodies or advancing their durable event cursors. A higher-version projection above a gap remains visible while sync safely repairs the missing range.
16. A cross-author answer to a pending question/review uses the answerer's current permissions and a new linked invocation; unrelated or completed replies cannot reactivate Maya. Approval buttons reject the wrong requester, stale version or expired proposal.
17. Pending Maya replies never inflate counts; delayed Check again performs no mutation; successful publication replaces the placeholder once even when SSE and HTTP arrive in either order.

Launch requires focused automated tests plus representative real-browser, assistive-technology and native-device evidence. Report local validation separately from production canary results. Monitor send/reconcile latency, duplicate/missing messages, unread drift, pending-operation age, notification suppression and duplication, receipt failures and native activation errors without recording message content by default.
