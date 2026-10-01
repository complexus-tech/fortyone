-- name: ListStoryCompletionTimeline :many
WITH created_events AS (
    SELECT CAST(scoped.created_at AS date) AS event_date, 1 AS created, 0 AS completed
    FROM stories AS scoped
    LEFT JOIN statuses AS status ON status.status_id = scoped.status_id
    WHERE scoped.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND scoped.deleted_at IS NULL
      AND scoped.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR scoped.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR scoped.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR scoped.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR scoped.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND scoped.created_at >= sqlc.arg(start_date) AND scoped.created_at <= sqlc.arg(end_date)
), completed_events AS (
    SELECT CAST(scoped.completed_at AS date) AS event_date, 0 AS created, 1 AS completed
    FROM stories AS scoped
    LEFT JOIN statuses AS status ON status.status_id = scoped.status_id
    WHERE scoped.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND scoped.deleted_at IS NULL
      AND scoped.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR scoped.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR scoped.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR scoped.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR scoped.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND status.category = 'completed'
      AND scoped.completed_at >= sqlc.arg(start_date) AND scoped.completed_at <= sqlc.arg(end_date)
), events AS (
    SELECT * FROM created_events
    UNION ALL
    SELECT * FROM completed_events
)
SELECT
    events.event_date AS date,
    CAST(SUM(events.created) AS int) AS created,
    CAST(SUM(events.completed) AS int) AS completed
FROM events
GROUP BY events.event_date
ORDER BY date;

-- name: ListObjectiveProgressTimeline :many
SELECT
    CAST(objective.created_at AS date) AS date,
    CAST(COUNT(objective.objective_id) AS int) AS total_objectives,
    CAST(COUNT(objective.objective_id) FILTER (WHERE status.category = 'completed') AS int) AS completed_objectives
FROM objectives AS objective
LEFT JOIN objective_statuses AS status ON status.status_id = objective.status_id
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND objective.created_at >= sqlc.arg(start_date)
  AND objective.created_at <= sqlc.arg(end_date)
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR objective.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR objective.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
GROUP BY CAST(objective.created_at AS date)
ORDER BY date;

-- name: ListTeamVelocityTimeline :many
SELECT
    CAST(story.completed_at AS date) AS date,
    story.team_id,
    CAST(COUNT(story.id) AS int) AS velocity
FROM stories AS story
INNER JOIN statuses AS status
    ON status.status_id = story.status_id
   AND status.category = 'completed'
WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.completed_at >= sqlc.arg(start_date)
  AND story.completed_at <= sqlc.arg(end_date)
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR story.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR story.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
GROUP BY CAST(story.completed_at AS date), story.team_id
ORDER BY date, story.team_id;

-- name: ListKeyMetricsTimeline :many
-- Cycle time is elapsed time from the first recorded entry into a started
-- state to completion. Missing start history is excluded, never replaced by
-- creation or last-edit time. Samples are grouped by the actual completion day.
WITH created_events AS (
    SELECT CAST(story.created_at AS date) AS event_date, story.assignee_id, 1 AS created,
           CAST(NULL AS numeric) AS cycle_days
    FROM stories AS story
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND story.deleted_at IS NULL
      AND story.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR story.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR story.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND story.created_at >= sqlc.arg(start_date) AND story.created_at <= sqlc.arg(end_date)
), completed_events AS (
    SELECT CAST(story.completed_at AS date) AS event_date, story.assignee_id, 0 AS created,
           EXTRACT(EPOCH FROM (story.completed_at - started.started_at)) / 86400 AS cycle_days
    FROM stories AS story
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    LEFT JOIN LATERAL (
        SELECT MIN(activity.created_at AT TIME ZONE 'UTC') AS started_at
        FROM story_activities AS activity
        INNER JOIN statuses AS started_status
            ON CAST(started_status.status_id AS text) = COALESCE(activity.new_value #>> '{}', activity.current_value)
           AND started_status.category = 'started'
    WHERE activity.story_id = story.id
          AND activity.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
          AND activity.field_changed = 'status_id'
          AND activity.created_at AT TIME ZONE 'UTC' >= story.created_at
          AND activity.created_at AT TIME ZONE 'UTC' <= story.completed_at
    ) AS started ON TRUE
    WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND story.deleted_at IS NULL
      AND story.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR story.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR story.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND status.category = 'completed'
      AND story.completed_at >= sqlc.arg(start_date) AND story.completed_at <= sqlc.arg(end_date)
), events AS (
    SELECT * FROM created_events
    UNION ALL
    SELECT * FROM completed_events
)
SELECT
    events.event_date AS date,
    CAST(COUNT(DISTINCT events.assignee_id) AS int) AS active_users,
    CAST(SUM(events.created) AS double precision) AS stories_per_day,
    CAST(ROUND(COALESCE(AVG(events.cycle_days), 0), 2) AS double precision) AS avg_cycle_time,
    CAST(COUNT(events.cycle_days) AS int) AS cycle_time_samples
FROM events
GROUP BY events.event_date
ORDER BY date;
