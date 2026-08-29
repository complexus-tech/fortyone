-- name: ListCalendarBusyWindows :many
SELECT busy.window_id,
       busy.connection_id,
       busy.workspace_id,
       busy.user_id,
       busy.provider,
       busy.provider_event_id,
       busy.calendar_id,
       busy.title,
       busy.start_at,
       busy.end_at,
       busy.status,
       busy.transparency,
       busy.is_private,
       busy.source_hash,
       busy.created_at,
       busy.updated_at
FROM calendar_busy_windows busy
INNER JOIN calendar_connections connection ON
    connection.connection_id = busy.connection_id
    AND connection.user_id = busy.user_id
    AND connection.revoked_at IS NULL
    AND connection.cleanup_pending_at IS NULL
WHERE busy.user_id = sqlc.arg(user_id)
  AND busy.start_at < sqlc.arg(end_at)
  AND busy.end_at > sqlc.arg(start_at)
ORDER BY busy.start_at ASC;

-- name: ListCalendarScheduleBlocks :many
SELECT
    block.block_id,
    block.workspace_id,
    block.user_id,
    story.id AS story_id,
    story.title AS story_title,
    CAST(COALESCE(CASE
        WHEN story.id IS NOT NULL AND team.code IS NOT NULL
            THEN team.code || '-' || CAST(story.sequence_id AS text)
        ELSE NULL
    END, '') AS text) AS story_code,
    status.color AS story_status_color,
    COALESCE(story.priority, '') AS story_priority,
    story.end_date AS story_end_date,
    team.team_id,
    team.name AS team_name,
    team.code AS team_code,
    block.block_type,
    CAST(CASE
        WHEN block.block_type = 'work' AND story.id IS NULL THEN 'Work'
        ELSE block.title
    END AS text) AS title,
    block.start_at,
    block.end_at,
    block.completed_at,
    CAST(block.completed_at IS NULL AND EXISTS (
        SELECT 1
        FROM calendar_busy_windows conflict_window
        INNER JOIN calendar_connections conflict_connection ON
            conflict_connection.connection_id = conflict_window.connection_id
            AND conflict_connection.user_id = conflict_window.user_id
            AND conflict_connection.revoked_at IS NULL
            AND conflict_connection.cleanup_pending_at IS NULL
        WHERE conflict_window.user_id = block.user_id
          AND conflict_window.start_at < block.end_at
          AND conflict_window.end_at > block.start_at
    ) AS boolean) AS has_conflict,
    block.is_locked,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_status END, '') AS text) AS auto_scheduling_status,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_reason END, '') AS text) AS auto_scheduling_reason,
    block.source,
    block.segment_index,
    block.external_provider,
    block.external_calendar_id,
    block.external_event_id,
    block.external_sync_hash,
    block.external_synced_at,
    block.created_at,
    block.updated_at,
    block.manual_override_at,
    block.manual_override_by
FROM calendar_schedule_blocks block
LEFT JOIN stories story ON
    story.id = block.story_id
    AND story.workspace_id = block.workspace_id
    AND story.deleted_at IS NULL
    AND EXISTS (
        SELECT 1
        FROM team_members viewer_membership
        WHERE viewer_membership.team_id = story.team_id
          AND viewer_membership.user_id = block.user_id
    )
LEFT JOIN teams team ON team.team_id = story.team_id
LEFT JOIN statuses status ON
    status.status_id = story.status_id
    AND status.team_id = story.team_id
WHERE block.workspace_id = sqlc.arg(workspace_id)
  AND block.user_id = sqlc.arg(user_id)
  AND block.start_at < sqlc.arg(end_at)
  AND block.end_at > sqlc.arg(start_at)
ORDER BY block.start_at ASC;

-- name: ListCalendarScheduleIssues :many
SELECT
    story.id AS story_id,
    story.title AS story_title,
    CAST(team.code || '-' || CAST(story.sequence_id AS text) AS text) AS story_code,
    team.team_id,
    team.name AS team_name,
    team.code AS team_code,
    story.estimated_duration_minutes,
    CAST(COALESCE(scheduled.scheduled_duration_minutes, 0) AS integer) AS scheduled_duration_minutes,
    CAST(GREATEST(
        COALESCE(story.estimated_duration_minutes, 0) - COALESCE(scheduled.scheduled_duration_minutes, 0),
        0
    ) AS integer) AS remaining_duration_minutes,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.updated_at
