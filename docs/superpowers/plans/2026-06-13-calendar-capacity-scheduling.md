# Calendar Capacity Scheduling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add calendar-aware delivery prediction and AI-assisted deadline updates, starting with Google Calendar and leaving a clean provider boundary for Outlook.

**Architecture:** Build a provider-neutral calendar layer in the Go API, store per-user calendar connections and busy/free windows, and feed that availability into a scheduling engine that can predict feasible delivery dates for stories, sprints, and objectives. Maya should own automated deadline changes through the existing system actor and every automated decision must leave a human-readable reason in activity metadata.

**Tech Stack:** Go API, PostgreSQL migrations, Google OAuth/Calendar API, existing `reports` analytics module, existing `stories` update/activity service, existing Maya system actor, Next.js Projects app, React Query, `nuqs`, existing Settings/Analytics UI patterns.

---

## Product Decisions

- Start with Google Calendar only.
- Design the provider interface so Outlook can be added without changing scheduling logic.
- Treat calendar access as opt-in per user.
- Default automated changes to "suggest only"; later workspace/team settings can allow Maya to apply changes automatically.
- Keep deadlines optional for manually created work, but require enough scheduling inputs for predictions: estimate/duration, assignee, working schedule, and either desired deadline or planning horizon.
- Use existing story fields first: `start_date`, `end_date`, `estimate_unit`, `assignee_id`, `sprint_id`, `objective_id`.
- Add metadata fields for prediction confidence and reason instead of replacing existing activity logs.
- Attribute automatic changes to Maya using the existing system actor `maya@fortyone.app`.

## Motion Reference Notes

Motion's official docs describe auto-scheduling around duration, deadline, start date, priority, recurrence, working schedules, and busy/free calendar events. Official help text says deadlines are automatically added so tasks do not get left behind, while some third-party reviews mention a "No deadline" option. For FortyOne, the better product decision is to allow no deadline for backlog capture, but only produce confident delivery predictions when the scheduling inputs are sufficient.

## Existing Code Anchors

- Server routing: `apps/server/internal/bootstrap/api/routes.go`
- Server service wiring: `apps/server/internal/bootstrap/api/services.go`
- Stories service and activity creation: `apps/server/internal/modules/stories/service/stories.go`
- Story DB fields: `apps/server/internal/migrations/000017_stories.up.sql`
- Story activity DB fields: `apps/server/internal/migrations/000018_story_activities.up.sql`
- Estimate units: `apps/server/internal/migrations/000052_estimate_units.up.sql`
- Maya/system actor: `apps/server/internal/migrations/000050_insert_system_user.up.sql`, `apps/server/internal/platform/actors/resolver.go`
- Reports/Pulse service: `apps/server/internal/modules/reports/service/pulse.go`, `apps/server/internal/modules/reports/service/workload.go`
- Reports routes: `apps/server/internal/modules/reports/http/routes.go`
- Google auth foundation: `apps/server/pkg/google/service.go`
- Integrations UI: `apps/projects/src/modules/settings/workspace/integrations/index.tsx`
- Analytics/Pulse UI: `apps/projects/src/modules/analytics/index.tsx`, `apps/projects/src/modules/analytics/components/pulse-report.tsx`
- Story update client paths: `apps/projects/src/modules/story/actions/update-story.ts`, `apps/projects/src/modules/story/hooks/update-mutation.ts`

## Phase 1: Calendar Connections

### Task 1: Add Calendar Integration Storage

**Files:**
- Create: `apps/server/internal/migrations/000065_calendar_integrations.up.sql`
- Create: `apps/server/internal/migrations/000065_calendar_integrations.down.sql`
- Create: `apps/server/internal/modules/calendar/repository/models.go`
- Create: `apps/server/internal/modules/calendar/repository/queries.go`
- Create: `apps/server/internal/modules/calendar/repository/commands.go`
- Create: `apps/server/internal/modules/calendar/repository/calendar.go`

