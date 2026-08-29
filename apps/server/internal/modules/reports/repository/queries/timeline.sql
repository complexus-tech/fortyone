-- name: ListStoryCompletionTimeline :many
SELECT
    CAST(story.created_at AS date) AS date,
    CAST(COUNT(story.id) AS int) AS created,
    CAST(COUNT(story.id) FILTER (
        WHERE status.category = 'completed'
          AND CAST(story.updated_at AS date) = CAST(story.created_at AS date)
    ) AS int) AS completed
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY CAST(story.created_at AS date)
ORDER BY date;

-- name: ListObjectiveProgressTimeline :many
SELECT
    CAST(objective.created_at AS date) AS date,
    CAST(COUNT(objective.objective_id) AS int) AS total_objectives,
    CAST(COUNT(objective.objective_id) FILTER (WHERE status.category = 'completed') AS int) AS completed_objectives
FROM objectives AS objective
LEFT JOIN objective_statuses AS status ON status.status_id = objective.status_id
WHERE objective.workspace_id = sqlc.arg(workspace_id)::uuid
  AND objective.created_at >= sqlc.arg(start_date)
  AND objective.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR objective.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR objective.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY CAST(objective.created_at AS date)
ORDER BY date;

-- name: ListTeamVelocityTimeline :many
SELECT
    CAST(story.updated_at AS date) AS date,
    story.team_id,
    CAST(COUNT(story.id) AS int) AS velocity
FROM stories AS story
INNER JOIN statuses AS status
    ON status.status_id = story.status_id
   AND status.category = 'completed'
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.updated_at >= sqlc.arg(start_date)
  AND story.updated_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY CAST(story.updated_at AS date), story.team_id
ORDER BY date, story.team_id;

-- name: ListKeyMetricsTimeline :many
SELECT
    CAST(story.created_at AS date) AS date,
    CAST(COUNT(DISTINCT story.assignee_id) AS int) AS active_users,
    CAST(COUNT(story.id) AS double precision) AS stories_per_day,
    CAST(ROUND(COALESCE(AVG(EXTRACT(EPOCH FROM (story.updated_at - story.created_at)) / 86400), 0), 2) AS double precision) AS avg_cycle_time
FROM stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY CAST(story.created_at AS date)
ORDER BY date;
