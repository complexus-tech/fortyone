-- name: GetSprintStoryBreakdown :one
SELECT
    COUNT(story.id)::bigint AS total,
    COUNT(story.id) FILTER (WHERE status.category = 'completed')::bigint AS completed,
    COUNT(story.id) FILTER (WHERE status.category = 'started')::bigint AS in_progress,
    COUNT(story.id) FILTER (WHERE status.category = 'unstarted')::bigint AS todo,
    COUNT(story.id) FILTER (WHERE status.category = 'paused')::bigint AS blocked,
    COUNT(story.id) FILTER (WHERE status.category = 'cancelled')::bigint AS cancelled
FROM sprints AS sprint
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = sprint.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
 AND workspace_member.role IN ('member', 'admin')
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS actor_team_member
  ON actor_team_member.team_id = sprint.team_id
 AND actor_team_member.user_id = workspace_member.user_id
LEFT JOIN stories AS story
  ON story.sprint_id = sprint.sprint_id
 AND story.workspace_id = sprint.workspace_id
 AND story.deleted_at IS NULL
 AND story.archived_at IS NULL
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE sprint.sprint_id = sqlc.arg(sprint_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id);

-- name: GetWorkspaceWorkingDays :one
SELECT COALESCE(
    (
        SELECT settings.working_days
        FROM workspace_settings AS settings
        WHERE settings.workspace_id = sqlc.arg(workspace_id)
    ),
    ARRAY[1, 2, 3, 4, 5]::smallint[]
)::smallint[] AS working_days
FROM workspace_members AS workspace_member
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
WHERE workspace_member.workspace_id = sqlc.arg(workspace_id)
  AND workspace_member.user_id = sqlc.arg(actor_id)
  AND workspace_member.role IN ('member', 'admin');

-- name: ListSprintTeamAllocation :many
SELECT
    member.user_id,
    member.username,
    COALESCE(member.avatar_url, '')::text AS avatar_url,
    COUNT(story.id)::bigint AS assigned,
    COUNT(story.id) FILTER (WHERE status.category = 'completed')::bigint AS completed
FROM sprints AS sprint
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = sprint.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
 AND workspace_member.role IN ('member', 'admin')
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS actor_team_member
  ON actor_team_member.team_id = sprint.team_id
 AND actor_team_member.user_id = workspace_member.user_id
JOIN users AS member ON member.is_active = TRUE
JOIN workspace_members AS member_workspace
  ON member_workspace.workspace_id = sprint.workspace_id
 AND member_workspace.user_id = member.user_id
 AND member_workspace.role IN ('member', 'admin')
JOIN team_members AS team_member
  ON team_member.team_id = sprint.team_id
 AND team_member.user_id = member.user_id