- [ ] Add `calendar_connections` with `connection_id`, `workspace_id`, `user_id`, `provider`, encrypted token payload, scopes, connected email, timezone, sync status, last synced time, created/updated timestamps, and revoked timestamp.
- [ ] Add `calendar_busy_windows` with `window_id`, `connection_id`, `workspace_id`, `user_id`, `provider_event_id`, `start_at`, `end_at`, `status`, `transparency`, `source_hash`, and timestamps.
- [ ] Add unique constraints for one active provider connection per user/workspace and idempotent provider event upserts.
- [ ] Add indexes on `(workspace_id, user_id, start_at, end_at)` for fast scheduling lookups.
- [ ] Write repository tests for upsert connection, revoke connection, upsert busy windows, and range queries.

### Task 2: Add Provider-Neutral Calendar Service

**Files:**
- Create: `apps/server/internal/modules/calendar/service/models.go`
- Create: `apps/server/internal/modules/calendar/service/calendar.go`
- Create: `apps/server/internal/modules/calendar/service/provider.go`
- Create: `apps/server/internal/modules/calendar/service/google.go`
- Modify: `apps/server/pkg/google/service.go`
- Modify: `apps/server/internal/bootstrap/api/services.go`
- Modify: `apps/server/internal/bootstrap/api/routes.go`

- [ ] Define `Provider` with `AuthCodeURL`, `ExchangeCode`, `ListBusyWindows`, and `Disconnect`.
- [ ] Extend Google OAuth config to support Calendar scopes separately from sign-in scopes.
- [ ] Use offline access for Google Calendar so background sync can refresh tokens.
- [ ] Store refresh/access tokens encrypted server-side; never expose them to the Projects app.
- [ ] Normalize Google events into provider-neutral busy windows using `busy` vs `free` transparency.
- [ ] Add service tests with a fake provider.

### Task 3: Add Calendar HTTP Endpoints

**Files:**
- Create: `apps/server/internal/modules/calendar/http/models.go`
- Create: `apps/server/internal/modules/calendar/http/routes.go`
- Create: `apps/server/internal/modules/calendar/http/calendar.go`
- Modify: `apps/server/internal/bootstrap/api/routes.go`

- [ ] `GET /workspaces/{workspaceSlug}/calendar/connections`
- [ ] `POST /workspaces/{workspaceSlug}/calendar/google/connect`
- [ ] `GET /workspaces/{workspaceSlug}/calendar/google/callback`
- [ ] `DELETE /workspaces/{workspaceSlug}/calendar/connections/{connectionId}`
- [ ] `POST /workspaces/{workspaceSlug}/calendar/sync`
- [ ] Restrict connection management to the signed-in user, with admins allowed to view team connection status but not tokens.

### Task 4: Add Calendar Settings UI

**Files:**
- Modify: `apps/projects/src/modules/settings/workspace/integrations/index.tsx`
- Create: `apps/projects/src/modules/settings/workspace/integrations/calendar/index.tsx`
- Create: `apps/projects/src/modules/settings/workspace/integrations/calendar/hooks.ts`
- Create: `apps/projects/src/modules/settings/workspace/integrations/calendar/actions.ts`
- Modify: `apps/projects/src/constants/keys.ts`

- [ ] Add Google Calendar card under workspace integrations.
- [ ] Show connected account, last sync time, and disconnect action.
- [ ] Add a "Sync now" action.
- [ ] Add copy explaining that FortyOne reads busy/free availability for planning and does not expose private event details.

## Phase 2: Capacity Model

### Task 5: Add Work Schedules and Scheduling Preferences

**Files:**
- Create: `apps/server/internal/migrations/000066_calendar_scheduling_preferences.up.sql`
- Create: `apps/server/internal/migrations/000066_calendar_scheduling_preferences.down.sql`
- Create: `apps/server/internal/modules/calendar/service/availability.go`
- Create: `apps/server/internal/modules/calendar/service/availability_test.go`
- Create: `apps/projects/src/modules/settings/account/calendar-preferences/index.tsx`

- [ ] Add per-user scheduling preferences: timezone, working days, daily start/end, focus block minimum, scheduling buffer, max scheduled work per day.
- [ ] Default to user timezone and Monday-Friday 09:00-17:00 when no preferences exist.
- [ ] Subtract busy calendar windows from working hours.
- [ ] Return available capacity windows by user/date.
- [ ] Test edge cases: all-day busy events, overlapping busy events, free events, timezone boundaries, and weekends.

