# Native team chat and Maya implementation plan

Status: Updated architecture and delivery plan. **SSE plus HTTP is the selected transport.** Remaining proposed policies and release gates are identified below. No feature implementation is authorised by this document.

Research and revision date: 16 October 2026. Initial capacity target: **1,000 concurrent users**, confirmed by the product owner. This revision refreshes the original research against repository reference `0ee7c80448b722eeca817e60b76c630cfdf957f1`, including the committed Maya comments implementation. Repository inspection establishes implementation context, not production deployment or load capacity.

## Intended outcome

FortyOne gains a dependable place for team conversations, with Chat in the main navigation. People discuss work, mention Maya, get grounded answers, turn a bounded conversation into tickets, claim ownership, and open the resulting work without losing the conversation. The quality bar is that people can trust both the conversation history and the work Maya performs.

The selected visual direction has the existing FortyOne outer navigation and command bar, a Chat navigation column, team conversations, a message stream, and an optional thread or ticket detail panel. The screenshots are concept references, not specifications for every field or state.

The following product corrections supersede the generated screenshots:

- Use Maya's existing persisted identity, name, avatar URL, and existing avatar component.
- Display **Maya**, without an appended AI badge or repeated AI label.
- Include Maya naturally in the team member avatar stack and mention picker.
- Remove “Maya is in every team” and similar explanatory chrome.
- Preserve the distinction between a system actor and a human internally. Maya has no human seat, online status, read receipt, or ability to log in.
- Use the workspace's configured work-item terminology in the UI. Existing backend entities remain stories; the current workspace displays tickets.

## Reading order

| Document                                                                  | Purpose                                                                                                                                     |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| [Product and interaction specification](01-product-and-interactions.md)   | Scope, exact behaviours, screenshots-to-product decisions, conversation rules, and the moments that should feel exceptional.                |
| [Backend data and realtime architecture](02-backend-data-and-realtime.md) | Module boundaries, schema, APIs, transactions, ordering, replay, authorisation, files, search, and retention.                               |
| [Maya orchestration and intelligence](03-maya-orchestration.md)           | Identity, context selection, shared-audience safety, tools, execution receipts, retries, streaming, models, cost controls, and evaluations. |
| [Clients and notifications](04-clients-and-notifications.md)              | Web, desktop, mobile, composer/feed/cache contracts, unread state, push, email, badges, and accessibility.                                  |
| [Delivery quality and operations](05-delivery-quality-and-operations.md)  | Dependency-ordered work packages, capacity assumptions, launch gates, testing, observability, rollout, recovery, and ownership.             |
| [Research and decision register](06-research-and-decisions.md)            | Current repository evidence, external primary sources, rejected alternatives, assumptions, and decisions to settle before implementation.   |

## Architecture direction

Keep the modular Go monolith. Add a first-class `chat` domain for native multi-user communication, use PostgreSQL as the authority, and use a transactional outbox for asynchronous effects. Extend the existing surface-neutral `agentinvocations` engine used by Maya in comments, with a trusted native-chat adapter and scoped messaging capabilities. Its durable background jobs continue when a browser closes. Chat adds its own context, audience and work-creation contracts; it does not introduce a second run lifecycle or approval store.

Use authenticated HTTP for sends, edits, reactions, read acknowledgements and other commands. Extend the existing shared **server-sent events (SSE)** connection for immediate server updates: actively viewed conversations receive full, authorised, versioned message projections; inactive conversations receive lightweight, coalesced activity and attention hints. This avoids a separate fetch for every visible message while limiting background fan-out. Reuse the connection across features rather than opening one stream per conversation. Mobile background delivery uses push and resynchronises on resume.

Durable HTTP sync recovers gaps and reconnects from persisted conversation events and explicit cursors. A live projection may render immediately, but must not advance a contiguous sync cursor past missing events. Redis Pub/Sub accelerates dispatch; it is not the replay store. Do not treat an open connection, a Redis publish, or a push-provider acceptance as proof that a user received or read a message.

SSE is selected because it fits the interaction and reuses infrastructure the team already understands. WebSockets are not inherently more expensive, and reaching 5,000, 10,000 or 50,000 concurrent users does not automatically require switching transports. Reconsider them only for a measured bidirectional, latency or client constraint. Connection count, message fan-out, database work, media and Maya usage determine capacity and cost more directly than the transport name.

The existing realtime feed, personal AI transcript storage, and document CRDT service each solve different problems. They provide reusable pieces, but none is the complete storage or delivery contract for native chat.

```mermaid
flowchart LR
  C[Web desktop and mobile] -->|Authenticated commands| API[Go chat application service]
  API -->|Atomic commit| DB[(PostgreSQL chat state events and outbox)]
  DB --> D[Outbox dispatcher]
  D --> R[Redis dispatch signals]
  R --> G[Authenticated SSE gateway]
  G -->|Active message updates and inactive hints| C
  C -->|Cursor replay and canonical reads| API
  D --> N[Notification projections and delivery]
  D --> A[Shared agentinvocations worker]
  A --> M[Native chat adapter and scoped capabilities]
  M --> S[Existing story and document services]
  A -->|Persist response and receipts| API
  A --> P[Model provider]
```

## The complete Maya experience

