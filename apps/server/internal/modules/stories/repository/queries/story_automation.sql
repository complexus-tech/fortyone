-- Story automation is application-clocked and transactionally fenced. Each
-- transition query removes its rows from the eligible set, so a transaction
-- retry cannot emit duplicate activities or audit events.

-- name: LockStoryAutomation :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(CAST(sqlc.arg(lock_name) AS text), 0)
);

-- name: ArchiveEligibleStoriesBatch :many
WITH candidates AS MATERIALIZED (
    SELECT
        story.id,
        story.workspace_id,
        story.team_id,
        story.status_id AS expected_status_id
    FROM public.stories AS story
    INNER JOIN public.statuses AS current_status
        ON current_status.status_id = story.status_id
       AND current_status.workspace_id = story.workspace_id
       AND current_status.team_id = story.team_id
    INNER JOIN public.team_story_automation_settings AS settings
        ON settings.workspace_id = story.workspace_id
       AND settings.team_id = story.team_id
    WHERE current_status.category IN ('completed', 'cancelled')
      AND settings.auto_archive_enabled = TRUE
      AND story.updated_at < timezone(
          'UTC',
          CAST(
              CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
              AS date
          ) - make_interval(months => settings.auto_archive_months)
      )
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
    ORDER BY story.updated_at, story.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF story SKIP LOCKED
)
UPDATE public.stories AS story
SET archived_at = CAST(sqlc.arg(as_of) AS timestamptz)
FROM candidates AS candidate
WHERE story.id = candidate.id
  AND story.workspace_id = candidate.workspace_id
  AND story.team_id = candidate.team_id
  AND story.status_id = candidate.expected_status_id
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
RETURNING story.id;

-- name: CloseEligibleStoriesBatch :many
WITH candidates AS MATERIALIZED (
    SELECT
        story.id,
        story.workspace_id,
        story.team_id,
        story.status_id AS expected_status_id,
        closed_status.status_id AS new_status_id
    FROM public.stories AS story
    INNER JOIN public.statuses AS current_status
        ON current_status.status_id = story.status_id
       AND current_status.workspace_id = story.workspace_id
       AND current_status.team_id = story.team_id
    INNER JOIN public.team_story_automation_settings AS settings
        ON settings.workspace_id = story.workspace_id
       AND settings.team_id = story.team_id
    CROSS JOIN LATERAL (
        SELECT candidate_status.status_id
        FROM public.statuses AS candidate_status
        WHERE candidate_status.workspace_id = story.workspace_id
          AND candidate_status.team_id = story.team_id
          AND candidate_status.category = 'cancelled'
        ORDER BY
            candidate_status.is_default DESC,
            candidate_status.order_index NULLS LAST,
            candidate_status.status_id
        LIMIT 1
    ) AS closed_status
    WHERE current_status.category IN ('unstarted', 'started')
      AND settings.auto_close_inactive_enabled = TRUE
      AND story.updated_at < timezone(
          'UTC',
          CAST(
              CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
              AS date
          ) - make_interval(months => settings.auto_close_inactive_months)
      )
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
    ORDER BY story.updated_at, story.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF story SKIP LOCKED
)
UPDATE public.stories AS story
SET
    status_id = candidate.new_status_id,
    updated_at = CAST(sqlc.arg(as_of) AS timestamptz)
FROM candidates AS candidate
WHERE story.id = candidate.id
  AND story.workspace_id = candidate.workspace_id
  AND story.team_id = candidate.team_id
  AND story.status_id = candidate.expected_status_id
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
RETURNING
    story.id,
    candidate.workspace_id,
    candidate.team_id,
    candidate.new_status_id AS status_id;