### Task 6: Convert Estimates Into Duration

**Files:**
- Create: `apps/server/internal/modules/scheduling/service/estimate.go`
- Create: `apps/server/internal/modules/scheduling/service/estimate_test.go`
- Modify: `apps/server/internal/modules/teamsettings/service/models.go`
- Modify: `apps/server/internal/modules/teamsettings/http/models.go`

- [ ] For `hours`, treat estimate value as hours.
- [ ] For `ideal_days`, convert to configured hours per day.
- [ ] For `points`, add team-level point-to-hours mapping with a sensible default.
- [ ] For `tshirt`, map sizes to hours via team settings.
- [ ] Mark predictions as low confidence when estimates are missing.

## Phase 3: Prediction Engine

### Task 7: Add Scheduling Service

**Files:**
- Create: `apps/server/internal/modules/scheduling/service/models.go`
- Create: `apps/server/internal/modules/scheduling/service/scheduling.go`
- Create: `apps/server/internal/modules/scheduling/service/scheduling_test.go`
- Create: `apps/server/internal/modules/scheduling/repository/queries.go`
- Create: `apps/server/internal/modules/scheduling/repository/scheduling.go`
- Modify: `apps/server/internal/bootstrap/api/services.go`

- [ ] Load candidate stories with assignee, priority, estimate, start/end dates, sprint, objective, blockers, and status category.
- [ ] Compute each assignee's remaining capacity from calendar availability.
- [ ] Respect blocked stories and predecessor relationships before scheduling dependent work.
- [ ] Sort work by urgent/high priority, existing deadline, sprint end date, objective deadline, then created date.
- [ ] Produce predicted start/end windows, feasible/unfeasible state, confidence, and reasons.
- [ ] Test prediction cases: one person, multiple people, blocked work, no estimate, no calendar connected, overloaded capacity, and existing hard deadline.

### Task 8: Add Prediction Endpoints

**Files:**
- Create: `apps/server/internal/modules/scheduling/http/models.go`
- Create: `apps/server/internal/modules/scheduling/http/routes.go`
- Create: `apps/server/internal/modules/scheduling/http/scheduling.go`
- Modify: `apps/server/internal/bootstrap/api/routes.go`

- [ ] `GET /workspaces/{workspaceSlug}/scheduling/predictions`
- [ ] `GET /workspaces/{workspaceSlug}/stories/{storyId}/schedule-prediction`
- [ ] `GET /workspaces/{workspaceSlug}/sprints/{sprintId}/schedule-prediction`
- [ ] `GET /workspaces/{workspaceSlug}/objectives/{objectiveId}/schedule-prediction`
- [ ] Support filters for team IDs, assignee IDs, sprint IDs, objective IDs, and date range.
- [ ] Return reason codes and human-readable reasons.

### Task 9: Feed Predictions Into Pulse and Analytics

**Files:**
- Modify: `apps/server/internal/modules/reports/service/pulse.go`
- Modify: `apps/server/internal/modules/reports/service/models.go`
- Modify: `apps/server/internal/modules/reports/http/models.go`
- Modify: `apps/projects/src/modules/analytics/components/pulse-report.tsx`
- Create: `apps/projects/src/modules/analytics/components/schedule-risk-panel.tsx`

- [ ] Add predicted late stories, predicted late sprints, capacity conflicts, and unscheduled work to Pulse.
- [ ] Add cards that link to filtered story lists or sprint/objective detail pages.
- [ ] Show why the item is at risk, not only that it is at risk.

## Phase 4: Maya Suggestions and Automated Changes

### Task 10: Add Deadline Suggestions

**Files:**
- Create: `apps/server/internal/migrations/000067_schedule_suggestions.up.sql`
- Create: `apps/server/internal/migrations/000067_schedule_suggestions.down.sql`
- Create: `apps/server/internal/modules/scheduling/service/suggestions.go`
- Create: `apps/server/internal/modules/scheduling/repository/suggestions.go`
- Create: `apps/server/internal/modules/scheduling/http/suggestions.go`

