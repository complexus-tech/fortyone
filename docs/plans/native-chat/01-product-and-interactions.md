# Product and interaction specification

## 1. Product promise

Chat should make a discussion easy to continue and easy to turn into accountable work. It must preserve who said what, what was decided, which work resulted, and what remains unresolved. Maya should help people complete that loop with fewer manual steps.

Use the selected global Chat layout. Preserve the actual application shell, theme, density, typography, navigation primitives, and ticket detail experience. Generated avatars, badge labels, placeholder teams, counts, dates, and text are illustrative. Use existing production components and real server-derived identity when implementing.

## 2. Navigation and information architecture

### Global navigation

Add Chat alongside My work in the app navigation, with the existing active-state styling. The navigation badge represents unresolved attention, such as direct mentions, DMs, followed-thread replies, and a requested Maya result. It is not the count of every message sent across every team.

The Chat column contains All unread, Threads, Teams, and Direct messages. Preserve ordered, stable navigation as unread activity changes; do not continually reorder the selected conversation beneath the pointer. A user may later pin a conversation. Search belongs in the Chat header and integrates with global search only when the same permission and deep-link contracts are implemented.

For the pilot, each eligible team has one canonical conversation. Team names, icons, archive state, and access come from the team domain. Renaming a team does not change its conversation ID or invalidate existing links. Joining or leaving a team updates the conversation list automatically.

### Route state

Use stable conversation IDs in routes, with separate thread and target-message parameters. Browser refresh, copied links, native notification clicks, back/forward, and workspace switching must restore the same context. A target message may require loading the page around it rather than scrolling from the newest page through all history.

Use one contextual side panel at a time: thread or ticket details. Opening a ticket from a thread preserves the thread location in navigation state; Back restores it. Each conversation and thread keeps its own draft and scroll anchor. Mobile pushes these surfaces onto a navigation stack rather than compressing three columns into a narrow screen.

## 3. Participants and access

Team conversation access follows current team membership and existing workspace eligibility. Public team visibility alone must not automatically grant the right to read chat if the team domain currently requires joining. Specify and test this boundary against actual team policies before launch.

The recommended team-history default is that a new active team member can read retained team history. Removed members lose server access immediately. Team archive makes the conversation read-only; restoring the team restores the same conversation. Team deletion enters the existing deletion lifecycle and invalidates pending runs and deliveries.

Maya appears as another avatar in the member stack, using the current system profile. Her accessible name is Maya. Her profile can explain available help using normal product language. Do not show a human presence dot, last-seen time, “read by Maya,” paid-seat count, or repeated AI badge. Do not insert fake human team membership solely to make the avatar render.

One-to-one human DMs have exactly two human participants in one workspace. Maya is not silently present and does not read them by default. For an initial private Maya entry, expose the existing personal Maya experience through an adapter or explicit navigation, without copying private transcript history into team chat. Group DMs need a separate participant-change/history decision; creating a new conversation when membership changes is the safest initial design.

## 4. Message behaviour

### Compose and send

- Rich text supports paragraphs, links, emphasis, inline code, code blocks, lists, mentions, and attachments. Set explicit byte, character, nesting, and attachment limits.
- Mentions store stable member IDs. Plain text that happens to spell `@Maya`, a quoted mention, or text inside a code block must not invoke the agent.
- Enter sends and Shift+Enter adds a line break, with an accessible preference if the existing composer supports it. IME composition must never accidentally send.
- Preserve unsent drafts when navigating. On a failed send, keep the exact content and attachments with Retry and Edit actions.
- Optimistically render a pending message with a stable client-generated ID. Reconcile HTTP acknowledgement and realtime arrival into that same row.
- Pending, accepted, failed, and retrying are distinct. “Sent” means server-committed; it does not mean every device received it.
- Pasting images creates real upload items with progress, cancel, failure, and scan/processing states. Do not embed raw base64 images in message content.

### Realtime delivery