FROM stories story
INNER JOIN teams team ON team.team_id = story.team_id
INNER JOIN statuses status ON status.status_id = story.status_id
INNER JOIN team_members membership ON
    membership.team_id = story.team_id
    AND membership.user_id = sqlc.arg(user_id)
LEFT JOIN LATERAL (
    SELECT CAST(COALESCE(
        SUM(EXTRACT(EPOCH FROM (schedule_block.end_at - schedule_block.start_at)) / 60),
        0
    ) AS integer) AS scheduled_duration_minutes
    FROM calendar_schedule_blocks schedule_block
    WHERE schedule_block.workspace_id = story.workspace_id
      AND schedule_block.user_id = story.assignee_id
      AND schedule_block.story_id = story.id
      AND schedule_block.source = 'maya'
      AND schedule_block.completed_at IS NULL
) scheduled ON TRUE
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.assignee_id = sqlc.arg(user_id)
  AND story.auto_scheduling_enabled = TRUE
  AND story.auto_scheduling_status = 'cannot_fit'
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
  AND story.completed_at IS NULL
  AND status.category NOT IN ('completed', 'cancelled')
ORDER BY story.auto_scheduling_updated_at DESC NULLS LAST, story.updated_at DESC, story.id;

-- name: ListSchedulingBlocksForUser :many
SELECT
    block.block_id,
    block.workspace_id,
    block.user_id,
    story.id AS story_id,
    story.title AS story_title,
    CAST(COALESCE(CASE
        WHEN story.id IS NOT NULL AND team.code IS NOT NULL
            THEN team.code || '-' || CAST(story.sequence_id AS text)
        ELSE NULL
    END, '') AS text) AS story_code,
    status.color AS story_status_color,
    COALESCE(story.priority, '') AS story_priority,
    story.end_date AS story_end_date,
    team.team_id,
    team.name AS team_name,
    team.code AS team_code,
    block.block_type,
    CAST(CASE
        WHEN block.block_type = 'work' AND story.id IS NULL THEN 'Work'
        ELSE block.title
    END AS text) AS title,
    block.start_at,
    block.end_at,
    block.completed_at,
    CAST(block.completed_at IS NULL AND EXISTS (
        SELECT 1
        FROM calendar_busy_windows conflict_window
        INNER JOIN calendar_connections conflict_connection ON
            conflict_connection.connection_id = conflict_window.connection_id
            AND conflict_connection.user_id = conflict_window.user_id
            AND conflict_connection.revoked_at IS NULL
            AND conflict_connection.cleanup_pending_at IS NULL
        WHERE conflict_window.user_id = block.user_id
          AND conflict_window.start_at < block.end_at
          AND conflict_window.end_at > block.start_at
    ) AS boolean) AS has_conflict,
    block.is_locked,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_status END, '') AS text) AS auto_scheduling_status,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_reason END, '') AS text) AS auto_scheduling_reason,
    block.source,
    block.segment_index,
    block.external_provider,
    block.external_calendar_id,
    block.external_event_id,
    block.external_sync_hash,
    block.external_synced_at,
    block.created_at,
    block.updated_at,
    block.manual_override_at,
    block.manual_override_by
FROM calendar_schedule_blocks block
LEFT JOIN stories story ON
    story.id = block.story_id
    AND story.workspace_id = block.workspace_id
    AND story.deleted_at IS NULL
    AND EXISTS (
        SELECT 1
        FROM team_members viewer_membership
        WHERE viewer_membership.team_id = story.team_id
          AND viewer_membership.user_id = block.user_id
    )
LEFT JOIN teams team ON team.team_id = story.team_id
LEFT JOIN statuses status ON
    status.status_id = story.status_id
    AND status.team_id = story.team_id
