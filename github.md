# GitHub Integration Plan

## Overview

This document outlines the implementation plan for integrating GitHub with our project management system. The integration will provide seamless connectivity between our stories/tasks and GitHub repositories, enabling developers to manage their workflow entirely within our platform while maintaining full GitHub functionality.

## Core Capabilities

### 1. Repository Access

- Read repositories the user has access to
- Workspace-level repository connections
- Team-scoped repository assignments
- Permission-based access control

### 2. Branch Information

- Create branches from stories
- Track branch activity and commits
- Monitor branch lifecycle (created → active → merged/deleted)
- Auto-sync branch status with story status

### 3. Pull Requests

- Create PRs from stories with pre-filled templates
- Sync PR status with story status
- Track PR lifecycle (draft → ready → approved → merged)
- Link PRs to stories for full traceability

### 4. Issues Sync

- Import GitHub issues as stories
- Export stories as GitHub issues
- Bi-directional status synchronization
- Comment synchronization

### 5. Webhooks

- Real-time notifications for PR events
- Branch creation/deletion events
- Issue updates and comments
- Commit notifications

### 6. Commit Information

- Link commits to stories via commit message parsing
- Track commit activity on story timeline
- Progress estimation based on commit frequency
- Developer activity tracking

## Technical Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React)                        │
├─────────────────────────────────────────────────────────────┤
│  • Story GitHub Integration UI                             │
│  • Repository Management                                   │
│  • Branch/PR Creation Forms                               │
│  • GitHub Activity Timeline                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Layer (Go)                           │
├─────────────────────────────────────────────────────────────┤
│  internal/handlers/integrationsgrp/                       │
│  ├── github.go           # GitHub integration endpoints    │
│  ├── webhooks.go         # Webhook handlers               │
│  ├── models.go           # API models                     │
│  └── routes.go           # Route definitions              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Business Logic (Go)                        │
├─────────────────────────────────────────────────────────────┤
│  internal/core/integrations/                              │
│  ├── integrations.go     # Integration service            │
│  ├── github.go           # GitHub-specific logic          │
│  ├── models.go           # Core models                    │
│  └── webhooks.go         # Webhook processing             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                GitHub Service (Go)                         │
├─────────────────────────────────────────────────────────────┤
│  pkg/github/                                              │
│  ├── service.go          # Main GitHub service            │
│  ├── auth.go             # OAuth flow handling            │
│  ├── repositories.go     # Repository operations          │
│  ├── branches.go         # Branch operations              │
│  ├── pullrequests.go     # PR operations                 │
│  ├── issues.go           # Issue operations               │
│  ├── webhooks.go         # Webhook management             │
│  └── commits.go          # Commit operations              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Data Layer (PostgreSQL)                   │
├─────────────────────────────────────────────────────────────┤
│  • user_github_integrations                               │
│  • github_repositories                                    │
│  • story_github_links                                     │
│  • github_webhook_events                                  │
│  • github_automation_preferences                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   GitHub API                               │
├─────────────────────────────────────────────────────────────┤
│  • REST API v4                                           │
│  • GraphQL API v4                                        │
│  • Webhooks                                              │
│  • OAuth Apps                                            │
└─────────────────────────────────────────────────────────────┘
```

## Database Schema

### Core Tables

```sql
-- User GitHub integrations (OAuth connections)
CREATE TABLE user_github_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    github_user_id TEXT NOT NULL,
    github_username TEXT NOT NULL,
    github_email TEXT,
    access_token TEXT NOT NULL, -- encrypted
    refresh_token TEXT, -- encrypted
    token_expires_at TIMESTAMP,
    scopes TEXT[] NOT NULL,
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, workspace_id)
);

-- Connected repositories
CREATE TABLE github_repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES user_github_integrations(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    github_repo_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    full_name TEXT NOT NULL,
    description TEXT,
    private BOOLEAN NOT NULL DEFAULT false,
    default_branch TEXT NOT NULL DEFAULT 'main',
    clone_url TEXT NOT NULL,
    ssh_url TEXT NOT NULL,
    webhook_id BIGINT, -- GitHub webhook ID
    webhook_secret TEXT, -- encrypted
    is_active BOOLEAN DEFAULT true,
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(workspace_id, github_repo_id)
);

-- Repository team assignments
CREATE TABLE repository_team_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES github_repositories(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(repository_id, team_id)
);

