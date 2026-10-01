-- name: GetWorkspaceMetrics :one
SELECT
    CAST(COUNT(DISTINCT story.id) FILTER (WHERE story.created_at >= sqlc.arg(start_date) AND story.created_at <= sqlc.arg(end_date)) AS int) AS total_stories,
    CAST(COUNT(DISTINCT story.id) FILTER (WHERE status.category = 'completed' AND story.created_at >= sqlc.arg(start_date) AND story.created_at <= sqlc.arg(end_date)) AS int) AS completed_stories,
    CAST(COUNT(DISTINCT story.id) FILTER (WHERE status.category = 'completed' AND story.completed_at >= sqlc.arg(start_date) AND story.completed_at <= sqlc.arg(end_date)) AS int) AS completed_in_period,
    CAST(COUNT(DISTINCT objective.objective_id) AS int) AS active_objectives,
    CAST(COUNT(DISTINCT sprint.sprint_id) AS int) AS active_sprints,
    CAST(COUNT(DISTINCT member.user_id) AS int) AS total_team_members
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
LEFT JOIN objectives AS objective ON objective.objective_id = story.objective_id
LEFT JOIN sprints AS sprint ON sprint.sprint_id = story.sprint_id
LEFT JOIN team_members AS member ON member.team_id = story.team_id
WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND (
      (story.created_at >= sqlc.arg(start_date) AND story.created_at <= sqlc.arg(end_date))
      OR (story.completed_at >= sqlc.arg(start_date) AND story.completed_at <= sqlc.arg(end_date))
  )
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR story.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR story.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])));

-- name: ListWorkspaceCompletionTrend :many
WITH created_events AS (
    SELECT DATE_TRUNC('week', scoped.created_at) AS event_week, 1 AS created, 0 AS completed
    FROM stories AS scoped
    LEFT JOIN statuses AS status ON status.status_id = scoped.status_id
    WHERE scoped.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND scoped.deleted_at IS NULL
      AND scoped.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR scoped.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR scoped.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR scoped.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR scoped.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND scoped.created_at >= sqlc.arg(start_date) AND scoped.created_at <= sqlc.arg(end_date)
), completed_events AS (
    SELECT DATE_TRUNC('week', scoped.completed_at) AS event_week, 0 AS created, 1 AS completed
    FROM stories AS scoped
    LEFT JOIN statuses AS status ON status.status_id = scoped.status_id
    WHERE scoped.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND scoped.deleted_at IS NULL
      AND scoped.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR scoped.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR scoped.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR scoped.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR scoped.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND status.category = 'completed'
      AND scoped.completed_at >= sqlc.arg(start_date) AND scoped.completed_at <= sqlc.arg(end_date)
), events AS (
    SELECT * FROM created_events
    UNION ALL
    SELECT * FROM completed_events
)
SELECT
    CAST(events.event_week AS timestamptz) AS week_start,
    CAST(SUM(events.completed) AS int) AS completed,
    CAST(SUM(events.created) AS int) AS total
FROM events
GROUP BY events.event_week
ORDER BY week_start;

-- name: ListWorkspaceVelocityTrend :many
SELECT
    TO_CHAR(DATE_TRUNC('week', story.completed_at), 'Mon DD') AS period,
    CAST(COUNT(DISTINCT story.id) AS int) AS velocity
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND status.category = 'completed'
  AND story.completed_at >= sqlc.arg(start_date)
  AND story.completed_at <= sqlc.arg(end_date)
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR story.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR story.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
GROUP BY DATE_TRUNC('week', story.completed_at)
ORDER BY DATE_TRUNC('week', story.completed_at);
