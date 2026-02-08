# AI Backend Expansion Plan (Go API)

## Document Status
- Owner: Backend Team
- Scope: `apps/server` only
- Last Updated: 2026-02-08
- Status: Proposed reference plan for implementation

## 1. Objective
Add backend AI capabilities to the Go API as an additive layer that increases product value (proactive insights, background intelligence, governance, cross-client reuse) **without replacing or changing the current AI UX/orchestration in `apps/projects`**.

## 2. Hard Constraints
1. Do not remove, replace, or re-architect current `apps/projects` AI chat behavior.
2. Do not break existing `chat-sessions` and `user memories` APIs.
3. New backend AI must be safe-by-default (authz, auditability, budget controls).
4. Rollout must be incremental and reversible.

## 3. Current State (Baseline)

### 3.1 What already exists in `apps/projects`
- Chat orchestration + model streaming in Next route handlers.
- Tool calling layer for stories/objectives/sprints/teams/navigation/theme.
- Suggestion generation endpoints for substories and key results.
- Voice transcription endpoint.

### 3.2 What already exists in `apps/server`
- Chat session persistence (`chat_sessions`, `chat_messages`).
- User memory CRUD (`user_memories`).
- Asynq worker + scheduler for recurring automation jobs.
- Notifications, reports, and search domains suitable for agent integration.

### 3.3 Why backend AI expansion is still needed
- Proactive intelligence is limited today (mostly user-initiated in chat).
- No centralized AI governance in Go for all clients.
- No backend audit/evaluation framework for AI outputs.
- No server-driven agent jobs for risk detection, weekly digests, workload health.

## 4. Product Outcomes
1. **Proactive Value**: users receive insights without manually asking chat.
2. **Operational Safety**: deterministic budgets, retries, observability, audit logs.
3. **Cross-Client Consistency**: same intelligence available to web/mobile/API consumers.
4. **Scalable Foundation**: reusable AI substrate for future features.

## 5. Non-Goals
1. Replacing `apps/projects/src/app/api/chat/route.ts`.
2. Moving all AI tool execution from Next to Go in phase 1.
3. Autonomous destructive actions without explicit policy and approvals.
4. Broad “general assistant” behavior outside project/work management.

## 6. Guiding Principles
1. **Additive, not disruptive**: current app AI remains source of truth for interactive chat.
2. **Policy-first**: authorization, quotas, and data boundaries before advanced generation.
3. **Evidence-first**: agent outputs must include traceable evidence references.
4. **Human-in-the-loop defaults**: recommendations first, automation later.
5. **Idempotent jobs**: all agent runs should be safely retryable.
6. **Cost-aware orchestration**: budget controls and provider fallback mandatory.

## 7. Target Capabilities (Backend AI Layer)

### 7.1 AI Platform Capabilities (foundation)
1. Provider abstraction (OpenAI + future providers) in Go.
2. Centralized policy engine (workspace-level feature flags and limits).
3. Prompt/version registry for reproducibility.
4. AI run lifecycle tracking (requested -> running -> completed/failed).
5. Structured output validation.
6. Audit logs with redaction controls.
7. Per-workspace and per-feature budget metering.

### 7.2 First Agent Features (business value)
1. Sprint Risk Agent (daily).
2. Objective Health Agent (weekly).
3. Workload/Capacity Agent (weekly).
4. Backlog Hygiene Agent (weekly).
5. Weekly Executive Digest Agent (weekly).

### 7.3 Delivery channels
1. In-app notifications (existing notifications domain).
2. Optional email digests (existing mailer/brevo stack).
3. Optional report snapshot records (reports domain).

## 8. Proposed Architecture (Go)

### 8.1 New domain modules
Create these under `internal/modules`:

1. `internal/modules/aiplatform`
- `service`: orchestration, model execution policies, budget checks.
- `repository`: AI run metadata, usage, provider errors.
- `http`: admin/operator endpoints (optional in phase 2+).

2. `internal/modules/agents`
- `service`: agent specs, execution pipeline, scheduling hooks.
- `repository`: agent configs, run results, evidence references.
- `http`: workspace settings + run history endpoints.

3. `internal/modules/insights`
- `service`: materialized insight feed + prioritization.
- `repository`: insight records and state transitions.
- `http`: query/acknowledge/dismiss actions.