-- Story-GitHub links (branches, PRs, issues)
CREATE TABLE story_github_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(story_id) ON DELETE CASCADE,
    repository_id UUID NOT NULL REFERENCES github_repositories(id) ON DELETE CASCADE,
    link_type TEXT NOT NULL CHECK (link_type IN ('branch', 'pull_request', 'issue')),
    github_id TEXT NOT NULL, -- branch name, PR number, or issue number
    github_url TEXT NOT NULL,
    title TEXT,
    state TEXT, -- branch: active/merged/deleted, PR: draft/open/closed, issue: open/closed
    metadata JSONB, -- additional GitHub-specific data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(story_id, repository_id, link_type, github_id)
);

-- GitHub automation preferences
CREATE TABLE github_automation_preferences (
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    auto_create_branch BOOLEAN DEFAULT false,
    auto_create_pr BOOLEAN DEFAULT false,
    auto_move_story_on_pr_merge BOOLEAN DEFAULT true,
    auto_assign_pr_reviewer BOOLEAN DEFAULT false,
    branch_naming_pattern TEXT DEFAULT 'story/{story-id}-{title}',
    pr_template TEXT,
    default_reviewer_ids UUID[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY(user_id, workspace_id)
);

-- Webhook event log
CREATE TABLE github_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID REFERENCES github_repositories(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    github_delivery_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(github_delivery_id)
);

-- Commit tracking
CREATE TABLE github_commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES github_repositories(id) ON DELETE CASCADE,
    story_id UUID REFERENCES stories(story_id) ON DELETE SET NULL,
    sha TEXT NOT NULL,
    message TEXT NOT NULL,
    author_username TEXT NOT NULL,
    author_email TEXT,
    committed_at TIMESTAMP NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(repository_id, sha)
);
```

### Indexes

```sql
-- Performance indexes
CREATE INDEX idx_user_github_integrations_user_workspace ON user_github_integrations(user_id, workspace_id);
CREATE INDEX idx_github_repositories_workspace ON github_repositories(workspace_id);
CREATE INDEX idx_github_repositories_integration ON github_repositories(integration_id);
CREATE INDEX idx_story_github_links_story ON story_github_links(story_id);
CREATE INDEX idx_story_github_links_repository ON story_github_links(repository_id);
CREATE INDEX idx_github_webhook_events_processed ON github_webhook_events(processed, created_at);
CREATE INDEX idx_github_commits_story ON github_commits(story_id);
CREATE INDEX idx_github_commits_repository_committed ON github_commits(repository_id, committed_at);
```

## API Endpoints

### Authentication & Setup

```
POST   /workspaces/{workspaceId}/integrations/github/connect
       # Initiate GitHub OAuth flow
       Request: { redirectUrl: string }
       Response: { authUrl: string, state: string }

POST   /workspaces/{workspaceId}/integrations/github/callback
       # Handle OAuth callback
       Request: { code: string, state: string }
       Response: { integration: GitHubIntegration, repositories: GitHubRepository[] }

DELETE /workspaces/{workspaceId}/integrations/github
       # Disconnect GitHub integration
       Response: 204 No Content

GET    /workspaces/{workspaceId}/integrations/github
       # Get integration status and connected repositories
       Response: {
         integration: GitHubIntegration | null,
         repositories: GitHubRepository[],
         preferences: GitHubAutomationPreferences
       }
```

### Repository Management

```
GET    /workspaces/{workspaceId}/integrations/github/repositories
       # List available repositories from GitHub
       Query: { refresh?: boolean }
       Response: { repositories: GitHubRepository[] }

POST   /workspaces/{workspaceId}/integrations/github/repositories/{repoId}/enable
       # Enable repository for workspace
       Request: { teamIds?: UUID[] }
       Response: { repository: GitHubRepository }

DELETE /workspaces/{workspaceId}/integrations/github/repositories/{repoId}
       # Disable repository
       Response: 204 No Content

PUT    /workspaces/{workspaceId}/integrations/github/repositories/{repoId}/teams
       # Update team assignments
       Request: { teamIds: UUID[] }
       Response: { repository: GitHubRepository }
```

### Story Integration

```
POST   /stories/{storyId}/github/branch
       # Create GitHub branch from story
       Request: {
         repositoryId: UUID,
         branchName?: string,
         baseBranch?: string
       }
       Response: { branch: GitHubBranch, link: StoryGitHubLink }

POST   /stories/{storyId}/github/pull-request
       # Create GitHub PR from story
       Request: {
         repositoryId: UUID,
         branchName: string,
         title?: string,
         description?: string,
         reviewers?: string[]
       }
       Response: { pullRequest: GitHubPullRequest, link: StoryGitHubLink }

GET    /stories/{storyId}/github/links
       # Get all GitHub links for story
       Response: { links: StoryGitHubLink[] }

DELETE /stories/{storyId}/github/links/{linkId}
       # Remove GitHub link
       Response: 204 No Content

