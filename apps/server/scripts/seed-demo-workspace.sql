\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE demo_seed_context ON COMMIT DROP AS
SELECT
    workspace.workspace_id,
    team.team_id,
    member.user_id
FROM workspaces AS workspace
JOIN LATERAL (
    SELECT teams.team_id
    FROM teams
    WHERE teams.workspace_id = workspace.workspace_id
    ORDER BY teams.created_at, teams.team_id
    LIMIT 1
) AS team ON true
JOIN LATERAL (
    SELECT workspace_members.user_id
    FROM workspace_members
    WHERE workspace_members.workspace_id = workspace.workspace_id
    ORDER BY
        CASE workspace_members.role WHEN 'admin' THEN 0 ELSE 1 END,
        workspace_members.created_at,
        workspace_members.user_id
    LIMIT 1
) AS member ON true
WHERE workspace.slug = :'workspace_slug';

DO $block$
BEGIN
    IF (SELECT COUNT(*) FROM demo_seed_context) <> 1 THEN
        RAISE EXCEPTION 'Demo seed requires one workspace, team, and member for the requested slug';
    END IF;
END
$block$;

INSERT INTO workspace_strategies (workspace_id, ultimate_goal, description)
SELECT
    context.workspace_id,
    'Make product planning effortless for growing teams',
    'A demo strategy connecting customer value, reliable delivery, and sustainable growth.'
FROM demo_seed_context AS context
ON CONFLICT (workspace_id) DO NOTHING;

INSERT INTO strategic_pillars (workspace_id, name, description, order_index)
SELECT context.workspace_id, seed.name, seed.description, seed.order_index
FROM demo_seed_context AS context
CROSS JOIN (
    VALUES
        ('Customer Love', 'Create an experience teams recommend without hesitation.', 1000),
        ('Operational Excellence', 'Ship predictably with a fast and dependable platform.', 2000),
        ('Sustainable Growth', 'Build repeatable acquisition and retention loops.', 3000)
) AS seed(name, description, order_index)
WHERE NOT EXISTS (
    SELECT 1
    FROM strategic_pillars AS existing
    WHERE existing.workspace_id = context.workspace_id
      AND existing.name = seed.name
);

INSERT INTO labels (name, team_id, workspace_id, color)
SELECT seed.name, context.team_id, context.workspace_id, seed.color
FROM demo_seed_context AS context
CROSS JOIN (
    VALUES
        ('Bug', '#ef4444'),
        ('Feature', '#6366f1'),
        ('Customer request', '#8b5cf6'),
        ('Design', '#ec4899'),
        ('Infrastructure', '#0ea5e9'),
        ('Analytics', '#14b8a6'),
        ('Quick win', '#f59e0b')
) AS seed(name, color)
WHERE NOT EXISTS (
    SELECT 1
    FROM labels AS existing
    WHERE existing.workspace_id = context.workspace_id
      AND existing.team_id = context.team_id
      AND existing.name = seed.name
);