-- name: InsertStoryAutoCloseActivities :execrows
WITH input AS (
    SELECT
        requested_story.story_id,
        requested_workspace.workspace_id,
        requested_team.team_id,
        requested_status.status_id
    FROM unnest(CAST(sqlc.arg(story_ids) AS uuid[]))
        WITH ORDINALITY AS requested_story(story_id, position)
    INNER JOIN unnest(CAST(sqlc.arg(workspace_ids) AS uuid[]))
        WITH ORDINALITY AS requested_workspace(workspace_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(team_ids) AS uuid[]))
        WITH ORDINALITY AS requested_team(team_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(status_ids) AS uuid[]))
        WITH ORDINALITY AS requested_status(status_id, position)
        USING (position)
)
INSERT INTO public.story_activities (
    story_id,
    user_id,
    activity_type,
    field_changed,
    current_value,
    reason,
    workspace_id,
    created_at
)
SELECT
    input.story_id,
    sqlc.arg(system_user_id),
    'update',
    'status_id',
    CAST(input.status_id AS text),
    sqlc.arg(reason),
    input.workspace_id,
    CAST(
        CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
        AS timestamp
    )
FROM input
INNER JOIN public.stories AS story
    ON story.id = input.story_id
   AND story.workspace_id = input.workspace_id
   AND story.team_id = input.team_id
   AND story.status_id = input.status_id
INNER JOIN public.statuses AS closed_status
    ON closed_status.status_id = input.status_id
   AND closed_status.workspace_id = input.workspace_id
   AND closed_status.team_id = input.team_id
   AND closed_status.category = 'cancelled'
WHERE story.deleted_at IS NULL
  AND story.archived_at IS NULL;

-- name: MigrateEligibleSprintStoriesBatch :many
WITH candidates AS MATERIALIZED (
    SELECT
        story.id,
        story.workspace_id,
        story.team_id,
        story.status_id AS expected_status_id,
        ended_sprint.sprint_id AS previous_sprint_id,
        next_sprint.sprint_id AS new_sprint_id
    FROM public.stories AS story
    INNER JOIN public.sprints AS ended_sprint
        ON ended_sprint.sprint_id = story.sprint_id
       AND ended_sprint.workspace_id = story.workspace_id
       AND ended_sprint.team_id = story.team_id
    INNER JOIN public.team_sprint_settings AS settings
        ON settings.workspace_id = story.workspace_id
       AND settings.team_id = story.team_id
    INNER JOIN public.statuses AS current_status
        ON current_status.status_id = story.status_id
       AND current_status.workspace_id = story.workspace_id
       AND current_status.team_id = story.team_id
    CROSS JOIN LATERAL (
        SELECT candidate_sprint.sprint_id
        FROM public.sprints AS candidate_sprint
        WHERE candidate_sprint.workspace_id = ended_sprint.workspace_id
          AND candidate_sprint.team_id = ended_sprint.team_id
          AND candidate_sprint.start_date > ended_sprint.end_date
          AND candidate_sprint.start_date <= (
              CAST(
                  CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
                  AS date
              ) + 30
          )
        ORDER BY candidate_sprint.start_date, candidate_sprint.sprint_id
        LIMIT 1
    ) AS next_sprint
    WHERE settings.move_incomplete_stories_enabled = TRUE
      AND ended_sprint.end_date >= (
          CAST(
              CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
              AS date
          ) - 1
      )
      AND ended_sprint.end_date < CAST(
          CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
          AS date
      )
      AND current_status.category IN ('backlog', 'unstarted', 'started')
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
    ORDER BY ended_sprint.end_date, ended_sprint.sprint_id, story.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF story SKIP LOCKED
)
UPDATE public.stories AS story
SET sprint_id = candidate.new_sprint_id
FROM candidates AS candidate
WHERE story.id = candidate.id
  AND story.workspace_id = candidate.workspace_id
  AND story.team_id = candidate.team_id
  AND story.status_id = candidate.expected_status_id
  AND story.sprint_id = candidate.previous_sprint_id
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
RETURNING
    story.id,
    candidate.workspace_id,
    candidate.team_id,
    candidate.previous_sprint_id,
    candidate.new_sprint_id;

