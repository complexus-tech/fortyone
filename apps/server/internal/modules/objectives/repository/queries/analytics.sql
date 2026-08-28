-- Every analytics query repeats workspace_id even after the read precondition.
-- This makes accidental reuse without the authorization check tenant-safe.

-- name: GetObjectivePriorityBreakdown :many
WITH authorized_objective AS (
	SELECT objective.objective_id
	FROM public.objectives AS objective
	INNER JOIN public.workspace_members AS membership
		ON membership.workspace_id = objective.workspace_id
	   AND membership.user_id = sqlc.arg(actor_id)
	INNER JOIN public.team_members AS team_membership
		ON team_membership.team_id = objective.team_id
	   AND team_membership.user_id = membership.user_id
	INNER JOIN public.users AS actor
		ON actor.user_id = membership.user_id
	   AND actor.is_active = TRUE
	WHERE objective.objective_id = sqlc.arg(objective_id)
	  AND objective.workspace_id = sqlc.arg(workspace_id)
	  AND membership.role IN ('member', 'admin')
)
SELECT
    CAST(COALESCE(story.priority, 'No Priority') AS text) AS priority,
    CAST(COUNT(*) AS integer) AS count
FROM public.stories AS story
INNER JOIN public.objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
INNER JOIN authorized_objective AS authorized
	ON authorized.objective_id = objective.objective_id