WITH objective_seed AS (
    SELECT *
    FROM (
        VALUES
            (1, 'Deliver a delightful onboarding experience', 'Help new teams reach their first meaningful outcome quickly.', 'Reduce time to first value and remove setup friction.', 'started', 'High', 'On Track', -30, 60),
            (2, 'Improve product reliability and performance', 'Make everyday workflows fast, predictable, and resilient.', 'Raise confidence in the platform under realistic load.', 'started', 'Urgent', 'At Risk', -25, 65),
            (3, 'Increase weekly team engagement', 'Create habits that bring teams back to plan, execute, and review work.', 'Grow meaningful weekly collaboration across active teams.', 'unstarted', 'High', 'On Track', -10, 80),
            (4, 'Launch actionable product analytics', 'Give product leaders trusted insight into delivery and outcomes.', 'Turn workspace activity into clear operating decisions.', 'unstarted', 'Medium', 'On Track', 5, 95),
            (5, 'Build a repeatable customer feedback loop', 'Connect incoming feedback to prioritised delivery and follow-up.', 'Close the loop from customer signal to shipped improvement.', 'paused', 'Medium', 'Off Track', -45, 45)
    ) AS data(seed_order, name, description, short_summary, status_category, priority, health, start_offset, end_offset)
),
missing AS (
    SELECT seed.*, context.workspace_id, context.team_id, context.user_id
    FROM objective_seed AS seed
    CROSS JOIN demo_seed_context AS context
    WHERE NOT EXISTS (
        SELECT 1
        FROM objectives AS existing
        WHERE existing.team_id = context.team_id
          AND existing.name = seed.name
    )
),
numbered AS (
    SELECT
        missing.*,
        COALESCE((SELECT MAX(sequence_id) FROM objectives WHERE team_id = missing.team_id), 0)
            + ROW_NUMBER() OVER (ORDER BY seed_order) AS sequence_id
    FROM missing
)
INSERT INTO objectives (
    sequence_id,
    name,
    description,
    short_summary,
    lead_user_id,
    team_id,
    workspace_id,
    start_date,
    end_date,
    is_private,
    status_id,
    priority,
    health,
    created_by
)
SELECT
    numbered.sequence_id,
    numbered.name,
    numbered.description,
    numbered.short_summary,
    numbered.user_id,
    numbered.team_id,
    numbered.workspace_id,
    CURRENT_DATE + numbered.start_offset,
    CURRENT_DATE + numbered.end_offset,
    false,
    status.status_id,
    numbered.priority,
    CAST(numbered.health AS objective_health_status),
    numbered.user_id
FROM numbered
JOIN objective_statuses AS status
  ON status.workspace_id = numbered.workspace_id
 AND status.category = numbered.status_category;

INSERT INTO team_objective_sequences (workspace_id, team_id, current_sequence)
SELECT context.workspace_id, context.team_id, COALESCE(MAX(objectives.sequence_id), 0)
FROM demo_seed_context AS context
LEFT JOIN objectives ON objectives.team_id = context.team_id
GROUP BY context.workspace_id, context.team_id
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = GREATEST(
    team_objective_sequences.current_sequence,
    EXCLUDED.current_sequence
);

WITH key_result_seed AS (
    SELECT *
    FROM (
        VALUES
            (1, 'Deliver a delightful onboarding experience', 'Reduce median setup time to 10 minutes', 'number', 32.0, 18.0, 10.0, -30, 60),
            (2, 'Deliver a delightful onboarding experience', 'Increase onboarding completion rate to 80%', 'percentage', 42.0, 63.0, 80.0, -30, 60),
            (3, 'Deliver a delightful onboarding experience', 'Publish an interactive getting-started checklist', 'boolean', 0.0, 1.0, 1.0, -20, 30),
            (4, 'Improve product reliability and performance', 'Reduce p95 API latency below 300 ms', 'number', 780.0, 460.0, 300.0, -25, 65),
            (5, 'Improve product reliability and performance', 'Maintain 99.95% successful request rate', 'percentage', 99.40, 99.82, 99.95, -25, 65),
            (6, 'Improve product reliability and performance', 'Resolve the top 10 recurring error signatures', 'number', 0.0, 6.0, 10.0, -20, 50),
            (7, 'Increase weekly team engagement', 'Increase weekly active teams by 25%', 'percentage', 0.0, 11.0, 25.0, -10, 80),
            (8, 'Increase weekly team engagement', 'Reach 60% weekly planning ritual adoption', 'percentage', 18.0, 34.0, 60.0, -10, 80),
            (9, 'Increase weekly team engagement', 'Run 12 customer workflow interviews', 'number', 0.0, 5.0, 12.0, -15, 55),
            (10, 'Launch actionable product analytics', 'Ship the workspace insights dashboard', 'boolean', 0.0, 0.0, 1.0, 5, 70),
            (11, 'Launch actionable product analytics', 'Validate 8 core metric definitions', 'number', 0.0, 3.0, 8.0, 0, 65),
            (12, 'Launch actionable product analytics', 'Achieve 95% event coverage for critical flows', 'percentage', 48.0, 71.0, 95.0, 0, 95),
            (13, 'Build a repeatable customer feedback loop', 'Triage 90% of new feedback within 3 days', 'percentage', 35.0, 58.0, 90.0, -45, 45),
            (14, 'Build a repeatable customer feedback loop', 'Link 40 feedback items to planned work', 'number', 4.0, 19.0, 40.0, -40, 40),
            (15, 'Build a repeatable customer feedback loop', 'Send outcome updates to 20 customers', 'number', 0.0, 7.0, 20.0, -30, 45)
    ) AS data(seed_order, objective_name, name, measurement_type, start_value, current_value, target_value, start_offset, end_offset)
),
missing AS (
    SELECT seed.*, objective.objective_id, objective.team_id, context.workspace_id, context.user_id
    FROM key_result_seed AS seed
    CROSS JOIN demo_seed_context AS context
    JOIN objectives AS objective
      ON objective.workspace_id = context.workspace_id
     AND objective.team_id = context.team_id
     AND objective.name = seed.objective_name
    WHERE NOT EXISTS (
        SELECT 1
        FROM key_results AS existing
        WHERE existing.objective_id = objective.objective_id
          AND existing.name = seed.name
    )
),
numbered AS (
    SELECT
        missing.*,
        COALESCE((SELECT MAX(sequence_id) FROM key_results WHERE team_id = missing.team_id), 0)
            + ROW_NUMBER() OVER (ORDER BY seed_order) AS sequence_id
    FROM missing
)
INSERT INTO key_results (
    objective_id,
    team_id,
    sequence_id,
    name,
    measurement_type,
    start_value,
    current_value,
    target_value,
    lead,
    start_date,
    end_date,
    created_by
)
SELECT
    numbered.objective_id,
    numbered.team_id,
    numbered.sequence_id,
    numbered.name,
    CAST(numbered.measurement_type AS measurement_type),
    numbered.start_value,
    numbered.current_value,
    numbered.target_value,
    numbered.user_id,
    CURRENT_DATE + numbered.start_offset,
    CURRENT_DATE + numbered.end_offset,
    numbered.user_id
