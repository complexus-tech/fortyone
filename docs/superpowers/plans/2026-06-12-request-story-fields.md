# Request Story Fields Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand integration requests with story-compatible planning fields and map them into created stories when a request is accepted.

**Architecture:** Add first-class request columns for story fields that already exist on `CoreNewStory`: estimate, objective, key result, sprint, start date, and end date. Keep provider-specific context in metadata, but render metadata links and attachments in the request body without adding request-only story concepts.

**Tech Stack:** Go API, PostgreSQL migrations, React/Next.js, TypeScript, existing `ui` and `icons` packages.

---

### Task 1: Backend Contract And Acceptance Mapping

**Files:**

- Create: `apps/server/internal/migrations/000064_integration_request_story_fields.up.sql`
- Create: `apps/server/internal/migrations/000064_integration_request_story_fields.down.sql`
- Modify: `apps/server/internal/modules/integrationrequests/service/models.go`
- Modify: `apps/server/internal/modules/integrationrequests/service/integrationrequests.go`
- Modify: `apps/server/internal/modules/integrationrequests/repository/repository.go`
- Modify: `apps/server/internal/modules/integrationrequests/http/models.go`
- Test: `apps/server/internal/modules/integrationrequests/service/integrationrequests_test.go`

- [ ] Write failing Go tests proving accepted requests pass estimate, objective, key result, sprint, start date, and end date into `CreateExternal`.
- [ ] Add database columns and repository read/write/update support.
- [ ] Extend service/http models and accept mapping.
- [ ] Run `go test ./internal/modules/integrationrequests/...`.

### Task 2: Provider Inputs

**Files:**

- Modify: `apps/server/internal/modules/github/service/github.go`
- Modify: `apps/server/internal/modules/slack/service/slack.go`

- [ ] Preserve GitHub metadata and keep priority behavior unchanged.
- [ ] Keep Slack request creation valid with the new optional fields.
- [ ] Do not invent provider-specific fields that cannot map to stories.

### Task 3: Frontend Request Editing

**Files:**

- Modify: `apps/projects/src/modules/integration-requests/types/index.ts`
- Modify: `apps/projects/src/modules/integration-requests/details.tsx`

- [ ] Add request fields to frontend types and update payload.
- [ ] Add editable deadline, start date, estimate, objective, and sprint controls to request properties.
- [ ] Keep key result supported in data types but do not expose a picker until there is an existing reusable key-result selector.
- [ ] Render source links and attachments from metadata in the body.

### Task 4: Verification

**Commands:**

- `go test ./internal/modules/integrationrequests/... ./internal/modules/github/... ./internal/modules/slack/...`
- `pnpm --filter projects exec tsc --noEmit --incremental false`
- `npx -y react-doctor@latest . --verbose --diff`

- [ ] Fix regressions from verification.
- [ ] Leave unrelated dirty worktree changes untouched.
