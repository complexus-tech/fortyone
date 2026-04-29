# Multi-Channel Messaging Integration Plan

## Document Status

- Status: Active implementation plan
- Scope: Multi-channel messaging integration for the `fortyone` monorepo
- Primary app surfaces:
  - `apps/projects` for workspace and team settings UI
  - `apps/server` for integration state, routing, permissions, and domain events
  - `apps/bot` (new) TypeScript Chat SDK runtime serving all messaging platforms
- Goal: Production-grade multi-channel integration that brings Maya and notifications to Slack, Microsoft Teams, WhatsApp, and other platforms through a single runtime

## Problem Statement

Teams need to interact with FortyOne from the messaging platforms they already use. This means receiving high-signal notifications, creating and managing work, and chatting with Maya directly from Slack, Teams, WhatsApp, and other channels.

Maya already exists as a full AI assistant with 80+ tools built on Vercel AI SDK. The goal is to extend Maya's reach beyond the web UI into external messaging platforms, while also delivering configurable notifications and interactive actions.

This should fit the existing workspace, team, story, comment, notification, and integration architecture. Each messaging platform should behave like a first-class product integration, not a disconnected bot.

## Product Scope

The integration should focus on five core workflows across all supported platforms:

1. AI assistant conversations (Maya)
2. Notifications
3. Story creation and lookup
4. Interactive actions
5. Link unfurling (where supported)

Each platform delivers these workflows at the level its capabilities allow. Slack gets the richest experience; WhatsApp gets a conversational-first experience within its constraints.

## Current Repository Reality

The best integration hooks already exist in the codebase:

- **Maya AI assistant** with 80+ tools (stories, sprints, objectives, teams, search, memory, etc.) built with Vercel AI SDK v6 (`streamText`, `tool()`, Zod schemas)
- **Redis Streams event system** publishing domain events (`story.created`, `story.updated`, `comment.created`, `user.mentioned`, etc.)
- **GitHub integration module** in Go showing the service/repository/http pattern for integrations
- **Workspace settings** with an existing integrations area
- **Team settings** with a natural place for channel routing and automation rules
- **System actor pattern** (`maya@fortyone.app`, `github@fortyone.app`) for non-browser actors

This means the primary work is building the Chat SDK runtime and the shared integration data layer, not rebuilding AI or event infrastructure.

## Guiding Principles

1. FortyOne owns product state, permissions, team routing, user identity mapping, and auditability for all providers.
2. Chat SDK owns messaging platform mechanics. One codebase, multiple adapters.
3. Maya's tools are the same everywhere. The AI layer does not fork per platform.
4. Keep installation state at the workspace level.
5. Keep channel routing and notification rules at the team level.
6. Degrade gracefully per platform. If a platform cannot do modals, use conversational flow. If it cannot do cards, use plain text.
7. Avoid noisy default behavior. Notifications should be configurable and high signal.
8. Use asynchronous delivery and retry logic for outbound messages.
9. Verify every inbound request using platform-native signature verification (handled by Chat SDK adapters).
10. Do not make the first release depend on every platform shipping simultaneously. Slack first, then expand.

## Architecture

### Overview

```
                        FortyOne Web UI
                              |
                         apps/projects
                          (Next.js)
                              |
                     POST /api/chat ──── Maya (AI SDK + tools)
                                              |
                                              | (same tools, different entry point)
                                              |
┌──────────────────────────────────────────────────────────────────┐
│                     apps/bot  (Chat SDK Runtime)                 │
│                                                                  │
│   ┌─────────┐   ┌─────────┐   ┌───────────┐   ┌───────────┐   │
│   │  Slack  │   │  Teams  │   │  WhatsApp │   │  Discord  │   │
│   │ adapter │   │ adapter │   │  adapter  │   │  adapter  │   │
│   └────┬────┘   └────┬────┘   └─────┬─────┘   └─────┬─────┘   │
│        └─────────────┴──────────────┴───────────────┘           │
│                             |                                    │
│                   Chat handlers layer                            │
│            (onNewMention, onSlashCommand, onAction,              │
│             onModalSubmit, onSubscribedMessage)                   │
│                             |                                    │
│                 Resolve platform user → FortyOne member           │
│                             |                                    │
│                 streamText() + Maya tools                         │
│                             |                                    │
│                 thread.post(result.fullStream)                    │
│                                                                  │
│   ┌──────────────────────────────────────┐                      │
│   │  Event Consumer (Redis Streams)      │                      │
│   │  domain events → notifications       │                      │
│   │  → team routing → thread.post()      │                      │
│   └──────────────────────────────────────┘                      │
└──────────────────────────────────────────────────────────────────┘
                              |
                     FortyOne Backend (Go)
                        apps/server
                  ┌──────────────────────┐
                  │ API for tool calls   │
                  │ Redis event stream   │
                  │ Installation records │
                  │ User link records    │
                  │ Permissions & scope  │
                  └──────────────────────┘
```

