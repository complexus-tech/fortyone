# Figma Integration Plan

## Document Status

- Status: Draft implementation plan
- Scope: Product-level Figma integration for design-to-workflow collaboration
- Primary app surfaces:
  - `apps/projects` for UI and settings
  - `apps/server` for OAuth, webhook ingestion if needed, metadata sync, and link enrichment
- Goal: Make Figma a first-class design source inside FortyOne without turning FortyOne into a design editor

## Problem Statement

Design-heavy teams need their design work connected to planning and execution. Today, a Figma integration should help users:

- link files and frames to work
- create work from designs
- keep design references visible in story context
- optionally notify teams when design artifacts change
- preserve design-to-delivery traceability

The right first version is not “edit Figma from FortyOne.” It is structured linking, context sync, and workflow acceleration around design artifacts.

## Recommended Product Positioning

The first meaningful Figma integration should focus on:

1. Rich Figma link previews
2. Linking files and frames to stories
3. Creating stories from selected Figma designs
4. Optional design review and handoff workflows
5. Optional notifications when linked Figma nodes change

That is much stronger and more realistic than trying to clone Figma behavior.

## Current Repository Reality

The repo already gives you strong places to attach this:

- Storys already support external links
- Story details already have activity, links, comments, and attachments surfaces
- Workspace settings can add integrations pages
- Backend already supports link metadata and attachment patterns

This means a Figma integration should become a dedicated structured link and context integration, not just another raw URL pasted into a generic links list.

## Guiding Principles

1. Treat Figma as a design system and artifact source, not as a task tracker.
2. Keep the first release focused on linking and context, not full sync.
3. Make node-level linking possible, not just file-level linking.
4. Preserve stable identifiers for Figma files and node ids.
5. Add workflow value around design review, handoff, and implementation readiness.
6. Build structured metadata and previews, not only simple hyperlinks.
7. Avoid over-fetching or expensive background sync unless the customer opts in.

## Target User Experience

### Workspace Admin

- Opens `Settings -> Workspace -> Integrations`
- Connects Figma
- Enables file metadata access and preview features
- Optionally enables design change notifications

### Product Manager or Designer

- Opens a story
- Links a Figma file or frame
- Sees a rich preview with file name, frame name, last editor, and updated time
- Uses a shortcut to create one or more stories from selected Figma frames
- Marks a story as `design ready` or `handoff ready`

### Engineer

- Opens a story
- Sees the linked Figma frame and context immediately
- Opens design directly from the story
- Uses the design as part of implementation workflow

## Functional Requirements

### Required for Version 1

- Connect Figma workspace or user via OAuth
- Resolve Figma file and node metadata from shared links
- Support linking Figma files and frames to stories
- Rich preview on linked Figma artifacts
- Create a story from a Figma link
- Optional bulk create stories from multiple selected Figma frames later
- Store stable Figma identifiers
- Show design metadata in the story UI
- Optional workspace and team-level notification preferences

### Recommended for Version 1.1

- Structured `design review` workflow support
- Handoff status automation
- Figma comment reference linking
- Design asset attachment import
- Linked frame update notifications
- Bulk import from a Figma page or section

### Out of Scope for First Release

- Editing Figma documents from FortyOne
- Full comment sync with Figma comments
- Live embedded canvases with editing
- Complete design-system token sync
- Code generation workflow inside product

## Recommended Architecture

### High-Level Model

- Workspace connects Figma
- Stories can link to Figma files and nodes
- Figma metadata is normalized and stored
- Story UI renders structured previews and change context
- Optional notification jobs watch linked artifacts for changes

### Module Layout

Create:

- `apps/server/internal/modules/figma/`

Suggested sub-packages:

- `service`
- `repository`
- `http`
- `client`
- `parsers`
- `preview`
- `sync`

Responsibilities:

- `service`: orchestration and business rules
- `repository`: Figma connections, files, nodes, and story links
- `http`: authenticated management endpoints
- `client`: Figma API access
- `parsers`: URL and node id extraction
- `preview`: normalized preview payload generation
- `sync`: optional metadata refresh jobs