WHERE block.user_id = sqlc.arg(user_id)
  AND block.start_at < sqlc.arg(end_at)
  AND block.end_at > sqlc.arg(start_at)
  AND block.completed_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM workspace_members owner_membership
      WHERE owner_membership.workspace_id = block.workspace_id
        AND owner_membership.user_id = block.user_id
  )
ORDER BY block.start_at ASC;

-- name: ListManualScheduleRescheduleEvents :many
SELECT event.next_start_at,
       event.timezone,
       event.created_at
FROM calendar_schedule_reschedule_events event
WHERE event.workspace_id = sqlc.arg(workspace_id)
  AND event.user_id = sqlc.arg(user_id)
  AND event.source = 'user'
  AND event.created_at >= sqlc.arg(since)
ORDER BY event.created_at DESC
LIMIT 100;

-- name: ListMayaScheduleBlocksForStory :many
SELECT
    block.block_id,
    block.workspace_id,
    block.user_id,
    story.id AS story_id,
    story.title AS story_title,
    CAST(COALESCE(CASE
        WHEN story.id IS NOT NULL AND team.code IS NOT NULL
            THEN team.code || '-' || CAST(story.sequence_id AS text)
        ELSE NULL
    END, '') AS text) AS story_code,
    status.color AS story_status_color,
    COALESCE(story.priority, '') AS story_priority,
    story.end_date AS story_end_date,
    team.team_id,
    team.name AS team_name,
    team.code AS team_code,
    block.block_type,
    CAST(CASE
        WHEN block.block_type = 'work' AND story.id IS NULL THEN 'Work'
        ELSE block.title
    END AS text) AS title,
    block.start_at,
    block.end_at,
    block.completed_at,
    CAST(block.completed_at IS NULL AND EXISTS (
        SELECT 1
        FROM calendar_busy_windows conflict_window
        INNER JOIN calendar_connections conflict_connection ON
            conflict_connection.connection_id = conflict_window.connection_id
            AND conflict_connection.user_id = conflict_window.user_id
            AND conflict_connection.revoked_at IS NULL
            AND conflict_connection.cleanup_pending_at IS NULL
        WHERE conflict_window.user_id = block.user_id
          AND conflict_window.start_at < block.end_at
          AND conflict_window.end_at > block.start_at
    ) AS boolean) AS has_conflict,
    block.is_locked,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_status END, '') AS text) AS auto_scheduling_status,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_reason END, '') AS text) AS auto_scheduling_reason,
    block.source,
    block.segment_index,
    block.external_provider,
    block.external_calendar_id,
    block.external_event_id,
    block.external_sync_hash,
    block.external_synced_at,
    block.created_at,
    block.updated_at,
    block.manual_override_at,
    block.manual_override_by
FROM calendar_schedule_blocks block
LEFT JOIN stories story ON
    story.id = block.story_id
    AND story.workspace_id = block.workspace_id
    AND story.deleted_at IS NULL
    AND EXISTS (
        SELECT 1
        FROM team_members viewer_membership
        WHERE viewer_membership.team_id = story.team_id
          AND viewer_membership.user_id = block.user_id
    )
LEFT JOIN teams team ON team.team_id = story.team_id
LEFT JOIN statuses status ON
    status.status_id = story.status_id
    AND status.team_id = story.team_id
WHERE block.workspace_id = sqlc.arg(workspace_id)
  AND block.user_id = sqlc.arg(user_id)
  AND block.story_id = sqlc.arg(story_id)
  AND block.source = 'maya'
  AND block.completed_at IS NULL
ORDER BY block.segment_index, block.start_at;

-- name: MayaScheduleOwnershipExists :one
SELECT EXISTS (
    SELECT 1
    FROM calendar_maya_schedule_ownerships ownership
    WHERE ownership.workspace_id = sqlc.arg(workspace_id)
      AND ownership.user_id = sqlc.arg(user_id)
      AND ownership.story_id = sqlc.arg(story_id)
);

-- name: CalendarScheduleStoryExists :one
SELECT EXISTS (
    SELECT 1
    FROM stories story
    INNER JOIN team_members membership ON
        membership.team_id = story.team_id
        AND membership.user_id = sqlc.arg(user_id)
    WHERE story.workspace_id = sqlc.arg(workspace_id)
      AND story.id = sqlc.arg(story_id)
      AND story.deleted_at IS NULL
);