### Always Owned by FortyOne (Go Backend)

These concerns remain in `apps/server` regardless of messaging platform:

- Workspace-to-provider installation ownership
- Channel/conversation inventory used by product settings
- Team-to-channel assignments
- Team-level notification and action policies
- Platform user to FortyOne member linking
- Delivery logs and audit trails
- Domain event publishing (Redis Streams)
- Story and comment permissions
- API routes for workspace and team settings

### Chat SDK Runtime (`apps/bot`)

A TypeScript application using Chat SDK with adapters for each platform. Responsibilities:

- Register and manage platform adapters (Slack, Teams, WhatsApp, Discord, etc.)
- Handle inbound webhooks, commands, actions, modals, and mentions
- Resolve platform users to FortyOne members
- Execute Maya's AI tools via `streamText()` with the same tool definitions used in the web UI
- Stream responses back to the originating platform via `thread.post(result.fullStream)`
- Consume Redis Streams domain events and dispatch notifications to assigned channels
- Manage thread subscriptions and state via Chat SDK state adapters (Redis or Postgres)

### AI Layer

Maya's tools are defined once and shared between the web chat route (`apps/projects/src/lib/ai/tools/`) and the bot runtime. The key difference is authentication:

- **Web UI**: `auth()` (NextAuth session) resolves the user
- **Bot runtime**: Platform identity → `channel_user_links` → FortyOne member → workspace-scoped API calls

The tool definitions, Zod schemas, system prompt, and model configuration are extracted into a shared package or imported directly. The AI SDK `streamText()` call is nearly identical in both entry points.

### Event-Driven Notifications

The Go backend already publishes domain events to Redis Streams:

- `story.created`, `story.updated`, `story.duplicated`
- `comment.created`, `comment.replied`
- `user.mentioned`
- `objective.updated`, `keyresult.updated`

The bot runtime subscribes to this stream and for each event:

1. Resolves which workspace the event belongs to
2. Looks up active provider installations for that workspace
3. Looks up team-to-channel assignments and notification settings
4. Formats the message using Chat SDK JSX (renders natively per platform)
5. Posts to the assigned channels via `thread.post()`

## Platform Capability Matrix

| Capability | Slack | Teams | WhatsApp | Discord |
|---|---|---|---|---|
| Maya AI chat | Full native streaming | Post+edit streaming | Buffered, 4096 char chunks | Post+edit streaming |
| Slash commands | Yes | No | No | Yes (via bot commands) |
| Modals (forms) | Native | No | No | No |
| Cards / rich messages | Block Kit | Adaptive Cards | Max 3 buttons, list messages | Embeds |
| Link unfurling | Yes | No | No | Embeds |
| Notifications (channels) | Yes | Yes | DMs only | Yes |
| Notifications (DMs) | Yes | Yes | Yes (24h window) | Yes |
| Reactions | Full | Read-only | Yes | Full |
| File uploads | Yes | Yes | Yes | Yes |
| Thread subscriptions | Yes | Yes | Limited | Yes |
| Multi-workspace OAuth | Built into adapter | Azure Bot (multi-tenant) | Permanent token | OAuth2 |
| Message editing | Yes | Yes | No | Yes |

### Platform-Specific Constraints

**Slack**: Richest experience. Native streaming, modals, unfurls, slash commands, Block Kit. Multi-workspace OAuth is built into the Chat SDK adapter. This is the launch platform.