### 8.2 Integration points with existing modules
1. `notifications`: publish agent insights as notifications.
2. `reports`: persist weekly digest outputs.
3. `search`: provide indexed facts for agent context.
4. `tasks` + worker scheduler: recurring execution.
5. `users` and `teams`: permission-scoped targeting.

### 8.3 High-level flow
1. Scheduler enqueues agent run task.
2. Agent service gathers deterministic facts from repositories.
3. Facts are normalized to a compact “analysis packet”.
4. AI platform executes prompt template with strict schema.
5. Output is validated and scored.
6. Insight is persisted and distributed via notifications/reports.
7. Metrics, traces, and budget usage recorded.

## 9. Data Model Plan

## 9.1 New tables

1. `ai_workspace_policies`
- Purpose: workspace-level toggles and budget controls.
- Fields:
  - `workspace_id uuid pk/fk`
  - `enabled boolean not null default false`
  - `monthly_token_budget bigint not null default 0`
  - `monthly_run_budget int not null default 0`
  - `allowed_features jsonb not null default '{}'`
  - `created_at timestamptz`
  - `updated_at timestamptz`

2. `ai_prompts`
- Purpose: versioned prompt templates and schemas.
- Fields:
  - `id uuid pk`
  - `key text not null` (e.g., `sprint_risk_v1`)
  - `version int not null`
  - `system_prompt text not null`
  - `instruction_prompt text not null`
  - `output_schema jsonb not null`
  - `status text not null` (`draft|active|deprecated`)
  - `created_by uuid`
  - `created_at`

3. `ai_runs`
- Purpose: canonical lifecycle of model executions.
- Fields:
  - `id uuid pk`
  - `workspace_id uuid not null`
  - `feature text not null` (e.g., `agent.sprint_risk`)
  - `provider text not null`
  - `model text not null`
  - `status text not null` (`queued|running|succeeded|failed|cancelled`)
  - `prompt_id uuid null`
  - `input_tokens int not null default 0`
  - `output_tokens int not null default 0`
  - `cost_micros bigint not null default 0`
  - `latency_ms int not null default 0`
  - `error_code text null`
  - `error_message text null`
  - `started_at timestamptz null`
  - `completed_at timestamptz null`
  - `created_at timestamptz not null default now()`

4. `agent_configs`
- Purpose: workspace-configured agent behavior.
- Fields:
  - `id uuid pk`
  - `workspace_id uuid not null`
  - `agent_key text not null` (`sprint_risk`, `objective_health`, ...)
  - `enabled boolean not null default false`
  - `schedule_cron text not null`
  - `thresholds jsonb not null default '{}'`
  - `delivery jsonb not null default '{}'`
  - `created_at`
  - `updated_at`
- Constraint: unique `(workspace_id, agent_key)`

5. `agent_runs`
- Purpose: business-level execution records for each agent cycle.
- Fields:
  - `id uuid pk`
  - `workspace_id uuid not null`
  - `agent_key text not null`
  - `status text not null` (`queued|running|succeeded|failed|partial`)
  - `trigger_type text not null` (`schedule|manual|event`)
  - `triggered_by uuid null`
  - `ai_run_id uuid null`
  - `summary text null`
  - `stats jsonb not null default '{}'`
  - `error text null`
  - `started_at`
  - `completed_at`
  - `created_at`

6. `agent_insights`
- Purpose: actionable outputs visible to users.
- Fields:
  - `id uuid pk`
  - `workspace_id uuid not null`
  - `agent_run_id uuid not null`
  - `agent_key text not null`
  - `severity text not null` (`low|medium|high|critical`)
  - `title text not null`
  - `body text not null`
  - `recommendations jsonb not null default '[]'`
  - `evidence jsonb not null default '[]'`
  - `entity_refs jsonb not null default '[]'`
  - `state text not null default 'open'` (`open|acknowledged|dismissed|resolved`)
  - `owner_user_id uuid null`
  - `created_at`
  - `updated_at`

7. `ai_usage_daily`
- Purpose: fast budget checks and reporting.
- Fields:
  - `workspace_id uuid not null`
  - `day date not null`
  - `feature text not null`
  - `run_count int not null default 0`
  - `input_tokens bigint not null default 0`
  - `output_tokens bigint not null default 0`
  - `cost_micros bigint not null default 0`
  - primary key `(workspace_id, day, feature)`

