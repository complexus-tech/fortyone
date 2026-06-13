# Top Tier Project Management App Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn FortyOne from a strong project management app into a differentiated, intelligence-driven operating system for planning, triage, execution, and reporting.

**Architecture:** Implement this as six independently shippable product slices: Command Center, Request Triage, Planning Intelligence, Impact Activity, Narrative Analytics, and Safe AI Actions. Each slice should expose backend report/action endpoints first, then frontend UI, then AI tools that consume the same backend contracts.

**Tech Stack:** Go API in `apps/server`, Next.js projects app in `apps/projects`, shared UI in `packages/ui`, Jest/React Testing Library for frontend tests, Go tests for backend services/repositories.

---

## Scope And Sequencing

Do not implement this as one branch. Ship in this order:

1. Command Center
2. Request Triage
3. Planning Intelligence
4. Impact Activity
5. Narrative Analytics
6. Safe AI Actions

Each phase should be releasable on its own and should not depend on hidden frontend calculations. Backend contracts must be reusable by the AI assistant and future analytics pages.

---

## Shared Product Principles

- Backend report endpoints should return decision-ready aggregates, not full unbounded lists.
- Frontend pages should show empty, loading, partial, and failed states deliberately.
- AI tools should call the same backend contracts as the UI.
- Any AI mutation must use preview -> confirmation -> execution.
- Filters and menu selectors must remain paginated and backend-searchable.
- Remove dead code at the end of each phase.

---

## File Structure Map

### Backend

- Modify: `apps/server/internal/modules/reports/service/models.go`
  - Add command-center, planning-risk, and narrative-report core models.
- Modify: `apps/server/internal/modules/reports/service/reports.go`
  - Add service methods that delegate to repository and derive risk summaries.
- Modify: `apps/server/internal/modules/reports/repository/queries.go`
  - Add SQL-backed report queries for command center, planning risk, and narrative analytics.
- Modify: `apps/server/internal/modules/reports/http/routes.go`
  - Add new analytics/report routes.
- Modify: `apps/server/internal/modules/reports/http/reports.go`
  - Add handlers, query parsing, and response mapping.
- Modify: `apps/server/internal/modules/reports/http/models.go`
  - Add app-facing response models.
- Modify: `apps/server/internal/modules/integrationrequests/...`
  - Add request triage scoring, duplicate lookup, and accept/decline reasons.
- Modify: `apps/server/internal/modules/stories/service/...`
  - Add safe action previews for story mutations when needed.

### Frontend

- Create: `apps/projects/src/modules/command-center/`
  - Main command-center page, hooks, components, cards, and risk sections.
- Modify: `apps/projects/src/modules/analytics/`
  - Add report-grade workload, delivery health, and objective risk views.
- Modify: `apps/projects/src/modules/integration-requests/`
  - Add triage queue, duplicate warnings, suggested fields, and batch actions.
- Modify: `apps/projects/src/modules/story/`
  - Add traceability panel and impact-aware activity presentation.
- Modify: `apps/projects/src/lib/ai/tools/`
  - Add report tools and safe mutation preview/confirm tools.
- Modify: `apps/projects/src/app/api/chat/system.ts`
  - Teach assistant when to use command-center reports, planning reports, and action previews.

### Tests

- Add Go service tests beside touched report and integration-request services.
- Add frontend pure helper tests for scoring, sorting, filters, and state transitions.
- Add component tests only where the behavior is hard to preserve by type-check alone.

---

## Phase 1: Command Center

**Outcome:** A workspace-level page that answers: what needs attention right now?

### Backend Contract

Add:

`GET /workspaces/{workspaceSlug}/analytics/command-center`

Response shape:

```ts
type CommandCenterReport = {
  generatedAt: string;
  summary: {
    openStories: number;
    overdueStories: number;
    blockedStories: number;
    untriagedRequests: number;
    overloadedMembers: number;
    atRiskObjectives: number;
    sprintRiskCount: number;
  };
  risks: CommandCenterRisk[];
  workload: WorkloadAnalysis;
  requests: {
    pending: number;
    highPriority: number;
    stale: number;
    byProvider: Array<{ provider: string; count: number }>;
  };
  objectives: Array<{
    objectiveId: string;
    name: string;
    health: "on_track" | "at_risk" | "blocked";
    reason: string;
  }>;
};

type CommandCenterRisk = {
  id: string;
  type:
    | "overdue_story"
    | "overloaded_member"
    | "stale_request"
    | "objective_risk"
    | "sprint_capacity"
    | "unestimated_work";
  severity: "low" | "medium" | "high" | "urgent";
  title: string;
  description: string;
  entityType: "story" | "request" | "objective" | "sprint" | "member";
  entityId: string;
  teamId?: string;
  recommendedAction: string;
};
```

