# Slack Integration Plan

## Document Status

- Status: Updated dual-path implementation plan
- Scope: Slack integration for the `fortyone` monorepo
- Primary app surfaces:
  - `apps/projects` for workspace and team settings UI
  - `apps/server` for integration state, routing, permissions, and domain events
  - optional TypeScript Slack runtime in `apps/projects` API routes or a dedicated app when Chat SDK is used
- Goal: Production-grade Slack integration that works with either:
  - a native Slack implementation in the Go backend
  - a hybrid Slack runtime powered by Chat SDK while FortyOne remains the source of truth

## Problem Statement

The product needs a Slack integration that does more than send generic notifications. Teams should be able to connect a Slack workspace, route channels to teams, receive high-signal updates, create and manage work from Slack, and optionally use Maya inside Slack for assistant-style workflows.

This should fit the existing workspace, team, story, comment, notification, and integration architecture already present in the repo. Slack should behave like a first-class product integration, not a disconnected bot.

## Product Scope

Slack is not GitHub. It should not try to be a full mirror of the product. The integration should focus on five strong workflows:

1. Notifications
2. Story creation and lookup
3. Link unfurling
4. Interactive actions from Slack
5. Optional assistant-style Slack conversations after identity linking exists

Optional deeper sync such as Slack thread mirroring can come after those foundations are stable.

## Current Repository Reality

The best integration hooks already exist in the codebase:

- Workspace settings already exposes an integrations area
- Team settings already has a natural place for team-level routing and automation rules
- Story creation, updates, comments, and notifications already move through backend services and event publishing
- Redis-backed async processing already exists for background work
- Maya already exists inside the product as an AI assistant with tool-based actions

This means Slack should be modeled as a proper integration module with a clear runtime boundary, not as raw incoming webhooks glued onto the frontend.

## Guiding Principles

1. FortyOne owns product state, permissions, team routing, and auditability in all approaches.
2. Slack delivery mechanics may be implemented natively or delegated to Chat SDK, but the product contract must stay the same.
3. Keep Slack connection state at the workspace level.
4. Keep channel routing and notification rules at the team level.
5. Avoid noisy default behavior. Slack should be configurable and high signal.
6. Use asynchronous delivery and retry logic for outbound messages.
7. Verify every inbound request using Slack signing secrets when operating natively, or the equivalent validated runtime contract when using Chat SDK.
8. Prefer interactive modals and actions over fragile slash-command argument parsing.
9. Do not make the first release depend on broad conversational sync.
10. Treat Maya in Slack as an additive surface, not a replacement for the main product UI.

## Architecture Strategy

### Core Decision

The integration plan should separate:

- the FortyOne integration core
- the Slack runtime implementation

That allows the same product model to work with both approaches.

### Always Owned by FortyOne

These concerns remain in FortyOne regardless of runtime choice:

- workspace-to-Slack installation ownership
- Slack channel inventory used by product settings
- team-to-channel assignments
- team-level notification and action policies
- Slack user to FortyOne member linking
- delivery logs and audit trails
- domain event translation into Slack-worthy messages
- story and comment permissions
- UI flows in workspace and team settings

### Slack Runtime Boundary

Define a runtime interface for the Slack-specific mechanics:

- start install
- complete install
- sync channels
- post message
- update message
- open DM
- open modal
- handle slash commands
- handle shortcuts
- handle interactivity payloads
- publish unfurls
- optionally run assistant conversations

FortyOne should depend on this interface, not directly on a concrete Slack implementation.

### Approach A: Native Slack Runtime in Go

Use this when:

- minimizing stack split matters
- the team wants all integration behavior in `apps/server`
- unfurls and custom Slack behavior need maximal control
- keeping install, event, and message handling in Go is preferred

Characteristics:

- Slack OAuth, events, commands, interactivity, and outbound messages all live in `apps/server/internal/modules/slack`
- bot tokens are stored by FortyOne
- all public Slack endpoints terminate in Go
- no additional Slack-specific TypeScript runtime is required

### Approach B: Hybrid Chat SDK Runtime

Use this when:

- fast iteration on DMs, commands, actions, modals, and assistant flows matters
- Maya in Slack is a first-order goal
- multi-provider chat surfaces are likely later
- a TypeScript runtime is acceptable

Characteristics:

- FortyOne still owns integration records, routing rules, user links, and delivery logs
- Chat SDK handles Slack app mechanics such as OAuth runtime behavior, DM handling, command handling, actions, modals, streaming, and assistant UX
- a TypeScript runtime lives in `apps/projects` API routes or a dedicated app
- the TypeScript runtime calls FortyOne APIs or internal services for domain reads and writes