-- name: InsertSprintMigrationActivities :execrows
WITH input AS (
    SELECT
        requested_story.story_id,
        requested_workspace.workspace_id,
        requested_team.team_id,
        requested_previous_sprint.previous_sprint_id,
        requested_new_sprint.new_sprint_id
    FROM unnest(CAST(sqlc.arg(story_ids) AS uuid[]))
        WITH ORDINALITY AS requested_story(story_id, position)
    INNER JOIN unnest(CAST(sqlc.arg(workspace_ids) AS uuid[]))
        WITH ORDINALITY AS requested_workspace(workspace_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(team_ids) AS uuid[]))
        WITH ORDINALITY AS requested_team(team_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(previous_sprint_ids) AS uuid[]))
        WITH ORDINALITY AS requested_previous_sprint(previous_sprint_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(new_sprint_ids) AS uuid[]))
        WITH ORDINALITY AS requested_new_sprint(new_sprint_id, position)
        USING (position)
)
INSERT INTO public.story_activities (
    story_id,
    user_id,
    activity_type,
    field_changed,
    current_value,
    old_value,
    new_value,
    reason,
    workspace_id,
    created_at
)
SELECT
    input.story_id,
    sqlc.arg(system_user_id),
    'update',
    'sprint_id',
    CAST(input.new_sprint_id AS text),
    to_jsonb(input.previous_sprint_id),
    to_jsonb(input.new_sprint_id),
    sqlc.arg(reason),
    input.workspace_id,
    CAST(
        CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC'
        AS timestamp
    )
FROM input
INNER JOIN public.stories AS story
    ON story.id = input.story_id
   AND story.workspace_id = input.workspace_id
   AND story.team_id = input.team_id
   AND story.sprint_id = input.new_sprint_id
INNER JOIN public.sprints AS previous_sprint
    ON previous_sprint.sprint_id = input.previous_sprint_id
   AND previous_sprint.workspace_id = input.workspace_id
   AND previous_sprint.team_id = input.team_id
INNER JOIN public.sprints AS new_sprint
    ON new_sprint.sprint_id = input.new_sprint_id
   AND new_sprint.workspace_id = input.workspace_id
   AND new_sprint.team_id = input.team_id
WHERE story.deleted_at IS NULL
  AND story.archived_at IS NULL;

-- name: InsertSprintMigrationAuditEvents :execrows
WITH input AS (
    SELECT
        requested_story.story_id,
        requested_workspace.workspace_id,
        requested_team.team_id,
        requested_previous_sprint.previous_sprint_id,
        requested_new_sprint.new_sprint_id
    FROM unnest(CAST(sqlc.arg(story_ids) AS uuid[]))
        WITH ORDINALITY AS requested_story(story_id, position)
    INNER JOIN unnest(CAST(sqlc.arg(workspace_ids) AS uuid[]))
        WITH ORDINALITY AS requested_workspace(workspace_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(team_ids) AS uuid[]))
        WITH ORDINALITY AS requested_team(team_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(previous_sprint_ids) AS uuid[]))
        WITH ORDINALITY AS requested_previous_sprint(previous_sprint_id, position)
        USING (position)
    INNER JOIN unnest(CAST(sqlc.arg(new_sprint_ids) AS uuid[]))
        WITH ORDINALITY AS requested_new_sprint(new_sprint_id, position)
        USING (position)
)
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
)
SELECT
    input.workspace_id,
    input.team_id,
    'automation',
    sqlc.arg(system_user_id),
    'story',
    input.story_id,
    'story.auto_moved_to_sprint',
    jsonb_build_object(
        'previous_sprint_id', input.previous_sprint_id,
        'new_sprint_id', input.new_sprint_id
    ),
    CAST(sqlc.arg(as_of) AS timestamptz)
FROM input
INNER JOIN public.stories AS story
    ON story.id = input.story_id
   AND story.workspace_id = input.workspace_id
   AND story.team_id = input.team_id
   AND story.sprint_id = input.new_sprint_id
INNER JOIN public.sprints AS previous_sprint
    ON previous_sprint.sprint_id = input.previous_sprint_id
   AND previous_sprint.workspace_id = input.workspace_id
   AND previous_sprint.team_id = input.team_id
INNER JOIN public.sprints AS new_sprint
    ON new_sprint.sprint_id = input.new_sprint_id
   AND new_sprint.workspace_id = input.workspace_id
   AND new_sprint.team_id = input.team_id
WHERE story.deleted_at IS NULL
  AND story.archived_at IS NULL;