## Data Model Design

### 1. `figma_connections`

Purpose:

- Store Figma OAuth connection for a workspace

Key fields:

- `id`
- `workspace_id`
- `figma_user_id`
- `figma_email`
- `access_token_encrypted`
- `refresh_token_encrypted`
- `expires_at`
- `connected_by_user_id`
- `is_active`
- `created_at`
- `updated_at`

### 2. `figma_files`

Purpose:

- Cache known Figma files linked into the workspace

Key fields:

- `id`
- `workspace_id`
- `figma_file_key`
- `name`
- `thumbnail_url`
- `last_modified`
- `last_editor_id`
- `team_name`
- `project_name`
- `url`
- `created_at`
- `updated_at`

### 3. `figma_nodes`

Purpose:

- Cache linked node-level metadata

Key fields:

- `id`
- `figma_file_id`
- `figma_node_id`
- `name`
- `node_type`
- `url`
- `thumbnail_url`
- `last_synced_at`
- `metadata`
- `created_at`
- `updated_at`

### 4. `story_figma_links`

Purpose:

- Link stories to files or nodes

Key fields:

- `id`
- `story_id`
- `workspace_id`
- `figma_file_id`
- `figma_node_id`
- `link_type`
- `is_primary`
- `created_by_user_id`
- `created_at`
- `updated_at`

Possible `link_type` values:

- `file`
- `frame`
- `component`
- `prototype`
- `design_system`

### 5. `figma_sync_preferences`

Purpose:

- Workspace-level Figma behavior settings

Key fields:

- `workspace_id`
- `enable_metadata_refresh`
- `enable_change_notifications`
- `notify_on_linked_frame_updates`
- `notify_on_file_renames`
- `refresh_interval_minutes`
- `created_at`
- `updated_at`

### 6. `figma_activity_snapshots`

Purpose:

- Store normalized snapshot state for change detection

Key fields:

- `id`
- `workspace_id`
- `figma_file_id`
- `figma_node_id`
- `snapshot_hash`
- `snapshot_payload`
- `captured_at`

## Supported Figma Objects

### Required Initially

- file links
- frame links
- component links

### Recommended Later

- sections
- pages
- prototypes
- comments
- design system libraries

## Product Workflows

### 1. Link Figma Artifact to Story

User flow:

- open story
- click `Add Figma link`
- paste Figma URL
- backend parses file key and node id
- backend fetches metadata
- story shows structured Figma preview

### 2. Create Story from Figma

User flow:

- user pastes a Figma frame link in creation flow
- product suggests title based on frame name
- metadata is attached to the story immediately

Recommended future enhancement:

- `Create multiple stories from selected frames`

### 3. Design Review Workflow

Optional team workflow:

- story has primary Figma frame
- team can mark:
  - `needs design`
  - `in review`
  - `approved`
  - `handoff ready`

This does not require Figma to own the workflow, only to be linked to it.

### 4. Design Change Awareness

If enabled:

- the system periodically refreshes linked file/node metadata
- if the linked artifact changed materially, add an activity entry or notification

## Figma URL Handling

The parser should support:

- file-level links
- node-level links
- branch links if they matter later
- make links only if needed later

Minimum parsed fields:

- file key
- node id if present
- original URL

The parser should normalize URLs into a stable canonical form.

## Preview Strategy

Rich preview should include:

- file name
- frame or node name
- file thumbnail or node image if available
- last modified time
- last editor when available
- direct open link

This should render much better than the generic external links component currently used for all links.

## Backend API Plan

All authenticated under:

- `/workspaces/{workspaceSlug}/integrations/figma`

Suggested routes:

- `GET /status`
- `POST /connect`
- `POST /resolve-link`
- `GET /files`
- `GET /files/{fileId}`
- `GET /nodes/{nodeId}`
- `POST /stories/{storyId}/links`
- `DELETE /stories/{storyId}/links/{linkId}`
- `PUT /preferences`
- `POST /refresh`
- `POST /disconnect`

## UI Plan

### Workspace Settings

Add integration detail page under workspace settings:

- connection status
- capability toggles
- refresh settings
- change notification settings

