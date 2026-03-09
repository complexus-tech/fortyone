# GitHub Integration Plan

## Document Status

- Status: Draft implementation plan
- Scope: New GitHub integration for the `fortyone` monorepo
- Primary app surfaces:
  - `apps/projects` for the product UI
  - `apps/server` for API, sync processing, persistence, and workers
- Goal: Production-grade GitHub App integration with inbound and outbound sync for the project management platform

## Problem Statement

The product needs a serious GitHub integration that supports more than simple linking. Teams should be able to connect repositories to their workspace and teams, ingest GitHub activity into FortyOne, and push relevant FortyOne changes back to GitHub. The integration must support:

- Repository connection at the workspace level
- Team-level repository assignment and automation rules
- Two-way sync for issues
- PR, branch, commit, and comment visibility on stories
- PR lifecycle automation on story state
- Reliable webhook ingestion, idempotent processing, and retryable outbound sync

This should replace the earlier partial GitHub work with a clearer architecture. Existing GitHub-shaped migrations in `apps/server/internal/migrations` should be treated as reference material, not a finished design.

## Current Repository Reality

The relevant extension points already exist:

- Workspace settings has a commented `Integrations` nav slot in `apps/projects/src/components/layouts/settings.tsx`
- Team settings already has an `Automations` area in `apps/projects/src/modules/settings/workspace/teams/management`
- Team settings API already exists in `apps/server/internal/modules/teamsettings/http/routes.go`
- Story create and update already publish domain events in `apps/server/internal/modules/stories/service/stories.go`
- Event types already flow through Redis Streams in `apps/server/pkg/events`, `apps/server/pkg/publisher`, and `apps/server/pkg/consumer`

There are also partial GitHub migrations already in the repo:

- `000034_github_installations`
- `000035_github_repositories`
- `000036_github_automation_preferences`
- `000037_github_webhook_events`
- `000038_github_commits`
- `000039_repository_team_assignments`
- `000040_team_github_automation_settings`
- `000041_story_github_links`

These are useful signals about intended domain concepts, but they should be revisited before implementation because they do not yet represent a complete sync model.

## Guiding Principles

1. Use a GitHub App, not PATs and not a plain OAuth app.
2. Model GitHub as a dedicated integration module, not as a few fields hidden inside team settings.
3. Keep workspace-level ownership separate from team-level routing rules.
4. Make issues the canonical two-way synced object first.
5. Treat pull requests, branches, commits, and reviews as linked engineering artifacts first, then add deeper automation incrementally.
6. Never process webhook payloads synchronously in the request path.
7. Design for idempotency from day one.
8. Prefer explicit mapping tables over inferred magic.
9. Use existing story service and event infrastructure instead of bypassing domain behavior.
10. Roll out in phases so the integration is useful early without being brittle.

## Target User Experience

### Workspace Admin

- Opens `Settings -> Workspace -> Integrations`
- Sees GitHub as an available provider
- Clicks `Connect GitHub`
- Installs the GitHub App into an org or user account
- Returns to FortyOne and sees installations and repositories
- Chooses which repositories are active
- Assigns repositories to one or more teams
- Configures sync defaults and automation policies

### Team Lead

- Opens team settings
- Chooses which connected repositories are relevant to the team
- Sets status mapping for issue and PR lifecycle
- Enables automation like:
  - Create GitHub issue when story is created
  - Move story to started when PR opens
  - Move story to completed when PR merges
  - Sync comments with GitHub issues

### Contributor

- Sees linked GitHub issue, PRs, branches, commits, and reviews on a story
- Can open the linked GitHub issue directly
- Can create a GitHub issue from a story if one does not exist
- Can see GitHub-originated activity inside the story activity feed

## Functional Requirements

### Required for Version 1

- Connect a workspace to one or more GitHub App installations
- Discover repositories available to each installation
- Assign repositories to one or more teams
- Inbound issue sync:
  - Opened
  - Edited
  - Closed
  - Reopened
  - Assignee changes
  - Label changes only if mapping is configured
- Outbound issue sync:
  - Story created
  - Story title/description updated
  - Story status changed
  - Assignee changed when identity mapping exists
- Inbound issue comments
- Outbound story comments to issue comments
- PR linkage to stories
- PR lifecycle visibility:
  - Open
  - Draft ready changes
  - Closed
  - Merged
- Commit and branch linkage to stories through story ref parsing
- Team-level automation rules for issue and PR status transitions
- Delivery logs, retries, idempotency, and health visibility