## 9.2 Migration strategy
1. Add tables with forward-only migrations.
2. Seed defaults: `enabled=false` for all policies/configs.
3. Backfill not required initially.
4. Add indexes after baseline write patterns validated.

## 10. API Plan (Go)

## 10.1 Workspace settings and visibility
1. `GET /workspaces/{workspaceSlug}/ai/policy`
2. `PUT /workspaces/{workspaceSlug}/ai/policy`
3. `GET /workspaces/{workspaceSlug}/agents/configs`
4. `PUT /workspaces/{workspaceSlug}/agents/configs/{agentKey}`

## 10.2 Insight consumption
1. `GET /workspaces/{workspaceSlug}/insights`
- filters: `state`, `agentKey`, `severity`, `teamId`, `assigneeId`
2. `PATCH /workspaces/{workspaceSlug}/insights/{id}`
- actions: `acknowledge`, `dismiss`, `resolve`, `assign_owner`

## 10.3 Operational endpoints (admin/operator)
1. `POST /workspaces/{workspaceSlug}/agents/{agentKey}/run`
- manual trigger
2. `GET /workspaces/{workspaceSlug}/agents/runs`
3. `GET /workspaces/{workspaceSlug}/agents/runs/{runId}`
4. `GET /workspaces/{workspaceSlug}/ai/usage`

## 10.4 Authz rules
1. Read insights: workspace members.
2. Update insight state: member+ (owner/admin for destructive actions).
3. Change policy/config: admin only.
4. Manual run: admin only (phase 1), then delegated roles optional.

## 11. Worker + Task Design

## 11.1 New task types
Add to `pkg/tasks`:
1. `TypeAgentSprintRiskRun = "agent:sprint_risk:run"`
2. `TypeAgentObjectiveHealthRun = "agent:objective_health:run"`
3. `TypeAgentWorkloadCapacityRun = "agent:workload_capacity:run"`
4. `TypeAgentBacklogHygieneRun = "agent:backlog_hygiene:run"`
5. `TypeAgentWeeklyDigestRun = "agent:weekly_digest:run"`

## 11.2 Scheduling defaults
Use current scheduler infrastructure:
1. Sprint Risk: daily 08:00 workspace timezone.
2. Objective Health: Monday 09:00.
3. Workload Capacity: Monday 09:15.
4. Backlog Hygiene: Tuesday 09:00.
5. Weekly Digest: Friday 16:00.

Note: timezone support requires per-workspace offset resolution at enqueue time.

## 11.3 Idempotency and retries
1. `agent_runs` unique lock key by `(workspace_id, agent_key, period_bucket)`.
2. Retry max 3 with exponential backoff.
3. Partial outputs allowed; mark run `partial` with explicit errors.

## 12. Agent Specifications

## 12.1 Sprint Risk Agent
- Inputs:
  - active sprints, sprint dates, story statuses, assignees, story update recency.
- Heuristics before AI:
  - overdue ratio, stalled stories, no-update threshold, scope growth.
- AI output schema:
  - list of risks with severity, rationale, affected entities, recommended actions.
- Delivery:
  - team leads + workspace admins notification.

## 12.2 Objective Health Agent
- Inputs:
  - objective status, KR current vs target, trend, timeframe remaining.
- Deterministic metrics:
  - projected attainment %, velocity gap, recent activity score.
- AI output:
  - at-risk objectives, confidence, recovery suggestions.

## 12.3 Workload/Capacity Agent
- Inputs:
  - assigned active stories per member, priority mix, due dates, throughput baseline.
- AI output:
  - overload and underload flags, rebalance suggestions.

## 12.4 Backlog Hygiene Agent
- Inputs:
  - backlog age, missing metadata, duplicate similarity candidates, stale priorities.
- AI output:
  - grouped cleanup suggestions with minimal-change recommendations.

## 12.5 Weekly Executive Digest Agent
- Inputs:
  - completed work, blockers, overdue deltas, trend changes.
- AI output:
  - concise summary + key metrics + next-week focus list.

## 13. AI Platform Service Contract