POST   /stories/{storyId}/github/sync
       # Manually sync story with GitHub
       Response: { links: StoryGitHubLink[] }
```

### Issue Sync

```
POST   /workspaces/{workspaceId}/integrations/github/sync-issues
       # Import GitHub issues as stories
       Request: {
         repositoryId: UUID,
         teamId?: UUID,
         labels?: string[],
         state?: 'open' | 'closed' | 'all'
       }
       Response: { imported: number, stories: Story[] }

POST   /stories/{storyId}/github/export-issue
       # Export story as GitHub issue
       Request: {
         repositoryId: UUID,
         labels?: string[]
       }
       Response: { issue: GitHubIssue, link: StoryGitHubLink }
```

### Automation & Preferences

```
GET    /workspaces/{workspaceId}/integrations/github/preferences
       # Get GitHub automation preferences
       Response: { preferences: GitHubAutomationPreferences }

PUT    /workspaces/{workspaceId}/integrations/github/preferences
       # Update GitHub automation preferences
       Request: { preferences: Partial<GitHubAutomationPreferences> }
       Response: { preferences: GitHubAutomationPreferences }
```

### Webhooks

```
POST   /integrations/github/webhook
       # GitHub webhook endpoint
       Headers: X-GitHub-Event, X-GitHub-Delivery, X-Hub-Signature-256
       Request: GitHub webhook payload
       Response: 200 OK

GET    /workspaces/{workspaceId}/integrations/github/webhook-events
       # List recent webhook events (admin only)
       Query: { limit?: number, offset?: number }
       Response: { events: GitHubWebhookEvent[], total: number }
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)

**Goals:** Basic GitHub OAuth and repository access

**Tasks:**

1. Create GitHub OAuth service (`pkg/github/auth.go`)
2. Implement basic GitHub API client (`pkg/github/service.go`)
3. Create database tables and migrations
4. Implement core integration models and services
5. Add OAuth flow endpoints
6. Basic repository listing and connection

**Deliverables:**

- Users can connect GitHub accounts
- List and connect repositories to workspaces
- Basic repository management UI

### Phase 2: Story-Branch Integration (Week 3-4)

**Goals:** Create and track branches from stories

**Tasks:**

1. Implement branch operations (`pkg/github/branches.go`)
2. Create story-GitHub link models and services
3. Add branch creation endpoints
4. Implement branch status tracking
5. Add branch creation UI to story pages
6. Create activity feed integration

**Deliverables:**

- Create GitHub branches from stories
- Track branch status and activity
- Story activity feed shows branch events
- Auto-sync branch status with story status

### Phase 3: Pull Request Integration (Week 5-6)

**Goals:** Create and track PRs from stories

**Tasks:**

1. Implement PR operations (`pkg/github/pullrequests.go`)
2. Add PR creation endpoints
3. Implement PR status tracking and sync
4. Create PR template system
5. Add PR creation UI
6. Implement PR status → story status automation

**Deliverables:**

- Create GitHub PRs from stories with templates
- Track PR status and sync with story status
- PR activity appears in story timeline
- Automated story status updates based on PR events

### Phase 4: Webhook System (Week 7-8)

**Goals:** Real-time GitHub event processing

**Tasks:**

1. Implement webhook management (`pkg/github/webhooks.go`)
2. Create webhook event processing system
3. Add webhook endpoints and verification
4. Implement event-driven story updates
5. Add webhook event logging and monitoring
6. Create error handling and retry logic

**Deliverables:**

- Auto-install webhooks on repository connection
- Real-time story updates from GitHub events
- Webhook event monitoring and logging
- Robust error handling and retry mechanisms

### Phase 5: Advanced Features (Week 9-10)

**Goals:** Issue sync, commit tracking, automation preferences

**Tasks:**

1. Implement issue sync operations (`pkg/github/issues.go`)
2. Add commit tracking (`pkg/github/commits.go`)
3. Create automation preferences system
4. Implement commit message parsing for story links
5. Add advanced automation rules
6. Create admin monitoring dashboard

**Deliverables:**

- Bi-directional issue-story sync
- Commit tracking and story linking
- Advanced automation preferences
- Comprehensive admin monitoring

### Phase 6: Polish & Optimization (Week 11-12)

**Goals:** Performance, UX improvements, monitoring

**Tasks:**

1. Performance optimization and caching
2. Enhanced error handling and user feedback
3. Comprehensive testing suite
4. Documentation and user guides
5. Monitoring and alerting setup
6. Security audit and hardening

**Deliverables:**

- Production-ready performance
- Complete test coverage
- User documentation
- Production monitoring
- Security validation

## User Workflows

### Workflow 1: Initial Setup