-- name: CalendarScheduleBlockConflicts :one
SELECT EXISTS (
    SELECT 1
    FROM calendar_schedule_blocks schedule_block
    WHERE schedule_block.user_id = sqlc.arg(user_id)
      AND schedule_block.completed_at IS NULL
      AND schedule_block.start_at < sqlc.arg(end_at)
      AND schedule_block.end_at > sqlc.arg(start_at)
      AND (
          CAST(sqlc.arg(exclude_block_id) AS uuid) = CAST('00000000-0000-0000-0000-000000000000' AS uuid)
          OR schedule_block.block_id <> CAST(sqlc.arg(exclude_block_id) AS uuid)
      )
    UNION ALL
    SELECT 1
    FROM calendar_busy_windows busy
    INNER JOIN calendar_connections connection ON
        connection.connection_id = busy.connection_id
        AND connection.user_id = busy.user_id
        AND connection.revoked_at IS NULL
        AND connection.cleanup_pending_at IS NULL
    WHERE busy.user_id = sqlc.arg(user_id)
      AND busy.start_at < sqlc.arg(end_at)
      AND busy.end_at > sqlc.arg(start_at)
);

-- name: GetCalendarScheduleBlock :one
SELECT
    block.block_id,
    block.workspace_id,
    block.user_id,
    story.id AS story_id,
    story.title AS story_title,
    CAST(COALESCE(CASE
        WHEN story.id IS NOT NULL AND team.code IS NOT NULL
            THEN team.code || '-' || CAST(story.sequence_id AS text)
        ELSE NULL
    END, '') AS text) AS story_code,
    status.color AS story_status_color,
    COALESCE(story.priority, '') AS story_priority,
    story.end_date AS story_end_date,
    team.team_id,
    team.name AS team_name,
    team.code AS team_code,
    block.block_type,
    CAST(CASE
        WHEN block.block_type = 'work' AND story.id IS NULL THEN 'Work'
        ELSE block.title
    END AS text) AS title,
    block.start_at,
    block.end_at,
    block.completed_at,
    CAST(block.completed_at IS NULL AND EXISTS (
        SELECT 1
        FROM calendar_busy_windows conflict_window
        INNER JOIN calendar_connections conflict_connection ON
            conflict_connection.connection_id = conflict_window.connection_id
            AND conflict_connection.user_id = conflict_window.user_id
            AND conflict_connection.revoked_at IS NULL
            AND conflict_connection.cleanup_pending_at IS NULL
        WHERE conflict_window.user_id = block.user_id
          AND conflict_window.start_at < block.end_at
          AND conflict_window.end_at > block.start_at
    ) AS boolean) AS has_conflict,
    block.is_locked,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_status END, '') AS text) AS auto_scheduling_status,
    CAST(COALESCE(CASE WHEN block.source = 'maya' THEN story.auto_scheduling_reason END, '') AS text) AS auto_scheduling_reason,
    block.source,
    block.segment_index,
    block.external_provider,
    block.external_calendar_id,
    block.external_event_id,
    block.external_sync_hash,
    block.external_synced_at,
    block.created_at,
    block.updated_at,
    block.manual_override_at,
    block.manual_override_by
FROM calendar_schedule_blocks block
LEFT JOIN stories story ON
    story.id = block.story_id
    AND story.workspace_id = block.workspace_id
    AND story.deleted_at IS NULL
    AND EXISTS (
        SELECT 1
        FROM team_members viewer_membership
        WHERE viewer_membership.team_id = story.team_id
          AND viewer_membership.user_id = block.user_id
    )
LEFT JOIN teams team ON team.team_id = story.team_id
LEFT JOIN statuses status ON
    status.status_id = story.status_id
    AND status.team_id = story.team_id
WHERE block.workspace_id = sqlc.arg(workspace_id)
  AND block.user_id = sqlc.arg(user_id)
  AND block.block_id = sqlc.arg(block_id)
LIMIT 1;
