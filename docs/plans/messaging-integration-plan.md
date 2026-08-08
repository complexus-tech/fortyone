# Go-Native Messaging Integration Plan

## Document Status

- Status: Active implementation plan
- Updated: 2026-08-08
- Launch provider: Slack
- Future providers: Microsoft Teams and WhatsApp
- Runtime: `apps/server` only
- Supersedes: the Chat SDK/`apps/bot` architecture and the runtime choices in
  `docs/plans/slack-integration-plan.md`

The removed Chat SDK plan remains useful only as product research. Do not
restore a second provider webhook runtime. The production operating contract is
[`docs/integration-runtime-contract.md`](../integration-runtime-contract.md),
and Slack deployment is governed by
[`docs/slack-production-runbook.md`](../slack-production-runbook.md).

## Outcome

FortyOne users can install Slack at workspace level, link their identity, turn
messages into stories or requests, use `/fortyone`, mention or DM Maya, and
receive correctly threaded responses. Team and dependent modal choices always
reflect the linked user's live FortyOne permissions.

The implementation establishes provider-neutral Go contracts so Teams and
WhatsApp can be added as adapters without duplicating installation,
authorization, inbox/outbox, assistant, or product logic.

## Product Scope

### Slack launch scope

1. Workspace OAuth installation and disconnection
2. Secure Slack-to-FortyOne account linking
3. **Create a FortyOne task** message shortcut
4. `/fortyone create [title]`
5. Dynamic create-task modal
6. Channel `@FortyOne` conversations with Maya
7. Direct-message conversations with Maya
8. Lifecycle handling for uninstall and token revocation
9. Request diagnostics, retry safety, and live admin controls

### Modal behavior

Team is the first field and is always selected. Its options contain only teams
the linked FortyOne user belongs to in the installed workspace. Changing team
immediately recalculates and rerenders:

- workflow status or request state;
- assignee;
- labels;
- objective;
- any later team-owned field.

The server clears stale dependent selections when they are not valid for the
new team. Submission rechecks membership and every selected identifier in a
single authorized domain operation; rendered Slack options are never treated
as authorization.

### Deliberate non-goals for the first release

- listening to every untagged channel or private-channel message;
- syncing complete Slack message history;
- mirroring every FortyOne notification into Slack by default;
- Microsoft Teams or WhatsApp launch in the same release;
- enabling Slack token rotation before refresh-token persistence is complete;
- reproducing Slack-specific modals on providers that do not support them.

## Architecture

```text
Slack HTTP ingress
  -> Slack request verifier and payload adapter
  -> durable messaging inbox / immediate acknowledgement
  -> messaging command processor
       -> installation + linked FortyOne actor
       -> normal workspace/team permission services
       -> story/request service or Maya assistant boundary
  -> durable messaging outbox
  -> Slack Web API adapter
```

### Go module boundaries

`apps/server/internal/modules/messaging` owns shared provider-neutral concepts:

- one-time OAuth/account-link nonces;
- inbound delivery identity and processing state;
- normalized conversations and message history;
- outbound delivery, idempotency, retry, and provider message mapping;
- normalized provider/actor/conversation/capability types.

`apps/server/internal/modules/slack` is the Slack adapter and Slack product
surface:

- Slack OAuth and encrypted installation credentials;
- Slack signature verification and payload parsing;
- events, commands, shortcuts, blocks, options, and modal rendering;
- Slack Web API calls and Slack error classification;
- OAuth-installer identity linking and single-use manual account binding;
- translation to/from messaging domain commands.

Existing FortyOne workspace, membership, story, request, workflow, label,
objective, and assistant services remain authoritative. Shared messaging code
must call them rather than recreating their rules.

### Future adapter contract

A Teams or WhatsApp adapter supplies:

- verified webhook ingress and normalized delivery IDs;
- install/account and user identity mapping;
- native message/card/form rendering;
- capability flags for threads, forms, ephemeral responses, streaming, and
  message updates;
- outbound delivery and standardized retry/error classification.

It does not add a new product database, permission implementation, or assistant
tool set.

## Canonical Slack Surface

The public API routes are:

- `GET /integrations/slack/setup`
- `POST /integrations/slack/events`
- `POST /integrations/slack/interactivity`
- `POST /integrations/slack/commands`

Workspace-authenticated management routes remain under
`/workspaces/{workspaceSlug}/integrations/slack`. Slack must never be configured
to call an `apps/bot` webhook or an `/internal/bot/slack/*` route.

The source-controlled configuration is
`apps/server/slack-app-manifest.yaml`. Its core scopes are
`app_mentions:read`, `chat:write`, `commands`, and `im:history`. Channel
inventory and user-directory/email scopes are outside the launch manifest.

## Security and Data Model

### Installation credentials

- Store Slack team/app/enterprise identifiers separately from credentials.
- Encrypt the complete credential payload with authenticated encryption and a
  versioned server-managed key.
- Keep an explicit revoked/disconnected timestamp.
- Persist the exact granted scopes returned by OAuth.
- Never log authorization codes, tokens, signing secrets, raw credential
  payloads, or connect nonces.