## 13.1 Core interface
```go
type ExecuteRequest struct {
    WorkspaceID uuid.UUID
    Feature     string
    PromptKey   string
    PromptVars  map[string]any
    Schema      json.RawMessage
    Provider    string
    Model       string
    Timeout     time.Duration
}

type ExecuteResponse struct {
    RunID        uuid.UUID
    Output       json.RawMessage
    InputTokens  int
    OutputTokens int
    CostMicros   int64
    LatencyMs    int
}
```

## 13.2 Required guarantees
1. Validate policy and budget before provider call.
2. Persist run start/end and usage regardless of success/failure.
3. Validate output against declared JSON schema.
4. Return typed domain errors (`ErrBudgetExceeded`, `ErrPolicyDisabled`, `ErrProviderTimeout`, etc).

## 14. Security, Privacy, and Compliance
1. Data minimization: prompt packets include only required fields.
2. PII controls: redact emails/names where unnecessary.
3. Auditability: every run and downstream action linked by IDs.
4. Retention policy:
  - raw prompt payloads optional and encrypted if stored.
  - run metadata retained longer than payloads.
5. No silent auto-mutations in phase 1.
6. Permission checks for every write endpoint.

## 15. Observability

## 15.1 Metrics
1. `ai_runs_total{feature,status,provider,model}`
2. `ai_run_latency_ms{feature,provider,model}`
3. `ai_tokens_total{feature,type=input|output}`
4. `ai_cost_micros_total{feature}`
5. `agent_insights_total{agentKey,severity,state}`
6. `agent_run_failures_total{agentKey,errorCode}`

## 15.2 Tracing
- Trace spans:
  - data gathering
  - prompt build
  - provider call
  - schema validation
  - persistence
  - notification fanout

## 15.3 Logging
- Structured logs with `workspace_id`, `agent_key`, `run_id`, `feature`.
- Never log sensitive prompt fields in plaintext.

## 16. Rollout Plan (Phased)

## Phase 0: Foundations (1-2 weeks)
Deliverables:
1. `aiplatform` skeleton module + provider adapter interface.
2. Migrations for `ai_workspace_policies`, `ai_runs`, `ai_usage_daily`, `ai_prompts`.
3. Budget and policy guardrails with tests.
4. Minimal operator endpoint to view run usage.

Exit criteria:
- Can execute a no-op/sandbox AI run end-to-end with persisted metadata.

## Phase 1: First production agent (Sprint Risk) (1-2 weeks)
Deliverables:
1. `agents` module with `sprint_risk` implementation.
2. Scheduled task + manual trigger endpoint.
3. Insight persistence + notification publishing.
4. Workspace settings for enable/disable and thresholds.

Exit criteria:
- Pilot workspace receives daily risk insights with acceptable precision.

## Phase 2: Weekly Digest + Objective Health (2 weeks)
Deliverables:
1. `weekly_digest` and `objective_health` agents.
2. Report persistence + notification/email outputs.
3. Basic run history UI API endpoints.

Exit criteria:
- Weekly digests reliably generated for pilot workspaces.

## Phase 3: Capacity + Backlog agents (2 weeks)
Deliverables:
1. workload/capacity analysis.
2. backlog hygiene recommendations.
3. evidence links and resolution state workflows.

Exit criteria:
- Teams can triage and resolve generated insights from API payloads.

## Phase 4: Hardening and scale (ongoing)
Deliverables:
1. provider fallback strategy.
2. quality evaluation harness and regression suite.
3. per-workspace tuning and anomaly alerting.

## 17. Implementation Backlog (Concrete)

### 17.1 Code scaffolding
1. `internal/modules/aiplatform/{service,repository,http}`
2. `internal/modules/agents/{service,repository,http}`
3. `internal/modules/insights/{service,repository,http}`
4. `internal/taskhandlers/agents.go`
5. `pkg/tasks/agents.go`

### 17.2 Bootstrap wiring
1. Register services in `internal/bootstrap/api/services.go`.
2. Register routes in `internal/bootstrap/api/routes.go`.
3. Register task handlers in worker bootstrap.
4. Register schedules in worker scheduler.

### 17.3 Migrations
1. add `ai_workspace_policies`.
2. add `ai_prompts`.
3. add `ai_runs`.
4. add `agent_configs`.
5. add `agent_runs`.
6. add `agent_insights`.
7. add `ai_usage_daily`.

