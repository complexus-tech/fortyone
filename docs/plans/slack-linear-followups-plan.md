# Slack Linear-Style Follow-Up Plan

## Document Status

- Status: Follow-up implementation plan after the Slack create-task modal alignment
- Scope: Linkbacks, unfurls, project channels, agent behavior, uploads, projects, and related settings
- Primary surfaces:
  - `apps/server/internal/modules/slack`
  - `apps/server/internal/modules/integrationrequests`
  - `apps/server/internal/modules/stories`
  - `apps/projects/src/modules/settings/workspace/integrations/slack`

## Current Baseline

The current Slack implementation now treats the Slack workspace as a workspace-level connection. Task creation no longer depends on Slack channel to FortyOne team mapping. The Slack modal is responsible for selecting the FortyOne team, and the status picker owns the decision between sending work to Requests or creating a task directly.

That means the next features should not reintroduce channel-to-team mapping as a prerequisite for task creation. Channels can still matter for notifications, project automation, thread sync, and link behavior, but task creation should stay modal-driven.

## Product Principles

- Workspace connection is global to the FortyOne workspace.
- Slash commands and message actions should use dynamic data from FortyOne.
- Slack should never infer the destination team from a channel when the user is explicitly creating work.
- Slack features that post into channels must be opt-in and auditable.
- Slack identity linking is required before personal or permission-sensitive actions.
- Assistant behavior is separate from core integration plumbing.
- Every inbound Slack event must be idempotent.

## Phase 1: Linkbacks

Goal: when a FortyOne task, request, or supported identifier is mentioned in Slack, the bot replies with the canonical FortyOne link once per Slack thread.

Backend work:

- Add a Slack event handler for message events where the bot is a member.
- Parse FortyOne identifiers and URLs from message text.
- Resolve identifiers against workspace-owned records.
- Add a linkback delivery table keyed by workspace, Slack team, channel, thread timestamp, entity type, and entity ID.
- Enforce a cooldown so the same entity is not repeatedly linked in the same thread.
- Skip private or restricted entities when the Slack actor cannot be mapped to an authorized FortyOne user.
- Post a compact reply with title, identifier, status, assignee, and open link.

Settings work:

- Add a workspace-level `Linkbacks` toggle.
- Keep it independent from task creation.
- Add copy explaining that linkbacks only work where the bot is present.

Testing:

- Unit-test identifier parsing.
- Unit-test idempotency and cooldown behavior.
- Integration-test that inactive Slack workspaces do not emit linkbacks.

## Phase 2: Unfurls

Goal: when a FortyOne URL is pasted in Slack, show a richer preview and optionally expose safe actions.

Backend work:

- Add Slack `link_shared` event support.
- Resolve URLs for stories, requests, documents, and later projects.
- Build unfurl payloads with title, status, team, assignee, and description excerpt.
- Add action buttons only when the Slack user is linked and authorized.
- Log unfurl attempts and failures for debugging.

Actions to support first:

- Open in FortyOne.
- Assign to me.
- Move to Request or convert Request to task only where domain permissions allow it.
- Sync Slack thread for the specific entity after thread sync exists.

Settings work:

- Add a workspace-level `Unfurls` toggle.
- Add a privacy note that restricted or private team data will not be unfurled unless the actor is authorized.

Testing:

- URL resolver tests for all supported routes.
- Permission tests for linked and unlinked Slack users.
- Event replay tests to ensure Slack retries are idempotent.

## Phase 3: Projects In Slack

Goal: let users select or link projects from Slack where FortyOne has a first-class project concept available to the task creation flow.

Backend work:

- Add repository methods for project search scoped by workspace and selected team.
- Extend Slack modal external select handlers to search projects by query.
- Add project ID to the create-task submission payload.
- Pass project ID into story creation once the story service supports it.
- Add project search result grouping for the link modal.

Slack modal work:

- Add optional `Project` external select after assignee and labels.
- Refresh project options when the selected team changes.
- Preserve selected project during modal refreshes when it remains valid.

Testing:

- Modal payload tests for project selection.
- Search tests for workspace/team scoping.
- Story creation tests proving project ID is persisted.

## Phase 4: Uploads

Goal: support files attached during Slack-driven task creation without losing provenance.

Backend work:

- Spike Slack file input and file event constraints against the current Slack API before implementation.
- Decide whether uploads come from modal file input, message attachments, or a follow-up action after task creation.
- Add an attachment ingestion path that stores Slack file metadata, downloads with the bot token, and saves through the existing FortyOne attachment/storage layer.
- Associate uploaded files with either the integration request or the created task.
- Add retry-safe file ingestion jobs for larger files.

Slack modal work:

- Add upload UI only after the API path is confirmed.
- Show privacy copy for what files the bot can access.

Testing:

- Unit-test Slack file metadata normalization.
- Integration-test storage failure handling.
- Ensure disconnected workspaces cannot download files.

## Phase 5: Project Channels

Goal: optionally create and manage Slack channels for FortyOne projects.

Backend work:

- Add settings for `project_channels_enabled` and `project_channel_prefix`.
- Add a project-created domain event consumer.
- Create a Slack channel using the configured prefix and normalized project name.
- Invite project members who have linked Slack identities.
- Store the channel mapping separately from team routing.
- Post high-signal project updates into the project channel.

Settings work:

- Add a `Project channels` card with enable toggle and prefix selector.
- Keep it below core connection settings because it is automation, not install health.

Testing:

- Prefix normalization tests.
- Duplicate channel creation tests.
- Member invite tests with linked and unlinked users.

## Phase 6: Agent Behavior

Goal: make Maya useful in Slack without confusing the core integration with the assistant surface.

Backend and runtime work:

- Require Slack user to FortyOne user linking before personal answers.
- Add workspace-level `Slack agent` enablement.
- Add `Slack workflow access` as a separate setting for workflow-triggered agent actions.
- Add workspace guidance text that is included in Slack-agent prompts.
- Use a Slack-specific actor context instead of the browser-authenticated web chat route.
- Gate code intelligence separately from general project/task help.

Supported first actions:

- Create a request from a message.
- Search tasks by natural language.
- Summarize a task thread.
- Answer "what am I working on" for the linked user.

Not first:

- Broad autonomous task mutation.
- Repository/code intelligence without explicit workspace enablement.
- Acting as another user.

Testing:

- Actor-context tests for linked and unlinked Slack users.
- Permission tests for workspace and team scopes.
- Prompt-context tests proving workspace guidance is included only when enabled.

## Data Model Additions

- `slack_user_links`: maps Slack users to FortyOne users per workspace.
- `slack_linkback_deliveries`: dedupes linkback replies per Slack thread/entity.
- `slack_unfurl_events`: audit trail for unfurls and failures.
- `slack_project_channels`: maps FortyOne projects to Slack channels.
- `slack_agent_settings`: workspace-level agent, workflow, code-intelligence, and guidance settings.
- `slack_file_imports`: tracks Slack file ingestion and attachment association.

## Recommended Order

1. Linkbacks, because they are low-interaction and validate event handling.
2. Unfurls, because they reuse URL resolution and permission checks.
3. Slack user linking, because it unlocks safe actions.
4. Projects in the create modal, once project persistence is confirmed.
5. Uploads, after Slack file API constraints are confirmed.
6. Project channels, because they create external Slack state.
7. Agent behavior, once identity linking and permissions are reliable.

## Open Decisions

- Whether project selection should be team-scoped or workspace-wide.
- Whether Slack file uploads should create attachments on requests immediately or wait until a request becomes a task.
- Whether linkbacks should support requests before requests have stable public identifiers.
- Whether project channels should be created automatically for all projects or only when enabled per project.
- Whether Slack agent responses should be allowed in public channels or restricted to DMs and threads.