- Make encryption-key rotation observable and test decryption of previous key
  versions before removing them.

### One-time state and links

OAuth state and account-connect links use high-entropy random nonces. Persist
only a hash plus bound workspace/user/provider data, short expiration, and
consumed timestamp. Consumption is atomic. Reject expired, replayed,
cross-workspace, and wrong-provider values.

### Request verification

All Slack POST routes verify the exact raw body using the Slack signing secret,
`X-Slack-Request-Timestamp`, and `X-Slack-Signature`. The accepted replay window
is five minutes. Verification happens before JSON/form decoding, enqueueing, or
side effects and uses constant-time comparison.

### Data minimization

Store normalized event metadata and only the message context required for the
product workflow. Recoverable provider bodies are encrypted in the database;
queue payloads contain only tenant-scoped event identities. Inbox, outbox, and
assistant conversation content expire after 30 days. Redact credentials,
authorization headers, connect values, and raw bodies from logs and tracing.

## Reliability Contract

### Three-second acknowledgement

Events, slash commands, shortcuts, block actions, external options, and modal
submissions must return a valid response within Slack's three-second deadline.
The synchronous path is limited to verification, minimal parsing, durable
deduplication, and acknowledgement. Slow product work and AI execution happen
after durable acceptance.

Opening a modal consumes a short-lived `trigger_id`; schedule it immediately
and bound the Slack call tightly. Never hold the HTTP response open for an AI
answer or story workflow.

### Inbound idempotency

- Events API deliveries deduplicate by Slack event/delivery identity.
- Interactive deliveries without a stable Slack event ID derive a canonical
  idempotency key from installation, payload type, trigger/view/message
  identity, actor, and action.
- A duplicate already accepted returns `2xx` without repeating the product
  action.
- Processing state transitions support safe worker retry and crash recovery.
- Do not assume ordering between `tokens_revoked`, `app_uninstalled`, messages,
  and retries.

### Outbound idempotency and rate limits

- Every outbound intent has a stable product idempotency key.
- Workers claim deliveries with concurrency-safe state transitions.
- Retry transient network/`5xx` failures with bounded exponential backoff and
  jitter.
- On HTTP `429`, obey Slack's `Retry-After` header for that method/workspace.
- Parse HTTP status and Slack JSON `ok`; an HTTP `200` with `ok: false` is an
  error.
- Authentication/revocation errors disable the installation or require
  reauthorization; they do not retry forever.
- Persist Slack request IDs and provider message timestamps for correlation and
  later updates.

## Identity and Authorization Flow

1. Resolve Slack installation by the payload's Slack workspace/enterprise
   identity.
2. Resolve the Slack user link inside the mapped FortyOne workspace.
3. Link the OAuth installer from Slack's authenticated installer identity.
4. Otherwise send an opaque, single-use manual connect URL.
5. Complete connection in an authenticated FortyOne session after validating
   all bound values.
6. Execute commands and Maya tools as the linked FortyOne user.

The core integration does not request Slack user-directory or email scopes.
Every other user links through the authenticated FortyOne flow.

## Slack Workflows

### Install and reinstall

1. A FortyOne workspace admin starts the install from integration settings.
2. The API creates one-time OAuth state and redirects to Slack.
3. Slack returns to `/integrations/slack/setup`.
4. The API atomically consumes state, exchanges the code, verifies the Slack
   response, encrypts credentials, and persists granted scopes.
5. The OAuth installer is linked transactionally to the initiating active
   FortyOne workspace member.
6. Scope changes require reauthorization; tokens do not gain new scopes from a
   manifest edit alone.

### Message shortcut

1. User chooses **Create a FortyOne task** on a Slack message.
2. The API verifies and acknowledges the `message_action` payload.
3. Resolve installation/user; send connect flow if unlinked.
4. Build title/description/source context from the shortcut's message payload.
5. Open the modal with the first permitted team selected.
6. On team change, acknowledge and update the view using `view_id` plus `hash`.
7. On submit, reauthorize all fields and create one FortyOne story or request.
8. Reply with a link in the source conversation/thread and persist the source
   mapping.

The shortcut payload already carries the selected message, so this flow does
not justify channel-history scopes.

### Slash command

- `/fortyone create [title]` opens the same modal and authorization flow.
- Empty or unsupported input returns concise ephemeral help.
- Unlinked users receive the manual connect action.
- Slash responses are visible only to the actor unless the resulting product
  action intentionally posts a shared confirmation.

### Mention and direct message

1. Normalize `app_mention` or `message.im`, excluding bot-authored/self events.
2. Persist/deduplicate before acknowledgement.
3. Resolve the linked FortyOne actor and conversation/thread mapping.
4. Invoke Maya with a bounded context and the same tool permissions used by the
   web product.
5. Deliver through the outbox to the original thread/conversation.
6. Store the provider message identity and a minimal conversation record.

Channel mentions require the app to be present in the channel. DMs operate in
the app's Messages tab. Do not subscribe to broad `message.channels` or
`message.groups` events for the launch release.