**Microsoft Teams**: Good experience. Adaptive Cards for rich messages. No modals or slash commands, so story creation is conversational (Maya walks through it). Streaming via post+edit pattern. Requires Azure Bot resource setup and Teams app manifest.

**WhatsApp**: Conversational-only experience. No modals, cards, or channels. DMs only. 24-hour messaging window — outside that window, only pre-approved template messages can be sent (requires Meta approval). Story creation is fully conversational. Notifications are DM-only and must respect the window. Identity linking is by phone number, not email.

**Discord**: Strong experience. Embeds for rich messages, bot commands, reactions. No modals. Good for developer/community-oriented workspaces.

## Target User Experience

### Workspace Admin

- Opens `Settings -> Workspace -> Integrations`
- Sees available messaging providers (Slack, Teams, WhatsApp, Discord)
- Clicks `Connect` on a provider
- Completes the provider-specific install flow
- Returns to FortyOne and sees the connected provider
- Syncs channels/conversations
- Configures which capabilities are enabled per provider

### Team Lead

- Opens team settings
- Assigns channels/conversations from connected providers to the team
- Configures which events post to which channels
- Sets notification filters (event type, priority)
- Optionally enables Maya assistant access for the team

### Contributor

- @mentions the bot in Slack/Teams/Discord or DMs on WhatsApp
- Asks Maya: "what am I working on this week?" and gets a streamed response
- Creates a story by saying "create a story for..." — Maya uses the `createStory` tool
- Gets notified in the right channel when stories are assigned, status changes, etc.
- Uses action buttons (where supported) to mark started, mark completed, or open in FortyOne
- Pastes a FortyOne link in Slack and gets a rich unfurl

## Functional Requirements

### Required for Version 1 (Slack)

- Connect a Slack workspace through OAuth install flow
- Persist installation metadata and health state
- Sync channels from Slack into FortyOne
- Assign channels to teams
- Maya AI assistant in Slack (mentions, DMs, slash commands)
- Configurable notifications for:
  - story created
  - story assigned
  - status changed
  - due date changed
  - comment added
- Link unfurling for FortyOne story URLs
- Slash command (`/fortyone`) to create a story via modal
- Global and message shortcuts for story creation
- Action buttons on notifications (mark started, mark completed, open in FortyOne)
- Delivery logs and retry handling
- Platform user to FortyOne member identity linking

### Version 1.1 (Slack Enhancements + Teams)

- Story search from Slack
- Comment-on-story from Slack modal
- Per-channel notification filtering
- DM notifications for direct assignments
- Microsoft Teams adapter with:
  - Maya AI chat
  - Channel notifications
  - Conversational story creation
  - Action buttons via Adaptive Cards

### Version 1.2 (WhatsApp + Discord)

- WhatsApp adapter with:
  - Maya AI chat (DM only)
  - DM notifications (within 24h window)
  - Template messages for outside-window notifications
  - Conversational story creation
- Discord adapter with:
  - Maya AI chat
  - Channel notifications
  - Bot commands
  - Embed-based rich messages

### Out of Scope for First Release

- Full bidirectional sync of messaging threads
- Slack Canvas, Workflow Builder, or Huddles
- Teams Copilot extensions
- WhatsApp Business API advanced features (catalogs, flows)
- Real-time presence sync
- Enterprise Grid or Azure AD org-wide controls

## Common Data Model

The data model is provider-agnostic with a `provider` discriminator.

### 1. `channel_installations`

Purpose: Store the messaging provider installation associated with a FortyOne workspace.

Key fields:

- `id` (UUID)
- `workspace_id` (FK)
- `provider` (enum: `slack`, `teams`, `whatsapp`, `discord`)
- `provider_team_id` (Slack team_id, Teams tenant_id, WhatsApp phone_number_id, Discord guild_id)
- `provider_team_name`
- `provider_team_domain`
- `bot_user_id`
- `credential_mode` (enum: `sdk_managed`, `self_managed`)
- `access_token_encrypted` (nullable, for self-managed mode)
- `scope`
- `provider_metadata` (JSONB, provider-specific fields)
- `capabilities` (JSONB, what this install supports: notifications, commands, unfurls, assistant, etc.)
- `is_active`
- `installed_by_user_id` (FK)
- `last_validated_at`
- `created_at`
- `updated_at`