### Architecture Recommendation

Build the plan so Phase 0 and Phase 1 are runtime-neutral. Choose the runtime implementation after the common data model and contracts are in place.

That keeps the product from being locked into either path too early.

## Target User Experience

### Workspace Admin

- Opens `Settings -> Workspace -> Integrations`
- Clicks `Connect Slack`
- Completes Slack install for a workspace
- Returns to FortyOne and sees the connected Slack workspace
- Syncs channels
- Chooses whether the integration supports notifications, commands, unfurls, interactive actions, and optionally assistant mode

### Team Lead

- Opens team settings
- Chooses one or more Slack channels for the team
- Configures which events should post to which channels
- Chooses whether Slack users can create stories, update status, or comment from Slack
- Optionally enables Slack assistant access for the team or workspace later

### Contributor

- Pastes a FortyOne story link in Slack and gets an unfurl
- Uses a slash command, shortcut, or assistant prompt to create a story
- Updates a story from Slack using buttons or a modal
- Gets relevant story notifications in the right channel
- Optionally asks Maya questions like "what am I working on this week?"

## Functional Requirements

### Required for Version 1

- Connect a Slack workspace through install flow
- Persist installation metadata and health state
- Sync channels from Slack into FortyOne
- Assign channels to teams
- Configurable notifications for:
  - story created
  - story assigned
  - status changed
  - due date changed
  - comment added
- Link unfurling for FortyOne story URLs
- Slash command or shortcut to create a story
- Modal-based story creation
- Slack action buttons for:
  - mark started
  - mark completed
  - open in FortyOne
  - assign to me when identity mapping exists
- Delivery logs and retry handling

### Strongly Recommended for Version 1.1

- Story search command from Slack
- Comment-on-story from Slack modal
- Per-channel notification filtering
- Slack user to FortyOne member identity linking
- Controlled DM notifications for direct assignments

### Optional Assistant Track

- Slack DM assistant backed by Maya
- Team or workspace summaries in Slack
- Personal "what am I working on this week?" responses
- Suggested prompts in Slack

This track should start only after Slack identity linking is reliable.

### Out of Scope for First Release

- Full bidirectional sync of every Slack thread in every channel
- Slack Canvas, Lists, Workflow Builder, or Huddles integrations
- Real-time presence sync
- Enterprise Grid-specific org-wide controls unless needed later
- Broad conversational access without identity linking

## Common Data Model

The data model should work for both runtime approaches.

### 1. `slack_workspaces`

Purpose:

- Store the Slack workspace installation associated with a FortyOne workspace

Key fields:

- `id`
- `workspace_id`
- `slack_team_id`
- `slack_team_name`
- `slack_team_domain`
- `bot_user_id`
- `runtime_kind`
- `credential_mode`
- `access_token_encrypted`
- `runtime_installation_ref`
- `scope`
- `runtime_metadata`
- `is_active`
- `installed_by_user_id`
- `last_validated_at`
- `created_at`
- `updated_at`

Notes:

- `runtime_kind` should indicate `native` or `chat_sdk`
- `credential_mode` should indicate whether FortyOne stores the token directly or the runtime manages credentials
- `access_token_encrypted` may be nullable when credentials are runtime-managed

### 2. `slack_channels`

Purpose:

- Inventory channels synced from Slack

Key fields:

- `id`
- `slack_workspace_id`
- `slack_channel_id`
- `name`
- `is_private`
- `is_archived`
- `is_member`
- `last_synced_at`
- `created_at`
- `updated_at`

### 3. `slack_team_channel_assignments`

Purpose:

- Map product teams to Slack channels

Key fields:

- `id`
- `team_id`
- `slack_channel_id`
- `assignment_type`
- `is_default`
- `created_at`
- `updated_at`

Possible `assignment_type` values:

- `notifications`
- `triage`
- `planning`
- `assistant_home` later if useful

### 4. `slack_team_settings`

Purpose:

- Team-level Slack notification and behavior controls

Key fields:

- `team_id`
- `notify_story_created`
- `notify_story_assigned`
- `notify_status_changed`
- `notify_comment_added`
- `notify_due_date_changed`
- `notify_only_high_priority`
- `allow_slack_story_creation`
- `allow_slack_status_updates`
- `allow_slack_comment_creation`
- `allow_thread_sync`
- `allow_slack_assistant`
- `default_channel_assignment_id`
- `created_at`
- `updated_at`

### 5. `slack_user_links`

Purpose:

- Map Slack users to FortyOne members