- [ ] Store suggestions for story deadline changes, sprint delivery date changes, and objective delivery date changes.
- [ ] Include `suggested_start_date`, `suggested_end_date`, `confidence`, `reason`, `reason_code`, `source`, `status`, and `created_by_actor_id`.
- [ ] De-duplicate open suggestions for the same entity/field when the recommendation has not changed.
- [ ] Add endpoints to list, accept, decline, and apply suggestions.

### Task 11: Add Activity Metadata

**Files:**
- Create: `apps/server/internal/migrations/000068_activity_metadata.up.sql`
- Create: `apps/server/internal/migrations/000068_activity_metadata.down.sql`
- Modify: `apps/server/internal/modules/stories/service/models.go`
- Modify: `apps/server/internal/modules/stories/service/stories.go`
- Modify: `apps/server/internal/modules/stories/repository/models.go`
- Modify: `apps/server/internal/modules/stories/repository/commands.go`
- Modify: `apps/projects/src/components/ui/activity.tsx`
- Modify: `apps/projects/src/modules/story/components/activities.tsx`

- [ ] Add `metadata jsonb` to `story_activities`.
- [ ] Add metadata fields: `source`, `reason`, `reasonCode`, `confidence`, `predictionId`, `suggestionId`, `appliedBy`.
- [ ] Extend `CoreActivity` and API models to include metadata.
- [ ] Render "Maya changed the delivery date because..." in activity feeds.
- [ ] Keep normal user changes visually unchanged.

### Task 12: Apply Suggestions As Maya

**Files:**
- Modify: `apps/server/internal/platform/actors/resolver.go`
- Modify: `apps/server/internal/bootstrap/api/runtime.go`
- Modify: `apps/server/internal/modules/scheduling/service/suggestions.go`
- Modify: `apps/server/internal/modules/stories/service/stories.go`

- [ ] Add `actors.KeyMaya` or alias `actors.KeySystem` clearly as Maya for scheduling automation.
- [ ] When applying a suggestion automatically, call story update paths with Maya's actor ID.
- [ ] Record activity metadata with the reason and prediction reference.
- [ ] Publish existing story updated events so notifications and GitHub sync keep working.
- [ ] Add tests proving the activity `user_id` is Maya's system user.

### Task 13: Add Automation Controls

**Files:**
- Create: `apps/server/internal/migrations/000069_scheduling_automation_preferences.up.sql`
- Create: `apps/server/internal/migrations/000069_scheduling_automation_preferences.down.sql`
- Create: `apps/projects/src/modules/settings/workspace/automation/scheduling.tsx`
- Modify: `apps/projects/src/modules/settings/workspace/teams/management/components/automations.tsx`

- [ ] Add workspace/team setting: `suggest_only`, `auto_apply_low_risk`, `auto_apply_all`, `disabled`.
- [ ] Default every workspace to `suggest_only`.
- [ ] Require admin permission to enable automatic changes.
- [ ] Add a maximum date movement guard, for example Maya cannot move a deadline by more than 14 days without manual approval.

## Phase 5: Background Sync and Jobs

### Task 14: Add Calendar Sync Worker

**Files:**
- Create: `apps/server/pkg/tasks/calendar.go`
- Create: `apps/server/internal/taskhandlers/calendar.go`
- Modify: `apps/server/internal/bootstrap/worker/handlers.go`
- Modify: `apps/server/internal/bootstrap/worker/scheduler.go`

- [ ] Enqueue calendar sync after connection.
- [ ] Schedule periodic sync for active connections.
- [ ] Use provider event hashes to avoid rewriting unchanged windows.
- [ ] Mark stale or failed connections without breaking the user account.

### Task 15: Add Prediction Worker

**Files:**
- Create: `apps/server/pkg/tasks/scheduling.go`
- Create: `apps/server/internal/taskhandlers/scheduling.go`
- Modify: `apps/server/internal/bootstrap/worker/handlers.go`
- Modify: `apps/server/internal/bootstrap/worker/scheduler.go`

