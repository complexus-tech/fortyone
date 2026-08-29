-- name: SupersedeStaleScheduleEventOutbox :exec
UPDATE calendar_schedule_event_outbox
SET processed_at = CURRENT_TIMESTAMP,
    last_error = 'Superseded by a newer schedule state.',
    updated_at = CURRENT_TIMESTAMP
WHERE schedule_block_id = sqlc.arg(schedule_block_id)
  AND processed_at IS NULL
  AND dedupe_key <> sqlc.arg(dedupe_key)
  AND (provider = sqlc.arg(provider) OR operation = 'upsert');

-- name: EnqueueScheduleEventOutbox :exec
INSERT INTO calendar_schedule_event_outbox (
    workspace_id,
    user_id,
    schedule_block_id,
    operation,
    provider,
    calendar_id,
    provider_event_id,
    payload,
    dedupe_key
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.narg(schedule_block_id),
    sqlc.arg(operation),
    sqlc.arg(provider),
    sqlc.arg(calendar_id),
    sqlc.arg(provider_event_id),
    CAST(sqlc.arg(payload) AS jsonb),
    sqlc.arg(dedupe_key)
)
ON CONFLICT (dedupe_key)
DO UPDATE SET
    payload = EXCLUDED.payload,
    processed_at = CASE
        WHEN CAST(sqlc.arg(reactivate_terminal) AS boolean)
          OR calendar_schedule_event_outbox.dead_lettered_at IS NULL
            THEN NULL
        ELSE calendar_schedule_event_outbox.processed_at
    END,
    dead_lettered_at = CASE
        WHEN CAST(sqlc.arg(reactivate_terminal) AS boolean) THEN NULL
        ELSE calendar_schedule_event_outbox.dead_lettered_at
    END,
    attempt_count = CASE
        WHEN CAST(sqlc.arg(reactivate_terminal) AS boolean)
          OR calendar_schedule_event_outbox.dead_lettered_at IS NULL
            THEN 0
        ELSE calendar_schedule_event_outbox.attempt_count
    END,
    last_error = CASE
        WHEN CAST(sqlc.arg(reactivate_terminal) AS boolean)
          OR calendar_schedule_event_outbox.dead_lettered_at IS NULL
            THEN NULL
        ELSE calendar_schedule_event_outbox.last_error
    END,
    available_at = CASE
        WHEN CAST(sqlc.arg(reactivate_terminal) AS boolean)
          OR calendar_schedule_event_outbox.dead_lettered_at IS NULL
            THEN CURRENT_TIMESTAMP
        ELSE calendar_schedule_event_outbox.available_at
    END,
    updated_at = CURRENT_TIMESTAMP;

-- name: ListReadyScheduleEventOutboxUsers :many
SELECT outbox.user_id
FROM calendar_schedule_event_outbox outbox
INNER JOIN calendar_connections connection ON
    connection.user_id = outbox.user_id
    AND connection.provider = outbox.provider
    AND connection.revoked_at IS NULL
    AND (
        connection.cleanup_pending_at IS NOT NULL
        OR (
            (connection.provider = 'google'
                AND CAST(sqlc.arg(google_owned_scope) AS text) = ANY(connection.scopes)
                AND CAST(sqlc.arg(google_read_scope) AS text) = ANY(connection.scopes))
            OR (connection.provider = 'microsoft'
                AND CAST(sqlc.arg(microsoft_write_scope) AS text) = ANY(connection.scopes))
        )
    )
WHERE outbox.processed_at IS NULL
  AND outbox.dead_lettered_at IS NULL
  AND outbox.available_at <= CURRENT_TIMESTAMP
GROUP BY outbox.user_id
ORDER BY MIN(outbox.available_at), outbox.user_id
LIMIT sqlc.arg(row_limit);

-- name: ClaimPendingScheduleEventOutbox :many
WITH ready AS (
    SELECT candidate.outbox_id
    FROM calendar_schedule_event_outbox candidate
    WHERE candidate.user_id = sqlc.arg(user_id)
      AND candidate.provider = sqlc.arg(provider)
      AND candidate.processed_at IS NULL
      AND candidate.dead_lettered_at IS NULL
      AND candidate.available_at <= CURRENT_TIMESTAMP
    ORDER BY candidate.created_at, candidate.outbox_id
    LIMIT sqlc.arg(row_limit)
    FOR UPDATE SKIP LOCKED
)
UPDATE calendar_schedule_event_outbox outbox
SET attempt_count = outbox.attempt_count + 1,
    available_at = CURRENT_TIMESTAMP + INTERVAL '5 minutes',
    updated_at = CURRENT_TIMESTAMP
FROM ready
WHERE outbox.outbox_id = ready.outbox_id
RETURNING outbox.outbox_id,
          outbox.workspace_id,
          outbox.user_id,
          outbox.schedule_block_id,
          outbox.operation,
          outbox.provider,
          outbox.calendar_id,
          outbox.provider_event_id,
          outbox.payload,
          outbox.dedupe_key,
          outbox.attempt_count;