Key fields:

- `id`
- `workspace_id`
- `user_id`
- `slack_user_id`
- `slack_username`
- `slack_display_name`
- `link_status`
- `created_at`
- `updated_at`

### 6. `slack_messages`

Purpose:

- Track outbound Slack messages created by the integration

Key fields:

- `id`
- `workspace_id`
- `team_id`
- `story_id`
- `comment_id`
- `slack_channel_id`
- `slack_ts`
- `message_type`
- `runtime_kind`
- `blocks`
- `status`
- `created_at`
- `updated_at`

### 7. `slack_event_deliveries`

Purpose:

- Store inbound Slack events and callback payloads or runtime callback records

Key fields:

- `id`
- `workspace_id`
- `slack_team_id`
- `event_type`
- `request_id`
- `runtime_kind`
- `payload`
- `signature_valid`
- `processed`
- `processed_at`
- `attempt_count`
- `last_error`
- `created_at`

### 8. `slack_install_sessions`

Purpose:

- Protect install flow and unify native OAuth state with runtime-managed install handoff

Key fields:

- `id`
- `workspace_id`
- `user_id`
- `runtime_kind`
- `state_token`
- `expires_at`
- `used_at`
- `created_at`

## Slack App Capabilities

These capabilities are needed regardless of runtime choice.

### OAuth Scopes

Recommended initial scopes:

- `commands`
- `chat:write`
- `channels:read`
- `groups:read` if private channels are included
- `links:read`
- `links:write`
- `users:read`
- `users:read.email` only if needed for identity linking

Only request what the product will actually use.

### Event Subscriptions

Recommended initial subscriptions:

- install and uninstall relevant events
- channel rename and channel deletion events
- message events only when thread sync is actually shipped

Keep initial event scope lean.

### Interactivity

Enable:

- block actions
- global shortcuts
- message shortcuts
- modal submissions

### Unfurling

Enable app unfurls for FortyOne story URLs.

## API Plan

### Workspace-Authenticated Routes

All under `/workspaces/{workspaceSlug}/integrations/slack`

Suggested routes:

- `GET /status`
- `POST /connect`
- `GET /channels`
- `POST /channels/sync`
- `POST /teams/{teamId}/assignments`
- `DELETE /teams/{teamId}/assignments/{assignmentId}`
- `GET /teams/{teamId}/settings`
- `PUT /teams/{teamId}/settings`
- `POST /disconnect`
- `POST /users/link`

These routes should not change based on runtime choice.

### Public Slack Routes

Suggested public contract:

- `GET /integrations/slack/oauth/callback`
- `POST /integrations/slack/events`
- `POST /integrations/slack/interactivity`
- `POST /integrations/slack/commands`
- `POST /integrations/slack/unfurls` if required by implementation details

Notes:

- In the native approach, these terminate in `apps/server`
- In the Chat SDK approach, these may terminate in the TypeScript runtime, but they should still map to the same product concepts and auditing flow

### Internal Runtime Contract

Define a runtime service interface with methods such as:

- `StartInstall`
- `CompleteInstall`
- `SyncChannels`
- `PostStoryNotification`
- `PostDirectAssignmentNotification`
- `OpenCreateStoryModal`
- `HandleCommand`
- `HandleInteraction`
- `PublishUnfurl`
- `SendAssistantReply`

The product services should call this interface rather than embedding Slack SDK logic directly.

## Install Flow

### Common Flow

1. User starts install from workspace settings.
2. FortyOne creates an install session bound to the workspace and user.
3. User is redirected to the Slack install flow.
4. Slack returns to the runtime callback.
5. Runtime validates the session and completes installation.
6. FortyOne records installation metadata and health state.
7. Initial channel sync runs.
8. UI reflects connected state.

### Native Variant

- FortyOne exchanges the OAuth code directly
- FortyOne stores the encrypted bot token
- FortyOne verifies future Slack requests directly

### Chat SDK Variant

- the runtime completes installation using Chat SDK
- FortyOne mirrors installation metadata needed for settings and routing
- credentials may remain runtime-managed
- FortyOne still records install success, health, and workspace linkage

## Notifications Design

### Delivery Source

Use existing domain events and notification-relevant service hooks, not frontend-only triggers.

Events that should be translated into Slack messages:

- story created
- story assigned
- story updated with meaningful fields
- comment created

This likely means using current event flow and adding a Slack-specific consumer or async dispatcher.

### Notification Filtering

Channel noise is the main risk. Support:

- team-scoped delivery
- event-type filtering
- priority filtering
- optional DM delivery for direct assignment events

