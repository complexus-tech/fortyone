# Slack Integration Plan

## Document Status

- Status: Draft implementation plan
- Scope: New Slack integration for the `fortyone` monorepo
- Primary app surfaces:
  - `apps/projects` for workspace and team settings UI
  - `apps/server` for OAuth, Slack events, interactive callbacks, and message delivery
- Goal: Production-grade Slack integration for notifications, commands, collaboration workflows, and optional thread-to-story syncing

## Problem Statement

The product needs a Slack integration that does more than send generic notifications. Teams should be able to connect a Slack workspace, route channels to teams, receive high-signal updates, create and manage work from Slack, and optionally sync discussion context back into FortyOne.

This should fit the existing workspace, team, story, comment, and notification architecture already present in the repo, not behave like a disconnected add-on.

## Recommended Product Scope

Slack is not GitHub. It should not try to be a full mirror of the product. The recommended Slack integration should focus on four strong workflows:

1. Notifications
2. Story creation and lookup
3. Link unfurling
4. Interactive actions from Slack

Optional deeper sync such as Slack thread mirroring can come after those foundations are stable.

## Current Repository Reality

The best integration hooks already exist in the codebase:

- Workspace settings navigation is easy to extend from `apps/projects/src/components/layouts/settings.tsx`
- Team settings already has a natural place for team-level routing and automation rules
- Story creation, updates, comments, and notifications already move through backend services and event publishing
- Redis-backed async processing already exists for background work

This means Slack should be modeled as a proper integration module, not as raw incoming webhooks glued onto the frontend.

## Guiding Principles

1. Use a Slack app with OAuth, event subscriptions, interactivity, and unfurls.
2. Keep Slack connection state at the workspace level.
3. Keep channel routing and notification rules at the team level.
4. Avoid noisy default behavior. Slack should be configurable and high signal.
5. Use asynchronous delivery and retry logic for outbound messages.
6. Verify every inbound request using Slack signing secrets.
7. Prefer interactive modals and actions over fragile slash-command argument parsing when possible.
8. Do not attempt full chat sync in the first release.

## Target User Experience

### Workspace Admin

- Opens `Settings -> Workspace -> Integrations`
- Clicks `Connect Slack`
- Completes Slack OAuth for a workspace
- Returns to FortyOne and sees the connected Slack workspace
- Syncs channels
- Chooses whether the integration should support notifications, commands, unfurls, and interactive actions

### Team Lead

- Opens team settings
- Chooses one or more Slack channels for the team
- Configures which events should post to which channels
- Chooses whether Slack users can create stories, update status, or comment from Slack

### Contributor

- Pastes a FortyOne story link in Slack and gets an unfurl
- Uses a slash command or message shortcut to create a story
- Updates a story from Slack using buttons or a modal
- Gets relevant story notifications in the right channel

## Functional Requirements

### Required for Version 1

- Connect a Slack workspace through OAuth
- Persist workspace bot token and basic metadata securely
- Sync channels from Slack into FortyOne
- Assign channels to teams
- Configurable notifications for:
  - story created
  - story assigned
  - status changed
  - due date changed
  - comment added
  - sprint started or ended later if desired
- Link unfurling for FortyOne story URLs
- Slash command or shortcut to create a story
- Modal-based story creation
- Slack action buttons for:
  - mark started
  - mark completed
  - open in FortyOne
  - assign to me when the identity mapping exists
- Delivery logs and retry handling

### Strongly Recommended for Version 1.1

- Story search command from Slack
- Comment-on-story from Slack modal
- Thread replies mirrored into story comments for explicitly connected channel threads
- Per-channel notification filtering
- Slack user to FortyOne member identity linking

### Out of Scope for First Release

- Full bidirectional sync of every Slack thread in every channel
- Slack Canvas, Lists, Workflow Builder, or Huddles integrations
- Full DM bot assistant
- Real-time presence sync
- Enterprise Grid-specific org-wide controls unless needed later

## Product Definition

### Core Slack Workflows

#### 1. Notifications

FortyOne posts clear, compact updates into mapped Slack channels.

Examples:

- New story created in Team A channel
- Story assigned to a member
- Story moved to completed
- High-priority item overdue

#### 2. Story Creation

Users can create a story from Slack via:

- slash command
- global shortcut
- message shortcut

#### 3. Unfurls

When a FortyOne link is pasted into Slack, Slack shows:

- story title
- status
- assignee
- priority
- due date
- team

#### 4. Interactive Actions

Slack message actions can:

- change story status
- assign the story
- open a modal to edit fields
- add a comment

## Recommended Architecture

### High-Level Model

- Workspace connects one Slack workspace
- Slack channels are synced into FortyOne
- Teams map to channels
- Slack messages are outbound notifications or interactive surfaces
- Inbound Slack actions call backend endpoints and then the normal story/comment services

### Module Layout

Create a dedicated module:

- `apps/server/internal/modules/slack/`

Suggested sub-packages:

- `service`
- `repository`
- `http`
- `client`
- `oauth`
- `events`
- `interactivity`
- `unfurls`

Responsibilities:

- `service`: core orchestration
- `repository`: Slack persistence and routing data
- `http`: OAuth and public Slack endpoints
- `client`: typed Slack API wrapper
- `oauth`: install and token exchange flow
- `events`: event subscription processing
- `interactivity`: button, modal, and shortcut handling
- `unfurls`: story link presentation logic

## Data Model Design

### 1. `slack_workspaces`

Purpose:

- Store the Slack workspace installation for a FortyOne workspace

Key fields:

- `id`
- `workspace_id`
- `slack_team_id`
- `slack_team_name`
- `slack_team_domain`
- `bot_user_id`
- `access_token_encrypted`
- `scope`
- `is_active`
- `installed_by_user_id`
- `created_at`
- `updated_at`

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
- `blocks`
- `status`
- `created_at`
- `updated_at`

### 7. `slack_event_deliveries`

Purpose:

- Store inbound Slack events and callback payloads

Key fields:

- `id`
- `workspace_id`
- `slack_team_id`
- `event_type`
- `request_id`
- `payload`
- `signature_valid`
- `processed`
- `processed_at`
- `attempt_count`
- `last_error`
- `created_at`

### 8. `slack_oauth_states`

Purpose:

- Protect OAuth installation flow against CSRF and stale callbacks

Key fields:

- `id`
- `workspace_id`
- `user_id`
- `state_token`
- `expires_at`
- `used_at`
- `created_at`

## Slack App Capabilities

### OAuth Scopes

Recommended initial scopes:

- `commands`
- `chat:write`
- `chat:write.customize` if needed later
- `channels:read`
- `groups:read`
- `links:read`
- `links:write`
- `users:read`
- `users:read.email` if identity linking needs it

Only request what the product will actually use.

### Event Subscriptions

Recommended initial event subscriptions:

- `app_uninstalled`
- `channel_deleted`
- `channel_rename`
- `message.channels` only if thread sync is enabled later
- `message.groups` only if private channel sync is enabled later

Keep initial event scope lean.

### Interactivity

Enable:

- Block actions
- Global shortcuts
- Message shortcuts
- Modal submissions

### Unfurling

Enable app unfurls for FortyOne story URLs.

## Backend API Plan

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

### Public Routes

Suggested routes:

- `GET /integrations/slack/oauth/callback`
- `POST /integrations/slack/events`
- `POST /integrations/slack/interactivity`
- `POST /integrations/slack/commands`

## OAuth Flow

### Install Flow

1. User starts install from workspace settings.
2. Backend creates a signed OAuth state entry bound to the workspace and user.
3. User is redirected to Slack OAuth.
4. Slack redirects to the callback.
5. Backend validates state.
6. Backend exchanges code for bot token.
7. Backend stores installation metadata and encrypted token.
8. Backend triggers initial channel sync.
9. UI reflects connected state.

### Reinstall and Uninstall

The system must handle:

- reinstall over existing connection
- revoked token
- Slack-side uninstall event
- workspace disconnect initiated from FortyOne

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
- optionally mention filtering later

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

However, parsing free text should be minimal. Prefer opening a modal after the command.

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

## Interactive Actions

### Supported Action Types

- status change buttons
- assign to me
- open edit modal
- add comment modal

### Handling Flow

1. Slack posts interaction payload.
2. Backend verifies Slack signature.
3. Backend resolves workspace, user, and target story.
4. Backend performs the action through normal services.
5. Backend updates the Slack message if necessary.

## Slack Thread Sync

This should be a controlled later phase, not default v1 behavior.

### Recommended V1.1 Behavior

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

So the product should support thread sync only when the thread was created by the integration and the team opted in.

## Identity Mapping

