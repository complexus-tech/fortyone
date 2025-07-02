# Slack Integration Plan

## Overview

This document outlines the implementation plan for integrating Slack with our project management system. The integration will enable seamless communication and notifications between our platform and Slack workspaces, allowing teams to stay informed about project updates, create and manage stories directly from Slack, and maintain context between conversations and work items.

## Core Capabilities

### 1. Notification System

- Send story updates to designated Slack channels
- Notify team members about assignments and mentions
- Daily/weekly digest summaries
- Custom notification triggers and filters
- @mention notifications for story activity

### 2. Story Management from Slack

- Create stories from Slack messages using slash commands
- Update story status through interactive buttons
- Link Slack threads to stories for context
- Quick story creation from pinned messages
- Convert Slack threads into actionable stories

### 3. Channel Integration

- Map teams to Slack channels
- Workspace-wide announcements
- Project-specific channels for focused discussions
- Automated channel creation for new projects/teams
- Archive/restore channel management

### 4. Slash Commands & Interactions

- `/story create` - Quick story creation
- `/story update` - Update story status/details
- `/story assign` - Assign stories to team members
- `/standup` - Daily standup automation
- `/projects status` - Get project overview

### 5. Interactive Components

- Story cards with action buttons in Slack
- Modal dialogs for detailed story creation/editing
- Approval workflows through Slack
- Status update buttons
- Quick filters and search

### 6. Threading & Context

- Link Slack conversations to stories
- Automatic thread creation for story discussions
- Context preservation between Slack and platform
- Message threading for ongoing story work
- Decision tracking in threads

## Technical Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Slack Workspace                         │
├─────────────────────────────────────────────────────────────┤
│  • Channels                                                │
│  • Slash Commands                                          │
│  • Interactive Components                                  │
│  • Bot Messages                                           │
│  • Events & Webhooks                                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Layer (Go)                           │
├─────────────────────────────────────────────────────────────┤
│  internal/handlers/integrationsgrp/                       │
│  ├── slack.go            # Slack integration endpoints     │
│  ├── slashcommands.go    # Slash command handlers         │
│  ├── interactions.go     # Interactive component handlers │
│  ├── events.go           # Slack event handlers           │
│  └── notifications.go    # Notification handlers          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Business Logic (Go)                        │
├─────────────────────────────────────────────────────────────┤
│  internal/core/integrations/                              │
│  ├── slack.go            # Slack-specific logic           │
│  ├── notifications.go    # Notification processing        │
│  ├── commands.go         # Command processing             │
│  └── threading.go        # Thread management              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Slack Service (Go)                         │
├─────────────────────────────────────────────────────────────┤
│  pkg/slack/                                               │
│  ├── service.go          # Main Slack service             │
│  ├── auth.go             # OAuth flow handling            │
│  ├── messages.go         # Message operations             │
│  ├── channels.go         # Channel operations             │
│  ├── users.go            # User operations                │
│  ├── commands.go         # Slash command handling         │
│  ├── interactions.go     # Interactive components         │
│  └── events.go           # Event processing               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Data Layer (PostgreSQL)                   │
├─────────────────────────────────────────────────────────────┤
│  • workspace_slack_integrations                           │
│  • slack_channels                                         │
│  • slack_users                                            │
│  • story_slack_threads                                    │
│  • slack_notification_preferences                         │
│  • slack_command_logs                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Slack API                               │
├─────────────────────────────────────────────────────────────┤
│  • Web API                                                │
│  • Events API                                             │
│  • Interactive Components                                 │
│  • Slash Commands                                         │
│  • OAuth                                                  │
└─────────────────────────────────────────────────────────────┘
```

## Database Schema

### Core Tables

```sql
-- Workspace Slack integrations
CREATE TABLE workspace_slack_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    slack_workspace_id TEXT NOT NULL,
    slack_workspace_name TEXT NOT NULL,
    bot_user_id TEXT NOT NULL,
    bot_access_token TEXT NOT NULL, -- encrypted
    user_access_token TEXT, -- encrypted (optional)
    installing_user_id UUID REFERENCES users(user_id),
    scopes TEXT[] NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(workspace_id),
    UNIQUE(slack_workspace_id)
);