### Tasks

- [ ] Write a failing Go service test in `apps/server/internal/modules/reports/service/command_center_test.go` that verifies risks are sorted by severity and generated even when optional sections are empty.
- [ ] Add `CoreCommandCenterReport`, `CoreCommandCenterRisk`, and summary models to `reports/service/models.go`.
- [ ] Add `GetCommandCenterReport` to the reports repository interface and service.
- [ ] Add repository queries for overdue stories, overloaded members, stale integration requests, at-risk objectives, and unestimated work.
- [ ] Add HTTP route and handler.
- [ ] Add frontend query `apps/projects/src/modules/command-center/queries/get-command-center.ts`.
- [ ] Add page route under the workspace app shell.
- [ ] Build the Command Center UI with sections:
  - Top risk summary strip
  - “Needs attention” ranked list
  - Workload mini report
  - Request triage mini report
  - Objective health mini report
- [ ] Add AI tool `apps/projects/src/lib/ai/tools/command-center.ts` that calls the same endpoint.
- [ ] Add system prompt guidance: use Command Center for broad “what should I focus on?” questions.

### Verification

- `go test ./internal/modules/reports/...`
- `pnpm --filter projects type-check`
- `pnpm --filter projects test`
- Scoped eslint on command-center and AI files.

---

## Phase 2: Request Triage

**Outcome:** Requests become an intelligent intake queue, not just a list.

### Backend Contract

Add or extend request list/detail responses with:

```ts
type IntegrationRequestTriage = {
  score: number;
  recommendation: "accept" | "decline" | "review";
  reasons: string[];
  suggestedStory: {
    title: string;
    description: string;
    priority: string | null;
    estimateValue: number | null;
    labelIds: string[];
    assigneeId: string | null;
  };
  duplicates: Array<{
    storyId: string;
    title: string;
    statusName: string;
    similarityReason: string;
  }>;
};
```

### Tasks

- [ ] Write failing service tests for request triage scoring:
  - high priority GitHub issue increases score
  - duplicate existing story changes recommendation to review
  - missing description lowers confidence
- [ ] Add triage model in `apps/server/internal/modules/integrationrequests/service/models.go`.
- [ ] Add duplicate lookup query using title/content similarity where available.
- [ ] Add request detail response fields for triage.
- [ ] Add batch endpoints:
  - `POST /integration-requests/batch/accept-preview`
  - `POST /integration-requests/batch/accept`
  - `POST /integration-requests/batch/decline`
- [ ] Add frontend triage queue layout:
  - provider filter
  - priority filter
  - stale filter
  - duplicate warning
  - suggested story fields
  - accept/decline reason capture
- [ ] Update request detail banner actions to show recommendation and duplicate warnings.
- [ ] Update AI request triage tool to surface score, recommendation, and duplicates.

### Verification

- `go test ./internal/modules/integrationrequests/...`
- `pnpm --filter projects type-check`
- `pnpm --filter projects test`

---

## Phase 3: Planning Intelligence

**Outcome:** The app can explain whether sprint/objective plans are realistic and suggest rebalancing.

### Backend Contract

Add:

`GET /workspaces/{workspaceSlug}/analytics/planning-risk`

Response:

```ts
type PlanningRiskReport = {
  scope: {
    teamIds?: string[];
    sprintIds?: string[];
    objectiveIds?: string[];
  };
  capacity: {
    totalEstimate: number;
    memberCount: number;
    averageEstimatePerMember: number;
    overloadedMembers: MemberWorkload[];
  };
  risks: Array<{
    type: "capacity" | "deadline" | "unestimated" | "dependency" | "priority";
    severity: "low" | "medium" | "high" | "urgent";
    title: string;
    explanation: string;
    storyIds: string[];
  }>;
  suggestions: Array<{
    action: "move_story" | "split_story" | "assign_member" | "reduce_scope";
    title: string;
    explanation: string;
    storyIds: string[];
    assigneeId?: string;
    targetSprintId?: string;
  }>;
};
```

### Tasks

- [ ] Reuse `WorkloadAnalysis` as the base input for member load.
- [ ] Add planning-risk service method that combines workload, sprint dates, objective dates, estimates, and priorities.
- [ ] Add tests for:
  - overloaded sprint
  - unestimated high-priority stories
  - overdue objective-linked work
- [ ] Add frontend planning report panel on sprint and objective pages.
- [ ] Add “simulate changes” UI:
  - select stories
  - choose move/split/reassign
  - preview resulting capacity