### Message Format

Slack messages should be concise and action-oriented:

- title
- status
- assignee
- priority
- due date if relevant
- direct link back to FortyOne

Use Block Kit, not plain text.

## Story Creation from Slack

### Slash Command

Recommended command:

- `/fortyone`

Supported initial intents:

- create story
- search story
- open help

Prefer opening a modal after the command instead of over-parsing free text.

### Global Shortcut

Recommended:

- `Create story in FortyOne`

Opens a modal with:

- team
- title
- description
- priority
- assignee
- status
- due date

### Message Shortcut

Recommended:

- `Create story from message`

Behavior:

- capture original Slack message text
- prefill story title and description
- store source link metadata

## Unfurl Design

When a story URL is pasted into Slack:

- validate the link points to a known workspace
- resolve the story through the backend
- check whether the requesting Slack workspace is connected to the same product workspace
- publish an unfurl with compact story data

Optional buttons:

- Open in FortyOne
- Mark started
- Mark completed
- Assign to me

Note:

- If the selected runtime does not provide strong unfurl support, keep unfurls implemented natively behind the same product contract

## Interactive Actions

### Supported Action Types

- status change buttons
- assign to me
- open edit modal
- add comment modal

### Handling Flow

1. Slack posts the interaction payload to the runtime.
2. The runtime validates authenticity.
3. FortyOne resolves workspace, user, and target story.
4. FortyOne performs the action through normal services.
5. The runtime updates the Slack message if necessary.

## Maya in Slack

### Goal

Allow users to ask assistant-style questions in Slack such as:

- what am I working on this week
- what is blocked in Team A
- summarize my sprint
- create a story from this request

### Preconditions

- Slack user linking must exist
- workspace and team access must be resolvable for the Slack user
- Maya tools must be runnable for a non-browser actor

### Recommended Implementation

- keep Maya logic and domain tools owned by FortyOne
- resolve Slack user to a FortyOne actor
- run assistant requests under that actor's permissions
- return the response through the Slack runtime

### Important Constraint

Do not assume the existing web chat route can be reused unchanged. Slack requests should use a Slack-specific actor flow rather than browser session auth.

## Slack Thread Sync

This should be a controlled later phase, not default v1 behavior.

### Recommended Later Behavior

- Only enable for channels explicitly mapped to a team
- Only sync replies made in threads started by FortyOne-posted notification messages
- Mirror thread replies into story comments
- Do not attempt to mirror arbitrary Slack conversation history

### Reasoning

Slack thread syncing gets messy fast:

- private discussions may not belong in the product
- edits and deletes are complex
- user identity mapping may be incomplete
- threads often drift away from the story context

## Identity Mapping

Slack actions like `Assign to me` and Maya-in-Slack require mapping Slack users to FortyOne members.

Recommended order:

1. Try email-based linking if permitted by scopes and user emails are available
2. Allow manual linking in settings
3. Store the result in `slack_user_links`

If no mapping exists:

- do not fail silently
- show a clear Slack error or ephemeral notice

## UI Plan

### Workspace Settings

Add a Slack integration details page under workspace settings with:

- connected workspace
- runtime kind
- install health
- channels
- capabilities enabled
- disconnect action

### Team Settings

Add a Slack section or adjacent integration tab with:

- channel assignments
- notification rules
- action permissions
- thread sync toggle
- assistant access toggle later

### Story UI

Slack does not need deep story embedding like GitHub. The story page should only show Slack-related metadata if useful:

- channels where the story was posted
- recent Slack-linked messages if thread sync is enabled

This can be deferred.

## Recommended Implementation Sequence

### Phase 0: Runtime-Neutral Foundation

- Add migrations for common Slack tables
- Add `apps/server/internal/modules/slackcore`
- Define runtime interface and product service contract
- Wire workspace and team settings routes
- Add workspace settings UI skeleton

### Phase 1: Install and Channel Sync

- Implement common install session flow
- Persist Slack workspace linkage and health state
- Sync channels
- Ship workspace status and channel management UI

### Phase 2: Notifications

- Add Slack delivery orchestration
- Add async dispatcher for story and comment events
- Add team-to-channel routing
- Ship channel notifications

### Phase 3A: Native Runtime Track

- Implement native Slack client in Go
- Implement OAuth callback in Go
- Implement command, interactivity, and event endpoints in Go
- Implement outbound posting and message updates in Go

### Phase 3B: Chat SDK Runtime Track