FROM numbered;

INSERT INTO team_key_result_sequences (workspace_id, team_id, current_sequence)
SELECT context.workspace_id, context.team_id, COALESCE(MAX(key_results.sequence_id), 0)
FROM demo_seed_context AS context
LEFT JOIN key_results ON key_results.team_id = context.team_id
GROUP BY context.workspace_id, context.team_id
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = GREATEST(
    team_key_result_sequences.current_sequence,
    EXCLUDED.current_sequence
);

INSERT INTO key_result_contributors (key_result_id, user_id)
SELECT key_result.id, context.user_id
FROM demo_seed_context AS context
JOIN objectives AS objective ON objective.workspace_id = context.workspace_id
JOIN key_results AS key_result ON key_result.objective_id = objective.objective_id
WHERE objective.name IN (
    'Deliver a delightful onboarding experience',
    'Improve product reliability and performance',
    'Increase weekly team engagement',
    'Launch actionable product analytics',
    'Build a repeatable customer feedback loop'
)
ON CONFLICT DO NOTHING;

INSERT INTO strategy_objective_alignments (objective_id, pillar_id, workspace_id)
SELECT objective.objective_id, pillar.pillar_id, context.workspace_id
FROM demo_seed_context AS context
JOIN objectives AS objective
  ON objective.workspace_id = context.workspace_id
JOIN strategic_pillars AS pillar
  ON pillar.workspace_id = context.workspace_id
 AND pillar.name = CASE objective.name
     WHEN 'Deliver a delightful onboarding experience' THEN 'Customer Love'
     WHEN 'Build a repeatable customer feedback loop' THEN 'Customer Love'
     WHEN 'Improve product reliability and performance' THEN 'Operational Excellence'
     WHEN 'Launch actionable product analytics' THEN 'Operational Excellence'
     WHEN 'Increase weekly team engagement' THEN 'Sustainable Growth'
 END
WHERE objective.name IN (
    'Deliver a delightful onboarding experience',
    'Improve product reliability and performance',
    'Increase weekly team engagement',
    'Launch actionable product analytics',
    'Build a repeatable customer feedback loop'
)
ON CONFLICT (objective_id) DO NOTHING;

