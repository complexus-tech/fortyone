-- name: ListSprints :many
WITH story_stats AS (
    SELECT
        story.sprint_id,
        COUNT(story.id) AS total_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'cancelled') AS cancelled_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'completed') AS completed_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'started') AS started_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'unstarted') AS unstarted_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'backlog') AS backlog_stories
    FROM stories AS story
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.sprint_id IS NOT NULL
    GROUP BY story.sprint_id
)
SELECT
    sprint.sprint_id,
    sprint.name,
    sprint.goal,
    sprint.objective_id,
    sprint.team_id,
    sprint.workspace_id,
    sprint.start_date,
    sprint.end_date,
    sprint.created_at,
    sprint.updated_at,
    sprint.schedule_managed_by_automation,
    COALESCE(stats.total_stories, 0)::bigint AS total_stories,
    COALESCE(stats.cancelled_stories, 0)::bigint AS cancelled_stories,
    COALESCE(stats.completed_stories, 0)::bigint AS completed_stories,
    COALESCE(stats.started_stories, 0)::bigint AS started_stories,
    COALESCE(stats.unstarted_stories, 0)::bigint AS unstarted_stories,
    COALESCE(stats.backlog_stories, 0)::bigint AS backlog_stories
FROM sprints AS sprint
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = sprint.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS team_member
  ON team_member.team_id = sprint.team_id
 AND team_member.user_id = workspace_member.user_id
LEFT JOIN story_stats AS stats ON stats.sprint_id = sprint.sprint_id
WHERE sprint.workspace_id = sqlc.arg(workspace_id)
  AND workspace_member.role IN ('member', 'admin')
  AND (sqlc.narg(sprint_id)::uuid IS NULL OR sprint.sprint_id = sqlc.narg(sprint_id))
  AND (sqlc.narg(objective_id)::uuid IS NULL OR sprint.objective_id = sqlc.narg(objective_id))
  AND (sqlc.narg(team_id)::uuid IS NULL OR sprint.team_id = sqlc.narg(team_id))
  AND (sqlc.narg(search)::text IS NULL OR sprint.name ILIKE '%' || sqlc.narg(search)::text || '%')
ORDER BY sprint.end_date DESC, sprint.sprint_id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: GetSprintByID :one
SELECT
    sprint.sprint_id,
    sprint.name,
    sprint.goal,
    sprint.objective_id,
    sprint.team_id,
    sprint.workspace_id,
    sprint.start_date,
    sprint.end_date,
    sprint.created_at,
    sprint.updated_at,
    sprint.schedule_managed_by_automation
FROM sprints AS sprint
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = sprint.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS team_member
  ON team_member.team_id = sprint.team_id
 AND team_member.user_id = workspace_member.user_id
WHERE sprint.sprint_id = sqlc.arg(sprint_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND workspace_member.role IN ('member', 'admin');

-- name: ListRunningSprints :many
WITH story_stats AS (
    SELECT
        story.sprint_id,
        COUNT(story.id) AS total_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'cancelled') AS cancelled_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'completed') AS completed_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'started') AS started_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'unstarted') AS unstarted_stories,
        COUNT(story.id) FILTER (WHERE status.category = 'backlog') AS backlog_stories
    FROM stories AS story
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.sprint_id IS NOT NULL
    GROUP BY story.sprint_id
)
SELECT
    sprint.sprint_id,
    sprint.name,
    sprint.goal,
    sprint.objective_id,
    sprint.team_id,
    sprint.workspace_id,
    sprint.start_date,
    sprint.end_date,
    sprint.created_at,
    sprint.updated_at,
    sprint.schedule_managed_by_automation,
    COALESCE(stats.total_stories, 0)::bigint AS total_stories,
    COALESCE(stats.cancelled_stories, 0)::bigint AS cancelled_stories,
    COALESCE(stats.completed_stories, 0)::bigint AS completed_stories,
    COALESCE(stats.started_stories, 0)::bigint AS started_stories,
    COALESCE(stats.unstarted_stories, 0)::bigint AS unstarted_stories,
    COALESCE(stats.backlog_stories, 0)::bigint AS backlog_stories
FROM sprints AS sprint
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = sprint.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS team_member
  ON team_member.team_id = sprint.team_id
 AND team_member.user_id = workspace_member.user_id
JOIN team_sprint_settings AS settings
  ON settings.team_id = sprint.team_id
 AND settings.workspace_id = sprint.workspace_id
 AND settings.auto_create_sprints = TRUE
LEFT JOIN story_stats AS stats ON stats.sprint_id = sprint.sprint_id
WHERE sprint.workspace_id = sqlc.arg(workspace_id)
  AND workspace_member.role IN ('member', 'admin')
  AND sprint.start_date <= sqlc.arg(today)::date
  AND sprint.end_date >= sqlc.arg(today)::date
ORDER BY sprint.end_date DESC, sprint.sprint_id DESC;