### Uninstall and revocation

`app_uninstalled` and `tokens_revoked` are idempotent lifecycle commands. Mark
credentials unusable, stop channel sync and outbound delivery, clear caches,
and surface reinstallation in FortyOne settings. Because event ordering is not
guaranteed, Web API authorization failures must reach the same safe state.

## Delivery Plan

### Phase 0: Canonical cutover

- Remove tracked `apps/bot` source and direct Slack webhook references.
- Commit the Go endpoint manifest and production runbook.
- Use one Slack app per environment with the same route/callback contract.
- Keep token rotation disabled.

Exit: no deploy, script, or active plan depends on the Chat SDK runtime.

### Phase 1: Shared messaging foundation

- Add provider-neutral nonce, inbound event, conversation/message, and outbound
  delivery persistence.
- Add versioned authenticated credential encryption.
- Add adapter interfaces, capabilities, and typed error classification.
- Add background processing, recovery, bounded retries, and observability.

Exit: repository/service tests prove atomic nonce consumption, deduplication,
concurrent claiming, retry classification, and idempotent delivery.

### Phase 2: Slack security and lifecycle

- Complete least-privilege OAuth and granted-scope persistence.
- Verify all signed routes from raw bodies.
- Process URL verification, retries, uninstall, and token-revocation events.
- Keep channel inventory and verified-email linking outside core launch scope.

Exit: invalid signatures/state/replays fail safely; uninstall and revoked tokens
disable the installation; core install works when all optional scopes are
denied.

### Phase 3: Task creation parity

- Wire message shortcut and slash command to one create-task workflow.
- Limit teams to linked-user memberships.
- Update dependent options live when team changes.
- Revalidate selections and create idempotently.
- Post linked confirmation and preserve source message/thread metadata.

Exit: the live matrix passes with multiple teams/workflows and delivery replays
create exactly one product record.

### Phase 4: Maya in Slack

- Normalize mention and DM conversations.
- Reuse Maya authorization/tool policies through an internal assistant boundary.
- Add bounded context, cancellation/timeouts, safe output rendering, and
  durable/threaded delivery.
- Add abuse/rate controls and content-retention policy.

Exit: a linked user gets a permission-correct answer in channels and DMs;
unlinked and unauthorized users get safe connect/denial responses; retries do
not duplicate answers.

### Phase 5: Production hardening

- Add admin health/status, install/reinstall state, backlog metrics, dead-letter
  recovery, and audit views.
- Load-test three-second acknowledgements and worker concurrency.
- Exercise `429`, `5xx`, timeout, revoked-token, and stale-view races.
- Complete privacy/support/distribution requirements and Slack review if
  required for commercial rollout.
- Roll out to internal, design-partner, then broader workspaces with kill
  switches and measurable gates.

Exit: on-call can identify and contain an incident without direct database
edits, and the production runbook has been executed end to end.

### Phase 6: Second provider proof

Implement a narrow Teams adapter after Slack is stable. Reuse the shared
installation, identity, inbox/outbox, assistant, and product command contracts;
only Teams verification, payload translation, capability selection, and
rendering are new.

Exit: the second provider ships without importing Slack types into shared domain
code or copying FortyOne authorization rules.

## Test Strategy

### Automated

- request signature fixtures and replay-window boundaries;
- OAuth/account-link state expiry, mismatch, single use, and concurrency;
- credential encryption round-trip, tamper failure, and key-version migration;
- duplicate Events API/interactive deliveries and worker crash recovery;
- core installation without channel inventory or user-directory scopes;
- team membership filtering and cross-team stale option rejection;
- `views.update` hash/stale-view handling;
- Web API HTTP/JSON errors, `Retry-After`, revocation, and log redaction;
- assistant actor/workspace isolation and outbound idempotency;
- adapter contract tests independent of Slack payload types.

### Live Slack acceptance

Follow `docs/slack-production-runbook.md`. A release is incomplete until a real
Slack development workspace has verified installation, shortcut, command,
dynamic modal, story creation, mention, DM, unlinked-user, least-privilege,
replay, rate-limit, uninstall, and reinstall flows.

## Launch Gates

- p95 acknowledgement is comfortably below three seconds and no synchronous AI
  call exists on ingress;
- duplicate product actions and duplicate outbound messages are zero in replay
  tests;
- cross-workspace/team authorization tests pass;
- credentials and signed links are absent from logs and traces;
- core functionality passes with optional scopes denied;
- lifecycle events and revoked-token errors converge on a disabled install;
- backlog age, retry counts, terminal failures, and provider rate limits are
  observable and alertable;
- support, privacy, data-retention, distribution, and Marketplace obligations
  are approved for the intended rollout.

## Completion Definition

The Slack integration is complete only when the Go API is the single canonical
ingress, installs are secure and workspace-scoped, users can create work from a
message or command with permission-correct dynamic options, Maya responds to
mentions and DMs, every slow/retryable action is durable and idempotent, and a
production operator can install, verify, revoke, recover, and audit the system
using the committed manifest and runbook.