WITH sprint_seed AS (
    SELECT *
    FROM (
        VALUES
            ('Demo: Foundation Sprint', 'Stabilise the core experience and close high-impact setup gaps.', 'Improve product reliability and performance', -28, -15),
            ('Demo: Activation Sprint', 'Help new teams complete setup and invite collaborators.', 'Deliver a delightful onboarding experience', -14, -1),
            ('Demo: Insights Sprint', 'Instrument the product and expose the first trusted metrics.', 'Launch actionable product analytics', 0, 13),
            ('Demo: Engagement Sprint', 'Build collaboration habits and customer feedback follow-up.', 'Increase weekly team engagement', 14, 27)
    ) AS data(name, goal, objective_name, start_offset, end_offset)
)
INSERT INTO sprints (name, goal, objective_id, team_id, workspace_id, start_date, end_date)
SELECT
    seed.name,
    seed.goal,
    objective.objective_id,
    context.team_id,
    context.workspace_id,
    CURRENT_DATE + seed.start_offset,
    CURRENT_DATE + seed.end_offset
FROM sprint_seed AS seed
CROSS JOIN demo_seed_context AS context
JOIN objectives AS objective
  ON objective.workspace_id = context.workspace_id
 AND objective.name = seed.objective_name
WHERE NOT EXISTS (
    SELECT 1
    FROM sprints AS existing
    WHERE existing.workspace_id = context.workspace_id
      AND existing.team_id = context.team_id
      AND existing.name = seed.name
);

CREATE TEMP TABLE demo_story_seed (
    seed_order integer PRIMARY KEY,
    title text NOT NULL,
    status_category text NOT NULL,
    priority text NOT NULL,
    estimate_unit smallint,
    start_offset integer,
    end_offset integer,
    objective_name text,
    key_result_name text,
    sprint_name text,
    parent_title text,
    completed_offset integer,
    description text NOT NULL
) ON COMMIT DROP;