- Add TypeScript Slack runtime in `apps/projects` or a dedicated app
- Implement install callback, commands, modals, actions, and DMs with Chat SDK
- Add internal calls from the runtime to FortyOne services
- Mirror install metadata and health back into FortyOne

### Phase 4: Story Creation and Search

- Add slash command and shortcut handling
- Add modal submit handling
- Support story creation and story search

### Phase 5: Unfurls and Interactive Actions

- Add link unfurls
- Add status buttons
- Add assign-to-me button
- Add comment modal

### Phase 6: Assistant Track

- Add Slack user linking hardening
- Add Maya actor resolution for Slack users
- Add DM assistant support
- Add summaries and personal workload prompts

### Phase 7: Optional Thread Sync

- Only for opted-in teams and integration-originated threads
- Mirror thread replies into story comments
- Add safeguards and auditability

## Testing Strategy

### Unit Tests

- install session validation
- Slack signature verification or runtime callback validation
- Block Kit payload generation
- channel routing resolution
- user identity mapping
- Maya actor resolution for Slack-linked users

### Repository Tests

- channel sync upserts
- assignment updates
- event delivery persistence
- message persistence

### Service Tests

- story created sends message to expected channel
- slash command opens modal
- modal submission creates story
- interactive action updates story state
- assistant request resolves the correct user and workspace

### Integration Tests

- install callback success flow
- unfurl request flow
- slash command flow
- interactive action flow
- DM assistant flow once shipped

## Operational Concerns

### Logging

Structured fields should include:

- workspace id
- slack workspace id
- channel id
- team id
- story id
- runtime kind
- callback type
- request id
- attempt count

### Metrics

Track:

- install success and failure
- channel sync counts
- message delivery success and failure
- interactive action success and failure
- modal submission failures
- unfurl counts
- assistant request counts and failures

### Retry Strategy

Outbound message delivery should:

- retry transient Slack API failures
- respect rate limits
- back off on `429`
- preserve enough context for replay

## Risks

### Risk: Divergent Implementations

Mitigation:

- keep one common product contract
- keep settings, data model, and auditability runtime-neutral
- isolate runtime-specific behavior behind a defined interface

### Risk: Channel Noise

Mitigation:

- make notifications opt-in by event type
- keep channel mapping team-specific
- support high-priority-only modes later

### Risk: Identity Mapping Gaps

Mitigation:

- allow manual user links
- make actions degrade gracefully
- do not ship broad assistant access before linking is reliable

### Risk: Overambitious Thread Sync

Mitigation:

- keep thread sync out of first release
- only enable for integration-originated threads later

### Risk: Slack Scope Bloat

Mitigation:

- request minimal scopes initially
- add scopes only when the feature actually ships

### Risk: Chat SDK Runtime Drift

Mitigation:

- keep core product state in FortyOne
- do not let runtime-managed credentials become the only source of install truth
- keep native fallback possible for critical paths like unfurls if necessary

## Recommended Milestones

### Milestone 1

- Runtime-neutral Slack core exists
- Workspace connects Slack
- Channels sync
- Team-to-channel assignments work

### Milestone 2

- Story and comment notifications post to Slack
- Delivery and retry tracking exists

### Milestone 3

- Slash command and shortcut-based story creation works
- Modal submission creates story successfully

### Milestone 4

- Link unfurls work
- Interactive actions can change status and assign stories

### Milestone 5

- Slack user linking is reliable
- Maya can answer personal and team summary questions in Slack

### Milestone 6

- Optional thread sync for integration-originated threads
- Richer action support and assistant prompts

## Open Questions

- Should a workspace support one Slack workspace only, or multiple installs later?
- Should private channel support be in the first release or follow after public channels?
- Which events are high enough signal to notify by default?
- Do we want slash commands only, or should shortcuts be the primary entry point?
- How much of comment syncing from Slack threads is actually desirable versus noisy?
- Is the team willing to own a long-term TypeScript Slack runtime if Chat SDK is selected?
- Do we want native unfurls even if the rest of Slack uses Chat SDK?

## Final Recommendation

Build Slack as a dedicated FortyOne integration with a runtime-neutral core and a pluggable Slack runtime.

Do not fork the product plan into separate "native Slack" and "Chat SDK Slack" documents. Keep one data model, one settings model, one audit model, and one set of milestones. Then choose the runtime implementation:

- choose native Go when stack simplicity and maximal control matter most
- choose Chat SDK when Slack assistant UX, faster bot iteration, and future multi-provider chat surfaces matter most

The strongest near-term path is to ship the common Slack core first, then decide whether Phase 3 is native or Chat SDK based on how important Maya-in-Slack is to the product roadmap.
