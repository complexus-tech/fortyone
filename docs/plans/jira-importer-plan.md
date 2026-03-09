# Jira Importer Plan

## Document Status

- Status: Draft implementation plan
- Scope: One-time and staged Jira import into FortyOne
- Primary app surfaces:
  - `apps/projects` for importer UI
  - `apps/server` for Jira auth, import jobs, mapping, and persistence
- Goal: Let teams migrate from Jira into FortyOne safely, visibly, and repeatably

## Problem Statement

Many potential customers will already have work living in Jira. If FortyOne forces manual recreation, adoption friction will be high. The product needs a Jira importer that can:

- Connect to a Jira Cloud workspace
- Discover projects, issue types, statuses, users, comments, and attachments
- Map Jira data into FortyOne workspace structures
- Run imports safely with preview, validation, idempotency, and progress tracking
- Support migration without requiring deep Jira sync forever

This should be an importer first, not a full ongoing bidirectional Jira integration.

## Recommended Product Positioning

The right first product is:

- `Jira importer`, not `full Jira sync`

Reason:

- Importers win migrations
- Full sync creates long-term complexity and product ambiguity
- You want teams to move into FortyOne, not stay split-brained forever

So the initial plan should support:

- preview
- selective import
- mapping
- import execution
- post-import reconciliation

It should not assume continuous two-way sync.

## Current Repository Reality

The repo already has the right base primitives:

- Workspace settings can add a new `Integrations` or `Import` flow
- Teams, statuses, labels, stories, comments, and members already exist as strong domain objects
- Story creation and updates already flow through backend services
- Async processing and scheduling already exist in `apps/server` via Redis and Asynq

That means the importer should be built as a proper module and job pipeline, not as a frontend-only wizard posting giant payloads.

## Guiding Principles

1. Treat Jira import as a migration workflow, not a long-term mirrored integration.
2. Never import directly into final tables without preview and mapping.
3. Preserve source identifiers so imports are idempotent and auditable.
4. Allow partial and staged imports by project.
5. Fail loudly on mapping conflicts and permissions issues.
6. Do not assume Jira workflows map 1:1 to FortyOne workflows.
7. Prefer admin-controlled import with team-level routing.
8. Support resumability for large imports.

## Target User Experience

### Workspace Admin

- Opens `Settings -> Workspace -> Imports` or `Integrations`
- Chooses `Import from Jira`
- Connects Jira Cloud site
- Picks Jira projects to import
- Reviews preview counts and mapping suggestions
- Maps:
  - Jira projects -> FortyOne teams
  - Jira statuses -> FortyOne statuses
  - Jira users -> FortyOne members
  - Jira priorities -> FortyOne priorities
- Runs import
- Watches progress and sees warnings/errors
- Reviews summary with imported, skipped, and failed items

### Team Lead

- Reviews imported team data
- Adjusts workflow mappings before final import
- Optionally imports comments, attachments, and labels

### Contributor

- Sees imported work already structured in FortyOne
- Can trace origin back to Jira when needed

## Functional Requirements

### Required for Version 1

- Jira Cloud OAuth or API token based connection
- Jira site and project discovery
- Import preview with counts
- Project-to-team mapping
- Status mapping
- User mapping
- Priority mapping
- Issue import into stories
- Comment import
- Labels import
- Optional attachment import
- Source link back to Jira issue
- Import run progress UI
- Import summary and error reporting
- Idempotent rerun support

### Recommended for Version 1.1

- Sprint import
- Epic import
- Parent/sub-task import
- Story point import
- Due date and assignee import refinements
- Rich text conversion improvements
- Attachment import with retries and quotas

### Out of Scope for First Release

- Full ongoing Jira sync
- Jira automation sync
- Jira board sync
- Jira dashboards
- Jira custom field parity for every field
- Jira Server/Data Center support unless there is clear demand

## Recommended Architecture

### High-Level Model

- Workspace connects a Jira source
- Admin creates an import configuration
- Import preview fetches and stages Jira metadata
- Admin confirms mappings
- Import job runs asynchronously
- Imported objects store source references for audit and rerun safety