### Story Detail

Add a Figma-specific section with:

- primary linked design
- additional linked frames
- preview card with thumbnail
- open in Figma button
- replace primary frame action

### Story Creation

Allow:

- create story from pasted Figma link
- optionally prefill title and metadata

### Search and Filters Later

Support filtering by:

- has Figma link
- linked file
- design review state

## Notifications and Activity

### Recommended Notification Cases

- linked frame updated materially
- linked file renamed
- design marked handoff ready later if workflow is added

### Activity Feed Integration

Add activity entries like:

- `Linked Figma frame`
- `Updated primary design reference`
- `Figma frame changed since last review`

Keep these as product activities, not generic external webhooks.

## Sync Strategy

### Version 1

- fetch metadata on demand when a link is created
- refresh metadata when a story is opened if stale enough

### Version 1.1

- scheduled refresh job for linked artifacts
- snapshot comparison for changes

### Avoid Early

- aggressive polling of every file in the workspace
- syncing entire design systems by default

## Story Model Recommendations

Figma integration should eventually have stronger structure than generic links.

Recommended options:

1. Introduce `story_figma_links` as a dedicated model
2. Keep generic external links for arbitrary URLs
3. Render Figma links through a special UI path

This is better than storing Figma only as an untyped external link.

## Design Review Workflow Possibility

This can become a strong product differentiator later.

Possible model:

- per-story design review state
- required linked design frame
- review checklist
- approval timestamp
- approved by member

This should be separate from the basic Figma linking release, but the basic design should not block it.

## Testing Strategy

### Unit Tests

- Figma URL parsing
- canonical URL normalization
- metadata transformation
- preview payload generation
- snapshot hashing

### Repository Tests

- file and node upserts
- story Figma link persistence
- sync preference updates

### Service Tests

- resolve link to file and node metadata
- create story link
- update primary link
- refresh stale metadata

### Integration Tests

- OAuth connect flow
- link resolution from story page
- preview rendering payload generation
- metadata refresh job

## Operational Concerns

### Logging

Structured fields:

- workspace id
- figma file key
- figma node id
- story id
- refresh reason
- connection id

### Metrics

Track:

- figma link resolutions
- linked files count
- linked nodes count
- refresh success/failure
- change notifications generated

### Rate Limiting

Figma API usage should be conservative:

- cache metadata
- refresh only stale or explicitly requested items
- avoid large recursive fetches by default

## Risks

### Risk: Product Value Is Too Weak If It Is Just “Paste a Link”

Mitigation:

- make previews structured
- support story creation from design
- support primary design references
- plan for review workflow extension

### Risk: Figma API Access Is Too Expensive for Background Polling

Mitigation:

- on-demand fetch first
- opt-in scheduled refresh later
- cache aggressively

### Risk: Node-Level Metadata Is Inconsistent Across Links

Mitigation:

- normalize all links through a parser
- store both raw and canonical identifiers

### Risk: Generic Link Model Hides Design Importance

Mitigation:

- add dedicated Figma data model and UI rendering

## Recommended Milestones

### Milestone 1

- connect Figma
- resolve file and node metadata
- link Figma artifacts to stories
- render rich previews

### Milestone 2

- create story from Figma link
- support primary design references
- add story activity entries

### Milestone 3

- scheduled metadata refresh
- linked design change notifications
- better frame and component handling

### Milestone 4

- design review workflow
- bulk frame-to-story creation
- optional comment reference linking

## Open Questions

- Do we need Figma OAuth in the first release, or is public/shared-link metadata enough for initial value?
- Should Figma links remain under story links UI initially or get a dedicated story design section immediately?
- Do we want design review state as a core feature or as metadata first?
- Should linked Figma updates create notifications by default or only activity entries?
- Do we need team-level Figma settings, or is workspace-level enough at first?

## Final Recommendation

Build Figma as a structured design context integration focused on rich previews, file and frame linking, story creation from design references, and eventual design review workflows. Do not start with heavy sync or full comment parity. The strongest first release is one that makes design artifacts visible, stable, and actionable inside story execution.
