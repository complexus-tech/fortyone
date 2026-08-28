-- name: ListSprintAutomationTeams :many
SELECT
    settings.workspace_id,
    settings.team_id
FROM public.team_sprint_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
WHERE settings.auto_create_sprints = TRUE
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR settings.workspace_id > sqlc.arg(after_workspace_id)
      OR (
          settings.workspace_id = sqlc.arg(after_workspace_id)
          AND settings.team_id > sqlc.arg(after_team_id)
      )
  )
ORDER BY settings.workspace_id, settings.team_id
LIMIT CAST(sqlc.arg(batch_size) AS integer);

-- name: LockSprintAutomation :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'sprint-automation:'
            || CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text)
            || ':'
            || CAST(CAST(sqlc.arg(team_id) AS uuid) AS text),
        0
    )
);

-- name: CountUpcomingSprintsForAutomation :one
SELECT CAST(COUNT(*) AS integer) AS upcoming_count
FROM (
    SELECT sprint.sprint_id
    FROM public.sprints AS sprint
    INNER JOIN public.teams AS team
        ON team.team_id = sprint.team_id
       AND team.workspace_id = sprint.workspace_id
    WHERE sprint.team_id = sqlc.arg(team_id)
      AND sprint.workspace_id = sqlc.arg(workspace_id)
      AND sprint.start_date > CAST(sqlc.arg(schedule_date) AS date)
    ORDER BY sprint.start_date, sprint.sprint_id
    LIMIT CAST(sqlc.arg(upcoming_target) AS integer)
) AS upcoming_sprints;

-- name: GetSprintAutomationScheduleBoundary :one
SELECT sprint.end_date
FROM public.sprints AS sprint
INNER JOIN public.teams AS team
    ON team.team_id = sprint.team_id
   AND team.workspace_id = sprint.workspace_id
WHERE sprint.team_id = sqlc.arg(team_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND sprint.end_date >= CAST(sqlc.arg(schedule_date) AS date)
ORDER BY sprint.end_date DESC, sprint.sprint_id
LIMIT 1
FOR SHARE OF sprint;

-- name: CreateAutomatedSprint :one
INSERT INTO public.sprints (
    name,
    team_id,
    workspace_id,
    start_date,
    end_date,
    schedule_managed_by_automation,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(name),
    sqlc.arg(team_id),
    sqlc.arg(workspace_id),
    CAST(sqlc.arg(start_date) AS date),
    CAST(sqlc.arg(end_date) AS date),
    TRUE,
    CAST(sqlc.arg(created_at) AS timestamptz),
    CAST(sqlc.arg(updated_at) AS timestamptz)
)
RETURNING sprint_id;

-- name: AdvanceSprintAutomationCounter :execrows
UPDATE public.team_sprint_settings AS settings
SET
    last_auto_sprint_number = settings.next_auto_sprint_number + CAST(sqlc.arg(created_count) AS integer) - 1,
    next_auto_sprint_number = settings.next_auto_sprint_number + CAST(sqlc.arg(created_count) AS integer),
    updated_at = CAST(sqlc.arg(updated_at) AS timestamptz) AT TIME ZONE 'UTC'
FROM public.teams AS team
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id)
  AND settings.auto_create_sprints = TRUE
  AND settings.next_auto_sprint_number = CAST(sqlc.arg(expected_next_number) AS integer)
  AND team.team_id = settings.team_id
  AND team.workspace_id = settings.workspace_id;

-- name: ListSprintAutomationInactivityCandidates :many
SELECT
    settings.workspace_id,
    settings.team_id
