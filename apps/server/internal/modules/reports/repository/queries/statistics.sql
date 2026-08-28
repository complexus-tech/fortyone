-- name: GetStoryStats :one
SELECT
    CAST(COUNT(*) FILTER (
        WHERE status.category = 'completed'
          AND story.assignee_id = sqlc.arg(actor_id)::uuid
    ) AS int) AS closed,
    CAST(COUNT(*) FILTER (
        WHERE story.end_date < CURRENT_DATE
          AND status.category NOT IN ('completed', 'cancelled')
          AND story.assignee_id = sqlc.arg(actor_id)::uuid
    ) AS int) AS overdue,
    CAST(COUNT(*) FILTER (
        WHERE status.category = 'started'
          AND story.assignee_id = sqlc.arg(actor_id)::uuid
    ) AS int) AS in_progress,
    CAST(COUNT(*) FILTER (WHERE story.reporter_id = sqlc.arg(actor_id)::uuid) AS int) AS created,
    CAST(COUNT(*) FILTER (WHERE story.assignee_id = sqlc.arg(actor_id)::uuid) AS int) AS assigned
FROM stories AS story
INNER JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date);

-- name: GetContributionStats :many
WITH dates AS (
    SELECT CAST(generate_series(
        CAST(sqlc.arg(start_date) AS date),
        CAST(sqlc.arg(end_date) AS date),
        INTERVAL '1 day'
    ) AS date) AS date
),
activity_counts AS (
    SELECT
        CAST(activity.created_at AS date) AS date,
        CAST(COUNT(*) AS int) AS contributions
    FROM story_activities AS activity
    WHERE activity.user_id = sqlc.arg(user_id)::uuid
      AND activity.workspace_id = sqlc.arg(workspace_id)::uuid
      AND activity.created_at >= sqlc.arg(start_date)
      AND activity.created_at <= sqlc.arg(end_date)
    GROUP BY CAST(activity.created_at AS date)
)
SELECT
    dates.date,
    CAST(COALESCE(activity_counts.contributions, 0) AS int) AS contributions
FROM dates
LEFT JOIN activity_counts ON activity_counts.date = dates.date
ORDER BY dates.date;

-- name: GetUserStats :one
SELECT
    CAST(COUNT(*) FILTER (WHERE story.assignee_id = sqlc.arg(user_id)::uuid) AS int) AS assigned_to_me,
    CAST(COUNT(*) FILTER (WHERE story.reporter_id = sqlc.arg(user_id)::uuid) AS int) AS created_by_me
FROM stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL;

-- name: GetStatusStats :many
WITH actor_teams AS (
    SELECT membership.team_id
    FROM team_members AS membership
    INNER JOIN teams AS team
        ON team.team_id = membership.team_id
       AND team.workspace_id = sqlc.arg(workspace_id)::uuid
    WHERE membership.user_id = sqlc.arg(actor_id)::uuid
)
SELECT
    status.name,
    CAST(COUNT(story.id) AS int) AS count
FROM statuses AS status
INNER JOIN stories AS story
    ON story.status_id = status.status_id
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
   AND story.team_id IN (SELECT actor_teams.team_id FROM actor_teams)
WHERE status.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (sqlc.narg(team_id)::uuid IS NULL OR story.team_id = sqlc.narg(team_id))
  AND (sqlc.narg(sprint_id)::uuid IS NULL OR story.sprint_id = sqlc.narg(sprint_id))
  AND (sqlc.narg(objective_id)::uuid IS NULL OR story.objective_id = sqlc.narg(objective_id))
GROUP BY status.status_id, status.name, status.order_index
ORDER BY status.order_index;

-- name: GetPriorityStats :many
WITH actor_teams AS (
    SELECT membership.team_id
    FROM team_members AS membership
    INNER JOIN teams AS team
        ON team.team_id = membership.team_id
       AND team.workspace_id = sqlc.arg(workspace_id)::uuid
    WHERE membership.user_id = sqlc.arg(actor_id)::uuid
)
SELECT
    CAST(COALESCE(CAST(story.priority AS text), 'No Priority') AS text) AS priority,
    CAST(COUNT(story.id) AS int) AS count
FROM stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.team_id IN (SELECT actor_teams.team_id FROM actor_teams)
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (sqlc.narg(team_id)::uuid IS NULL OR story.team_id = sqlc.narg(team_id))
  AND (sqlc.narg(sprint_id)::uuid IS NULL OR story.sprint_id = sqlc.narg(sprint_id))
  AND (sqlc.narg(objective_id)::uuid IS NULL OR story.objective_id = sqlc.narg(objective_id))
GROUP BY story.priority
ORDER BY CASE CAST(COALESCE(CAST(story.priority AS text), 'No Priority') AS text)
    WHEN 'Urgent' THEN 1
    WHEN 'High' THEN 2
    WHEN 'Medium' THEN 3
    WHEN 'Low' THEN 4
    WHEN 'No Priority' THEN 5
    ELSE 6
END;