INSERT INTO demo_story_seed VALUES
    (1, 'Demo: Audit the first-run workspace experience', 'completed', 'High', 3, -28, -25, 'Deliver a delightful onboarding experience', 'Reduce median setup time to 10 minutes', 'Demo: Foundation Sprint', NULL, -25, 'Walk through account creation, workspace setup, and the first created story; capture every source of hesitation.'),
    (2, 'Demo: Design the onboarding progress checklist', 'completed', 'High', 5, -24, -18, 'Deliver a delightful onboarding experience', 'Publish an interactive getting-started checklist', 'Demo: Foundation Sprint', NULL, -18, 'Create a compact checklist that celebrates progress and always offers one clear next action.'),
    (3, 'Demo: Implement checklist persistence', 'started', 'High', 5, -15, 4, 'Deliver a delightful onboarding experience', 'Publish an interactive getting-started checklist', 'Demo: Activation Sprint', 'Demo: Design the onboarding progress checklist', NULL, 'Persist checklist completion per workspace member and keep the state consistent across devices.'),
    (4, 'Demo: Add sample project creation', 'unstarted', 'Medium', 3, 2, 10, 'Deliver a delightful onboarding experience', 'Increase onboarding completion rate to 80%', 'Demo: Insights Sprint', NULL, NULL, 'Offer a realistic starter project so new teams can explore the product before importing their own work.'),
    (5, 'Demo: Measure time to first created story', 'started', 'High', 3, -12, 6, 'Deliver a delightful onboarding experience', 'Reduce median setup time to 10 minutes', 'Demo: Activation Sprint', NULL, NULL, 'Capture the elapsed time between workspace creation and the first meaningful story.'),
    (6, 'Demo: Improve empty states across planning views', 'backlog', 'Medium', 5, 8, 22, 'Deliver a delightful onboarding experience', 'Increase onboarding completion rate to 80%', 'Demo: Engagement Sprint', NULL, NULL, 'Replace dead ends with contextual guidance and a direct action for Objectives, Roadmap, and Backlog.'),
    (7, 'Demo: Add invite teammates success confirmation', 'completed', 'Medium', 2, -20, -16, 'Deliver a delightful onboarding experience', 'Increase onboarding completion rate to 80%', 'Demo: Foundation Sprint', NULL, -16, 'Confirm successful invites and explain what each invited teammate should expect next.'),
    (8, 'Demo: Profile slow workspace dashboard queries', 'completed', 'Urgent', 5, -28, -22, 'Improve product reliability and performance', 'Reduce p95 API latency below 300 ms', 'Demo: Foundation Sprint', NULL, -22, 'Record query plans under realistic data volume and identify the highest-cost joins.'),
    (9, 'Demo: Add indexes for story list filters', 'started', 'High', 5, -18, -4, 'Improve product reliability and performance', 'Reduce p95 API latency below 300 ms', 'Demo: Activation Sprint', NULL, NULL, 'Optimise the common workspace, team, status, assignee, and sprint filter paths.'),
    (10, 'Demo: Cache workspace navigation counts', 'paused', 'High', 3, -10, 5, 'Improve product reliability and performance', 'Reduce p95 API latency below 300 ms', 'Demo: Activation Sprint', NULL, NULL, 'Cache expensive sidebar counts with explicit invalidation after story mutations.'),
    (11, 'Demo: Add API latency service-level indicators', 'started', 'High', 3, -7, 8, 'Improve product reliability and performance', 'Maintain 99.95% successful request rate', 'Demo: Insights Sprint', NULL, NULL, 'Track latency and error rate by route with useful percentiles and workspace-safe dimensions.'),
    (12, 'Demo: Investigate intermittent realtime disconnects', 'paused', 'Urgent', 8, -6, 3, 'Improve product reliability and performance', 'Resolve the top 10 recurring error signatures', 'Demo: Insights Sprint', NULL, NULL, 'Reproduce disconnects, classify the failure modes, and add enough telemetry to isolate the cause.'),
    (13, 'Demo: Add graceful reconnect with backoff', 'unstarted', 'High', 5, 4, 13, 'Improve product reliability and performance', 'Resolve the top 10 recurring error signatures', 'Demo: Insights Sprint', 'Demo: Investigate intermittent realtime disconnects', NULL, 'Reconnect without losing pending updates and make connection state visible to the user.'),
    (14, 'Demo: Remove a deprecated activity endpoint', 'cancelled', 'Low', 2, -12, -8, 'Improve product reliability and performance', 'Resolve the top 10 recurring error signatures', 'Demo: Activation Sprint', NULL, NULL, 'This work was superseded by the unified activity feed and is retained to test cancelled-state UI.'),
    (15, 'Demo: Interview teams about weekly planning', 'completed', 'High', 5, -21, -12, 'Increase weekly team engagement', 'Run 12 customer workflow interviews', 'Demo: Foundation Sprint', NULL, -12, 'Learn how teams prepare, run, and follow up on their weekly planning ritual.'),
    (16, 'Demo: Prototype a weekly planning digest', 'started', 'Medium', 5, -8, 7, 'Increase weekly team engagement', 'Reach 60% weekly planning ritual adoption', 'Demo: Activation Sprint', NULL, NULL, 'Summarise priorities, overdue work, risks, and recent wins before the team planning session.'),
    (17, 'Demo: Add recurring planning reminder settings', 'unstarted', 'Medium', 3, 10, 21, 'Increase weekly team engagement', 'Reach 60% weekly planning ritual adoption', 'Demo: Engagement Sprint', NULL, NULL, 'Let workspace members choose a planning day, time, timezone, and delivery channel.'),
    (18, 'Demo: Surface stale in-progress work', 'backlog', 'High', 3, 15, 27, 'Increase weekly team engagement', 'Increase weekly active teams by 25%', 'Demo: Engagement Sprint', NULL, NULL, 'Call attention to stories with no meaningful activity so teams can unblock or re-prioritise them.'),
    (19, 'Demo: Add lightweight end-of-week reflection', 'backlog', 'Low', 3, 18, 30, 'Increase weekly team engagement', 'Increase weekly active teams by 25%', 'Demo: Engagement Sprint', NULL, NULL, 'Prompt teams to record wins, risks, and one improvement for the next week.'),
    (20, 'Demo: Define the active-team metric contract', 'completed', 'High', 3, -10, -5, 'Launch actionable product analytics', 'Validate 8 core metric definitions', 'Demo: Activation Sprint', NULL, -5, 'Document the event, actor, workspace, time window, exclusions, and expected reconciliation query.'),
    (21, 'Demo: Instrument objective creation funnel', 'started', 'High', 5, 0, 8, 'Launch actionable product analytics', 'Achieve 95% event coverage for critical flows', 'Demo: Insights Sprint', NULL, NULL, 'Capture viewed, started, validation-failed, created, and abandoned events for objective creation.'),
    (22, 'Demo: Build workspace insights overview', 'started', 'Urgent', 8, 1, 13, 'Launch actionable product analytics', 'Ship the workspace insights dashboard', 'Demo: Insights Sprint', NULL, NULL, 'Create an answer-first dashboard covering delivery health, engagement, and strategy execution.'),
    (23, 'Demo: Add date-range and team filters to insights', 'unstarted', 'Medium', 5, 7, 15, 'Launch actionable product analytics', 'Ship the workspace insights dashboard', 'Demo: Insights Sprint', 'Demo: Build workspace insights overview', NULL, 'Support comparable periods and stable filter state without triggering redundant requests.'),
    (24, 'Demo: Reconcile dashboard totals with source queries', 'backlog', 'High', 5, 12, 24, 'Launch actionable product analytics', 'Validate 8 core metric definitions', 'Demo: Engagement Sprint', NULL, NULL, 'Add a repeatable QA checklist for totals, filters, empty states, and timezone boundaries.'),
    (25, 'Demo: Create feedback triage views', 'completed', 'High', 5, -32, -24, 'Build a repeatable customer feedback loop', 'Triage 90% of new feedback within 3 days', 'Demo: Foundation Sprint', NULL, -24, 'Provide views for new, needs context, planned, shipped, and closed feedback.'),
    (26, 'Demo: Suggest related stories from feedback', 'started', 'High', 8, -13, 9, 'Build a repeatable customer feedback loop', 'Link 40 feedback items to planned work', 'Demo: Activation Sprint', NULL, NULL, 'Rank likely related stories while keeping the final linking decision with the product manager.'),
    (27, 'Demo: Add customer outcome update composer', 'unstarted', 'Medium', 5, 9, 23, 'Build a repeatable customer feedback loop', 'Send outcome updates to 20 customers', 'Demo: Engagement Sprint', NULL, NULL, 'Draft concise updates that explain what changed and why it matters to the customer.'),
    (28, 'Demo: Track feedback response time', 'backlog', 'Medium', 3, 15, 29, 'Build a repeatable customer feedback loop', 'Triage 90% of new feedback within 3 days', 'Demo: Engagement Sprint', NULL, NULL, 'Measure time from feedback submission to the first meaningful product-team response.'),
    (29, 'Demo: Fix keyboard focus in command menus', 'started', 'High', 3, -4, 3, NULL, NULL, 'Demo: Insights Sprint', NULL, NULL, 'Preserve focus through filtering, nested menus, selection, and escape navigation.'),
    (30, 'Demo: Polish dark-mode contrast in detail panels', 'unstarted', 'Medium', 3, 3, 11, NULL, NULL, 'Demo: Insights Sprint', NULL, NULL, 'Review borders, muted text, hover states, and selected rows against the shared theme tokens.'),
    (31, 'Demo: Document the local development setup', 'backlog', 'Low', 2, 20, 32, NULL, NULL, NULL, NULL, NULL, 'Write a short path from clone to a working web app, API, worker, database, and test suite.'),
    (32, 'Demo: Upgrade dependency health reporting', 'cancelled', 'Low', 3, -18, -10, NULL, NULL, 'Demo: Foundation Sprint', NULL, NULL, 'Superseded experiment retained as sample cancelled work for list, board, and analytics testing.');