FROM public.team_sprint_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
WHERE settings.auto_create_sprints = TRUE
  AND team.created_at <= sqlc.arg(team_created_before)
  AND settings.updated_at <= (
      CAST(sqlc.arg(settings_updated_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR settings.workspace_id > sqlc.arg(after_workspace_id)
      OR (
          settings.workspace_id = sqlc.arg(after_workspace_id)
          AND settings.team_id > sqlc.arg(after_team_id)
      )
  )
ORDER BY settings.workspace_id, settings.team_id
LIMIT CAST(sqlc.arg(batch_size) AS integer);

-- name: GetSprintAutomationInactivitySnapshot :one
SELECT
    team.name,
    team.created_at AS team_created_at,
    settings.updated_at AS settings_updated_at,
    CAST(latest_story.created_at IS NOT NULL AS boolean) AS has_latest_human_story,
    COALESCE(latest_story.created_at, team.created_at) AS latest_human_story_at,
    CAST(latest_sprint_change.created_at IS NOT NULL AS boolean) AS has_latest_human_sprint_change,
    COALESCE(latest_sprint_change.created_at, team.created_at) AS latest_human_sprint_change_at
FROM public.team_sprint_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
LEFT JOIN LATERAL (
        SELECT story.created_at
        FROM public.stories AS story
        INNER JOIN public.users AS reporter
            ON reporter.user_id = story.reporter_id
           AND reporter.is_system = FALSE
        WHERE story.team_id = settings.team_id
          AND story.workspace_id = settings.workspace_id
          AND story.deleted_at IS NULL
          AND story.archived_at IS NULL
          AND story.is_draft = FALSE
        ORDER BY story.created_at DESC, story.id
        LIMIT 1
    ) AS latest_story ON TRUE
LEFT JOIN LATERAL (
        SELECT activity.created_at
        FROM public.story_activities AS activity
        INNER JOIN public.stories AS story
            ON story.id = activity.story_id
           AND story.team_id = settings.team_id
           AND story.workspace_id = settings.workspace_id
        INNER JOIN public.users AS actor
            ON actor.user_id = activity.user_id
           AND actor.is_system = FALSE
        WHERE activity.field_changed = 'sprint_id'
          AND story.deleted_at IS NULL
          AND story.archived_at IS NULL
        ORDER BY activity.created_at DESC, activity.activity_id DESC
        LIMIT 1
    ) AS latest_sprint_change ON TRUE
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id)
  AND settings.auto_create_sprints = TRUE;

-- name: DisableSprintAutomationIfInactive :execrows
UPDATE public.team_sprint_settings AS settings
SET
    auto_create_sprints = FALSE,
    auto_create_disabled_at = CAST(sqlc.arg(disabled_at) AS timestamptz),
    auto_create_disabled_reason = sqlc.arg(disabled_reason),
    updated_at = CAST(sqlc.arg(disabled_at) AS timestamptz) AT TIME ZONE 'UTC'
FROM public.teams AS team
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id)
  AND settings.auto_create_sprints = TRUE
  AND team.team_id = settings.team_id
  AND team.workspace_id = settings.workspace_id
  AND team.created_at <= sqlc.arg(team_created_before)
  AND settings.updated_at <= (
      CAST(sqlc.arg(settings_updated_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.stories AS story
      INNER JOIN public.users AS reporter
          ON reporter.user_id = story.reporter_id
         AND reporter.is_system = FALSE
      WHERE story.team_id = settings.team_id
        AND story.workspace_id = settings.workspace_id
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND story.is_draft = FALSE
        AND story.created_at >= sqlc.arg(activity_before)
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_activities AS activity
      INNER JOIN public.stories AS story
          ON story.id = activity.story_id
         AND story.team_id = settings.team_id
         AND story.workspace_id = settings.workspace_id
      INNER JOIN public.users AS actor
          ON actor.user_id = activity.user_id
         AND actor.is_system = FALSE
      WHERE activity.field_changed = 'sprint_id'
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND activity.created_at >= sqlc.arg(activity_before)
  );

-- name: InsertSprintAutomationAuditEvent :exec
INSERT INTO public.audit_events (
    workspace_id,
    team_id,
    actor_type,
    actor_id,
    entity_type,
    entity_id,
    event_type,
    metadata,
    created_at
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    'automation',
    NULL,
    sqlc.arg(entity_type),
    sqlc.arg(entity_id),
    sqlc.arg(event_type),
    CAST(sqlc.arg(metadata) AS jsonb),
    CAST(sqlc.arg(created_at) AS timestamptz)
);