-- Slack channels
CREATE TABLE slack_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES workspace_slack_integrations(id) ON DELETE CASCADE,
    slack_channel_id TEXT NOT NULL,
    name TEXT NOT NULL,
    is_private BOOLEAN DEFAULT false,
    is_archived BOOLEAN DEFAULT false,
    purpose TEXT,
    topic TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(integration_id, slack_channel_id)
);

-- Team-channel mappings
CREATE TABLE team_slack_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
    slack_channel_id UUID NOT NULL REFERENCES slack_channels(id) ON DELETE CASCADE,
    notification_types TEXT[] DEFAULT '{}', -- story_created, story_updated, etc.
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(team_id, slack_channel_id)
);

-- Slack users
CREATE TABLE slack_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES workspace_slack_integrations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
    slack_user_id TEXT NOT NULL,
    slack_username TEXT NOT NULL,
    display_name TEXT,
    real_name TEXT,
    email TEXT,
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(integration_id, slack_user_id)
);

-- Story-Slack thread links
CREATE TABLE story_slack_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(story_id) ON DELETE CASCADE,
    slack_channel_id UUID NOT NULL REFERENCES slack_channels(id) ON DELETE CASCADE,
    slack_thread_ts TEXT NOT NULL,
    slack_message_ts TEXT NOT NULL,
    thread_type TEXT NOT NULL CHECK (thread_type IN ('discussion', 'notification', 'creation')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(story_id, slack_channel_id, slack_thread_ts)
);

