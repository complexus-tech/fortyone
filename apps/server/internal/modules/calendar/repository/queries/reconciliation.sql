-- name: LockMayaScheduleStoryVersion :one
SELECT story.id
FROM stories story
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.id = sqlc.arg(story_id)
  AND story.updated_at = sqlc.arg(expected_updated_at)
FOR UPDATE OF story;

-- name: LockEligibleMayaScheduleStory :one
SELECT story.id
FROM stories story
INNER JOIN statuses status ON
    status.status_id = story.status_id
INNER JOIN users selected_user ON
    selected_user.user_id = sqlc.arg(user_id)
    AND selected_user.is_active = TRUE
INNER JOIN workspace_members membership ON
    membership.workspace_id = story.workspace_id
    AND membership.user_id = sqlc.arg(user_id)
INNER JOIN team_members team_membership ON
    team_membership.team_id = story.team_id
    AND team_membership.user_id = sqlc.arg(user_id)
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.id = sqlc.arg(story_id)
  AND story.auto_scheduling_enabled = TRUE
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
  AND story.completed_at IS NULL
  AND status.category NOT IN ('completed', 'cancelled')
FOR UPDATE OF story
FOR SHARE OF status, selected_user, membership, team_membership;

-- name: ListExistingMayaScheduleSegments :many
SELECT block.block_id,
       block.segment_index,
       block.title,
       block.start_at,
       block.end_at,
       block.is_locked,
       block.external_provider,
       block.external_calendar_id,
       block.external_event_id,
       block.external_sync_hash,
       block.manual_override_at,
       block.manual_override_by
FROM calendar_schedule_blocks block
WHERE block.workspace_id = sqlc.arg(workspace_id)
  AND block.user_id = sqlc.arg(user_id)
  AND block.story_id = sqlc.arg(story_id)
  AND block.source = 'maya'
  AND block.completed_at IS NULL
FOR UPDATE;

-- name: MayaScheduleSegmentConflicts :one
SELECT EXISTS (
    SELECT 1
    FROM calendar_schedule_blocks block
    WHERE block.user_id = sqlc.arg(user_id)
      AND block.completed_at IS NULL
      AND block.start_at < sqlc.arg(end_at)
      AND block.end_at > sqlc.arg(start_at)
      AND NOT (block.block_id = ANY(CAST(sqlc.arg(excluded_block_ids) AS uuid[])))
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

-- name: UpdateMayaScheduleSegment :exec
UPDATE calendar_schedule_blocks
SET title = CAST(sqlc.arg(title) AS text),
    start_at = sqlc.arg(start_at),
    end_at = sqlc.arg(end_at),
    is_locked = sqlc.arg(is_locked),
    external_provider = CAST(sqlc.arg(external_provider) AS text),
    external_calendar_id = 'primary',
    external_event_id = CAST(sqlc.arg(external_event_id) AS text),
    updated_at = CASE
        WHEN title <> CAST(sqlc.arg(title) AS text)
          OR start_at <> sqlc.arg(start_at)
          OR end_at <> sqlc.arg(end_at)
          OR is_locked <> sqlc.arg(is_locked)
            THEN CURRENT_TIMESTAMP
        ELSE updated_at
    END
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND story_id = sqlc.arg(story_id)
  AND segment_index = sqlc.arg(segment_index)
  AND source = 'maya';

-- name: CreateMayaScheduleSegment :exec
INSERT INTO calendar_schedule_blocks (
    block_id,
    workspace_id,
    user_id,
    story_id,
    block_type,
    title,
    start_at,
    end_at,
    is_locked,
    source,
    segment_index,
    external_provider,
    external_calendar_id,
    external_event_id
) VALUES (
    sqlc.arg(block_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(story_id),
    'work',
    sqlc.arg(title),
    sqlc.arg(start_at),
    sqlc.arg(end_at),
    sqlc.arg(is_locked),
    'maya',
    sqlc.arg(segment_index),
    CAST(sqlc.arg(external_provider) AS text),
    'primary',
    CAST(sqlc.arg(external_event_id) AS text)
);

-- name: DeleteMayaScheduleSegment :exec
DELETE FROM calendar_schedule_blocks
WHERE block_id = sqlc.arg(block_id);

-- name: RetainMayaScheduleOwnership :exec
INSERT INTO calendar_maya_schedule_ownerships (
    workspace_id,
    story_id,
    user_id
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(user_id)
)
ON CONFLICT (workspace_id, story_id)
DO UPDATE SET
    user_id = EXCLUDED.user_id,
    updated_at = CURRENT_TIMESTAMP,
    recovery_attempted_at = NULL;

-- name: ReleaseMayaScheduleOwnership :exec
DELETE FROM calendar_maya_schedule_ownerships
WHERE workspace_id = sqlc.arg(workspace_id)
  AND story_id = sqlc.arg(story_id)
  AND user_id = sqlc.arg(user_id);

-- name: ListCalendarWriteDestinations :many
SELECT connection.provider,
       connection.is_primary,
       CASE
           WHEN connection.provider = 'google'
               THEN CAST(sqlc.arg(google_read_scope) AS text) = ANY(connection.scopes)
                    AND CAST(sqlc.arg(google_owned_scope) AS text) = ANY(connection.scopes)
           WHEN connection.provider = 'microsoft'
               THEN CAST(sqlc.arg(microsoft_write_scope) AS text) = ANY(connection.scopes)
           ELSE FALSE
       END AS can_write
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
ORDER BY connection.is_primary DESC, connection.created_at, connection.connection_id;