The selected transport is HTTP for commands and the existing shared SSE connection for incoming updates. A person viewing a conversation receives the authorised, versioned message content immediately, without waiting for a periodic refresh or an HTTP fetch per new message. Inactive conversations receive grouped activity/unread hints; opening one loads its canonical history. Typing signals are throttled and temporary. They must not compete with durable message delivery.

On reconnect, the client restores missing events through durable cursor sync and deduplicates live arrivals, acknowledgements and replay. Receiving a later message may update the visible timeline, but cannot mark an unseen gap as synchronised or read. A quiet connection-status indicator explains recovery without losing drafts or moving a person away from history. Mobile background updates use notifications; returning to the app resynchronises before claiming the conversation is current.

### Threads

Use one reply level with a stable root message. Replies can reference another reply, but remain in the same thread. The root shows a reply count, recent participant avatars, and unread reply state. Opening the thread fetches its own cursor-paginated feed. Root and thread composers have separate drafts and mention scopes.

Maya invoked in a thread replies in that thread. A channel-level invocation creates a reply thread under the request; the channel row shows a compact result preview and current status. This preserves the screenshots' discoverability without filling the main channel with long agent transcripts. A short creation receipt can be visible on the root; detailed explanations remain in the thread.

### Edit and delete

Authors may edit their own messages using optimistic concurrency. An edit increments a revision, preserves a visible edited marker, updates search, and emits a durable event. If another revision wins, show the current content and preserve the user's attempted edit for reconciliation.

Deleting a root leaves a tombstone while retained replies exist. Deleting a reply does not destroy siblings. Deletion removes content from normal reads, search, previews, attachment access where unreferenced, future AI context, and pending notification excerpts. Audit metadata preserves actor, operation, and time under the retention policy; it is not permission to keep deleted plaintext forever.

Edits do not automatically re-run Maya. Adding a Maya mention through an edit offers an explicit Ask Maya action. Removing/deleting an unstarted triggering message cancels its queued run. For a run already in progress, see the cancellation and revision rules in the AI plan; completed tickets are never silently deleted because the source message was deleted.

### Reactions, pins, and saved state

Reactions are idempotent add/remove actions keyed by message, user, and canonical emoji identifier. Treat server counts as authoritative. A reaction creates no default push or email. Pins are shared conversation state with a permission rule and audit entry. Saved messages are private user state and must not become shared bookmarks accidentally.

## 5. Maya behaviours

### Natural discovery

Typing `@` opens the existing mention pattern, containing Maya and authorised human members. Maya's row uses her existing avatar and name. A short capability description in the picker or profile is sufficient. There is no persistent “Maya is in every team” banner.

First use can offer a small number of relevant suggestions in an empty composer or help surface: “Summarise this thread,” “What did we decide?” and “Create tasks from these messages.” These do not appear as extra cards on every conversation.

### Natural follow-up and review

Reuse the conversational behaviour already present in Maya comments. A person can reply directly to a specific pending Maya clarification or review without another mention. The server resolves that exact pending interaction; ordinary conversation and replies to completed Maya messages do not silently start new agent work. Agent-authored messages never invoke Maya recursively.

The answering human owns the new invocation and supplies its current permissions. Another authorised teammate can answer a clarification or say “Yes, go ahead” to an eligible review, but does not inherit the original requester's authority. Exact assent creates a replacement proposal for that person, preserving the reviewed values, resource versions and original expiry, and revalidating access, destination audience and source freshness. “Yes, but change the deadline” is a new instruction that must be evaluated again. Keep the resulting human/Maya audit attribution visible.

The explicit approval button remains restricted to the original requester and the exact proposal version. Reviews expire after ten minutes, matching the shared comments engine; expiration requires a fresh review instead of silently extending permission. This is separate from the 24-hour lifetime of the server-issued pre-visit catch-up checkpoint.

### Questions and summaries