WITH missing AS (
    SELECT seed.*, context.workspace_id, context.team_id, context.user_id
    FROM demo_story_seed AS seed
    CROSS JOIN demo_seed_context AS context
    WHERE NOT EXISTS (
        SELECT 1
        FROM stories AS existing
        WHERE existing.workspace_id = context.workspace_id
          AND existing.team_id = context.team_id
          AND existing.title = seed.title
          AND existing.deleted_at IS NULL
    )
),
numbered AS (
    SELECT
        missing.*,
        COALESCE((SELECT MAX(sequence_id) FROM stories WHERE team_id = missing.team_id), 0)
            + ROW_NUMBER() OVER (ORDER BY seed_order) AS sequence_id
    FROM missing
)
INSERT INTO stories (
    sequence_id,
    team_id,
    title,
    description,
    description_html,
    objective_id,
    key_result_id,
    status_id,
    assignee_id,
    reporter_id,
    priority,
    sprint_id,
    workspace_id,
    start_date,
    end_date,
    completed_at,
    estimate,
    estimate_unit
)
SELECT
    numbered.sequence_id,
    numbered.team_id,
    numbered.title,
    numbered.description,
    '<p>' || numbered.description || '</p>',
    objective.objective_id,
    key_result.id,
    status.status_id,
    numbered.user_id,
    numbered.user_id,
    numbered.priority,
    sprint.sprint_id,
    numbered.workspace_id,
    CASE WHEN numbered.start_offset IS NULL THEN NULL ELSE CURRENT_DATE + numbered.start_offset END,
    CASE WHEN numbered.end_offset IS NULL THEN NULL ELSE CURRENT_DATE + numbered.end_offset END,
    CASE WHEN numbered.completed_offset IS NULL THEN NULL ELSE CURRENT_TIMESTAMP + numbered.completed_offset * INTERVAL '1 day' END,
    numbered.estimate_unit,
    numbered.estimate_unit
