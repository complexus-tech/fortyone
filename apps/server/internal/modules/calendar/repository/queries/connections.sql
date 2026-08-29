-- name: LockCalendarUser :exec
SELECT pg_advisory_xact_lock(
    hashtextextended('calendar:' || CAST(CAST(sqlc.arg(user_id) AS uuid) AS text), 0)
);

-- name: ListCalendarConnectionsByUser :many
SELECT connection.*
FROM calendar_connections connection
WHERE connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
  AND connection.user_id = sqlc.arg(user_id)
ORDER BY connection.created_at DESC;

-- name: ListCalendarConnectionsByWorkspace :many
SELECT connection.*
FROM calendar_connections connection
WHERE connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
  AND connection.workspace_id = sqlc.arg(workspace_id)
ORDER BY connection.created_at DESC;

-- name: GetOwnedCalendarConnection :one
SELECT connection.*
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.connection_id = sqlc.arg(connection_id)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
LIMIT 1;

-- name: GetActiveCalendarConnection :one
SELECT connection.*
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.provider = sqlc.arg(provider)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
LIMIT 1;

-- name: GetCalendarConnection :one
SELECT connection.*
FROM calendar_connections connection
WHERE connection.connection_id = sqlc.arg(connection_id)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL;

-- name: GetScheduleEventDispatchConnection :one
SELECT connection.*, CAST(connection.cleanup_pending_at IS NOT NULL AS boolean) AS cleanup_pending
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.provider = COALESCE((
      SELECT outbox.provider
      FROM calendar_schedule_event_outbox outbox
      WHERE outbox.user_id = sqlc.arg(user_id)
        AND outbox.processed_at IS NULL
        AND outbox.dead_lettered_at IS NULL
        AND EXISTS (
            SELECT 1
            FROM calendar_connections provider_connection
            WHERE provider_connection.user_id = outbox.user_id
              AND provider_connection.provider = outbox.provider
              AND provider_connection.revoked_at IS NULL
        )
      ORDER BY outbox.created_at, outbox.outbox_id
      LIMIT 1
  ), connection.provider)
  AND connection.revoked_at IS NULL
ORDER BY connection.cleanup_pending_at NULLS FIRST
LIMIT 1;

-- name: WorkspaceCalendarMemberExists :one
SELECT EXISTS (
    SELECT 1
    FROM workspace_members membership
    INNER JOIN users actor ON actor.user_id = membership.user_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(user_id)
      AND actor.is_active = TRUE
);

-- name: ListCalendarConnectionsNeedingWatch :many
SELECT connection.*
FROM calendar_connections connection
WHERE connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
  AND (
      connection.notification_channel_id IS NULL
      OR connection.notification_expires_at IS NULL
      OR connection.notification_expires_at <= sqlc.arg(renew_before)
  )
ORDER BY connection.notification_expires_at NULLS FIRST, connection.connection_id;

-- name: LockExistingCalendarConnection :one
SELECT connection.provider_account_id,
       CAST(connection.cleanup_pending_at IS NOT NULL AS boolean) AS cleanup_pending
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.provider = sqlc.arg(provider)
  AND connection.revoked_at IS NULL
FOR UPDATE;

-- name: UpsertCalendarConnection :one
INSERT INTO calendar_connections (
    workspace_id,
    user_id,
    credential_generation,
    provider_account_id,
    provider,
    is_primary,
    connected_email,
    timezone,
    token_payload,
    scopes,
    sync_status,
    sync_error,
    revoked_at,
    updated_at
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(credential_generation),
    sqlc.arg(provider_account_id),
    sqlc.arg(provider),
    (
        CAST(sqlc.arg(can_write) AS boolean)
        AND NOT EXISTS (
            SELECT 1
            FROM calendar_connections primary_connection
            WHERE primary_connection.user_id = sqlc.arg(user_id)
              AND primary_connection.is_primary = TRUE
        )
    ),
    sqlc.arg(connected_email),
    sqlc.arg(timezone),
    sqlc.arg(token_payload),
    CAST(sqlc.arg(scopes) AS text[]),
    'connected',
    NULL,
    NULL,
    CURRENT_TIMESTAMP
)
ON CONFLICT (user_id, provider)
WHERE revoked_at IS NULL
DO UPDATE SET
    credential_generation = EXCLUDED.credential_generation,
    is_primary = calendar_connections.is_primary OR EXCLUDED.is_primary,
    provider_account_id = EXCLUDED.provider_account_id,
    connected_email = EXCLUDED.connected_email,
    timezone = EXCLUDED.timezone,
    token_payload = EXCLUDED.token_payload,
    scopes = EXCLUDED.scopes,
    sync_status = 'connected',
    sync_error = NULL,
    sync_token = NULL,
    notification_channel_id = NULL,
    notification_resource_id = NULL,
    notification_expires_at = NULL,
    last_synced_at = CASE
        WHEN calendar_connections.provider_account_id <> ''
          AND calendar_connections.provider_account_id = EXCLUDED.provider_account_id
            THEN calendar_connections.last_synced_at
        ELSE NULL
    END,
    updated_at = CURRENT_TIMESTAMP