-- Slack notification preferences
CREATE TABLE slack_notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(team_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
    notification_type TEXT NOT NULL,
    slack_channel_id UUID REFERENCES slack_channels(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT true,
    mention_users BOOLEAN DEFAULT false,
    thread_replies BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Either workspace-wide, team-specific, or user-specific
    CHECK (
        (team_id IS NULL AND user_id IS NULL) OR  -- workspace-wide
        (team_id IS NOT NULL AND user_id IS NULL) OR  -- team-specific
        (team_id IS NULL AND user_id IS NOT NULL)     -- user-specific
    )
);

-- Slash command logs
CREATE TABLE slack_command_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES workspace_slack_integrations(id) ON DELETE CASCADE,
    slack_user_id TEXT NOT NULL,
    user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
    command TEXT NOT NULL,
    parameters TEXT,
    slack_channel_id TEXT NOT NULL,
    response_type TEXT, -- ephemeral, in_channel
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    execution_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Slack event queue (for async processing)
CREATE TABLE slack_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES workspace_slack_integrations(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    event_ts TEXT NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMP,
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(integration_id, event_ts)
);
```

### Indexes

```sql
-- Performance indexes
CREATE INDEX idx_workspace_slack_integrations_workspace ON workspace_slack_integrations(workspace_id);
CREATE INDEX idx_slack_channels_integration ON slack_channels(integration_id);
CREATE INDEX idx_team_slack_channels_team ON team_slack_channels(team_id);
CREATE INDEX idx_team_slack_channels_channel ON team_slack_channels(slack_channel_id);
CREATE INDEX idx_slack_users_integration_user ON slack_users(integration_id, user_id);
CREATE INDEX idx_story_slack_threads_story ON story_slack_threads(story_id);
CREATE INDEX idx_slack_notification_preferences_workspace ON slack_notification_preferences(workspace_id);
CREATE INDEX idx_slack_command_logs_integration_created ON slack_command_logs(integration_id, created_at);
CREATE INDEX idx_slack_events_processed ON slack_events(processed, created_at);
```

## API Endpoints

### Integration Setup

```
POST   /workspaces/{workspaceId}/integrations/slack/connect
       # Initiate Slack OAuth flow
       Request: { redirectUrl: string }
       Response: { authUrl: string, state: string }

POST   /workspaces/{workspaceId}/integrations/slack/callback
       # Handle OAuth callback
       Request: { code: string, state: string }
       Response: { integration: SlackIntegration, channels: SlackChannel[] }

DELETE /workspaces/{workspaceId}/integrations/slack
       # Disconnect Slack integration
       Response: 204 No Content

GET    /workspaces/{workspaceId}/integrations/slack
       # Get integration status and configuration
       Response: {
         integration: SlackIntegration | null,
         channels: SlackChannel[],
         users: SlackUser[]
       }
```

### Channel Management

```
GET    /workspaces/{workspaceId}/integrations/slack/channels
       # List Slack channels
       Query: { refresh?: boolean, includePrivate?: boolean }
       Response: { channels: SlackChannel[] }

POST   /workspaces/{workspaceId}/integrations/slack/channels/sync
       # Sync channels from Slack
       Response: { synced: number, channels: SlackChannel[] }

PUT    /teams/{teamId}/slack/channels
       # Configure team-channel mappings
       Request: {
         channelMappings: {
           channelId: UUID,
           notificationTypes: string[],
           isPrimary: boolean
         }[]
       }
       Response: { mappings: TeamSlackChannel[] }

DELETE /teams/{teamId}/slack/channels/{channelId}
       # Remove team-channel mapping
       Response: 204 No Content
```

### Notification Preferences

```
GET    /workspaces/{workspaceId}/integrations/slack/notifications
       # Get notification preferences
       Query: { teamId?: UUID, userId?: UUID }
       Response: { preferences: SlackNotificationPreference[] }

PUT    /workspaces/{workspaceId}/integrations/slack/notifications
       # Update notification preferences
       Request: { preferences: SlackNotificationPreference[] }
       Response: { preferences: SlackNotificationPreference[] }

POST   /workspaces/{workspaceId}/integrations/slack/notifications/test
       # Send test notification
       Request: { channelId: UUID, notificationType: string }
       Response: { success: boolean, messageTs?: string }
```

### Slash Commands & Interactions

```
POST   /integrations/slack/commands
       # Handle all slash commands
       Headers: X-Slack-Signature, X-Slack-Request-Timestamp
       Request: Slack command payload
       Response: Slack response format

POST   /integrations/slack/interactions
       # Handle interactive components
       Headers: X-Slack-Signature, X-Slack-Request-Timestamp
       Request: Slack interaction payload
       Response: Slack response format

POST   /integrations/slack/events
       # Handle Slack events
       Headers: X-Slack-Signature, X-Slack-Request-Timestamp
       Request: Slack event payload
       Response: { challenge?: string } or 200 OK
```

### Story Integration

```
POST   /stories/{storyId}/slack/notify
       # Send story notification to Slack
       Request: {
         channelIds?: UUID[],
         message?: string,
         mentionUsers?: boolean,
         createThread?: boolean
       }
       Response: { notifications: SlackNotification[] }

POST   /stories/{storyId}/slack/thread
       # Create dedicated Slack thread for story
       Request: { channelId: UUID }
       Response: { thread: StorySlackThread }

GET    /stories/{storyId}/slack/threads
       # Get Slack threads for story
       Response: { threads: StorySlackThread[] }
```

### User Management

```
GET    /workspaces/{workspaceId}/integrations/slack/users
       # List Slack users and mappings
       Response: { users: SlackUser[] }

POST   /workspaces/{workspaceId}/integrations/slack/users/sync
       # Sync users from Slack
       Response: { synced: number, users: SlackUser[] }

PUT    /users/{userId}/slack/link
       # Link user to Slack account
       Request: { slackUserId: string }
       Response: { user: SlackUser }

DELETE /users/{userId}/slack/link
       # Unlink user from Slack account
       Response: 204 No Content
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)

**Goals:** Basic Slack OAuth and channel access

**Tasks:**

1. Create Slack OAuth service (`pkg/slack/auth.go`)
2. Implement basic Slack API client (`pkg/slack/service.go`)
3. Create database tables and migrations
4. Implement core integration models and services
5. Add OAuth flow endpoints
6. Basic channel listing and user sync