Useful answers distinguish decisions, proposed ideas, owners, due dates, open questions, and disagreements. Never turn an uncertain suggestion into a firm commitment. A summary should say what it covers and link to specific messages or work items. “View 10 source messages” opens the exact selection's source identities under current access checks. Edited sources show an indicator; deleted sources show tombstones. The view cannot reveal forbidden old plaintext from retained AI evidence.

“Summarise since I was last here” uses the requester's root and relevant thread read boundaries captured before the current visit advances them, not a guessed timestamp. State whether unopened threads are included. “Summarise this thread” excludes unrelated channel discussions. “Last 10 messages” is defined explicitly in the AI document and excludes operational status updates.

Maya may answer using current team tickets when asked about progress, but must distinguish the state of work now from what people said at the time. A linked source and freshness timestamp prevent an old chat claim from being presented as current ticket status.

### Conversation to work

For a clear creation request, Maya identifies discrete outcomes, avoids duplicates, uses valid team workflow defaults, preserves unresolved questions, and creates bounded work items under the requester’s authority. Do not invent owners, dates, estimates, or certainty.

Inline previews contain the real identifier, title, current status, relevant assignment, and at most a small amount of supporting information. They are typed resource references with authorised hydration, not model-generated HTML. Click opens the existing detail surface. Inaccessible/deleted work renders an appropriate neutral placeholder without leaking its title.

“Assign to me” is a normal conditional story mutation. It only succeeds if the item is still unassigned and the caller can take it. Two simultaneous clicks cannot overwrite the winner. The other person sees the current owner and can use the normal explicit reassignment flow if permitted.

### Progress and recovery

Show a brief status within 500 ms of accepted invocation when conditions permit: Reading conversation, Checking related work, Creating 3 tickets. These labels describe observable work, not hidden model reasoning. Replace the status with the result rather than appending multiple status messages.

Reuse the existing comment pending-state distinction: progress is not a stored human-visible message that increments reply counts, refreshing status does not restart a run, and delayed work has a truthful recovery state. Live chat progress and safe answer blocks arrive over SSE; the existing comments polling loop is not the primary chat delivery mechanism. Answer streaming is a new chat capability and releases only content that has passed the destination's access checks.

For an uncertain result, show “Checking whether the ticket was created” while recovering the receipt. Do not show a generic Retry that could repeat the action. If only some items completed, show each outcome and offer Retry remaining with stable item IDs.

Closing the browser does not cancel a run. Stop is an explicit action. After cancellation, report which actions had already committed. A reversible follow-up may be offered, but it must obey the ordinary domain rules and must not erase edits another teammate made afterward.

## 6. The experiences that should impress users

| Experience                  | Observable behaviour                                                                                                            | Required foundation                                                                      |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Catch up quickly            | A concise, source-linked summary separates decisions from unresolved questions and starts from the user's actual read boundary. | Accurate read cursors, deterministic context, source citations, summary evaluations.     |
| Preserve intent             | Tickets contain useful acceptance criteria derived from the discussion, original constraints, and open questions.               | Typed extraction, evidence links, domain validation, action-quality review.              |
| Avoid repeated work         | Maya recognises that a discussion already produced a ticket and offers to open or update it.                                    | Conversation-resource links, scoped search, duplicate detection, explicit update policy. |
| Keep context visible        | Opening or claiming a ticket preserves the chat and source messages.                                                            | Shared routes, live previews, conditional assignment, side-panel history.                |
| Continue across devices     | A message or Maya run started on desktop appears correctly on mobile after reconnect.                                           | Durable state, replay, per-user cursors, client outbox, run snapshots.                   |
| Stay quiet when appropriate | One useful result and one relevant notification replace a cascade of bot posts and alerts.                                      | Run-result coalescing and unified attention policy.                                      |

Later intelligence can include decision history, recurring opt-in team briefings, and suggestions that work is already complete or blocked. These should follow the same source and permission contracts. Do not make unsolicited monitoring, automatic external messages, or hidden personal-memory ingestion prerequisites for the core experience.

## 7. Empty, loading, failure, and boundary states

Design all of these before implementation acceptance:

| State                                | Expected experience                                                                                                                                                                                    |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| No team conversations                | Explain eligibility and offer existing join-team navigation if allowed.                                                                                                                                |
| New conversation                     | Quiet introduction, member avatars, composer, optional contextual Maya prompts.                                                                                                                        |
| Initial load                         | Stable layout skeleton; avoid fabricated message rows or changing scroll anchors.                                                                                                                      |
| Loading older messages               | Preserve the first visible message and offset while prepending.                                                                                                                                        |
| User reading history                 | New-message indicator; do not force-scroll to the bottom.                                                                                                                                              |
| Reconnecting                         | Small connection status; retain draft and accepted history.                                                                                                                                            |
| Offline send                         | The internal pilot can preserve a draft with “Reconnect to send.” The completed experience uses a durable local outbox with explicit queued state and transmits when an authorised connection returns. |
| Permission removed                   | Stop subscriptions, clear cached previews/drafts as policy requires, explain access loss without leaking content.                                                                                      |
| Archived team                        | History readable under policy, composer disabled, no new Maya actions.                                                                                                                                 |
| Attachment scanning                  | File remains visible as processing; no download or AI extraction until clean.                                                                                                                          |
| Maya unavailable or budget exhausted | Human chat continues; explain Maya's availability and preserve the request.                                                                                                                            |
| Maya ambiguous request               | One concise inline question with actionable choices.                                                                                                                                                   |
| Model failure                        | Honest failure state, preserved context selection, safe retry based on run/action receipts.                                                                                                            |
| Notification target deleted          | Open the conversation with a removed-message marker, never a generic broken page.                                                                                                                      |
| Search has no result                 | Clear scoped empty state; never imply inaccessible content exists.                                                                                                                                     |

## 8. Accessibility and input quality

Provide a documented keyboard path through conversation navigation, messages, thread controls, composer, and details. Focus returns to the invoking message when a panel closes. Announce new messages only when appropriate; do not read every streaming token aloud. Use a single polite completion announcement for Maya and concise status updates.

Virtualisation must preserve meaningful DOM order, focus, and screen-reader navigation. Validate the feed pattern before adopting it; an accessible paginated mode is preferable to a nominal ARIA role with broken focus. Test actual VoiceOver and NVDA interactions, reduced motion, high zoom, contrast, keyboard-only use, IME composition, touch targets, and long/localised text.

## 9. Product acceptance scenarios

1. Two team members exchange messages and replies across web and mobile; arrival order, revisions, and unread state agree after a reconnect.
2. A third member opens the channel without opening a thread; thread replies remain unread.
3. Joseph mentions Maya, selects the last 10 messages, and receives a summary whose sources match that exact selection.
4. An explicit request creates three tickets; retrying the HTTP request and restarting the worker create no duplicates.
5. Two people claim one unassigned ticket; one succeeds and both see the same owner.
6. A user with access to a private document asks about it in a team chat that lacks access; no private title, snippet, attachment, citation, or notification leaks.
7. Removing the requester while Maya runs prevents further actions and unsafe shared output.
8. A desktop recipient reading the exact thread receives no redundant alert; a mobile recipient sees one actionable notification and lands at the correct message.
9. Once the durable outbox gate is enabled, a pending message survives app restart, then sends once after reconnect; switching accounts cannot expose the draft. The earlier draft-only pilot must not display a queued-send promise.
10. Disabling Maya leaves human chat, search, attachments, and history functioning.
11. An active conversation renders incoming messages directly from SSE; an inactive one updates attention without receiving every full message body. Reconnect repairs a deliberately dropped event without duplication or an unread/read gap.
12. A teammate answers a specific pending Maya question without another mention. The resulting work uses that teammate's authority and attribution. An unrelated reply does not invoke Maya.
13. Another teammate's exact conversational assent preserves the reviewed values and original expiry under a new authorised invocation; changing a field requires a new instruction. They cannot use the original requester's approval button, and an expired or stale proposal does not execute.