Slack actions like `Assign to me` require mapping Slack users to FortyOne members.

Recommended order:

1. Try email-based linking if permitted by scopes and user emails are available
2. Allow manual linking in settings
3. Store result in `slack_user_links`

If no mapping exists:

- do not fail silently
- show a clear Slack error or ephemeral notice

## UI Plan

### Workspace Settings

Add:

- `apps/projects/src/app/[workspaceSlug]/settings/workspace/integrations/page.tsx`
- Slack provider card
- Slack integration details page

Sections:

- connected workspace
- OAuth health
- channels
- capabilities enabled
- disconnect action

### Team Settings

Add a Slack section or adjacent integration tab with:

- channel assignments
- notification rules
- action permissions
- thread sync toggle

### Story UI

Slack does not need deep story embedding like GitHub. The story page should only show Slack-related metadata if useful:

- channels where the story was posted
- recent Slack-linked messages if thread sync is enabled

This can be deferred.

## Recommended File-Level Implementation Sequence

### Phase 1: Foundation

- Add migrations
- Add Slack module scaffolding in `apps/server/internal/modules/slack`
- Wire routes in `apps/server/internal/bootstrap/api/routes.go`
- Wire services in `apps/server/internal/bootstrap/api/services.go`

### Phase 2: OAuth and Channel Sync

- Implement OAuth start and callback
- Persist Slack workspace
- Sync channels
- Add workspace settings UI

### Phase 3: Notifications

- Add Slack delivery service
- Add async dispatcher for story and comment events
- Add team-to-channel routing
- Ship channel notifications

### Phase 4: Commands, Shortcuts, and Modals

- Add slash command endpoint
- Add shortcuts and modal submit handling
- Support story creation and story search

### Phase 5: Unfurls and Interactive Actions

- Add link unfurls
- Add status buttons
- Add assign-to-me button
- Add comment modal

### Phase 6: Optional Thread Sync

- Only for opted-in teams and integration-originated threads
- Mirror thread replies into story comments
- Add safeguards and auditability

## Testing Strategy

### Unit Tests

- Slack signature verification
- OAuth state validation
- Block Kit payload generation
- channel routing resolution
- user identity mapping

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

### Integration Tests

- OAuth callback success flow
- unfurl request flow
- slash command flow
- interactive action flow

## Operational Concerns

### Logging

Structured fields should include:

- workspace id
- slack workspace id
- channel id
- team id
- story id
- callback type
- request id
- attempt count

### Metrics

Track:

- message delivery success and failure
- interactive action success and failure
- modal submission failures
- unfurl counts
- channel sync counts
- install and uninstall counts

### Retry Strategy

Outbound message delivery should:

- retry transient Slack API failures
- respect rate limits
- back off on `429`
- preserve enough context for replay

## Risks

### Risk: Channel Noise

Mitigation:

- make notifications opt-in by event type
- keep channel mapping team-specific
- support high-priority-only modes later

### Risk: Identity Mapping Gaps

Mitigation:

- allow manual user links
- make interactive actions degrade gracefully

### Risk: Overambitious Thread Sync

Mitigation:

- keep thread sync out of first release
- only enable for integration-originated threads in later phases

### Risk: Slack Scope Bloat

Mitigation:

- request minimal scopes initially
- add scopes only when the feature actually ships

## Recommended Milestones

### Milestone 1

- Workspace connects Slack
- Channels sync
- Team-to-channel assignments work

### Milestone 2

- Story and comment notifications post to Slack
- Delivery and retry tracking exists

### Milestone 3

- Slash command and shortcut-based story creation
- Modal submission creates story successfully

### Milestone 4

- Link unfurls work
- Interactive actions can change status and assign stories

### Milestone 5

- Optional thread sync for integration-originated threads
- Identity linking and richer action support

## Open Questions

- Should a workspace support one Slack workspace only, or multiple installs later?
- Should private channel support be in the first release or follow after public channels?
- Which events are high enough signal to notify by default?
- Do we want slash commands only, or should shortcuts be the primary entry point?
- How much of comment syncing from Slack threads is actually desirable versus noisy?

## Final Recommendation

Build Slack as a dedicated integration centered on workspace OAuth, team-to-channel routing, high-signal notifications, link unfurls, and interactive story workflows. Do not attempt broad conversational sync at first. The best first release is one that makes FortyOne visible and actionable inside Slack without turning Slack into a second full copy of the application.