FROM numbered
JOIN statuses AS status
  ON status.team_id = numbered.team_id
 AND status.category = numbered.status_category
LEFT JOIN objectives AS objective
  ON objective.workspace_id = numbered.workspace_id
 AND objective.team_id = numbered.team_id
 AND objective.name = numbered.objective_name
LEFT JOIN key_results AS key_result
  ON key_result.objective_id = objective.objective_id
 AND key_result.name = numbered.key_result_name
LEFT JOIN sprints AS sprint
  ON sprint.workspace_id = numbered.workspace_id
 AND sprint.team_id = numbered.team_id
 AND sprint.name = numbered.sprint_name;

UPDATE stories AS child
SET parent_id = parent.id
FROM demo_story_seed AS seed
CROSS JOIN demo_seed_context AS context
JOIN stories AS parent
  ON parent.workspace_id = context.workspace_id
 AND parent.team_id = context.team_id
 AND parent.title = seed.parent_title
WHERE seed.parent_title IS NOT NULL
  AND child.workspace_id = context.workspace_id
  AND child.team_id = context.team_id
  AND child.title = seed.title
  AND child.parent_id IS NULL;

UPDATE stories AS blocked
SET blocked_by_id = blocker.id
FROM demo_seed_context AS context
JOIN stories AS blocker
  ON blocker.workspace_id = context.workspace_id
 AND blocker.team_id = context.team_id
 AND blocker.title = 'Demo: Investigate intermittent realtime disconnects'
WHERE blocked.workspace_id = context.workspace_id
  AND blocked.team_id = context.team_id
  AND blocked.title = 'Demo: Add graceful reconnect with backoff'
  AND blocked.blocked_by_id IS NULL;

INSERT INTO team_story_sequences (workspace_id, team_id, current_sequence)
SELECT context.workspace_id, context.team_id, COALESCE(MAX(stories.sequence_id), 0)
FROM demo_seed_context AS context
LEFT JOIN stories ON stories.team_id = context.team_id
GROUP BY context.workspace_id, context.team_id
ON CONFLICT (workspace_id, team_id)
DO UPDATE SET current_sequence = GREATEST(
    team_story_sequences.current_sequence,
    EXCLUDED.current_sequence
);