- [ ] Add AI planning report tool.
- [ ] Add safe AI preview action for proposed sprint rebalancing.

### Verification

- Go reports tests
- Frontend type-check and test suite
- Manual check on sprint and objective pages

---

## Phase 4: Impact Activity

**Outcome:** Activity stops reading like raw audit logs and starts reading like product impact.

### New Activity Presentation

Examples:

- “Priority changed from Medium to High. Sprint risk increased because the story is unestimated.”
- “Estimate changed from 2 to 8 points. Joseph is now above the overload threshold.”
- “Accepted from GitHub request #123. Linked source and imported 4 comments.”

### Tasks

- [ ] Add an activity enrichment layer in frontend helpers first.
- [ ] Write tests for activity sentence generation.
- [ ] Add backend activity context only where frontend cannot derive it safely.
- [ ] Update story activity, request activity, and updates tab to use enriched copy.
- [ ] Add icons consistently: estimate, priority, status, request/source, assignee.
- [ ] Add source trace links for GitHub/Slack/Intercom-originated stories.

### Verification

- Jest tests for activity formatter.
- Scoped eslint.
- Manual review on story detail and request detail.

---

## Phase 5: Narrative Analytics

**Outcome:** Analytics becomes report-grade, with conclusions and next actions.

### Reports To Add

1. Workload report
2. Delivery health report
3. Objective progress report
4. Request intake report
5. Team throughput report

### Tasks

- [ ] Add report schemas in `apps/projects/src/modules/analytics/types/index.ts`.
- [ ] Add backend report endpoints where data is not already available.
- [ ] Build reusable report components:
  - summary cards
  - insight list
  - trend chart
  - recommended actions
- [ ] Add “copy report” and “export report” actions.
- [ ] Add AI report tools that return structured report data, not prose-only answers.
- [ ] Add report freshness timestamps and partial-data warnings.

### Verification

- Go report tests.
- Frontend type-check.
- Component/helper tests for report formatting.

---

## Phase 6: Safe AI Actions

**Outcome:** The assistant can act, but only through transparent previews and confirmation.

### Action Flow

Every mutation must follow:

1. User asks for change.
2. AI generates preview.
3. UI shows exactly what will change.
4. User confirms.
5. Backend executes.
6. Activity logs record AI-assisted action.

### Actions To Support First

- Accept selected requests.
- Decline selected requests.
- Create story from request.
- Reassign stories.
- Change priority.
- Change estimate.
- Move stories to sprint.
- Add labels.

### Tasks

- [ ] Create shared AI action preview type.
- [ ] Add backend preview endpoints for bulk mutations.
- [ ] Add confirmation UI in chat side panel.
- [ ] Add mutation execution tools that require preview IDs.
- [ ] Add activity log entries for AI-assisted actions.
- [ ] Add permission checks before previews and before execution.

### Verification

- Backend tests for preview/execution mismatch.
- Frontend tests for confirmation state machine.
- Manual chat flow checks.

---

## Design QA Checklist

Before each phase is considered done:

- [ ] Loading state is not visually jumpy.
- [ ] Empty state says what to do next.
- [ ] Error state gives a recovery action.
- [ ] AI and UI use the same backend contracts.
- [ ] Filters are paginated where lists can grow.
- [ ] No unbounded frontend list fetching for analytics/reporting.
- [ ] No dead helpers, unused types, or duplicated query logic.
- [ ] Chat-open layout still works.
- [ ] Mobile layout does not overlap controls.

---

## Recommended Execution Strategy

Use one branch per phase:

1. `codex/command-center`
2. `codex/request-triage`
3. `codex/planning-intelligence`
4. `codex/impact-activity`
5. `codex/narrative-analytics`
6. `codex/safe-ai-actions`

Each branch should finish with:

```bash
pnpm --filter projects type-check
pnpm --filter projects test
pnpm --filter projects exec eslint <touched frontend files>
cd apps/server && go test ./internal/modules/<touched modules>/...
```

---

## Self-Review

Spec coverage:

- Command center: covered in Phase 1.
- Request triage: covered in Phase 2.
- Planning intelligence: covered in Phase 3.
- Impact activity: covered in Phase 4.
- Narrative analytics: covered in Phase 5.
- Safe AI actions: covered in Phase 6.
- Keyboard-first speed: should be added as a follow-up after core flows stabilize, because shortcuts depend on final screen/action placement.

Known deliberate deferral:

- Keyboard-first command menu is not included as a standalone phase yet. Add it after Phase 6 so it can command the finished action surface instead of being rewired repeatedly.