### Module Layout

Create:

- `apps/server/internal/modules/jiraimport/`

Suggested sub-packages:

- `service`
- `repository`
- `http`
- `client`
- `mapping`
- `jobs`
- `converters`

Responsibilities:

- `service`: orchestration of preview and import
- `repository`: import runs, mappings, staged data, source references
- `http`: authenticated routes for preview and execution
- `client`: Jira API wrapper
- `mapping`: field and workflow translation
- `jobs`: background import execution
- `converters`: Jira text, issue, comment, and attachment conversion

## Data Model Design

### 1. `jira_connections`

Purpose:

- Store Jira workspace/site connection metadata

Key fields:

- `id`
- `workspace_id`
- `jira_cloud_id`
- `site_url`
- `site_name`
- `auth_type`
- `access_token_encrypted`
- `refresh_token_encrypted`
- `expires_at`
- `connected_by_user_id`
- `is_active`
- `created_at`
- `updated_at`

### 2. `jira_import_runs`

Purpose:

- Track import attempts

Key fields:

- `id`
- `workspace_id`
- `jira_connection_id`
- `status`
- `started_by_user_id`
- `mode`
- `selected_projects`
- `include_comments`
- `include_attachments`
- `include_labels`
- `include_sprints`
- `include_epics`
- `total_items`
- `processed_items`
- `succeeded_items`
- `failed_items`
- `warning_count`
- `started_at`
- `completed_at`
- `last_error`
- `created_at`
- `updated_at`

### 3. `jira_project_mappings`

Purpose:

- Map Jira projects to FortyOne teams

Key fields:

- `id`
- `import_run_id`
- `jira_project_id`
- `jira_project_key`
- `jira_project_name`
- `team_id`
- `create_team_if_missing`
- `created_at`
- `updated_at`

### 4. `jira_status_mappings`

Purpose:

- Map Jira statuses to FortyOne statuses

Key fields:

- `id`
- `import_run_id`
- `jira_project_id`
- `jira_status_id`
- `jira_status_name`
- `jira_status_category`
- `fortyone_status_id`
- `created_at`
- `updated_at`

### 5. `jira_user_mappings`

Purpose:

- Map Jira users to workspace members

Key fields:

- `id`
- `import_run_id`
- `jira_account_id`
- `jira_display_name`
- `jira_email`
- `user_id`
- `resolution_type`
- `created_at`
- `updated_at`

### 6. `jira_source_links`

Purpose:

- Preserve Jira-to-FortyOne object linkage

Key fields:

- `id`
- `workspace_id`
- `team_id`
- `story_id`
- `comment_id`
- `jira_issue_id`
- `jira_issue_key`
- `jira_comment_id`
- `jira_url`
- `source_type`
- `metadata`
- `created_at`
- `updated_at`

### 7. `jira_import_errors`

Purpose:

- Store row-level or entity-level errors

Key fields:

- `id`
- `import_run_id`
- `jira_entity_type`
- `jira_entity_id`
- `jira_entity_key`
- `stage`
- `error_code`
- `error_message`
- `payload`
- `created_at`

### 8. `jira_import_staging`

Purpose:

- Cache preview or staged raw payloads for large imports

Key fields:

- `id`
- `import_run_id`
- `entity_type`
- `source_id`
- `payload`
- `normalized_payload`
- `created_at`

## Jira Objects to Support

### Required Initially

- Projects
- Users
- Statuses
- Issues
- Comments
- Labels
- Attachments optionally

### Recommended Next

- Epics
- Sub-tasks
- Sprints
- Components
- Issue links
- Story points when available

### Low Priority

- Custom fields beyond selected mapped ones
- Worklogs
- Automation rules
- Watchers

## Object Mapping Strategy

### Jira Project -> FortyOne Team

Default recommendation:

- One Jira project maps to one FortyOne team

Admin can choose:

- map to an existing team
- create a new team during import

### Jira Issue -> FortyOne Story

Default field mapping:

- summary -> title
- description -> description / descriptionHTML
- assignee -> assigneeId if mapped
- reporter -> reporterId if mapped, else importer user or system user
- priority -> FortyOne priority
- due date -> endDate
- created -> createdAt when allowed, or store in metadata
- updated -> updatedAt when allowed, or store in metadata
- labels -> labels
- parent/sub-task -> parentId relationship

### Jira Status -> FortyOne Status

This is the hardest mapping area.

Approach:

- suggest mappings using Jira status category:
  - `To Do` -> `unstarted` or `backlog`
  - `In Progress` -> `started`
  - `Done` -> `completed`
- require user review before execution

### Jira User -> FortyOne Member

Resolution order:

1. exact email match
2. explicit admin mapping
3. leave unmapped and import with null assignee

### Jira Comments -> FortyOne Comments

Import comments as historical comments with:

- original author if mapped
- otherwise importer/system attribution plus original author in metadata
- original created timestamp stored in metadata if product comment model cannot safely backfill

### Jira Attachments -> FortyOne Attachments

If enabled:

- fetch attachment metadata and binary
- upload through existing attachment pipeline
- maintain original filename and source link

## Import Modes

### Mode 1: Preview

Fetch and display:

- projects available
- issue counts
- users
- statuses
- labels
- sample issues

No stories are created.

### Mode 2: Dry Run Validation

Validate:

- auth
- project access
- mapping completeness
- conflicting teams or statuses
- missing users
- attachment size concerns

### Mode 3: Full Import

Perform staged import asynchronously.

### Mode 4: Incremental Retry

Re-run failed or skipped entities from a previous import run.

## Import Execution Flow

1. Admin creates a new import run.
2. Backend validates Jira connection.
3. Backend fetches selected project metadata.
4. Metadata is staged in DB.
5. Admin reviews and edits mappings.
6. Admin confirms import.
7. Background job begins:
   - import teams if needed
   - import statuses if needed
   - import users mappings or unresolved placeholders
   - import issues
   - import issue relationships
   - import comments
   - import labels
   - import attachments if enabled
8. Progress updates are written to `jira_import_runs`.
9. UI polls for progress.
10. Completion report is shown.

## Import Order

Recommended order:

1. projects and team resolution
2. statuses
3. users
4. labels
5. epics if supported
6. issues without parent dependencies first if needed
7. sub-tasks / parent links
8. comments
9. attachments
10. cross-links

## UI Plan

### Workspace Settings

Add a new page:

- `apps/projects/src/app/[workspaceSlug]/settings/workspace/imports/page.tsx`

Or place Jira inside:

- `settings/workspace/integrations`

### Jira Import Wizard

Suggested steps:

1. Connect Jira
2. Select projects
3. Review preview
4. Map projects to teams
5. Map statuses
6. Map users
7. Choose import options
8. Confirm and start
9. Watch progress
10. Review results

### Results UI

Show:

- imported stories count
- skipped count
- failed count
- missing user mappings
- unmapped statuses
- attachment failures
- links back to imported teams

## API Plan

All authenticated under:

- `/workspaces/{workspaceSlug}/imports/jira`

Suggested routes:

- `POST /connect`
- `GET /status`
- `GET /projects`
- `POST /preview`
- `GET /preview/{runId}`
- `PUT /preview/{runId}/project-mappings`
- `PUT /preview/{runId}/status-mappings`
- `PUT /preview/{runId}/user-mappings`
- `POST /runs/{runId}/execute`
- `GET /runs/{runId}`
- `GET /runs/{runId}/errors`
- `POST /runs/{runId}/retry`
- `POST /disconnect`

## Background Job Design

Use Asynq for import execution.

Suggested task types:

- `jira:preview:build`
- `jira:import:run`
- `jira:import:attachments`
- `jira:import:retry`

Each job should:

- be resumable
- checkpoint progress
- write errors to DB
- avoid duplicate story creation with source links

## Idempotency and Safety

This is mandatory.

### Story Idempotency

Before creating a story:

- check if `jira_issue_id` or `jira_issue_key` already exists in `jira_source_links`
- if yes, skip or update based on mode

