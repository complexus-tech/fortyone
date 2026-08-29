-- name: ListStoryStatusBreakdown :many
SELECT
    status.name AS status_name,
    CAST(COUNT(story.id) AS int) AS count,
    story.team_id
FROM stories AS story
INNER JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
GROUP BY status.status_id, status.name, status.order_index, story.team_id
ORDER BY status.order_index, story.team_id;

-- name: ListStoryPriorityDistribution :many
SELECT
    CAST(COALESCE(CAST(story.priority AS text), 'No Priority') AS text) AS priority,
    CAST(COUNT(story.id) AS int) AS count
FROM stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
GROUP BY story.priority
ORDER BY CASE CAST(COALESCE(CAST(story.priority AS text), 'No Priority') AS text)
    WHEN 'Urgent' THEN 1
    WHEN 'High' THEN 2
    WHEN 'Medium' THEN 3
    WHEN 'Low' THEN 4
    WHEN 'No Priority' THEN 5
    ELSE 6
END;

-- name: ListStoryCompletionByTeam :many
SELECT
    team.team_id,
    team.name AS team_name,
    CAST(COUNT(story.id) AS int) AS total,
    CAST(COUNT(story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed
FROM teams AS team
LEFT JOIN stories AS story
    ON story.team_id = team.team_id
   AND story.workspace_id = sqlc.arg(workspace_id)::uuid
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
   AND story.created_at >= sqlc.arg(start_date)
   AND story.created_at <= sqlc.arg(end_date)
   AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
   AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR team.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY team.team_id, team.name
ORDER BY team.name, team.team_id;

-- name: ListStoryBurndown :many
SELECT
    CAST(story.updated_at AS date) AS completion_date,
    CAST(COUNT(story.id) AS int) AS remaining
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND status.category = 'completed'
  AND story.updated_at >= sqlc.arg(start_date)
  AND story.updated_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
GROUP BY CAST(story.updated_at AS date)
ORDER BY completion_date;