### 17.4 Config/env vars
Add to `apps/server/.env.example`:
1. `APP_AI_ENABLED=false`
2. `APP_AI_PROVIDER=openai`
3. `APP_AI_OPENAI_API_KEY=`
4. `APP_AI_DEFAULT_MODEL=`
5. `APP_AI_REQUEST_TIMEOUT=30s`
6. `APP_AI_MAX_RETRIES=3`
7. `APP_AI_MAX_PARALLEL_RUNS=5`
8. `APP_AI_BUDGET_ENFORCEMENT=true`

## 18. Testing Strategy

## 18.1 Unit tests
1. Policy checks (disabled, exhausted budget, allowed feature mismatch).
2. Prompt compilation and schema validation.
3. Agent heuristics and risk scoring.
4. Insight state transitions and permission checks.

## 18.2 Integration tests
1. Repository read/write for all new tables.
2. End-to-end worker execution with mocked provider.
3. Notification/report fanout behavior.
4. Retry/idempotency behavior under transient failures.

## 18.3 Contract tests
1. JSON schema compatibility for each agent output.
2. API response shape snapshot tests.

## 18.4 Load tests
1. N workspaces x M agents daily simulation.
2. Token/cost budget protection under burst conditions.

## 19. Failure Modes and Mitigations
1. Provider outage -> fallback model/provider, enqueue retry, degrade gracefully.
2. Budget exhaustion -> skip run with policy error and notify admins.
3. Invalid output schema -> mark run failed; do not publish insights.
4. Noisy/low-quality output -> confidence threshold + suppression rules.
5. Duplicate runs -> idempotency keys and period locks.

## 20. Governance and Controls
1. Workspace admins can toggle each agent.
2. Workspace admins can set sensitivity/thresholds.
3. Central kill switch (`APP_AI_ENABLED=false`).
4. All agent actions recorded in audit logs.

## 21. Acceptance Criteria (Program-Level)
1. Existing `apps/projects` AI behavior remains unchanged in production.
2. At least one backend agent produces actionable weekly value for pilot users.
3. Run costs and token usage are visible and enforceable by workspace.
4. Insight APIs are stable and permission-safe.
5. Operational on-call playbook exists for failures.

## 22. Open Decisions
1. Provider strategy for phase 1: OpenAI only or OpenAI + fallback.
2. Storage of raw prompts/responses: full, redacted, or metadata-only.
3. Workspace timezone source of truth for scheduling.
4. Initial default thresholds for each agent.
5. Whether email digests are opt-in or opt-out for admins.

## 23. Suggested First Implementation Slice
1. Build `aiplatform` with mocked provider + `ai_runs` persistence.
2. Implement `sprint_risk` agent with deterministic packet + strict JSON schema.
3. Publish insights to notifications only (no email) for first pilot.
4. Add metrics dashboards and alerting before expanding to more agents.

## 24. Notes for Future Convergence (Optional, Not Required)
If later desired, `apps/projects` AI routes can call Go AI APIs for central policy enforcement while preserving the same UX. This is explicitly out of current scope and should be treated as a separate project.

---

## Appendix A: Example Agent Output Schema (Sprint Risk)
```json
{
  "type": "object",
  "required": ["summary", "risks"],
  "properties": {
    "summary": { "type": "string" },
    "risks": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["severity", "title", "reason", "recommendations"],
        "properties": {
          "severity": { "type": "string", "enum": ["low", "medium", "high", "critical"] },
          "title": { "type": "string" },
          "reason": { "type": "string" },
          "recommendations": { "type": "array", "items": { "type": "string" } },
          "entityRefs": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["type", "id"],
              "properties": {
                "type": { "type": "string" },
                "id": { "type": "string" }
              }
            }
          }
        }
      }
    }
  }
}
```

## Appendix B: Suggested Task Queue Assignment
1. Agent execution tasks -> `automation` queue.
2. Insight fanout tasks -> `notifications` queue.
3. Backfill/recompute tasks -> `cleanup` or dedicated `ai-low` queue if added.

## Appendix C: Rollback Plan
1. Disable all agents via policy and env kill switch.
2. Stop scheduler registrations for agent task types.
3. Keep tables in place for postmortem and audit, no destructive rollback.