**Deliverables:**

- Workspaces can connect to Slack
- List and sync Slack channels
- Basic user mapping between platforms
- Channel management UI

### Phase 2: Notification System (Week 3-4)

**Goals:** Send notifications from platform to Slack

**Tasks:**

1. Implement message operations (`pkg/slack/messages.go`)
2. Create notification preference system
3. Integrate with existing event system
4. Add team-channel mapping functionality
5. Implement notification templates
6. Create notification testing tools

**Deliverables:**

- Story updates posted to Slack channels
- Configurable notification preferences
- Team-specific channel assignments
- Rich message formatting with story details

### Phase 3: Slash Commands (Week 5-6)

**Goals:** Basic story management from Slack

**Tasks:**

1. Implement slash command handling (`pkg/slack/commands.go`)
2. Create command processing system
3. Add story creation from Slack
4. Implement status update commands
5. Add user authentication and authorization
6. Create command help and documentation

**Deliverables:**

- `/story create` command for quick story creation
- `/story update` for status changes
- `/story assign` for assignments
- User authentication in Slack context

### Phase 4: Interactive Components (Week 7-8)

**Goals:** Rich interactive story management

**Tasks:**

1. Implement interactive components (`pkg/slack/interactions.go`)
2. Create modal dialogs for detailed forms
3. Add button-based status updates
4. Implement approval workflows
5. Create story cards with actions
6. Add search and filtering interfaces

**Deliverables:**

- Interactive story cards in Slack
- Modal forms for detailed story creation/editing
- Action buttons for quick updates
- Rich story browsing in Slack

### Phase 5: Threading & Context (Week 9-10)

**Goals:** Link conversations to stories and maintain context

**Tasks:**

1. Implement thread management
2. Create story-thread linking system
3. Add conversation-to-story conversion
4. Implement context preservation
5. Create thread-based discussions
6. Add conversation search and history

**Deliverables:**

- Slack threads linked to stories
- Convert conversations to actionable stories
- Context preservation between platforms
- Thread-based story discussions

### Phase 6: Advanced Features (Week 11-12)

**Goals:** Automation, reporting, and optimization

**Tasks:**

1. Implement daily standup automation
2. Create digest and summary reports
3. Add advanced notification rules
4. Implement message parsing for auto-linking
5. Create analytics and usage reporting
6. Performance optimization and monitoring

**Deliverables:**

- Automated daily standups
- Weekly/monthly digest reports
- Smart message parsing and linking
- Comprehensive analytics dashboard
- Production-ready performance

## User Workflows

### Workflow 1: Initial Setup

```
1. Workspace Admin goes to Settings → Integrations
2. Clicks "Connect Slack"
3. Completes OAuth flow and grants permissions
4. System syncs channels and users
5. Admin maps teams to relevant Slack channels
6. Configures notification preferences
7. Team members link their Slack accounts
```

### Workflow 2: Story Notifications

```
1. Developer creates/updates a story
2. System checks notification preferences
3. Formatted notification sent to team's Slack channel
4. Team members receive @mention notifications
5. Discussion happens in Slack thread
6. Updates flow back to story activity feed
```

### Workflow 3: Quick Story Creation

```
1. Team member types `/story create Bug in login form`
2. Slack shows modal with story details form
3. User fills in description, assigns team/member
4. Story created in platform
5. Confirmation sent to Slack with story link
6. Optional: Create dedicated thread for discussion
```

### Workflow 4: Status Updates

```
1. Developer working on story gets Slack notification
2. Clicks "Mark In Progress" button in message
3. Story status updates in platform
4. Team receives update notification
5. Activity logged in story timeline
```

### Workflow 5: Daily Standup

```
1. `/standup` command triggered (manually or scheduled)
2. System gathers team member stories and updates
3. Formatted standup report posted to team channel
4. Team members can respond with additional updates
5. Standup summary logged for historical reference
```

