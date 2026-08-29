-- name: ListCalendarEvents :many
SELECT
    event.event_id,
    event.connection_id,
    event.provider,
    event.calendar_id,
    event.provider_event_id,
    event.title,
    event.location,
    event.meeting_url,
    event.html_link,
    event.start_at,
    event.end_at,
    event.is_all_day,
    event.start_date,
    event.end_date,
    event.is_private,
    event.created_at,
    event.updated_at
FROM calendar_events event
INNER JOIN calendar_connections connection ON
    connection.connection_id = event.connection_id
    AND connection.user_id = event.user_id
    AND connection.revoked_at IS NULL
    AND connection.cleanup_pending_at IS NULL
    AND (
        (connection.provider = sqlc.arg(google_provider)
            AND CAST(sqlc.arg(google_read_scope) AS text) = ANY(connection.scopes))
        OR (connection.provider = sqlc.arg(microsoft_provider)
            AND CAST(sqlc.arg(microsoft_read_scope) AS text) = ANY(connection.scopes))
    )
WHERE event.user_id = sqlc.arg(user_id)
  AND event.start_at < sqlc.arg(end_at)
  AND event.end_at > sqlc.arg(start_at)
ORDER BY event.start_at ASC, event.event_id ASC;

-- name: GetCalendarEvent :one
SELECT event.*
FROM calendar_events event
INNER JOIN calendar_connections connection ON
    connection.connection_id = event.connection_id
    AND connection.user_id = event.user_id
    AND connection.revoked_at IS NULL
    AND connection.cleanup_pending_at IS NULL
    AND (
        (connection.provider = sqlc.arg(google_provider)
            AND CAST(sqlc.arg(google_read_scope) AS text) = ANY(connection.scopes))
        OR (connection.provider = sqlc.arg(microsoft_provider)
            AND CAST(sqlc.arg(microsoft_read_scope) AS text) = ANY(connection.scopes))
    )
WHERE event.user_id = sqlc.arg(user_id)
  AND event.event_id = sqlc.arg(event_id)
LIMIT 1;

-- name: UpsertCalendarEvent :exec
INSERT INTO calendar_events (
    connection_id,
    workspace_id,
    user_id,
    provider,
    calendar_id,
    provider_event_id,
    title,
    description,
    location,
    meeting_url,
    html_link,
    organizer,
    attendees,
    attendees_omitted,
    is_all_day,
    start_date,
    end_date,
    start_at,
    end_at,
    visibility,
    is_private,
    source_hash,
    sync_generation,
    updated_at
) VALUES (
    sqlc.arg(connection_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(provider),
    sqlc.arg(calendar_id),
    sqlc.arg(provider_event_id),
    sqlc.narg(title),
    sqlc.narg(description),
    sqlc.narg(location),
    sqlc.narg(meeting_url),
    sqlc.narg(html_link),
    CAST(sqlc.narg(organizer) AS jsonb),
    CAST(sqlc.arg(attendees) AS jsonb),
    sqlc.arg(attendees_omitted),
    sqlc.arg(is_all_day),
    CAST(sqlc.narg(start_date) AS date),
    CAST(sqlc.narg(end_date) AS date),
    sqlc.arg(start_at),
    sqlc.arg(end_at),
    sqlc.arg(visibility),
    sqlc.arg(is_private),
    sqlc.arg(source_hash),
    sqlc.arg(sync_generation),
    CURRENT_TIMESTAMP
)
ON CONFLICT (connection_id, calendar_id, provider_event_id)
DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    location = EXCLUDED.location,
    meeting_url = EXCLUDED.meeting_url,
    html_link = EXCLUDED.html_link,
    organizer = EXCLUDED.organizer,
    attendees = EXCLUDED.attendees,
    attendees_omitted = EXCLUDED.attendees_omitted,
    is_all_day = EXCLUDED.is_all_day,
    start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    start_at = EXCLUDED.start_at,
    end_at = EXCLUDED.end_at,
    visibility = EXCLUDED.visibility,
    is_private = EXCLUDED.is_private,
    source_hash = EXCLUDED.source_hash,
    sync_generation = EXCLUDED.sync_generation,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteStaleCalendarEvents :exec
DELETE FROM calendar_events
WHERE connection_id = sqlc.arg(connection_id)
  AND sync_generation <> sqlc.arg(sync_generation);

-- name: DeleteCalendarBusyWindows :exec
DELETE FROM calendar_busy_windows
WHERE connection_id = sqlc.arg(connection_id);

-- name: InsertCalendarBusyWindow :exec
INSERT INTO calendar_busy_windows (
    connection_id,
    workspace_id,
    user_id,
    provider,
    provider_event_id,
    calendar_id,
    title,
    start_at,
    end_at,
    status,
    transparency,
    is_private,
    source_hash,
    updated_at
) VALUES (
    sqlc.arg(connection_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(provider),
    sqlc.arg(provider_event_id),
    sqlc.narg(calendar_id),
    sqlc.narg(title),
    sqlc.arg(start_at),
    sqlc.arg(end_at),
    sqlc.arg(status),
    sqlc.arg(transparency),
    sqlc.arg(is_private),
    sqlc.arg(source_hash),
    CURRENT_TIMESTAMP
);

-- name: LockCalendarConnectionForIncrementalSync :one
SELECT connection.connection_id
FROM calendar_connections connection
WHERE connection.connection_id = sqlc.arg(connection_id)
  AND connection.workspace_id = sqlc.arg(workspace_id)
  AND connection.credential_generation = sqlc.arg(credential_generation)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
FOR UPDATE;

-- name: DeleteChangedCalendarEvent :exec
DELETE FROM calendar_events
WHERE connection_id = sqlc.arg(connection_id)
  AND calendar_id = 'primary'
  AND provider_event_id = sqlc.arg(provider_event_id);

-- name: DeleteChangedCalendarBusyWindow :exec
DELETE FROM calendar_busy_windows
WHERE connection_id = sqlc.arg(connection_id)
  AND provider_event_id = sqlc.arg(provider_event_id);

-- name: InvalidateDeletedManagedScheduleEvent :exec
UPDATE calendar_schedule_blocks
SET external_sync_hash = NULL,
    external_synced_at = NULL
WHERE user_id = sqlc.arg(user_id)
  AND external_provider = CAST(sqlc.arg(provider) AS text)
  AND external_event_id = CAST(sqlc.arg(event_id) AS text);

-- name: LockChangedManagedScheduleEvent :one
SELECT block.block_id,
       block.workspace_id,
       block.story_id,
       block.title,
       block.start_at,
       block.end_at
FROM calendar_schedule_blocks block
WHERE block.user_id = sqlc.arg(user_id)
  AND block.external_provider = CAST(sqlc.arg(provider) AS text)
  AND block.external_event_id = CAST(sqlc.arg(event_id) AS text)
  AND block.completed_at IS NULL
FOR UPDATE;

-- name: InvalidateChangedManagedScheduleEvent :exec
UPDATE calendar_schedule_blocks
SET external_sync_hash = NULL,
    external_synced_at = NULL
WHERE block_id = sqlc.arg(block_id);

-- name: StoreIncrementalCalendarSyncToken :exec
UPDATE calendar_connections
SET sync_token = NULLIF(CAST(sqlc.arg(sync_token) AS text), ''),
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id);