RETURNING calendar_connections.*;

-- name: LockCalendarConnectionForPrimary :one
SELECT connection.*
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.connection_id = sqlc.arg(connection_id)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
FOR UPDATE;

-- name: ClearPrimaryCalendarConnection :exec
UPDATE calendar_connections
SET is_primary = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND is_primary = TRUE;

-- name: SetPrimaryCalendarConnection :one
UPDATE calendar_connections
SET is_primary = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id)
RETURNING calendar_connections.*;

-- name: UpdateCalendarConnectionToken :execrows
UPDATE calendar_connections
SET token_payload = sqlc.arg(token_payload),
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id)
  AND credential_generation = sqlc.arg(credential_generation)
  AND revoked_at IS NULL;

-- name: BeginCalendarConnectionSync :one
UPDATE calendar_connections
SET credential_generation = sqlc.arg(next_credential_generation),
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND provider = sqlc.arg(provider)
  AND credential_generation = sqlc.arg(current_credential_generation)
  AND revoked_at IS NULL
  AND cleanup_pending_at IS NULL
RETURNING calendar_connections.*;

-- name: LockCalendarConnectionForRevocation :one
SELECT connection.provider, connection.is_primary
FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.connection_id = sqlc.arg(connection_id)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
FOR UPDATE;

-- name: MarkCalendarConnectionCleanupPending :execrows
UPDATE calendar_connections
SET cleanup_pending_at = CURRENT_TIMESTAMP,
    is_primary = FALSE,
    sync_status = 'revoked',
    sync_error = NULL,
    sync_token = NULL,
    notification_channel_id = NULL,
    notification_resource_id = NULL,
    notification_expires_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND connection_id = sqlc.arg(connection_id)
  AND revoked_at IS NULL
  AND cleanup_pending_at IS NULL;

-- name: ReactivateCalendarOutboxAfterAuthorizationRefresh :exec
UPDATE calendar_schedule_event_outbox
SET dead_lettered_at = NULL,
    attempt_count = 0,
    available_at = CURRENT_TIMESTAMP,
    last_error = 'Retrying after calendar authorization refresh.',
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND provider = sqlc.arg(provider)
  AND processed_at IS NULL
  AND dead_lettered_at IS NOT NULL;

-- name: ReactivateCalendarOutboxForCleanup :exec
UPDATE calendar_schedule_event_outbox
SET dead_lettered_at = NULL,
    attempt_count = 0,
    available_at = CURRENT_TIMESTAMP,
    last_error = 'Retrying provider cleanup during calendar disconnect.',
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND provider = sqlc.arg(provider)
  AND processed_at IS NULL;

-- name: ListMayaScheduleMirrorsForCleanup :many
SELECT block.block_id,
       block.workspace_id,
       block.story_id,
       block.external_calendar_id,
       block.external_event_id
FROM calendar_schedule_blocks block
WHERE block.user_id = sqlc.arg(user_id)
  AND block.source = 'maya'
  AND block.external_provider = CAST(sqlc.arg(provider) AS text)
  AND block.external_event_id IS NOT NULL
ORDER BY block.block_id
FOR UPDATE;

-- name: PromoteReplacementPrimaryCalendarConnection :exec
UPDATE calendar_connections
SET is_primary = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = (
    SELECT candidate.connection_id
    FROM calendar_connections candidate
    WHERE candidate.user_id = sqlc.arg(user_id)
      AND candidate.revoked_at IS NULL
      AND candidate.cleanup_pending_at IS NULL
      AND (
          (candidate.provider = 'google'
              AND CAST(sqlc.arg(google_read_scope) AS text) = ANY(candidate.scopes)
              AND CAST(sqlc.arg(google_owned_scope) AS text) = ANY(candidate.scopes))
          OR (candidate.provider = 'microsoft'
              AND CAST(sqlc.arg(microsoft_write_scope) AS text) = ANY(candidate.scopes))
      )
    ORDER BY candidate.created_at, candidate.connection_id
    LIMIT 1
);