-- name: ScheduleEventUpsertIsCurrent :one
SELECT EXISTS (
    SELECT 1
    FROM calendar_schedule_blocks block
    INNER JOIN calendar_maya_schedule_ownerships ownership ON
        ownership.workspace_id = block.workspace_id
        AND ownership.story_id = block.story_id
        AND ownership.user_id = block.user_id
    INNER JOIN stories story ON
        story.id = block.story_id
        AND story.workspace_id = block.workspace_id
    INNER JOIN statuses status ON status.status_id = story.status_id
    INNER JOIN users owner_user ON
        owner_user.user_id = block.user_id
        AND owner_user.is_active = TRUE
    INNER JOIN workspace_members workspace_member ON
        workspace_member.workspace_id = block.workspace_id
        AND workspace_member.user_id = block.user_id
    INNER JOIN team_members team_member ON
        team_member.team_id = story.team_id
        AND team_member.user_id = block.user_id
    LEFT JOIN team_sprint_settings team_settings ON
        team_settings.team_id = story.team_id
        AND team_settings.workspace_id = story.workspace_id
    LEFT JOIN sprints sprint ON sprint.sprint_id = story.sprint_id
    WHERE block.block_id = sqlc.arg(block_id)
      AND block.workspace_id = sqlc.arg(workspace_id)
      AND block.user_id = sqlc.arg(user_id)
      AND block.source = 'maya'
      AND block.completed_at IS NULL
      AND block.external_provider = CAST(sqlc.arg(provider) AS text)
      AND block.external_calendar_id = CAST(sqlc.arg(calendar_id) AS text)
      AND block.external_event_id = CAST(sqlc.arg(provider_event_id) AS text)
      AND block.title = sqlc.arg(title)
      AND story.title = sqlc.arg(title)
      AND block.start_at = sqlc.arg(start_at)
      AND block.end_at = sqlc.arg(end_at)
      AND story.assignee_id = block.user_id
      AND story.auto_scheduling_enabled = TRUE
      AND ownership.updated_at >= story.updated_at
      AND (team_settings.updated_at IS NULL OR ownership.updated_at >= team_settings.updated_at)
      AND (sprint.updated_at IS NULL OR ownership.updated_at >= sprint.updated_at)
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.completed_at IS NULL
      AND status.category NOT IN ('completed', 'cancelled')
);

-- name: MarkScheduleEventOutboxProcessed :exec
UPDATE calendar_schedule_event_outbox
SET processed_at = CURRENT_TIMESTAMP,
    last_error = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = sqlc.arg(outbox_id);

-- name: MarkScheduleBlockMirrored :exec
UPDATE calendar_schedule_blocks
SET external_provider = CAST(sqlc.arg(provider) AS text),
    external_calendar_id = CAST(sqlc.arg(calendar_id) AS text),
    external_event_id = CAST(sqlc.arg(provider_event_id) AS text),
    external_sync_hash = CAST(sqlc.arg(sync_hash) AS text),
    external_synced_at = CURRENT_TIMESTAMP
WHERE block_id = sqlc.arg(block_id)
  AND completed_at IS NULL;

-- name: MarkScheduleEventOutboxFailed :exec
UPDATE calendar_schedule_event_outbox
SET last_error = sqlc.arg(last_error),
    dead_lettered_at = CASE
        WHEN CAST(sqlc.arg(permanent) AS boolean) OR attempt_count >= 8 THEN CURRENT_TIMESTAMP
        ELSE NULL
    END,
    available_at = CASE
        WHEN CAST(sqlc.arg(permanent) AS boolean) OR attempt_count >= 8 THEN available_at
        ELSE CURRENT_TIMESTAMP + CASE
            WHEN attempt_count <= 1 THEN INTERVAL '1 minute'
            WHEN attempt_count = 2 THEN INTERVAL '2 minutes'
            WHEN attempt_count = 3 THEN INTERVAL '4 minutes'
            WHEN attempt_count = 4 THEN INTERVAL '8 minutes'
            WHEN attempt_count = 5 THEN INTERVAL '16 minutes'
            WHEN attempt_count = 6 THEN INTERVAL '32 minutes'
            ELSE INTERVAL '1 hour'
        END
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = sqlc.arg(outbox_id)
  AND processed_at IS NULL;

-- name: ReleaseScheduleEventOutbox :exec
UPDATE calendar_schedule_event_outbox
SET attempt_count = GREATEST(attempt_count - 1, 0),
    available_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = ANY(CAST(sqlc.arg(outbox_ids) AS uuid[]))
  AND processed_at IS NULL;

-- name: DeleteDrainedCleanupPendingCalendarConnection :exec
DELETE FROM calendar_connections connection
WHERE connection.user_id = sqlc.arg(user_id)
  AND connection.provider = sqlc.arg(provider)
  AND connection.cleanup_pending_at IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM calendar_schedule_event_outbox outbox
      WHERE outbox.user_id = connection.user_id
        AND outbox.provider = connection.provider
        AND outbox.processed_at IS NULL
        AND outbox.dead_lettered_at IS NULL
  );