LEFT JOIN stories AS story
  ON story.assignee_id = member.user_id
 AND story.sprint_id = sprint.sprint_id
 AND story.workspace_id = sprint.workspace_id
 AND story.deleted_at IS NULL
 AND story.archived_at IS NULL
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE sprint.sprint_id = sqlc.arg(sprint_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND sprint.team_id = sqlc.arg(team_id)
GROUP BY member.user_id, member.username, member.avatar_url
ORDER BY assigned DESC, member.username, member.user_id;

-- name: ListSprintBurndownChanges :many
WITH params AS (
    SELECT
        sqlc.arg(workspace_id)::uuid AS workspace_id,
        sqlc.arg(sprint_id)::uuid AS sprint_id,
        sqlc.arg(sprint_id)::text AS sprint_id_text,
        sqlc.arg(start_date)::timestamp AS start_date,
        sqlc.arg(end_date)::timestamp AS end_date,
        sqlc.arg(actor_id)::uuid AS actor_id
),
authorized_sprint AS (
    SELECT sprint.sprint_id
    FROM sprints AS sprint
    JOIN params
      ON params.sprint_id = sprint.sprint_id
     AND params.workspace_id = sprint.workspace_id
    JOIN workspace_members AS workspace_member
      ON workspace_member.workspace_id = sprint.workspace_id
     AND workspace_member.user_id = params.actor_id
     AND workspace_member.role IN ('member', 'admin')
    JOIN users AS actor
      ON actor.user_id = workspace_member.user_id
     AND actor.is_active = TRUE
    JOIN team_members AS team_member
      ON team_member.team_id = sprint.team_id
     AND team_member.user_id = workspace_member.user_id
),
date_series AS (
    SELECT generate_series(
        (SELECT start_date::date FROM params),
        (SELECT end_date::date FROM params),
        INTERVAL '1 day'
    )::date AS burn_date
    FROM authorized_sprint
),
initial_stories_list AS (
    SELECT story.id
    FROM stories AS story
    JOIN params ON params.workspace_id = story.workspace_id
    WHERE story.sprint_id = params.sprint_id
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
    EXCEPT
    SELECT activity.story_id
    FROM story_activities AS activity
    JOIN stories AS story ON story.id = activity.story_id
    JOIN params ON params.workspace_id = story.workspace_id
    WHERE activity.field_changed = 'sprint_id'
      AND (
          NULLIF(activity.new_value #>> '{}', 'null')::uuid = params.sprint_id
          OR (activity.new_value IS NULL AND activity.current_value = params.sprint_id_text)
      )
      AND activity.created_at >= params.start_date
    UNION
    SELECT activity.story_id
    FROM story_activities AS activity
    JOIN stories AS story ON story.id = activity.story_id
    JOIN params ON params.workspace_id = story.workspace_id
    WHERE activity.field_changed = 'sprint_id'
      AND (
          NULLIF(activity.old_value #>> '{}', 'null')::uuid = params.sprint_id
          OR (
              activity.old_value IS NULL
              AND activity.activity_type = 'update'
              AND activity.current_value <> params.sprint_id_text
          )
      )
      AND activity.created_at >= params.start_date
),
daily_scope_changes AS (
    SELECT
        activity.created_at::date AS event_date,
        SUM(
            CASE
                WHEN NULLIF(activity.new_value #>> '{}', 'null')::uuid = params.sprint_id
                  OR (activity.new_value IS NULL AND activity.current_value = params.sprint_id_text) THEN 1
                WHEN NULLIF(activity.old_value #>> '{}', 'null')::uuid = params.sprint_id
                  OR (
                      activity.old_value IS NULL
                      AND activity.activity_type = 'update'
                      AND activity.current_value <> params.sprint_id_text
                  ) THEN -1
                ELSE 0
            END
        )::bigint AS delta
    FROM story_activities AS activity
    JOIN stories AS story ON story.id = activity.story_id
    JOIN params ON params.workspace_id = story.workspace_id
    WHERE activity.field_changed = 'sprint_id'
      AND activity.created_at >= params.start_date
      AND activity.created_at <= params.end_date + INTERVAL '1 day'
    GROUP BY activity.created_at::date
),
stories_ever_in_sprint AS (
    SELECT story.id
    FROM stories AS story
    JOIN params ON params.workspace_id = story.workspace_id
    WHERE story.sprint_id = params.sprint_id
    UNION
    SELECT activity.story_id
    FROM story_activities AS activity
    JOIN stories AS story ON story.id = activity.story_id
    JOIN params ON params.workspace_id = story.workspace_id
    WHERE activity.field_changed = 'sprint_id'
      AND (
          NULLIF(activity.old_value #>> '{}', 'null')::uuid = params.sprint_id
          OR (activity.old_value IS NULL AND activity.current_value <> params.sprint_id_text)
      )
),
daily_completion_changes AS (
    SELECT
        activity.created_at::date AS event_date,
        SUM(
            CASE
                WHEN new_status.category = 'completed'
                  AND (old_status.category IS NULL OR old_status.category <> 'completed') THEN 1
                WHEN old_status.category = 'completed'
                  AND (new_status.category IS NULL OR new_status.category <> 'completed') THEN -1
                ELSE 0
            END
        )::bigint AS delta
    FROM story_activities AS activity
    JOIN params ON TRUE
    JOIN stories_ever_in_sprint AS scoped_story ON scoped_story.id = activity.story_id
    LEFT JOIN statuses AS old_status ON NULLIF(activity.old_value #>> '{}', 'null')::uuid = old_status.status_id
    LEFT JOIN statuses AS new_status ON NULLIF(activity.new_value #>> '{}', 'null')::uuid = new_status.status_id
    WHERE activity.field_changed = 'status_id'
      AND activity.created_at >= params.start_date
      AND activity.created_at <= params.end_date + INTERVAL '1 day'
    GROUP BY activity.created_at::date
),
initial_completed_list AS (
    SELECT initial_story.id
    FROM initial_stories_list AS initial_story
    JOIN stories AS story ON story.id = initial_story.id
    JOIN statuses AS status ON status.status_id = story.status_id
    WHERE status.category = 'completed'
    EXCEPT
    SELECT activity.story_id
    FROM story_activities AS activity
    JOIN params ON TRUE
    JOIN stories_ever_in_sprint AS scoped_story ON scoped_story.id = activity.story_id
    JOIN statuses AS status ON NULLIF(activity.new_value #>> '{}', 'null')::uuid = status.status_id
    WHERE activity.field_changed = 'status_id'
      AND status.category = 'completed'
      AND activity.created_at >= params.start_date
    UNION
    SELECT activity.story_id
    FROM story_activities AS activity
    JOIN params ON TRUE
    JOIN stories_ever_in_sprint AS scoped_story ON scoped_story.id = activity.story_id
    JOIN statuses AS status ON NULLIF(activity.old_value #>> '{}', 'null')::uuid = status.status_id
    WHERE activity.field_changed = 'status_id'
      AND status.category = 'completed'
      AND activity.created_at >= params.start_date
),
initial_scope AS (
    SELECT COUNT(*)::bigint AS story_count FROM initial_stories_list
),
initial_completions AS (
    SELECT COUNT(*)::bigint AS completion_count FROM initial_completed_list
)
SELECT
    date_series.burn_date AS event_date,
    COALESCE(scope_changes.delta, 0)::bigint AS scope_delta,
    COALESCE(completion_changes.delta, 0)::bigint AS completion_delta,
    initial_scope.story_count AS initial_stories,
    initial_completions.completion_count AS initial_completed
FROM date_series
CROSS JOIN initial_scope
CROSS JOIN initial_completions
LEFT JOIN daily_scope_changes AS scope_changes ON scope_changes.event_date = date_series.burn_date
LEFT JOIN daily_completion_changes AS completion_changes ON completion_changes.event_date = date_series.burn_date
ORDER BY date_series.burn_date;