Notes:

- `credential_mode` = `sdk_managed` means Chat SDK adapter stores/manages tokens (default for Slack multi-workspace)
- `credential_mode` = `self_managed` means FortyOne stores the encrypted token directly
- `capabilities` allows the product to know what features this installation supports, since platforms differ

### 2. `channel_conversations`

Purpose: Inventory channels/conversations synced from the messaging provider.

Key fields:

- `id` (UUID)
- `installation_id` (FK)
- `provider_channel_id`
- `name`
- `is_private`
- `is_archived`
- `is_member` (bot is a member)
- `conversation_type` (enum: `channel`, `group`, `dm`)
- `last_synced_at`
- `created_at`
- `updated_at`

### 3. `channel_team_assignments`

Purpose: Map product teams to messaging channels/conversations.

Key fields:

- `id` (UUID)
- `team_id` (FK)
- `conversation_id` (FK to `channel_conversations`)
- `assignment_type` (enum: `notifications`, `triage`, `planning`)
- `is_default`
- `created_at`
- `updated_at`

### 4. `channel_team_settings`

Purpose: Team-level notification and behavior controls per provider.

Key fields:

- `id` (UUID)
- `team_id` (FK)
- `installation_id` (FK)
- `notify_story_created` (bool)
- `notify_story_assigned` (bool)
- `notify_status_changed` (bool)
- `notify_comment_added` (bool)
- `notify_due_date_changed` (bool)
- `notify_only_high_priority` (bool)
- `allow_story_creation` (bool)
- `allow_status_updates` (bool)
- `allow_comment_creation` (bool)
- `allow_assistant` (bool)
- `default_assignment_id` (FK to `channel_team_assignments`, nullable)
- `created_at`
- `updated_at`

### 5. `channel_user_links`

Purpose: Map messaging platform users to FortyOne members.

Key fields:

- `id` (UUID)
- `installation_id` (FK)
- `user_id` (FK to FortyOne users)
- `provider_user_id` (Slack user_id, Teams AAD object_id, WhatsApp wa_id, Discord user_id)
- `provider_username`
- `provider_display_name`
- `link_method` (enum: `email_auto`, `manual`, `phone`)
- `link_status` (enum: `active`, `pending`, `unlinked`)
- `created_at`
- `updated_at`

### 6. `channel_messages`

Purpose: Track outbound messages sent by the integration.

Key fields:

- `id` (UUID)
- `installation_id` (FK)
- `team_id` (FK, nullable)
- `story_id` (FK, nullable)
- `comment_id` (FK, nullable)
- `conversation_id` (FK)
- `provider_message_id` (Slack ts, Teams message_id, etc.)
- `message_type` (enum: `notification`, `assistant_response`, `unfurl`, `action_update`)
- `content_blocks` (JSONB)
- `status` (enum: `sent`, `delivered`, `failed`, `retrying`)
- `created_at`
- `updated_at`

### 7. `channel_event_deliveries`

Purpose: Store inbound events and callback payloads from messaging providers.

Key fields:

- `id` (UUID)
- `installation_id` (FK)
- `event_type`
- `request_id`
- `payload` (JSONB)
- `signature_valid` (bool)
- `processed` (bool)
- `processed_at`
- `attempt_count`
- `last_error`
- `created_at`

### 8. `channel_install_sessions`

Purpose: Protect install flows with CSRF state tokens.

Key fields:

- `id` (UUID)
- `workspace_id` (FK)
- `user_id` (FK)
- `provider` (enum)
- `state_token`
- `expires_at`
- `used_at`
- `created_at`

## API Plan

### Workspace-Authenticated Routes

All under `/workspaces/{workspaceSlug}/integrations/messaging`