WHERE story.objective_id = sqlc.arg(objective_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
GROUP BY story.priority
ORDER BY CASE COALESCE(story.priority, 'No Priority')
    WHEN 'Urgent' THEN 1
    WHEN 'High' THEN 2
    WHEN 'Medium' THEN 3
    WHEN 'Low' THEN 4
    ELSE 5
END;

-- name: GetObjectiveProgressBreakdown :one
WITH authorized_objective AS (
	SELECT objective.objective_id
	FROM public.objectives AS objective
	INNER JOIN public.workspace_members AS membership
		ON membership.workspace_id = objective.workspace_id
	   AND membership.user_id = sqlc.arg(actor_id)
	INNER JOIN public.team_members AS team_membership
		ON team_membership.team_id = objective.team_id
	   AND team_membership.user_id = membership.user_id
	INNER JOIN public.users AS actor
		ON actor.user_id = membership.user_id
	   AND actor.is_active = TRUE
	WHERE objective.objective_id = sqlc.arg(objective_id)
	  AND objective.workspace_id = sqlc.arg(workspace_id)
	  AND membership.role IN ('member', 'admin')
)
SELECT
    CAST(COUNT(*) AS integer) AS total,
    CAST(COUNT(*) FILTER (WHERE status.category = 'completed') AS integer) AS completed,
    CAST(COUNT(*) FILTER (WHERE status.category = 'started') AS integer) AS in_progress,
    CAST(COUNT(*) FILTER (WHERE status.category = 'unstarted') AS integer) AS todo,
    CAST(COUNT(*) FILTER (WHERE status.category = 'blocked') AS integer) AS blocked,
    CAST(COUNT(*) FILTER (WHERE status.category = 'cancelled') AS integer) AS cancelled
FROM public.stories AS story
INNER JOIN public.objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
INNER JOIN authorized_objective AS authorized
	ON authorized.objective_id = objective.objective_id
INNER JOIN public.statuses AS status
    ON status.status_id = story.status_id
   AND status.workspace_id = story.workspace_id
WHERE story.objective_id = sqlc.arg(objective_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL;

-- name: GetObjectiveTeamAllocation :many
WITH authorized_objective AS (
	SELECT objective.objective_id
	FROM public.objectives AS objective
	INNER JOIN public.workspace_members AS membership
		ON membership.workspace_id = objective.workspace_id
	   AND membership.user_id = sqlc.arg(actor_id)
	INNER JOIN public.team_members AS team_membership
		ON team_membership.team_id = objective.team_id
	   AND team_membership.user_id = membership.user_id
	INNER JOIN public.users AS actor
		ON actor.user_id = membership.user_id
	   AND actor.is_active = TRUE
	WHERE objective.objective_id = sqlc.arg(objective_id)
	  AND objective.workspace_id = sqlc.arg(workspace_id)
	  AND membership.role IN ('member', 'admin')
)
SELECT
    account.user_id,
    account.username,
    account.avatar_url,
    CAST(COUNT(story.id) AS integer) AS assigned,
    CAST(COUNT(story.id) FILTER (WHERE status.category = 'completed') AS integer) AS completed
FROM public.stories AS story
INNER JOIN public.objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
INNER JOIN authorized_objective AS authorized
	ON authorized.objective_id = objective.objective_id
INNER JOIN public.users AS account ON account.user_id = story.assignee_id
LEFT JOIN public.statuses AS status
    ON status.status_id = story.status_id
   AND status.workspace_id = story.workspace_id
WHERE story.objective_id = sqlc.arg(objective_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
  AND account.is_active = TRUE
GROUP BY account.user_id, account.username, account.avatar_url
ORDER BY assigned DESC, account.username, account.user_id;

-- name: GetObjectiveProgressChart :many
WITH authorized_objective AS (
	SELECT objective.objective_id
	FROM public.objectives AS objective
	INNER JOIN public.workspace_members AS membership
		ON membership.workspace_id = objective.workspace_id
	   AND membership.user_id = sqlc.arg(actor_id)
	INNER JOIN public.team_members AS team_membership
		ON team_membership.team_id = objective.team_id
	   AND team_membership.user_id = membership.user_id
	INNER JOIN public.users AS actor
		ON actor.user_id = membership.user_id
	   AND actor.is_active = TRUE
	WHERE objective.objective_id = sqlc.arg(objective_id)
	  AND objective.workspace_id = sqlc.arg(workspace_id)
	  AND membership.role IN ('member', 'admin')
), date_series AS (
    SELECT CAST(generate_series(
        CAST(sqlc.arg(chart_start) AS date),
        CAST(sqlc.arg(chart_end) AS date),
        INTERVAL '1 day'
    ) AS date) AS completion_date
), objective_story AS (
    SELECT story.*
    FROM public.stories AS story
    INNER JOIN public.objectives AS objective
        ON objective.objective_id = story.objective_id
       AND objective.workspace_id = story.workspace_id
	INNER JOIN authorized_objective AS authorized
		ON authorized.objective_id = objective.objective_id
    WHERE story.objective_id = sqlc.arg(objective_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
), story_status_change AS (
    SELECT
        story.id AS story_id,
        CAST(activity.created_at AS date) AS change_date,
        status.category,
        ROW_NUMBER() OVER (
            PARTITION BY story.id, CAST(activity.created_at AS date)
            ORDER BY activity.created_at DESC, activity.activity_id DESC
        ) AS row_number
    FROM objective_story AS story
    INNER JOIN public.story_activities AS activity ON activity.story_id = story.id
    INNER JOIN public.statuses AS status
        ON status.status_id = CASE
            WHEN activity.current_value ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
			THEN CAST(activity.current_value AS uuid)
            ELSE NULL
        END
       AND status.workspace_id = story.workspace_id
    WHERE activity.activity_type = 'update'
      AND activity.field_changed = 'status_id'
	  AND activity.workspace_id = story.workspace_id
	  AND activity.current_value ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
), latest_status_by_date AS (
    SELECT
        dates.completion_date,
        story.id AS story_id,
        COALESCE((
            SELECT change.category
            FROM story_status_change AS change
            WHERE change.story_id = story.id
              AND change.change_date <= dates.completion_date
              AND change.row_number = 1
            ORDER BY change.change_date DESC
            LIMIT 1
        ), 'unstarted') AS status_category
    FROM date_series AS dates
    CROSS JOIN objective_story AS story
    WHERE story.created_at < dates.completion_date + INTERVAL '1 day'
)
SELECT
    dates.completion_date,
    CAST(COALESCE(COUNT(status.story_id) FILTER (WHERE status.status_category = 'completed'), 0) AS integer) AS stories_completed,
    CAST(COALESCE(COUNT(status.story_id) FILTER (WHERE status.status_category = 'started'), 0) AS integer) AS stories_in_progress,
    CAST(COALESCE(COUNT(status.story_id), 0) AS integer) AS total_stories
FROM date_series AS dates
LEFT JOIN latest_status_by_date AS status ON status.completion_date = dates.completion_date
GROUP BY dates.completion_date
ORDER BY dates.completion_date;