```
1. Workspace Admin goes to Settings → Integrations
2. Clicks "Connect GitHub"
3. Completes OAuth flow
4. Selects repositories to connect
5. Assigns repositories to teams (optional)
6. Configures automation preferences
```

### Workflow 2: Story Development

```
1. Developer assigned to story
2. Clicks "Create Branch" on story
3. System creates GitHub branch with naming pattern
4. Story status auto-updates to "In Progress"
5. Developer works on branch, commits code
6. Commits appear in story activity feed
7. Developer clicks "Create Pull Request"
8. PR created with story context and template
9. Story status updates to "In Review"
10. PR approved and merged
11. Story status auto-updates to "Done"
```

### Workflow 3: Issue Import

```
1. Team Lead goes to workspace settings
2. Clicks "Import GitHub Issues"
3. Selects repository and filters
4. Reviews issue preview
5. Confirms import
6. GitHub issues become stories in selected team
7. Ongoing sync keeps them in sync
```

### Workflow 4: External Collaboration

```
1. Internal story needs external contractor work
2. Click "Export to GitHub Issue"
3. GitHub issue created for external visibility
4. External contributor works on GitHub
5. Issue updates sync back to internal story
6. Team tracks progress in familiar interface
```

## Security Considerations

### OAuth Security

- Use state parameter to prevent CSRF attacks
- Validate redirect URLs
- Secure token storage with encryption at rest
- Implement token refresh logic
- Scope limitation (request minimal required permissions)

### Webhook Security

- Verify webhook signatures using HMAC-SHA256
- Use random webhook secrets per repository
- Implement rate limiting on webhook endpoints
- Log all webhook events for audit trail
- Validate webhook payload structure

### Data Protection

- Encrypt GitHub tokens using AES-256
- Use secure key management (environment variables/vault)
- Implement proper access controls
- Regular security audits
- Comply with data retention policies

### API Security

- Rate limiting for GitHub API calls
- Proper error handling (don't leak sensitive info)
- Validate all inputs
- Implement proper authentication checks
- Use HTTPS for all communications

## Monitoring & Observability

### Metrics to Track

- GitHub API rate limit usage
- Webhook processing success/failure rates
- Integration connection health
- Story-GitHub sync accuracy
- User adoption rates

### Logging

- All GitHub API calls
- Webhook events and processing
- OAuth flows and failures
- Integration errors and warnings
- Performance metrics

### Alerting

- GitHub API rate limit approaching
- Webhook processing failures
- Integration connection failures
- Unusual error rates
- Security-related events

## Testing Strategy

### Unit Tests

- GitHub service operations
- Webhook event processing
- Model validation and business logic
- API endpoint handlers

### Integration Tests

- GitHub API interactions (using test repositories)
- Webhook flow end-to-end
- OAuth flow simulation
- Database operations

### E2E Tests

- Complete user workflows
- Story-GitHub sync scenarios
- Error handling and recovery
- Performance under load

## Configuration

### Environment Variables

```bash
# GitHub App Configuration
GITHUB_CLIENT_ID=your_github_app_client_id
GITHUB_CLIENT_SECRET=your_github_app_client_secret
GITHUB_WEBHOOK_SECRET=your_webhook_secret

# Encryption
GITHUB_TOKEN_ENCRYPTION_KEY=your_encryption_key

# API Configuration
GITHUB_API_BASE_URL=https://api.github.com
GITHUB_API_TIMEOUT=30s
GITHUB_RATE_LIMIT_BUFFER=100 # requests to keep in reserve

# Webhook Configuration
GITHUB_WEBHOOK_BASE_URL=https://yourapp.com/integrations/github/webhook
```

### GitHub App Setup

1. Create GitHub App with required permissions:

   - Repository permissions: Contents (read/write), Metadata (read), Pull requests (read/write), Issues (read/write)
   - Organization permissions: Members (read)
   - User permissions: Email addresses (read)

2. Configure webhook URL and events:

   - Push events
   - Pull request events
   - Issues events
   - Branch creation/deletion events

3. Generate and securely store client secret and webhook secret

## Success Metrics

### Technical Metrics

- 99.9% webhook processing success rate
- <100ms average API response time
- <5% GitHub API rate limit usage
- Zero data loss in sync operations

### User Adoption Metrics

- 80% of active developers connect GitHub
- 60% of stories have associated GitHub activity
- 40% reduction in context switching between tools
- 90% user satisfaction with integration

### Business Metrics

- 20% improvement in development velocity
- 30% better visibility into development progress
- 50% reduction in manual status updates
- Improved compliance and audit trails

This integration will transform how development teams manage their workflow by creating a seamless bridge between project management and code development, maintaining the benefits of both systems while eliminating the overhead of context switching.