- [ ] Recompute predictions after story estimate, assignee, priority, start date, end date, sprint, objective, or status changes.
- [ ] Recompute predictions after calendar sync.
- [ ] Create or update suggestions when predicted dates differ from committed dates.
- [ ] Apply suggestions only when automation settings allow it.

## Phase 6: UI Surfaces

### Task 16: Add Story Detail Schedule Insight

**Files:**
- Modify: `apps/projects/src/modules/story/components/options.tsx`
- Modify: `apps/projects/src/modules/story/components/main-details.tsx`
- Create: `apps/projects/src/modules/story/components/schedule-insight.tsx`
- Create: `apps/projects/src/modules/story/hooks/schedule-prediction.ts`

- [ ] Show predicted completion date beside existing date/estimate fields.
- [ ] Show reason text and confidence.
- [ ] Show accept/decline controls when Maya has a suggestion.
- [ ] Link to the user's calendar connection settings when availability is missing.

### Task 17: Add Analytics Scheduling Tab or Section

**Files:**
- Modify: `apps/projects/src/modules/analytics/index.tsx`
- Create: `apps/projects/src/modules/analytics/components/scheduling-report.tsx`
- Create: `apps/projects/src/modules/analytics/hooks/scheduling-predictions.ts`
- Modify: `apps/projects/src/modules/analytics/types/index.ts`

- [ ] Add scheduling insights to Workspace pulse first.
- [ ] If the page gets crowded, add a third Analytics tab named `Scheduling`.
- [ ] Include capacity by member, predicted late work, unestimated work, unconnected calendars, and recommended deadline changes.

### Task 18: Add Maya Tools

**Files:**
- Create: `apps/projects/src/lib/ai/tools/scheduling.ts`
- Modify: `apps/projects/src/lib/ai/tools/index.ts`
- Modify: `apps/projects/src/app/api/chat/system.ts`

- [ ] Add tools to get scheduling predictions, list schedule suggestions, apply a suggestion, decline a suggestion, and explain a deadline change.
- [ ] Make Maya answer manager questions like "Can we finish this sprint on time?" and "Who is over capacity next week?"
- [ ] Require explicit confirmation before Maya applies a schedule change from chat unless workspace automation settings already allow it.

## Phase 7: Marketing

### Task 19: Update Landing Page After Product Is Real

**Files:**
- Modify: `apps/landing/src/app/(marketing)/page.tsx`
- Modify landing components discovered in the current implementation pass.

- [ ] Only update marketing after Google Calendar connection, predictions, and Maya activity attribution are implemented.
- [ ] Position this as "calendar-aware project planning" and "Maya predicts delivery risk before deadlines slip."
- [ ] Include a real product screenshot or generated product mock based on the shipped UI.

## Acceptance Criteria

- A user can connect Google Calendar.
- FortyOne can sync busy/free calendar windows without exposing private event titles/descriptions.
- The scheduling engine can predict story, sprint, and objective delivery risk using estimates, assignees, priorities, dependencies, and capacity.
- Maya can create deadline suggestions with clear reasons.
- Any automatic deadline change is attributed to Maya in activities.
- The activity feed displays why Maya changed a date.
- Workspace admins can keep the system in suggest-only mode or enable guarded auto-apply.
- Pulse/Analytics shows calendar-aware delivery risk.
- The implementation leaves a clear Outlook provider boundary.

## Recommended Execution Order

1. Calendar connection storage and Google OAuth.
2. Busy/free sync.
3. Availability calculation.
4. Estimate-to-duration conversion.
5. Prediction endpoint without auto-write behavior.
6. Pulse/Analytics read-only scheduling insight.
7. Suggestions table and accept/decline flow.
8. Maya-attributed automatic application with activity metadata.
9. Maya chat tools.
10. Landing page update.

## Risks

- Calendar privacy: store only busy/free windows unless the user explicitly allows event details later.
- Bad predictions from missing estimates: show low confidence and push users to estimate work.
- Over-aggressive automation: default to suggest-only and add movement limits.
- OAuth token security: encrypt tokens and never return them to the UI.
- User trust: every Maya change needs an explanation and a way to undo/decline future suggestions.