- `GET /providers` -- list available providers and their connection status
- `GET /providers/{provider}/status` -- detailed status for a specific provider
- `POST /providers/{provider}/connect` -- start install flow
- `POST /providers/{provider}/disconnect` -- disconnect a provider
- `GET /providers/{provider}/conversations` -- list synced conversations
- `POST /providers/{provider}/conversations/sync` -- trigger conversation sync
- `POST /teams/{teamId}/assignments` -- assign a conversation to a team
- `DELETE /teams/{teamId}/assignments/{assignmentId}` -- remove assignment
- `GET /teams/{teamId}/settings` -- get team notification/behavior settings
- `PUT /teams/{teamId}/settings` -- update team settings
- `POST /users/link` -- manually link a platform user to a FortyOne member
- `DELETE /users/link/{linkId}` -- remove a user link
- `GET /users/links` -- list user links for the workspace

### Public Webhook Routes

These terminate in the Chat SDK runtime (`apps/bot`) and are registered per-adapter:

- `POST /webhooks/slack/events`
- `POST /webhooks/slack/interactivity`
- `POST /webhooks/slack/commands`
- `GET /webhooks/slack/oauth/callback`
- `POST /webhooks/teams/messages`
- `GET /webhooks/whatsapp/verify`
- `POST /webhooks/whatsapp/messages`
- `POST /webhooks/discord/interactions`

Chat SDK adapters handle signature verification, payload normalization, and routing to the appropriate handler.

## Install Flow

### Slack

1. User clicks `Connect Slack` in workspace settings.
2. FortyOne creates an install session with a state token.
3. User is redirected to Slack OAuth V2 consent screen.
4. Slack redirects to the Chat SDK runtime's OAuth callback.
5. Chat SDK adapter completes the token exchange and stores credentials.
6. The runtime calls back to FortyOne API to record the installation.
7. Initial channel sync runs.
8. UI reflects connected state.

Multi-workspace OAuth is built into `@chat-adapter/slack`. Omit the bot token, provide `SLACK_CLIENT_ID` + `SLACK_CLIENT_SECRET`. Tokens are stored encrypted (AES-256-GCM) in the state adapter and resolved dynamically by `team_id`.

### Microsoft Teams

1. User clicks `Connect Teams` in workspace settings.
2. FortyOne creates an install session.
3. User installs the Teams app via the Teams admin center or app catalog.
4. Azure Bot Framework handles authentication (client secret or federated identity).
5. The runtime detects the new tenant and records the installation.
6. Channel sync runs.

Requires an Azure Bot resource and a Teams app manifest. Multi-tenant is the default.

### WhatsApp

1. User clicks `Connect WhatsApp` in workspace settings.
2. User enters WhatsApp Business API credentials (access token, phone number ID, app secret).
3. FortyOne stores credentials and activates the adapter.
4. Webhook verification handshake completes.
5. No channel sync needed (WhatsApp is DM-only).

No OAuth flow. Uses a permanent Meta system user token.

### Discord

1. User clicks `Connect Discord` in workspace settings.
2. User is redirected to Discord OAuth2 to authorize the bot for their server.
3. Bot joins the server.
4. Channel sync runs.

## Notifications Design

### Event Source

The Go backend already publishes domain events to Redis Streams. The bot runtime subscribes to the stream and translates events into platform messages.

Events that trigger notifications:

- `story.created`
- `story.updated` (status change, assignment change, due date change)
- `comment.created`
- `comment.replied`
- `user.mentioned`

### Notification Routing

For each event:

1. Resolve the workspace from the event payload.
2. Look up active installations for that workspace.
3. Look up team-to-channel assignments and team notification settings.
4. Filter by event type, priority, and team scope.
5. Format the message using Chat SDK JSX components.
6. Post to each assigned channel.

### Message Format

Notifications use Chat SDK JSX, which renders natively per platform:

```tsx
<Card>
  <Section>
    <CardText>**Story created**: {story.title}</CardText>
    <Fields>
      <Field title="Status">{story.status}</Field>
      <Field title="Priority">{story.priority}</Field>
      <Field title="Assignee">{story.assignee || "Unassigned"}</Field>
    </Fields>
  </Section>
  <Actions>
    <Button actionId="mark_started" value={story.id}>Mark Started</Button>
    <Button actionId="mark_completed" value={story.id}>Mark Completed</Button>
    <LinkButton url={storyUrl}>Open in FortyOne</LinkButton>
  </Actions>
</Card>
```