-- name: DeleteCalendarConnectionEvents :exec
DELETE FROM calendar_events
WHERE connection_id = sqlc.arg(connection_id);

-- name: DeleteCalendarConnectionBusyWindows :exec
DELETE FROM calendar_busy_windows
WHERE connection_id = sqlc.arg(connection_id);

-- name: DeleteDrainedCalendarConnection :exec
DELETE FROM calendar_connections connection
WHERE connection.connection_id = sqlc.arg(connection_id)
  AND connection.cleanup_pending_at IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM calendar_schedule_event_outbox outbox
      WHERE outbox.user_id = connection.user_id
        AND outbox.provider = connection.provider
        AND outbox.processed_at IS NULL
        AND outbox.dead_lettered_at IS NULL
  );

-- name: LockCalendarConnectionForSnapshot :one
SELECT connection.connection_id
FROM calendar_connections connection
WHERE connection.connection_id = sqlc.arg(connection_id)
  AND connection.workspace_id = sqlc.arg(workspace_id)
  AND connection.user_id = sqlc.arg(user_id)
  AND connection.provider = sqlc.arg(provider)
  AND connection.credential_generation = sqlc.arg(credential_generation)
  AND connection.revoked_at IS NULL
  AND connection.cleanup_pending_at IS NULL
FOR UPDATE;

-- name: RemoveCalendarConnectionScope :exec
UPDATE calendar_connections
SET scopes = array_remove(scopes, CAST(sqlc.arg(scope) AS text)),
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id);

-- name: InvalidateMayaScheduleMirrorHashes :exec
UPDATE calendar_schedule_blocks
SET external_sync_hash = NULL,
    external_synced_at = NULL
WHERE user_id = sqlc.arg(user_id)
  AND source = 'maya'
  AND external_provider = CAST(sqlc.arg(provider) AS text)
  AND external_sync_hash IS NOT NULL;

-- name: DetachMayaScheduleMirrors :exec
UPDATE calendar_schedule_blocks
SET external_provider = NULL,
    external_calendar_id = NULL,
    external_event_id = NULL,
    external_sync_hash = NULL,
    external_synced_at = NULL
WHERE user_id = sqlc.arg(user_id)
  AND source = 'maya'
  AND external_provider = CAST(sqlc.arg(provider) AS text);

-- name: UpdateCalendarSnapshotMetadata :exec
UPDATE calendar_connections
SET sync_token = NULLIF(CAST(sqlc.arg(sync_token) AS text), ''),
    timezone = COALESCE(NULLIF(BTRIM(CAST(sqlc.arg(timezone) AS text)), ''), timezone),
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id);

-- name: MarkCalendarConnectionSynced :execrows
UPDATE calendar_connections
SET sync_status = 'synced',
    sync_error = NULL,
    last_synced_at = sqlc.arg(synced_at),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND connection_id = sqlc.arg(connection_id)
  AND credential_generation = sqlc.arg(credential_generation)
  AND revoked_at IS NULL
  AND cleanup_pending_at IS NULL;

-- name: MarkCalendarConnectionSyncFailed :execrows
UPDATE calendar_connections
SET sync_status = 'failed',
    sync_error = sqlc.arg(sync_error),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND connection_id = sqlc.arg(connection_id)
  AND credential_generation = sqlc.arg(credential_generation)
  AND revoked_at IS NULL
  AND cleanup_pending_at IS NULL;

-- name: SetCalendarNotificationChannel :execrows
UPDATE calendar_connections
SET notification_channel_id = sqlc.arg(channel_id),
    notification_resource_id = sqlc.arg(resource_id),
    notification_expires_at = sqlc.arg(expires_at),
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND credential_generation = sqlc.arg(credential_generation)
  AND revoked_at IS NULL
  AND cleanup_pending_at IS NULL;

-- name: ClearCalendarNotificationChannel :exec
UPDATE calendar_connections
SET notification_channel_id = NULL,
    notification_resource_id = NULL,
    notification_expires_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE connection_id = sqlc.arg(connection_id);