1. Joseph posts “@Maya create tasks from the last 10 messages.” The structured mention identifies the existing Maya principal.
2. The message, its trigger, and the shared engine's durable invocation commit together. Joseph sees his accepted message and a brief Maya status.
3. The server freezes the exact preceding message IDs and revisions in the current channel or thread. “Last 10” does not drift as new messages arrive.
4. The context builder excludes unavailable, deleted, private, or audience-incompatible material. Maya gets the current team's permitted workflow metadata and a small, task-specific tool set.
5. Maya identifies useful work, checks for previously created or matching tickets, and returns a typed action plan. The server validates scope, fields, limits, and the requester's current authority.
6. A clear, bounded, same-team creation request can execute under the proposed native-chat action policy. Ambiguous instructions produce a targeted question; suggestions and broader changes use an explicit preview.
7. Each committed action has a durable receipt. Worker retries recover those receipts instead of creating the same tickets again.
8. Maya's answer displays actual ticket IDs and live, permission-checked previews. Users can assign an unassigned ticket to themselves through a conditional mutation, or open its details beside the chat.
9. The ticket preserves Joseph as the authorising human and records Maya's role separately, with links to the source conversation. Notification aggregation presents one coherent result instead of a burst of overlapping alerts.

The existing comments implementation already supplies durable invocation, fenced execution, approval, conversational continuation and recovery primitives. Chat extends that foundation with frozen message selections, shared-audience checks, citations, typed ticket cards and safe streaming. Independently committed batch items need an explicit extension to the shared execution/checkpoint contract; they are not an existing capability to assume.

People can also reply naturally to a specific pending Maya clarification or review without mentioning her again. A different person's reply becomes a new invocation under that person's current authority. Exact conversational approval preserves the reviewed values, resource versions and expiry; a changed instruction starts a new request. The approval button remains available only to the original requester. Reviews use the existing ten-minute expiry; the separate pre-visit catch-up checkpoint lasts 24 hours.

The detailed policy, including cancellation and partial failure, is in [Maya orchestration](03-maya-orchestration.md). The bounded creation fast path is an explicit native-chat extension; it must not silently alter comments, personal Maya or Slack behaviour.

## Scope and delivery sequence

The destination includes team conversations, threads, reactions, attachments, mentions, search, saved/read state, direct messages, useful notifications, and Maya. Deliver these as testable vertical slices.

| Stage                                       | Usable result                                                                                                   | Entry to next stage                                                              |
| ------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Foundation                                  | Stable contracts, authorisation model, identity, event ledger, replay, and measured transport prototype.        | Concurrency and failure tests demonstrate durable ordering and tenant isolation. |
| Human team chat                             | Internal teams can exchange messages, reply, react, attach files, reconnect, search, and manage attention.      | Cross-device consistency and notification policy pass the pilot matrix.          |
| Maya read capabilities                      | Adapt the existing engine for grounded questions, thread summaries, catch-up and natural clarification replies. | Source and shared-audience evaluations meet launch thresholds.                   |
| Maya work creation                          | Explicit conversation-to-ticket flow, real previews, conditional assignment, audit, and recovery.               | Crash/retry, ambiguity, authorisation, and action-quality gates pass.            |
| Private communication and client completion | One-to-one human DMs, a private Maya entry point, mobile/desktop parity, notification tuning.                   | Privacy and lifecycle matrix passes across all clients.                          |
| Controlled general availability             | Tested 1,000-user operating target, runbooks, budgets, retention, support, and staged rollout.                  | Production canary evidence meets the stated release gates.                       |

Human DMs are part of the overall plan because they appear in the selected direction. They are a separate release gate, not a reason to delay the first team pilot. Group DMs, arbitrary private channels, voice/video, federation, full Slack history import, and automatic monitoring of every conversation remain later work.

The 1,000-user target defines the first load-test and operating envelope, not a product limit or a claim of measured capacity. [Delivery and operations](05-delivery-quality-and-operations.md) includes explicit 5,000-, 10,000- and 50,000-user checkpoints, connection assumptions, database/worker limits and hosting cost drivers. Growth requires measured scaling and budget review; a monthly hosting bill cannot be inferred from peak concurrent users alone.

## Non-negotiable contracts

- A successful send means the canonical message and required durable effects are committed.
- A retry with the same logical send identity cannot create another message.
- A retry or worker crash cannot repeat a committed Maya action.
- Every shared answer is safe for its destination audience, not merely accessible to its requester.
- Read state follows what a person actually viewed; receiving a network event or opening a notification is insufficient.
- A channel read does not mark unseen thread replies as read.
- File access, search, previews, notifications, and replay use the same current membership boundaries as messages.
- Removing a member stops new access and pending deliveries. Cached content is evicted on the clients we control; previously viewed information cannot be retracted from a person.
- Human messaging remains usable when Maya or its model provider is unavailable.
- No paid infrastructure or production capacity is assumed to exist merely because the repository contains an implementation.

## What must be decided before coding

Adopt the proposed native-chat action policy and history/retention defaults; confirm team-chat availability for guests and machine credentials; choose the attachment scanning service; validate the selected SSE path on the actual hosting and native clients; and assign the work-package owners. The SSE transport decision is settled for this plan; the spike qualifies its implementation and capacity. The [decision register](06-research-and-decisions.md) supplies a recommended default and consequence for the remaining choices.

## Completion definition

The feature is complete when the behaviour, correctness, AI evaluations, accessibility, client support, migration, operations, and production canary gates all pass. A working screenshot, a successful model demo, passing unit tests, or a deployed API alone does not meet this definition.
