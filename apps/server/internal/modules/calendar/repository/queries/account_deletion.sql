-- name: DeleteRevokedAccountCalendarConnections :exec
DELETE FROM calendar_connections
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
  AND revoked_at IS NOT NULL;

-- name: ScrubAccountCalendarCleanup :exec
UPDATE calendar_connections
SET connected_email = '', provider_account_id = '', timezone = 'UTC',
    sync_error = NULL, sync_token = NULL,
    notification_channel_id = NULL, notification_resource_id = NULL,
    notification_expires_at = NULL
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
  AND cleanup_pending_at IS NOT NULL;

-- name: ScrubAccountCalendarOutbox :exec
UPDATE calendar_schedule_event_outbox
SET payload = CAST('{}' AS jsonb), last_error = NULL
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid);

-- name: AccountCalendarCleanupPending :one
SELECT EXISTS (
    SELECT 1 FROM calendar_connections
    WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
);
