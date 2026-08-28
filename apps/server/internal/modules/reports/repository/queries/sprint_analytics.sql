-- name: ListSprintProgress :many
SELECT
    sprint.sprint_id,
    sprint.name AS sprint_name,
    sprint.team_id,
    CAST(COUNT(story.id) AS int) AS total,
    CAST(COUNT(story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed,
    CASE
        WHEN sprint.end_date < CURRENT_DATE THEN 'Completed'
        WHEN sprint.start_date > CURRENT_DATE THEN 'Not Started'
        WHEN sprint.start_date <= CURRENT_DATE AND sprint.end_date >= CURRENT_DATE THEN 'Active'
        ELSE 'Unknown'
    END AS status
FROM sprints AS sprint
LEFT JOIN stories AS story
    ON story.sprint_id = sprint.sprint_id
   AND story.workspace_id = sqlc.arg(workspace_id)::uuid
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE sprint.workspace_id = sqlc.arg(workspace_id)::uuid
  AND sprint.created_at >= sqlc.arg(start_date)
  AND sprint.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR sprint.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR sprint.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
GROUP BY sprint.sprint_id, sprint.name, sprint.team_id, sprint.start_date, sprint.end_date
ORDER BY sprint.name, sprint.sprint_id;

-- name: ListCombinedSprintBurndown :many
SELECT
    CAST(story.updated_at AS date) AS date,
    CAST(COUNT(story.id) FILTER (WHERE story.created_at <= CAST(story.updated_at AS date)) AS int) AS planned,
    CAST(COUNT(story.id) AS int) AS actual
FROM stories AS story
INNER JOIN sprints AS sprint
    ON sprint.sprint_id = story.sprint_id
   AND sprint.workspace_id = sqlc.arg(workspace_id)::uuid
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND status.category = 'completed'
  AND story.updated_at >= sqlc.arg(start_date)
  AND story.updated_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR sprint.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR sprint.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
GROUP BY CAST(story.updated_at AS date)
ORDER BY date;

-- name: ListSprintTeamAllocation :many
SELECT
    team.team_id,
    team.name AS team_name,
    CAST(COUNT(DISTINCT sprint.sprint_id) AS int) AS active_sprints,
    CAST(COUNT(story.id) AS int) AS total_stories,
    CAST(COUNT(story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed_stories
FROM teams AS team
LEFT JOIN sprints AS sprint
    ON sprint.team_id = team.team_id
   AND sprint.workspace_id = sqlc.arg(workspace_id)::uuid
   AND sprint.created_at >= sqlc.arg(start_date)
   AND sprint.created_at <= sqlc.arg(end_date)
   AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR sprint.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
LEFT JOIN stories AS story
    ON story.sprint_id = sprint.sprint_id
   AND story.workspace_id = sqlc.arg(workspace_id)::uuid
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR team.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY team.team_id, team.name
ORDER BY team.name, team.team_id;

-- name: ListSprintHealth :many
WITH sprint_status AS (
    SELECT CASE
        WHEN sprint.end_date < CURRENT_DATE THEN 'Completed'
        WHEN sprint.start_date > CURRENT_DATE THEN 'Not Started'
        WHEN sprint.end_date <= CURRENT_DATE + INTERVAL '3 days' THEN 'At Risk'
        WHEN sprint.start_date <= CURRENT_DATE AND sprint.end_date >= CURRENT_DATE THEN 'On Track'
        ELSE 'Unknown'
    END AS status
    FROM sprints AS sprint
    WHERE sprint.workspace_id = sqlc.arg(workspace_id)::uuid
      AND sprint.created_at >= sqlc.arg(start_date)
      AND sprint.created_at <= sqlc.arg(end_date)
      AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR sprint.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
      AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR sprint.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
)
SELECT
    sprint_status.status,
    CAST(COUNT(*) AS int) AS count
FROM sprint_status
GROUP BY sprint_status.status
ORDER BY CASE sprint_status.status
    WHEN 'On Track' THEN 1
    WHEN 'At Risk' THEN 2
    WHEN 'Not Started' THEN 3
    WHEN 'Completed' THEN 4
    ELSE 5
END;

