-- Objective reads always carry both the tenant and the current actor. The
-- internal flag exists only for trusted background consumers that already
-- hold a tenant-scoped objective identifier.

-- name: ListObjectives :many
WITH visible_objectives AS (
    SELECT objective.*
    FROM public.objectives AS objective
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = objective.workspace_id
       AND membership.user_id = sqlc.arg(actor_id)
    INNER JOIN public.team_members AS team_membership
        ON team_membership.team_id = objective.team_id
       AND team_membership.user_id = sqlc.arg(actor_id)
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE objective.workspace_id = sqlc.arg(workspace_id)
	  AND membership.role IN ('member', 'admin')
      AND (
          CAST(sqlc.narg(objective_id) AS uuid) IS NULL
          OR objective.objective_id = CAST(sqlc.narg(objective_id) AS uuid)
      )
      AND (
          CAST(sqlc.narg(team_id) AS uuid) IS NULL
          OR objective.team_id = CAST(sqlc.narg(team_id) AS uuid)
      )
      AND (
          CAST(sqlc.arg(search) AS text) = ''
          OR objective.name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
      )
), story_schedule AS (
    SELECT
        story.id,
        story.sequence_id,
        story.title,
        story.objective_id,
        CASE
            WHEN MIN(block.start_at) IS NULL THEN story.start_date
            WHEN story.start_date IS NULL THEN CAST(MIN(block.start_at) AS date)
            ELSE LEAST(story.start_date, CAST(MIN(block.start_at) AS date))
        END AS forecast_start_date,
        CASE
            WHEN MAX(block.end_at) IS NULL THEN story.end_date
            WHEN story.end_date IS NULL THEN CAST(MAX(block.end_at) AS date)
            ELSE GREATEST(story.end_date, CAST(MAX(block.end_at) AS date))
        END AS forecast_end_date,
        CASE
            WHEN MAX(block.end_at) IS NOT NULL
                AND (story.end_date IS NULL OR CAST(MAX(block.end_at) AS date) >= story.end_date)
            THEN 'calendar'
            ELSE 'planning'
        END AS forecast_source
    FROM public.stories AS story
    INNER JOIN public.statuses AS status
        ON status.status_id = story.status_id
       AND status.workspace_id = story.workspace_id
    LEFT JOIN public.calendar_schedule_blocks AS block
        ON block.story_id = story.id
       AND block.workspace_id = story.workspace_id
       AND block.block_type = 'work'
    INNER JOIN visible_objectives AS objective
        ON objective.objective_id = story.objective_id
       AND objective.workspace_id = story.workspace_id
    WHERE story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND status.category NOT IN ('completed', 'cancelled')
    GROUP BY story.id, story.sequence_id, story.title, story.objective_id, story.start_date, story.end_date
), ranked_story_schedule AS (
    SELECT
        story_schedule.*,
        ROW_NUMBER() OVER (
            PARTITION BY objective_id
            ORDER BY forecast_end_date DESC, sequence_id DESC, id
        ) AS forecast_rank
    FROM story_schedule
    WHERE forecast_end_date IS NOT NULL
), objective_schedule AS (
    SELECT
        objective_id,
        MIN(forecast_start_date) AS forecast_start_date,
        MAX(forecast_end_date) AS forecast_end_date
    FROM story_schedule
    GROUP BY objective_id
), story_stats AS (
    SELECT
        objective.objective_id,
        COUNT(story.id) AS total,
        COUNT(story.id) FILTER (WHERE status.category = 'cancelled') AS cancelled,
        COUNT(story.id) FILTER (WHERE status.category = 'completed') AS completed,
        COUNT(story.id) FILTER (WHERE status.category = 'started') AS started,
        COUNT(story.id) FILTER (WHERE status.category = 'unstarted') AS unstarted,
        COUNT(story.id) FILTER (WHERE status.category = 'backlog') AS backlog
    FROM visible_objectives AS objective
    LEFT JOIN public.stories AS story
        ON story.objective_id = objective.objective_id
       AND story.workspace_id = objective.workspace_id
       AND story.deleted_at IS NULL
       AND story.archived_at IS NULL
    LEFT JOIN public.statuses AS status ON status.status_id = story.status_id
       AND status.workspace_id = story.workspace_id
    GROUP BY objective.objective_id
), key_result_stats AS (
    SELECT key_result.objective_id, COUNT(*) AS total
    FROM public.key_results AS key_result
    INNER JOIN visible_objectives AS objective ON objective.objective_id = key_result.objective_id
    GROUP BY key_result.objective_id
)
SELECT
    objective.objective_id,
    objective.sequence_id,
    objective.name,
    objective.description,
    objective.short_summary,
    objective.lead_user_id,
    objective.team_id,
    objective.workspace_id,
    objective.start_date,
    objective.end_date,
    objective.is_private,
    objective.created_at,
    objective.updated_at,
    objective.status_id,
    objective.priority,
    CAST(COALESCE(CAST(objective.health AS text), '') AS text) AS health,
    objective.color,
    objective.created_by,
    CAST(COALESCE(key_result_stats.total, 0) AS integer) AS key_result_count,
    CAST(COALESCE(story_stats.total, 0) AS integer) AS total_stories,
    CAST(COALESCE(story_stats.cancelled, 0) AS integer) AS cancelled_stories,
    CAST(COALESCE(story_stats.completed, 0) AS integer) AS completed_stories,
    CAST(COALESCE(story_stats.started, 0) AS integer) AS started_stories,
    CAST(COALESCE(story_stats.unstarted, 0) AS integer) AS unstarted_stories,
    CAST(COALESCE(story_stats.backlog, 0) AS integer) AS backlog_stories,
    CAST(COALESCE(CAST(objective_schedule.forecast_start_date AS date), CAST('1970-01-01' AS date)) AS date) AS forecast_start_date,
    CAST(objective_schedule.forecast_start_date IS NOT NULL AS boolean) AS has_forecast_start_date,
    CAST(COALESCE(CAST(objective_schedule.forecast_end_date AS date), CAST('1970-01-01' AS date)) AS date) AS forecast_end_date,
    CAST(objective_schedule.forecast_end_date IS NOT NULL AS boolean) AS has_forecast_end_date,
    ranked_story_schedule.id AS forecast_cause_id,
    ranked_story_schedule.sequence_id AS forecast_cause_sequence_id,
    ranked_story_schedule.title AS forecast_cause_title,
    ranked_story_schedule.forecast_source AS forecast_cause_source