### Comment Idempotency

Before creating a comment:

- check if `jira_comment_id` already exists

### Attachment Idempotency

Before downloading and uploading:

- check whether the attachment source id already exists for the story

## Text Conversion

Jira descriptions may use Atlassian document format or older markup.

Create a dedicated converter layer to:

- translate Jira rich text into HTML acceptable for FortyOne
- fall back safely when formatting cannot be preserved
- preserve source raw payload in metadata when needed

## Attachment Strategy

Attachments can be expensive and failure-prone.

Recommended first release behavior:

- optional toggle
- size guardrails
- per-file retry
- clear failure reporting

Do not let one bad attachment fail the entire import.

## Priority Mapping

Jira priorities vary widely.

Recommended default mapping:

- Highest / Blocker / Critical -> `Urgent`
- High -> `High`
- Medium -> `Medium`
- Low / Lowest -> `Low`
- Unknown -> `No Priority`

Allow override before import execution.

## Epic and Parent-Child Mapping

If epics are imported:

- Jira epic can map to FortyOne epic if that product concept is stable enough
- If epic support is not ready for importer V1, store epic reference in metadata and import underlying issues normally

For sub-tasks:

- create parent stories first
- then link imported sub-tasks via `parentId`

## Sprint Import

This should come after the base importer.

Recommended first support:

- import sprint names and dates
- map imported issues to imported sprints

Do not block base issue import on sprint support.

## Auditability

Every imported story should preserve origin metadata:

- Jira issue key
- Jira issue URL
- source project
- original timestamps where relevant
- import run id

This should be visible in story metadata or at least stored for admin debugging.

## Testing Strategy

### Unit Tests

- status mapping suggestions
- priority mapping
- text conversion
- idempotency checks
- user matching logic

### Repository Tests

- import run creation and updates
- source link de-dupe behavior
- error persistence
- progress checkpointing

### Service Tests

- preview generation
- execution orchestration
- issue import into story
- comment import
- attachment import fallback handling

### Integration Tests

- full preview flow
- full import flow for a small Jira sample project
- rerun behavior on partially imported data

## Operational Concerns

### Logging

Structured fields:

- workspace id
- import run id
- jira project id
- jira issue key
- team id
- stage
- attempt count

### Metrics

Track:

- preview runs
- import runs started/completed/failed
- issues imported
- comments imported
- attachments imported
- mapping conflicts
- unresolved users

### Replay and Recovery

Support:

- retry failed rows
- rerun from a checkpoint
- export import error report

## Risks

### Risk: Workflow Mapping Is Wrong

Mitigation:

- require explicit mapping review
- provide strong defaults but not automatic blind acceptance

### Risk: Large Imports Time Out

Mitigation:

- background jobs
- staging
- pagination
- checkpointing

### Risk: Duplicate Imports

Mitigation:

- persistent source link table
- import-run aware idempotency checks

### Risk: Rich Text Conversion Is Messy

Mitigation:

- preserve raw content in metadata
- convert conservatively
- do not block import on imperfect formatting

### Risk: User Matching Is Incomplete

Mitigation:

- allow null assignee
- show unresolved users clearly
- support manual mapping

## Recommended Milestones

### Milestone 1

- Jira connection
- project discovery
- preview generation

### Milestone 2

- project/status/user mapping
- import run creation
- issue import

### Milestone 3

- comments, labels, and source links
- import summaries and retry flow

### Milestone 4

- attachments
- parent-child relationships
- basic sprint import

## Open Questions

- Do we support Jira Cloud only at first?
- Should epics map to FortyOne epics immediately or remain metadata first?
- Should imported comments retain original timestamps as first-class values or only in metadata?
- Do we create missing users in FortyOne, or only map existing workspace members?
- Should imported issues default to existing teams or create new teams by default?

## Final Recommendation

Build Jira as a migration-grade importer with strong preview, mapping, asynchronous execution, source-link idempotency, and clear error reporting. Do not build continuous two-way Jira sync first. The fastest win is helping teams leave Jira safely and land in FortyOne with their structure preserved.