WITH assignments(title, label_name) AS (
    VALUES
        ('Demo: Audit the first-run workspace experience', 'Customer request'),
        ('Demo: Design the onboarding progress checklist', 'Design'),
        ('Demo: Implement checklist persistence', 'Feature'),
        ('Demo: Add sample project creation', 'Feature'),
        ('Demo: Improve empty states across planning views', 'Design'),
        ('Demo: Add invite teammates success confirmation', 'Quick win'),
        ('Demo: Profile slow workspace dashboard queries', 'Infrastructure'),
        ('Demo: Add indexes for story list filters', 'Infrastructure'),
        ('Demo: Cache workspace navigation counts', 'Infrastructure'),
        ('Demo: Investigate intermittent realtime disconnects', 'Bug'),
        ('Demo: Add graceful reconnect with backoff', 'Bug'),
        ('Demo: Prototype a weekly planning digest', 'Feature'),
        ('Demo: Build workspace insights overview', 'Analytics'),
        ('Demo: Add date-range and team filters to insights', 'Analytics'),
        ('Demo: Reconcile dashboard totals with source queries', 'Analytics'),
        ('Demo: Create feedback triage views', 'Customer request'),
        ('Demo: Suggest related stories from feedback', 'Customer request'),
        ('Demo: Fix keyboard focus in command menus', 'Bug'),
        ('Demo: Fix keyboard focus in command menus', 'Quick win'),
        ('Demo: Polish dark-mode contrast in detail panels', 'Design')
)
INSERT INTO story_labels (story_id, label_id)
SELECT story.id, label.label_id
FROM assignments
CROSS JOIN demo_seed_context AS context
JOIN stories AS story
  ON story.workspace_id = context.workspace_id
 AND story.team_id = context.team_id
 AND story.title = assignments.title
JOIN labels AS label
  ON label.workspace_id = context.workspace_id
 AND label.team_id = context.team_id
 AND label.name = assignments.label_name
ON CONFLICT DO NOTHING;

WITH comment_seed(title, content) AS (
    VALUES
        ('Demo: Implement checklist persistence', '[Demo seed] The data model is ready; the remaining work is wiring optimistic updates.'),
        ('Demo: Add indexes for story list filters', '[Demo seed] Query plans look much better with realistic workspace volume.'),
        ('Demo: Investigate intermittent realtime disconnects', '[Demo seed] Reproduced after a laptop wakes from sleep; capturing the reconnect trace next.'),
        ('Demo: Build workspace insights overview', '[Demo seed] Keep the first screen answer-first: delivery health, engagement, then drill-downs.'),
        ('Demo: Suggest related stories from feedback', '[Demo seed] The initial ranking should explain why each story is considered related.'),
        ('Demo: Fix keyboard focus in command menus', '[Demo seed] Please verify arrow keys, Home/End, Escape, and focus restoration.'),
        ('Demo: Polish dark-mode contrast in detail panels', '[Demo seed] The muted border token is too subtle against the elevated surface.'),
        ('Demo: Prototype a weekly planning digest', '[Demo seed] Include overdue work only when it has an owner and a clear next action.'),
        ('Demo: Add sample project creation', '[Demo seed] Use believable content and make it obvious that the sample can be deleted.'),
        ('Demo: Track feedback response time', '[Demo seed] Define whether automated acknowledgements count before implementing the metric.')
)
INSERT INTO story_comments (content, story_id, commenter_id)
SELECT comment_seed.content, story.id, context.user_id
FROM comment_seed
CROSS JOIN demo_seed_context AS context
JOIN stories AS story
  ON story.workspace_id = context.workspace_id
 AND story.team_id = context.team_id
 AND story.title = comment_seed.title
WHERE NOT EXISTS (
    SELECT 1
    FROM story_comments AS existing
    WHERE existing.story_id = story.id
      AND existing.content = comment_seed.content
);

COMMIT;

WITH target AS (
    SELECT workspace_id FROM workspaces WHERE slug = :'workspace_slug'
)
SELECT 'objectives' AS entity, COUNT(*) AS total
FROM objectives JOIN target USING (workspace_id)
UNION ALL
SELECT 'key_results', COUNT(*)
FROM key_results
JOIN objectives ON objectives.objective_id = key_results.objective_id
JOIN target ON target.workspace_id = objectives.workspace_id
UNION ALL
SELECT 'sprints', COUNT(*)
FROM sprints JOIN target USING (workspace_id)
UNION ALL
SELECT 'labels', COUNT(*)
FROM labels JOIN target USING (workspace_id)
UNION ALL
SELECT 'stories', COUNT(*)
FROM stories JOIN target USING (workspace_id)
WHERE stories.deleted_at IS NULL
UNION ALL
SELECT 'strategic_pillars', COUNT(*)
FROM strategic_pillars JOIN target USING (workspace_id)
ORDER BY entity;