This renders as Block Kit on Slack, Adaptive Card on Teams, interactive buttons on WhatsApp (max 3), and an embed on Discord.

### Filtering

- Team-scoped delivery (only post to channels assigned to the story's team)
- Event-type filtering (per team settings)
- Priority filtering (optional high-priority-only mode)
- DM delivery for direct assignment events (where platform supports it)

## Maya in Messaging Platforms

### How It Works

When a user @mentions the bot or sends a DM:

1. Chat SDK handler fires (`onNewMention` or `onSubscribedMessage`).
2. The runtime resolves the platform user to a FortyOne member via `channel_user_links`.
3. If no link exists, respond with an ephemeral message explaining how to link their account.
4. If linked, construct the AI SDK context (workspace, teams, user, memories, subscription).
5. Call `streamText()` with the same Maya tools and system prompt used in the web UI.
6. Stream the response to the platform via `thread.post(result.fullStream)`.
7. Subscribe to the thread for multi-turn conversation via `thread.subscribe()`.

### Tool Execution

Maya's 80+ tools work unchanged. The tools use workspace-scoped API calls, and the workspace is resolved from the linked user's membership. Tools that navigate the web UI (`navigation`, `theme`) are excluded from the bot runtime's tool set since they are not applicable outside the browser.

### Conversational Story Creation

On platforms without modals (Teams, WhatsApp, Discord), story creation is conversational:

- User: "create a story for fixing the login bug on the mobile app"
- Maya: calls `createStory` tool with title and description inferred from the message
- Maya: "Created story PROJ-142: Fix login bug on mobile app. Priority: Medium, Status: Backlog. Want me to assign it to someone or change the priority?"

On Slack, the `/fortyone create` command opens a native modal with form fields.

### Slack-Specific Features

- **Slash command**: `/fortyone` with subcommands (create, search, help)
- **Global shortcut**: "Create story in FortyOne" opens a modal
- **Message shortcut**: "Create story from message" captures the Slack message text as story description
- **App Home**: summary of assigned stories and recent activity

## Link Unfurling

Supported on Slack only (other platforms use embeds or plain links).

When a FortyOne story URL is pasted in Slack:

1. Slack sends the URL to the bot's unfurl handler.
2. The bot resolves the story via the FortyOne API.
3. Verifies the Slack workspace is connected to the same FortyOne workspace.
4. Publishes an unfurl with compact story data and action buttons.

## Interactive Actions

### Supported Actions

- Status change buttons (mark started, mark completed)
- Assign to me
- Open in FortyOne (link button)
- Add comment (modal on Slack, conversational elsewhere)

### Handling Flow

1. Chat SDK handler fires (`onAction` or `onModalSubmit`).
2. Resolve the platform user to a FortyOne member.
3. Resolve the target story from the action payload.
4. Execute the action through the FortyOne API using the linked user's permissions.
5. Update the original message to reflect the new state.

## Identity Mapping

Platform actions and Maya require mapping platform users to FortyOne members.

### Linking Methods

1. **Email auto-link**: If the platform provides the user's email (Slack with `users:read.email` scope, Teams via Azure AD), attempt automatic matching against FortyOne user emails.
2. **Manual link**: User links their account via FortyOne settings or a bot command (`/fortyone link`).
3. **Phone link**: WhatsApp only. User provides their WhatsApp number in FortyOne settings.

### When No Link Exists

- Do not fail silently.
- Respond with a clear message explaining how to link their account.
- Provide a direct link to the FortyOne settings page for manual linking.
- Actions that require identity (assign to me, Maya queries) are blocked until linked.
- Notifications that reference a specific user include "link your account" prompts.

## UI Plan

### Workspace Settings: Integrations Page

A provider-agnostic integrations page:

- List of available providers with connect/disconnect state
- For each connected provider:
  - Provider name and connected workspace/tenant/phone
  - Installation health status
  - Synced conversations count
  - Enabled capabilities
  - Last validated timestamp
  - Disconnect action

### Team Settings: Messaging Section

Per-team configuration:

- Channel/conversation assignments (across all connected providers)
- Notification rules (which events, which channels)
- Action permissions (story creation, status updates, comments)
- Maya assistant toggle
- Provider-specific overrides if needed

### Story UI

Minimal Slack/messaging metadata on stories:

- Channels where the story was posted (if useful)
- Link to Slack thread (if thread sync is enabled later)

This can be deferred.

## Implementation Sequence

### Phase 0: Common Foundation

- Add database migrations for `channel_*` tables
- Add `apps/server/internal/modules/messaging` with repository and service layers
- Wire workspace-authenticated API routes for provider management
- Add workspace settings UI skeleton for messaging integrations
- Extract shared Maya tool definitions into importable package

### Phase 1: Chat SDK Runtime + Slack Install

- Create `apps/bot` with Chat SDK, Slack adapter, and Redis state
- Implement Slack OAuth install flow via Chat SDK adapter
- Wire install callback to record installation in FortyOne
- Implement channel sync
- Ship workspace status and channel management UI
- Identity linking (email auto-link + manual)

### Phase 2: Maya in Slack

- Wire Maya tools into the bot runtime
- Implement `onNewMention` and DM handlers
- Implement `onSubscribedMessage` for multi-turn conversations
- Implement user resolution (platform user -> FortyOne member -> workspace context)
- Stream AI responses to Slack via `thread.post(result.fullStream)`
- Add `/fortyone` slash command with modal-based story creation

### Phase 3: Slack Notifications + Actions

- Add Redis Streams consumer in the bot runtime
- Implement notification routing (event -> team -> channel)
- Build JSX notification templates
- Implement action button handlers (status change, assign, comment)
- Implement link unfurling
- Add delivery logging and retry handling

### Phase 4: Microsoft Teams

- Add Teams adapter to the bot runtime
- Set up Azure Bot resource and Teams app manifest
- Implement Teams-specific install flow
- Adapt notification templates (Adaptive Cards)
- Maya chat via Teams (post+edit streaming)
- Conversational story creation (no modals on Teams)

### Phase 5: WhatsApp

- Add WhatsApp adapter to the bot runtime
- Implement credential setup flow in workspace settings
- Maya chat via WhatsApp DMs
- DM notifications (within 24h window)
- Set up template messages for outside-window notifications (requires Meta approval)
- Conversational story creation
- Phone-based identity linking

### Phase 6: Discord + Additional Providers

- Add Discord adapter
- Bot commands, embeds, channel notifications
- Evaluate Telegram, Google Chat based on demand

### Phase 7: Advanced Features

- Story search commands across all platforms
- Richer action support (comment modals on Slack, inline editing)
- Per-channel notification filtering
- Thread sync (opt-in, integration-originated threads only, Slack first)
- Workspace/team summaries and personal workload prompts across platforms
- App Home on Slack with story dashboard

## Testing Strategy

### Unit Tests

- User identity resolution per provider
- Notification routing logic (event -> team -> channel)
- JSX notification template rendering per platform
- Install session validation
- Tool set filtering (exclude web-only tools from bot runtime)

### Integration Tests

- Slack OAuth install flow end-to-end
- Slash command -> modal -> story creation flow
- @mention -> Maya response flow
- Domain event -> notification delivery flow
- Action button -> story state change flow
- Link unfurl flow
- Multi-workspace token resolution

### Platform-Specific Tests

- Teams Adaptive Card rendering
- WhatsApp message chunking and 24h window handling
- Discord embed rendering
- Cross-platform notification parity

## Operational Concerns

### Logging

Structured fields:

- `workspace_id`
- `provider`
- `installation_id`
- `provider_team_id`
- `provider_channel_id`
- `team_id`
- `story_id`
- `event_type`
- `request_id`
- `attempt_count`

### Metrics

Track per provider:

- Install success and failure
- Conversation sync counts
- Message delivery success and failure
- Maya request counts and latency
- Interactive action success and failure
- Unfurl counts (Slack)
- Identity link rates

### Retry Strategy

Outbound message delivery should:

- Retry transient API failures with exponential backoff
- Respect platform rate limits (back off on 429)
- Preserve enough context for replay
- Log failed deliveries for manual inspection
- WhatsApp: do not retry outside the 24h window unless using template messages

### Deployment

The bot runtime (`apps/bot`) is a standalone Node.js process:

- Runs independently from `apps/projects` and `apps/server`
- Connects to the same Redis and Postgres instances
- Scales horizontally (Chat SDK state adapters handle distributed locking)
- Health check endpoint for orchestration
- Environment variables per provider (adapter auto-detects from env)

## Risks

### Risk: Platform Capability Divergence

Mitigation: Capability matrix stored per installation. UI and bot runtime check capabilities before offering features. Graceful degradation is a first-class design concern, not an afterthought.

### Risk: Channel Noise

Mitigation: Notifications are opt-in by event type. Channel mapping is team-specific. High-priority-only mode available.

### Risk: Identity Mapping Gaps

Mitigation: Email auto-link for Slack/Teams. Manual link always available. Actions degrade gracefully (show link prompt instead of failing). Maya blocked until identity is confirmed.

### Risk: WhatsApp 24h Window

Mitigation: Template messages for outside-window notifications. Clear UI indicating which notifications may be delayed. Do not promise real-time WhatsApp delivery.

### Risk: Tool Divergence Between Web and Bot

Mitigation: Single tool definition source. Web-only tools (navigation, theme) are explicitly excluded via a filter. All other tools are shared unchanged.

### Risk: Scaling Bot Runtime

Mitigation: Chat SDK state adapters provide distributed locking. Redis-backed state handles multi-instance deployment. Stateless handler design allows horizontal scaling.

### Risk: Provider API Changes

Mitigation: Chat SDK adapters abstract provider APIs. Updates to adapters absorb breaking changes. Pin adapter versions and test before upgrading.

## Recommended Milestones

### Milestone 1: Slack Foundation

- Bot runtime exists with Slack adapter
- Workspace connects Slack
- Channels sync
- Team-to-channel assignments work
- Identity linking works

### Milestone 2: Maya in Slack

- Users can chat with Maya in Slack (mentions, DMs, threads)
- Slash command story creation with modal
- Message shortcut story creation

### Milestone 3: Slack Notifications + Actions

- Story and comment notifications post to assigned channels
- Action buttons work (status change, assign)
- Link unfurling works
- Delivery tracking and retry

### Milestone 4: Microsoft Teams

- Teams adapter connected
- Maya chat works in Teams
- Channel notifications with Adaptive Cards
- Conversational story creation

### Milestone 5: WhatsApp

- WhatsApp adapter connected
- Maya chat via DMs
- DM notifications within 24h window
- Template messages for outside-window

### Milestone 6: Expansion

- Discord adapter
- Cross-platform notification parity
- Advanced features (thread sync, summaries, rich actions)

## Open Questions

- Should a workspace support multiple installations of the same provider (e.g., two Slack workspaces)?
- Should private channel support be in the first Slack release or follow after public channels?
- Which domain events are high enough signal to notify by default vs. requiring explicit opt-in?
- For WhatsApp template messages: who owns template creation and Meta approval workflow?
- How should the bot runtime authenticate against the FortyOne API? Service account? Internal token? Direct database access?
- Should `apps/bot` live in the monorepo or be a separate deployment? (Recommendation: monorepo, separate deploy)
- How much of the Maya system prompt needs to change for messaging contexts (shorter responses, no markdown-heavy formatting)?
- Should we support per-user provider preferences (e.g., "notify me on Slack, not Teams")?

## Final Recommendation

Build a single Chat SDK runtime (`apps/bot`) that serves all messaging platforms through adapters. Reuse Maya's existing AI SDK tools unchanged. Let the Go backend own state, permissions, and events. Let Chat SDK own platform mechanics.

Ship Slack first (richest experience, most demanded). Follow with Teams (enterprise), then WhatsApp (mobile-first teams), then Discord (community/dev teams). Each new provider is an adapter addition + install flow + UI, not a rewrite.

The strongest path: Phase 0-1 (foundation + Slack install), Phase 2 (Maya in Slack — the flagship feature), Phase 3 (notifications + actions), then Phase 4+ (expand to Teams, WhatsApp, Discord).