### Workflow 6: Issue Triage

```
1. Customer reports issue in support channel
2. Support team uses `/story create` to convert to story
3. Story automatically assigned to appropriate team
4. Notification sent to team's development channel
5. Developer picks up story and starts work
6. Updates flow back to original support thread
```

## Slack App Configuration

### Required Scopes

**Bot Token Scopes:**

- `chat:write` - Send messages as bot
- `chat:write.public` - Send messages to public channels
- `channels:read` - Access public channel information
- `groups:read` - Access private channel information (if needed)
- `users:read` - Access user information
- `users:read.email` - Access user email addresses
- `commands` - Handle slash commands
- `im:write` - Send direct messages
- `reactions:write` - Add reactions to messages

**User Token Scopes (optional):**

- `identity.basic` - Access user identity
- `identity.email` - Access user email

### Events Subscription

**Bot Events:**

- `message.channels` - Messages in channels
- `message.groups` - Messages in private channels
- `message.im` - Direct messages
- `app_mention` - When bot is mentioned
- `team_join` - New team members
- `channel_created` - New channels
- `channel_rename` - Channel renames

### Interactive Components

- **Message Buttons**: Quick actions on story notifications
- **Modal Dialogs**: Detailed forms for story creation/editing
- **Select Menus**: Team/assignee selection
- **Date Pickers**: Due dates and scheduling

### Slash Commands

```
/story - Main story management command
  ├── create [title] - Create new story
  ├── update [id] [status] - Update story status
  ├── assign [id] [@user] - Assign story
  ├── search [query] - Search stories
  └── help - Show command help

/standup - Daily standup management
  ├── start - Start standup session
  ├── status - Get current status
  └── summary - Show team summary

/projects - Project overview commands
  ├── status [team] - Show project status
  ├── deadlines - Show upcoming deadlines
  └── metrics - Show team metrics
```

## Security Considerations

### OAuth Security

- Validate state parameter in OAuth flow
- Secure token storage with encryption
- Implement proper scope validation
- Regular token refresh and validation
- Audit OAuth permissions regularly

### Webhook Security

- Verify Slack request signatures
- Validate request timestamps (prevent replay attacks)
- Rate limiting on webhook endpoints
- Sanitize and validate all inputs
- Log security-related events

### Data Protection

- Encrypt Slack tokens at rest
- Implement proper access controls
- Don't store sensitive message content
- Comply with data retention policies
- Regular security audits

### Command Security

- Authenticate users for sensitive commands
- Validate user permissions for actions
- Rate limit command usage
- Log all command executions
- Prevent command injection attacks

## Monitoring & Observability

### Metrics to Track

- Slack API rate limit usage
- Message delivery success rates
- Command usage and performance
- User engagement with notifications
- Integration health and uptime

### Logging

- All Slack API calls and responses
- Command executions and results
- Notification deliveries
- Authentication events
- Error conditions and failures

### Alerting

- Slack API rate limits approaching
- Command processing failures
- Authentication failures
- Unusual usage patterns
- Integration disconnections

## Testing Strategy

### Unit Tests

- Slack service operations
- Command processing logic
- Message formatting and templates
- Authentication and authorization

### Integration Tests

- Slack API interactions
- OAuth flow simulation
- Webhook processing
- Database operations

### E2E Tests

- Complete user workflows
- Command execution flows
- Notification delivery
- Interactive component handling

## Success Metrics

### Technical Metrics

- 99.9% notification delivery success rate
- <500ms average command response time
- <10% Slack API rate limit usage
- Zero message delivery failures

### User Adoption Metrics

- 70% of teams connect Slack channels
- 80% of notifications delivered to Slack
- 50% of stories created via Slack commands
- 90% user satisfaction with integration

### Business Metrics

- 30% reduction in context switching
- 25% faster issue resolution
- 40% improvement in team communication
- Better project visibility and transparency

This Slack integration will enhance team collaboration by bringing project management directly into the communication flow, reducing context switching and ensuring everyone stays informed about project progress.
