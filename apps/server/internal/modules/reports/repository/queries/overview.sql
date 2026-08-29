-- name: GetWorkspaceMetrics :one
SELECT
    CAST(COUNT(DISTINCT story.id) AS int) AS total_stories,
    CAST(COUNT(DISTINCT story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed_stories,
    CAST(COUNT(DISTINCT objective.objective_id) AS int) AS active_objectives,
    CAST(COUNT(DISTINCT sprint.sprint_id) AS int) AS active_sprints,
    CAST(COUNT(DISTINCT member.user_id) AS int) AS total_team_members
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
LEFT JOIN objectives AS objective ON objective.objective_id = story.objective_id
LEFT JOIN sprints AS sprint ON sprint.sprint_id = story.sprint_id
LEFT JOIN team_members AS member ON member.team_id = story.team_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]));

-- name: ListWorkspaceCompletionTrend :many
SELECT
    CAST(DATE_TRUNC('week', story.created_at) AS timestamptz) AS week_start,
    CAST(COUNT(DISTINCT story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed,
    CAST(COUNT(DISTINCT story.id) AS int) AS total
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY DATE_TRUNC('week', story.created_at)
ORDER BY week_start;

-- name: ListWorkspaceVelocityTrend :many
SELECT
    TO_CHAR(DATE_TRUNC('week', story.updated_at), 'Mon DD') AS period,
    CAST(COUNT(DISTINCT story.id) AS int) AS velocity
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND status.category = 'completed'
  AND story.updated_at >= sqlc.arg(start_date)
  AND story.updated_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY DATE_TRUNC('week', story.updated_at)
ORDER BY DATE_TRUNC('week', story.updated_at);