FROM visible_objectives AS objective
LEFT JOIN story_stats ON story_stats.objective_id = objective.objective_id
LEFT JOIN key_result_stats ON key_result_stats.objective_id = objective.objective_id
LEFT JOIN objective_schedule ON objective_schedule.objective_id = objective.objective_id
LEFT JOIN ranked_story_schedule
    ON ranked_story_schedule.objective_id = objective.objective_id
   AND ranked_story_schedule.forecast_rank = 1
ORDER BY objective.created_at DESC, objective.objective_id
LIMIT NULLIF(CAST(sqlc.arg(result_limit) AS integer), 0)
OFFSET CAST(sqlc.arg(result_offset) AS integer);

-- name: GetObjective :one
WITH visible_objective AS (
    SELECT objective.*
    FROM public.objectives AS objective
    WHERE objective.objective_id = sqlc.arg(objective_id)
      AND objective.workspace_id = sqlc.arg(workspace_id)
      AND (
          CAST(sqlc.arg(internal_access) AS boolean)
          OR EXISTS (
              SELECT 1
              FROM public.workspace_members AS membership
              INNER JOIN public.team_members AS team_membership
                  ON team_membership.team_id = objective.team_id
                 AND team_membership.user_id = membership.user_id
              INNER JOIN public.users AS actor
                  ON actor.user_id = membership.user_id
                 AND actor.is_active = TRUE
              WHERE membership.workspace_id = objective.workspace_id
                AND membership.user_id = sqlc.arg(actor_id)
				AND membership.role IN ('member', 'admin')
          )
      )
), story_schedule AS (
    SELECT
        story.id,
        story.sequence_id,
        story.title,
        story.objective_id,
        CASE
            WHEN MIN(block.start_at) IS NULL THEN story.start_date
            WHEN story.start_date IS NULL THEN CAST(MIN(block.start_at) AS date)
            ELSE LEAST(story.start_date, CAST(MIN(block.start_at) AS date))
        END AS forecast_start_date,
        CASE
            WHEN MAX(block.end_at) IS NULL THEN story.end_date
            WHEN story.end_date IS NULL THEN CAST(MAX(block.end_at) AS date)
            ELSE GREATEST(story.end_date, CAST(MAX(block.end_at) AS date))
        END AS forecast_end_date,
        CASE
            WHEN MAX(block.end_at) IS NOT NULL
                AND (story.end_date IS NULL OR CAST(MAX(block.end_at) AS date) >= story.end_date)
            THEN 'calendar'
            ELSE 'planning'
        END AS forecast_source
    FROM public.stories AS story
    INNER JOIN public.statuses AS status
        ON status.status_id = story.status_id
       AND status.workspace_id = story.workspace_id
    LEFT JOIN public.calendar_schedule_blocks AS block
        ON block.story_id = story.id
       AND block.workspace_id = story.workspace_id
       AND block.block_type = 'work'
    INNER JOIN visible_objective AS objective
        ON objective.objective_id = story.objective_id
       AND objective.workspace_id = story.workspace_id
    WHERE story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND status.category NOT IN ('completed', 'cancelled')
    GROUP BY story.id, story.sequence_id, story.title, story.objective_id, story.start_date, story.end_date
), ranked_story_schedule AS (
    SELECT
        story_schedule.*,
        ROW_NUMBER() OVER (
            PARTITION BY objective_id
            ORDER BY forecast_end_date DESC, sequence_id DESC, id
        ) AS forecast_rank
    FROM story_schedule
    WHERE forecast_end_date IS NOT NULL
), objective_schedule AS (
    SELECT objective_id, MIN(forecast_start_date) AS forecast_start_date, MAX(forecast_end_date) AS forecast_end_date
    FROM story_schedule
    GROUP BY objective_id
), story_stats AS (
    SELECT
        objective.objective_id,
        COUNT(story.id) AS total,
        COUNT(story.id) FILTER (WHERE status.category = 'cancelled') AS cancelled,
        COUNT(story.id) FILTER (WHERE status.category = 'completed') AS completed,
        COUNT(story.id) FILTER (WHERE status.category = 'started') AS started,
        COUNT(story.id) FILTER (WHERE status.category = 'unstarted') AS unstarted,
        COUNT(story.id) FILTER (WHERE status.category = 'backlog') AS backlog
    FROM visible_objective AS objective
    LEFT JOIN public.stories AS story
        ON story.objective_id = objective.objective_id
       AND story.workspace_id = objective.workspace_id
       AND story.deleted_at IS NULL
       AND story.archived_at IS NULL
    LEFT JOIN public.statuses AS status ON status.status_id = story.status_id
       AND status.workspace_id = story.workspace_id
    GROUP BY objective.objective_id
), key_result_stats AS (
    SELECT key_result.objective_id, COUNT(*) AS total
    FROM public.key_results AS key_result
    INNER JOIN visible_objective AS objective ON objective.objective_id = key_result.objective_id
    GROUP BY key_result.objective_id
)
SELECT
    objective.objective_id,
    objective.sequence_id,
    objective.name,
    objective.description,
    objective.short_summary,
    objective.lead_user_id,
    objective.team_id,
    objective.workspace_id,
    objective.start_date,
    objective.end_date,
    objective.is_private,
    objective.created_at,
    objective.updated_at,
    objective.status_id,
    objective.priority,
    CAST(COALESCE(CAST(objective.health AS text), '') AS text) AS health,
    objective.color,
    objective.created_by,
    CAST(COALESCE(key_result_stats.total, 0) AS integer) AS key_result_count,
    CAST(COALESCE(story_stats.total, 0) AS integer) AS total_stories,
    CAST(COALESCE(story_stats.cancelled, 0) AS integer) AS cancelled_stories,
    CAST(COALESCE(story_stats.completed, 0) AS integer) AS completed_stories,
    CAST(COALESCE(story_stats.started, 0) AS integer) AS started_stories,
    CAST(COALESCE(story_stats.unstarted, 0) AS integer) AS unstarted_stories,
    CAST(COALESCE(story_stats.backlog, 0) AS integer) AS backlog_stories,
    CAST(COALESCE(CAST(objective_schedule.forecast_start_date AS date), CAST('1970-01-01' AS date)) AS date) AS forecast_start_date,
    CAST(objective_schedule.forecast_start_date IS NOT NULL AS boolean) AS has_forecast_start_date,
    CAST(COALESCE(CAST(objective_schedule.forecast_end_date AS date), CAST('1970-01-01' AS date)) AS date) AS forecast_end_date,
    CAST(objective_schedule.forecast_end_date IS NOT NULL AS boolean) AS has_forecast_end_date,
    ranked_story_schedule.id AS forecast_cause_id,
    ranked_story_schedule.sequence_id AS forecast_cause_sequence_id,
    ranked_story_schedule.title AS forecast_cause_title,
    ranked_story_schedule.forecast_source AS forecast_cause_source
FROM visible_objective AS objective
LEFT JOIN story_stats ON story_stats.objective_id = objective.objective_id
LEFT JOIN key_result_stats ON key_result_stats.objective_id = objective.objective_id
LEFT JOIN objective_schedule ON objective_schedule.objective_id = objective.objective_id
LEFT JOIN ranked_story_schedule
    ON ranked_story_schedule.objective_id = objective.objective_id
   AND ranked_story_schedule.forecast_rank = 1;

-- name: CanReadObjective :one
SELECT EXISTS (
    SELECT 1
    FROM public.objectives AS objective
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = objective.workspace_id
       AND membership.user_id = sqlc.arg(actor_id)
    INNER JOIN public.team_members AS team_membership
        ON team_membership.team_id = objective.team_id
       AND team_membership.user_id = sqlc.arg(actor_id)
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE objective.objective_id = sqlc.arg(objective_id)
      AND objective.workspace_id = sqlc.arg(workspace_id)
	  AND membership.role IN ('member', 'admin')
) AS can_read;