### Should Be Included Soon After V1

- Pull request review and review comment ingestion
- Identity linking between GitHub users and FortyOne members
- Branch creation from stories
- PR creation helpers
- Label mapping
- Bulk initial import for existing GitHub issues

### Explicitly Out of Scope for First Release

- Full project board sync
- Milestone sync
- Check run sync
- GitHub Discussions sync
- Wiki sync
- Code scanning or security alerts
- Full review thread editing parity between Slack, FortyOne, and GitHub

## Recommended Architecture

### High-Level Model

- Workspace owns GitHub installations
- Installations expose repositories
- Repositories are assigned to teams
- Stories can be linked to GitHub issues, PRs, branches, commits, comments, and reviews
- Team policies define how GitHub state maps to FortyOne workflow state
- Event-driven processors handle outbound sync
- Webhook processors handle inbound sync

### Module Layout

Create a dedicated server module:

- `apps/server/internal/modules/github/`

Suggested sub-packages:

- `service`
- `repository`
- `http`
- `domain` or `models`
- `client`
- `webhooks`
- `sync`

Suggested responsibilities:

- `service`: orchestration, workflow rules, idempotency coordination
- `repository`: persistence and query methods
- `http`: authenticated management routes and public webhook/install routes
- `client`: GitHub App auth, REST calls, pagination, retries
- `webhooks`: signature verification, event routing, payload parsing
- `sync`: outbound translators and processors

## Data Model Design

The current migrations have the right nouns but not the full structure. The integration should use a more explicit schema.

### 1. `github_installations`

Purpose:

- Track every GitHub App installation connected to a workspace

Key fields:

- `id`
- `workspace_id`
- `installation_id`
- `github_app_id`
- `account_id`
- `account_login`
- `account_type`
- `repository_selection`
- `permissions`
- `events`
- `is_active`
- `suspended_at`
- `suspended_by`
- `installed_by_user_id`
- `created_at`
- `updated_at`

Important change from current migration:

- Do not enforce one installation per workspace. A workspace may need multiple installations across orgs.

### 2. `github_repositories`

Purpose:

- Inventory repositories available through installations

Key fields:

- `id`
- `workspace_id`
- `installation_id`
- `github_repo_id`
- `name`
- `full_name`
- `owner_login`
- `description`
- `private`
- `default_branch`
- `html_url`
- `clone_url`
- `ssh_url`
- `is_active`
- `archived`
- `disabled`
- `last_catalog_synced_at`
- `created_at`
- `updated_at`

Important note:

- Prefer app-level webhook handling. Do not rely on per-repo webhook secrets as the main design.

### 3. `github_repository_team_assignments`

Purpose:

- Map repositories to one or more teams

Key fields:

- `id`
- `repository_id`
- `team_id`
- `is_primary`
- `default_issue_status_id`
- `default_closed_status_id`
- `default_merged_status_id`
- `created_at`
- `updated_at`

This expands the current `repository_team_assignments` concept into a more explicit policy surface.

### 4. `github_team_sync_settings`

Purpose:

- Store per-team GitHub behavior

Key fields:

- `team_id`
- `create_issue_on_story_create`
- `sync_issue_title`
- `sync_issue_body`
- `sync_issue_comments`
- `sync_issue_state`
- `sync_issue_assignees`
- `link_pull_requests`
- `move_story_on_pr_open`
- `move_story_on_pr_ready_for_review`
- `move_story_on_pr_merge`
- `pr_open_status_id`
- `pr_review_status_id`
- `pr_merged_status_id`
- `closed_issue_status_id`
- `reopened_issue_status_id`
- `branch_naming_pattern`
- `story_reference_pattern`
- `created_at`
- `updated_at`

This should replace the overly narrow `team_github_automation_settings` shape.

### 5. `github_object_links`

Purpose:

- Canonical mapping between FortyOne objects and GitHub objects

This is more flexible than only storing `story_github_links`.

Key fields:

- `id`
- `workspace_id`
- `team_id`
- `story_id`
- `comment_id`
- `repository_id`
- `github_object_type`
- `github_node_id`
- `github_numeric_id`
- `github_number`
- `github_url`
- `github_parent_numeric_id`
- `is_canonical`
- `metadata`
- `last_inbound_synced_at`
- `last_outbound_synced_at`
- `created_at`
- `updated_at`

Supported `github_object_type` values:

- `issue`
- `issue_comment`
- `pull_request`
- `pull_request_review`
- `pull_request_review_comment`
- `branch`
- `commit`

### 6. `github_webhook_deliveries`

Purpose:

- Raw webhook ingest audit and replay safety

Key fields:

- `id`
- `delivery_id`
- `event_type`
- `installation_id`
- `repository_id`
- `signature_valid`
- `payload`
- `received_at`
- `processed`
- `processed_at`
- `processing_attempts`
- `last_error`

### 7. `github_sync_outbox`

Purpose:

- Queue and track outbound GitHub operations

Key fields:

- `id`
- `workspace_id`
- `team_id`
- `story_id`
- `comment_id`
- `repository_id`
- `operation_type`
- `dedupe_key`
- `payload`
- `status`
- `attempt_count`
- `scheduled_at`
- `last_attempted_at`
- `completed_at`
- `last_error`

Operation examples:

- `issue.create`
- `issue.update`
- `issue.close`
- `issue.reopen`
- `issue.comment.create`
- `issue.comment.update`

### 8. `github_user_links`

Purpose:

- Map GitHub identities to FortyOne members

Key fields:

- `id`
- `workspace_id`
- `user_id`
- `github_user_id`
- `github_login`
- `created_at`
- `updated_at`

## Workflow and State Mapping

### Story to Issue

Recommended canonical mapping:

- FortyOne story = source-of-truth work item in product context
- GitHub issue = engineering execution mirror or linked issue

Rules:

- A story may have zero or one canonical GitHub issue link
- A story may have many PR, branch, and commit links
- PRs should not replace issues as the canonical synced object in V1

### GitHub Issue State to Story State

Do not hardcode `closed = completed`.

Instead:

- Map GitHub open/reopened/closed into team-configured target statuses
- Store mapping in `github_team_sync_settings`

Example default:

- `issue.opened` -> first `unstarted` status
- `issue.reopened` -> first `started` or `unstarted` status
- `issue.closed` -> first `completed` status

### PR Lifecycle to Story State

Recommended automation options:

- PR opened -> move to `started`
- PR ready for review -> move to review-ready status
- PR merged -> move to `completed`
- PR closed unmerged -> optionally revert or leave unchanged

All of these should be opt-in at team level.

## GitHub App Setup

### Required GitHub App Permissions

Repository permissions:

- Issues: read/write
- Pull requests: read/write
- Contents: read
- Metadata: read
- Commit statuses: read only if needed later

Events:

- `installation`
- `installation_repositories`
- `issues`
- `issue_comment`
- `pull_request`
- `pull_request_review`
- `pull_request_review_comment`
- `push`

### Secret Management

Add environment variables in server config:

- `GITHUB_APP_ID`
- `GITHUB_APP_PRIVATE_KEY`
- `GITHUB_WEBHOOK_SECRET`
- `GITHUB_CLIENT_ID`
- `GITHUB_CLIENT_SECRET`
- `GITHUB_APP_NAME`
- `GITHUB_APP_URL`

If the install callback uses OAuth state or setup action redirects, also include:

- `GITHUB_SETUP_REDIRECT_URL`

## Backend API Plan

### Workspace-Authenticated Routes

All under `/workspaces/{workspaceSlug}/integrations/github`

Suggested routes:

- `GET /status`
- `GET /installations`
- `POST /installations/sync`
- `GET /repositories`
- `POST /repositories/sync`
- `PUT /repositories/{repositoryId}`
- `POST /repositories/{repositoryId}/assignments`
- `DELETE /repositories/{repositoryId}/assignments/{assignmentId}`
- `GET /teams/{teamId}/settings`
- `PUT /teams/{teamId}/settings`
- `POST /stories/{storyId}/create-issue`
- `POST /stories/{storyId}/link-issue`
- `DELETE /stories/{storyId}/links/{linkId}`
- `POST /disconnect`

### Public Routes

Suggested routes:

- `POST /integrations/github/webhooks`
- `GET /integrations/github/install/callback`
- `POST /integrations/github/install/callback`

## Inbound Sync Design

### Webhook Request Path

1. Receive request on public webhook endpoint.
2. Validate GitHub signature using `GITHUB_WEBHOOK_SECRET`.
3. Read `X-GitHub-Delivery`, `X-GitHub-Event`, and installation data.
4. Persist payload into `github_webhook_deliveries`.
5. Return success quickly.
6. Enqueue async processing task.

### Event Routing

Route by event type:

- `installation`
- `installation_repositories`
- `issues`
- `issue_comment`
- `pull_request`
- `pull_request_review`
- `pull_request_review_comment`
- `push`

### Inbound Issues

#### `issues.opened`

- Resolve installation and repository
- Resolve assigned team from repository assignment
- Create story through the story service, not direct SQL shortcuts
- Create canonical `issue` link in `github_object_links`
- Record inbound origin metadata so outbound sync does not echo the same change back

#### `issues.edited`

- Resolve linked story
- Update title/body only if changed
- Preserve loop suppression markers

#### `issues.closed`

- Resolve linked story
- Apply mapped closed status if configured

#### `issues.reopened`

- Resolve linked story
- Apply mapped reopened status if configured

#### `issues.assigned` and `issues.unassigned`

- Only sync if the GitHub user maps to a workspace member
- If no mapping exists, do nothing and log clearly

### Inbound Issue Comments

- Resolve story by canonical issue link
- Create or update mirrored story comment
- Mark the comment as GitHub-originated
- Support delete handling by marking the mirrored comment removed, not hard-deleting blindly

### Inbound Pull Requests

On PR events:

- Parse story reference from:
  - branch name
  - PR title
  - PR body
- Link PR to matching story if found
- Update PR metadata on the story
- Trigger configured automation on:
  - opened
  - ready_for_review
  - closed
  - merged

### Inbound Reviews and Review Comments

Version 1 behavior:

- Ingest as GitHub-originated activity entries or linked artifacts
- Do not try to make them fully editable mirrored story comments yet

### Inbound Push Events

- Parse branch name and commit messages for story refs like `TEAM-123`
- Upsert branch and commit links to matched story
- Expose them in the story UI

## Outbound Sync Design

### Source of Truth for Outbound Work

Use existing domain events, not direct DB triggers.

Current events already exist for:

- story created
- story updated
- comment created

Required event additions:

- comment updated
- comment deleted
- story deleted or archived if needed later
- more explicit sync metadata on story updates

### Outbound Processing Flow

1. Story or comment change occurs in normal app workflow.
2. Domain event is published to Redis Stream.
3. GitHub outbound consumer reads relevant events.
4. Consumer resolves whether the story has:
   - canonical issue link
   - configured repo assignment
   - sync enabled for the changed field
5. Consumer writes a `github_sync_outbox` item.
6. Worker executes the GitHub API call.
7. Success or failure is persisted.

### Outbound Story Create

When enabled and no canonical issue exists:

- Create GitHub issue in the team’s default repository
- Title = story title
- Body = rendered story description plus story URL back to FortyOne
- Store canonical issue link

### Outbound Story Update

Supported field mappings:

- `title` -> issue title
- `description` and `descriptionHTML` -> issue body
- `status_id` -> issue open/closed when mapping exists
- `assignee_id` -> issue assignee when user mapping exists

### Outbound Comments

When enabled:

- Story comment create -> issue comment create
- Comment edits and deletes should be added only when the app gets explicit comment updated/deleted events

## Loop Prevention and Idempotency

This is mandatory.

### Inbound

- De-dupe by `X-GitHub-Delivery`
- Store processed status and last error
- Ignore deliveries already processed successfully

### Outbound

- Every outbound operation must have a stable `dedupe_key`
- Store GitHub response metadata
- Track last outbound change fingerprint

### Echo Suppression

When the app updates GitHub and GitHub sends a webhook back:

- Compare against stored sync metadata
- Ignore if the change fingerprint matches the just-sent operation
- Expire suppression windows after a short safe interval

## UI Plan

### Workspace Settings

Add a new page:

- `apps/projects/src/app/[workspaceSlug]/settings/workspace/integrations/page.tsx`

Add a new module:

- `apps/projects/src/modules/settings/workspace/integrations/`

GitHub card should show:

- Installed / not installed
- Number of installations
- Number of active repositories
- Last sync
- Health state
- Connect / manage button

### GitHub Workspace Detail

Sections:

- Connection state
- Installations
- Repositories
- Repository assignment
- Sync defaults
- Health and errors

### Team Settings

Extend current team management with a GitHub subsection or adjacent tab:

- Default repository
- Story -> issue sync toggles
- Issue -> story sync toggles
- PR automation toggles
- Status mapping
- Branch naming rules

### Story Detail UI

Add GitHub-specific UI elements:

- Canonical GitHub issue panel
- Linked PR list
- Linked branches
- Linked commits
- GitHub-originated activities
- Action buttons:
  - Create issue
  - Link existing issue
  - Open in GitHub

## Recommended File-Level Implementation Sequence

### Phase 1: Foundation

- Add new migrations
- Add GitHub module scaffolding in `apps/server/internal/modules/github`
- Register routes in `apps/server/internal/bootstrap/api/routes.go`
- Wire service construction in `apps/server/internal/bootstrap/api/services.go`

### Phase 2: Install and Repository Catalog

- Implement GitHub App auth client
- Implement install callback
- Implement installation sync
- Implement repository inventory sync
- Add workspace settings UI for GitHub

### Phase 3: Inbound Issues

- Implement webhook endpoint
- Persist deliveries
- Process `issues` events
- Create stories via story service
- Link stories and issues

### Phase 4: Outbound Issues

- Extend outbound consumer
- Add GitHub-specific consumer or worker
- Translate story create and update events into issue API calls
- Add canonical issue creation from story UI

### Phase 5: Comments and PRs

- Add issue comment sync
- Add PR linking and automation
- Add commit and branch visibility

### Phase 6: Hardening

- Replay tools
- Dead-letter handling
- Admin health UI
- Alerting and metrics

## Testing Strategy

### Unit Tests

- GitHub signature verification
- Story ref parsing
- Status mapping resolution
- Loop suppression logic
- Dedupe key generation

### Repository Tests

- Installation upsert
- Repository sync
- Object link upsert
- Webhook delivery persistence
- Outbox retry state changes

### Service Tests

- Issue opened creates story
- Issue edited updates story
- Story created creates issue
- Story updated updates issue
- PR merged moves story when enabled

### Integration Tests

- Install flow callback
- Webhook handler acceptance and queueing
- End-to-end issue sync
- Comment sync both ways

## Operational Concerns

### Logging

Log with structured fields:

- workspace id
- installation id
- repository id
- team id
- story id
- github delivery id
- operation type
- attempt count

### Metrics

Track:

- webhook deliveries by event type
- processing latency
- failed deliveries
- failed outbound operations
- repos synced
- stories linked
- PR automation counts

### Replay and Recovery

Support:

- replaying failed webhook deliveries
- retrying failed outbound operations
- re-syncing installation repositories
- rebuilding links from GitHub history for a repository if needed

## Risks

### Risk: Status Mapping Is Too Rigid

Mitigation:

- Keep status mapping per team and configurable
- Never hardcode GitHub state into one global workflow assumption

### Risk: Duplicate Story Creation from Webhooks

Mitigation:

- De-dupe by delivery id and canonical issue link
- Check for existing issue link before creating a story

### Risk: Sync Loops

Mitigation:

- Use suppression metadata and operation fingerprints
- Persist outbound dedupe keys

### Risk: Identity Mapping Is Incomplete

Mitigation:

- Make assignee syncing optional
- Fallback to no-op and visible warnings

### Risk: PR Review Thread Mapping Is Overcomplicated

Mitigation:

- Keep review threads as linked activity first
- Do not promise full parity in V1

## Recommended Milestones

### Milestone 1

- Workspace can connect GitHub App
- Repositories sync into FortyOne
- Repositories can be assigned to teams

### Milestone 2

- GitHub issue opened creates story
- Story created creates GitHub issue
- Canonical issue link visible on story

### Milestone 3

- Story updates sync to issue
- Issue updates sync to story
- Closed and reopened issue state mapping works

### Milestone 4

- Issue comments sync both ways
- PR links appear on story
- PR merge automation works

### Milestone 5

- Health dashboards
- Retries and replay
- User mapping and assignee sync

## Open Questions

- Can a repository be assigned to multiple teams in the first release, or should one team be primary and others read-only?
- Should a story be allowed to create multiple GitHub issues across repositories, or do we enforce one canonical issue only?
- Do we want label mapping in the first meaningful release, or postpone until workflows stabilize?
- Should PR creation from FortyOne be included in the first release or only linked after GitHub-side creation?
- Do we want initial historical import, or only sync from the moment of connection onward?

## Final Recommendation

Build GitHub integration as a dedicated module centered on a GitHub App, repository-to-team assignment, canonical issue sync, event-driven outbound processing, and webhook-driven inbound processing. Keep issue sync as the core two-way contract, and layer PR, commit, branch, and review behavior on top once the issue path is stable. Do not continue from the earlier partial implementation directly without redesigning the schema and integration boundaries first.
