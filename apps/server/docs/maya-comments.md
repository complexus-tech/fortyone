# Maya in story comments

People can select Maya in the existing mention picker and ask a question or request an update to the current story. Maya answers through ordinary comments, reply threads, and notifications. There are no visible queue states, progress cards, or special reply composers.

After the request is accepted, a temporary Maya reply shows “Thinking…” with small pulsing dots. It stays in the thread until her published reply arrives, without becoming a stored comment or increasing reply counts. Collapsed threads show the same feedback in their reply summary. After 30 seconds the label becomes “Still working on this…”; after three minutes it becomes static and offers “Check again”, which refreshes the request and comments without executing another action. Connection failures also stop the animation, and reduced-motion preferences keep the dots still.

## Conversation behavior

- A structured Maya mention on a new comment creates a durable invocation in the same database transaction as the comment. A display name, plain `@Maya` string, quoted mention, or human named Maya is not an agent target.
- Clear requests such as “close this task”, “assign this to Amara”, or “rewrite the description” can apply directly under the requesting member's current permissions.
- An ambiguous request produces a normal question. Any current member with access may answer that specific Maya comment without mentioning Maya again. The answering member owns and authorizes the new invocation.
- A requested suggestion produces a normal Maya reply containing an expandable review of the exact proposed changes. The original requester may use its approval action. A clear conversational answer such as “yes, please” or “go ahead” creates an exact replacement proposal under the answering member's permissions. It retains the original values, story-version precondition, and expiry.
- A revised answer such as “yes, but change the deadline too” is treated as a new instruction, never as approval of the previous proposal.
- Maya's replies use her existing system identity. Normal reply notifications go to the person she answered. Generated comments cannot recursively invoke the agent.

## Thread presentation and read state

Single-reply threads and the most recently active discussion open automatically. Explicit expand/collapse choices take precedence, and a newly active thread never closes an already open discussion or discards its draft. Collapsed threads show up to four actual participants, the reply count, and an unread dot. Expanded threads show the replies directly, with quiet Reply and Hide replies actions.

An add-reaction button follows Reply and appears on comment hover or keyboard focus; it remains reachable on touch screens. Only reactions people have chosen appear as persistent chips; selecting the same emoji aggregates its count, and selecting your existing reaction removes it. The picker contains 80 searchable choices grouped by expressions, feedback, celebration, and work. Mentions use the same foreground text, balanced padding, and subtle background in the editor and published content.

Maya's Markdown is converted to safe comment HTML before publication, including emphasis, lists, links, quotes, code, and tables. Legacy plain comments containing Markdown are formatted at display time. Existing rich text, typed mentions, and code are preserved; active HTML and unsafe links are not accepted as rendered content.

Read state is persisted per person and root thread. Only an expanded discussion visible in the focused viewport advances its reply cutoff; opening the story alone does not consume collapsed replies. The existing story cutoff applies to root comments. Server cutoffs are monotonic, capped at server time, and scoped to the canonical root in the current story. New arrivals after a saved cutoff remain unread.

## Supported actions and context

The native comment adapter restricts reads and writes to the owning story. Supported updates cover title, description, priority, status (including close/reopen), assignee, dates, time estimates, focus-block duration, and scheduling flags. Team statuses and active team members provide grounded names and IDs.

Sprint, objective, key-result, label, relationship, deletion, cross-story, and bulk actions are not exposed by this adapter. Existing Slack and other integrations retain their own capabilities. Attachment extraction and personal Maya chat history are not added to comment context.

Description replacements use plain text, escape generated markup, and preserve safe existing reference links and embedded media. Both the text and HTML fields update together. The model must supply the exact timestamp from its story read, and execution checks that timestamp again. Large descriptions that cannot be read completely are not rewritten.

## Reusable execution boundary

`internal/modules/agentinvocations` owns the surface-neutral lifecycle: source identity and revision, requesting user, conversation and entity bindings, durable proposals, fenced leases, approval, continuation, action receipts, publication, and recovery. It does not import the story or comment modules.

`internal/bootstrap/agentadapter` binds that lifecycle to comments. It validates mention identity, selects bounded visible context, checks workspace entitlement and live member/admin access, invokes existing messaging tools, and publishes comments through the normal comment service. `internal/bootstrap/worker/agent_invocations.go` composes the real services.

A future native-message adapter should:

1. Save the message and invocation together, with a distinct surface identifier and stable source ID.
2. Derive context and authority on the server from that message, its conversation, and its current audience.
3. Implement the runtime authorization, preparation, application, and publication ports; bind actions to the answering user, not to Maya's system privileges.
4. Publish ordinary messages through the message domain and deduplicate notification delivery.
5. Reuse the lifecycle, continuation, fencing, usage accounting, and receipt machinery. Extend source-scoped tools only when their target catalog and shared-audience policy are available.

This change does not create channels, native messages, or a second chat system.

## Reliability and authorization

- The queued database row is the durable dispatch record. Redis wake-up accelerates processing; a periodic recovery scan handles missed dispatches and expired leases.
- Comment creation IDs survive browser/offline retries. Replays retain the original source, context, and continuation binding and cannot execute twice.
- Model calls occur outside transactions. Action application and its receipt commit in one fenced database unit. Publication is separate and uses deterministic IDs, so a missing reply can be recovered without repeating the action.
- Proposal, successful-result, and failure replies have separate durable identities. Notification retries do not regenerate responses or repeat mutations.
- Authorization checks current workspace membership, role, team access, feature entitlement, source existence, and source revision before work and again within execution/publication transactions. A role downgrade blocks queued work.
- Story activity retains the acting human, before/after values, timestamp, and the reason “Updated with Maya from a comment.” The invocation receipt also links the source, executor, and changed fields.
- Provider errors and credentials are never copied into comments. Safe failures appear as ordinary Maya replies when source access still permits publication.

## Configuration and deployment

Apply migrations 221–226 before deploying the API and worker. These add durable invocations, native stored update proposals, native usage-event linkage, per-thread read cutoffs, and expanded reaction catalogs. Preserve populated receipts and repair schema forward.

`APP_MAYA_COMMENTS_ENABLED` controls availability and defaults to true. The existing OpenAI configuration, workspace Maya entitlement, messaging mutation key, rate limits, and `OPENAI_ASSISTANT_WORKSPACE_TOKENS_PER_DAY` remain required. Both API and worker must use compatible configuration. The worker must run for replies to be processed after comments are saved.

The current bounds are a two-minute lease, a ninety-second processing attempt, a sixty-second model preparation budget, three automatic attempts, a ten-minute proposal expiry, and bounded thread/description context. User controls use version checks; they do not replace current permission checks.

## Verification

The focused test suites cover structured mention parsing, source edits and access loss, live role downgrades, duplicate creation, queue outages, cross-author clarification and approval, proposal publication, stale description reads, rollback, duplicate execution receipts, prolonged delivery recovery, native usage deduplication, Markdown rendering, threaded read state, reactions, and editor draft preservation. PostgreSQL tests use isolated test databases.

Local app validation exercised the actual configured model and worker composition on PRO-98: closing the story, asking a clarification, applying a description from an ordinary untagged reply, reviewing an exact title proposal, and accepting it with “Yes, please.” The last step applied the stored proposal without another model call. Persisted replies, single notifications, human audit attribution, before/after values, and live title refresh were verified. Multi-author reactions aggregate correctly, and viewing replies clears their persisted unread marker. These checks are local validation, not deployment verification.
