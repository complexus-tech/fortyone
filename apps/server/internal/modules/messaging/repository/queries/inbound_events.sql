-- name: InsertInboundEvent :one
INSERT INTO messaging_inbound_events (
    provider,
    workspace_id,
    installation_generation,
    external_workspace_id,
    external_event_id,
    event_type,
    payload_encrypted
) VALUES (
    sqlc.arg(provider),
    sqlc.narg(workspace_id),
    sqlc.narg(installation_generation),
    sqlc.arg(external_workspace_id),
    sqlc.arg(external_event_id),
    sqlc.arg(event_type),
    NULLIF(CAST(sqlc.arg(payload_encrypted) AS text), '')
)
ON CONFLICT (provider, external_workspace_id, external_event_id) DO NOTHING
RETURNING id,
          workspace_id,
          installation_generation,
          external_workspace_id,
          external_event_id,
          status,
          attempt_count,
          recovery_generation,
          processed_at,
          payload_encrypted;

-- name: BackfillInboundEventPayload :one
UPDATE messaging_inbound_events AS event
SET payload_encrypted = COALESCE(
    event.payload_encrypted,
    NULLIF(CAST(sqlc.arg(payload_encrypted) AS text), '')
)
WHERE event.provider = sqlc.arg(provider)
  AND event.external_workspace_id = sqlc.arg(external_workspace_id)
  AND event.external_event_id = sqlc.arg(external_event_id)
RETURNING event.id,
          event.workspace_id,
          event.installation_generation,
          event.external_workspace_id,
          event.external_event_id,
          event.status,
          event.attempt_count,
          event.recovery_generation,
          event.processed_at,
          event.payload_encrypted;

-- name: ClaimRecoverableInboundEvents :many
WITH candidates AS (
    SELECT candidate.id
    FROM messaging_inbound_events AS candidate
    WHERE candidate.provider = sqlc.arg(filter_provider)
      AND candidate.payload_encrypted IS NOT NULL
      AND candidate.attempt_count < 20
      AND (
          (candidate.status = 'pending' AND candidate.updated_at < NOW() - INTERVAL '30 seconds')
          OR (candidate.status = 'failed' AND candidate.updated_at < NOW() - INTERVAL '5 minutes')
          OR (
              candidate.status = 'processing'
              AND candidate.updated_at
                  < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
          )
      )
      AND (
          candidate.recovery_enqueued_at IS NULL
          OR candidate.recovery_enqueued_at < NOW() - (
              (
                  sqlc.arg(recovery_base_seconds)
                  * POWER(
                      2,
                      LEAST(candidate.recovery_generation, CAST(sqlc.arg(recovery_max_shift) AS integer))
                  )
              ) * INTERVAL '1 second'
          )
      )
    ORDER BY COALESCE(candidate.recovery_enqueued_at, candidate.received_at), candidate.received_at
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(row_limit)
)
UPDATE messaging_inbound_events AS event
SET recovery_generation = event.recovery_generation + 1,
    recovery_enqueued_at = NOW()
FROM candidates
WHERE event.id = candidates.id
RETURNING event.id,
          event.workspace_id,
          event.installation_generation,
          event.external_workspace_id,
          event.external_event_id,
          event.status,
          event.attempt_count,
          event.recovery_generation,
          event.processed_at,
          event.payload_encrypted;

-- name: MarkInboundEventQueued :execrows
UPDATE messaging_inbound_events
SET recovery_enqueued_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status IN ('pending', 'failed', 'processing');

-- name: ReleaseInboundEventRecovery :exec
UPDATE messaging_inbound_events
SET recovery_enqueued_at = NULL
WHERE id = sqlc.arg(id)
  AND recovery_generation = sqlc.arg(recovery_generation);

-- name: GetInboundEvent :one
SELECT id,
       workspace_id,
       installation_generation,
       external_workspace_id,
       external_event_id,
       status,
       attempt_count,
       recovery_generation,
       processed_at,
       payload_encrypted
FROM messaging_inbound_events
WHERE provider = sqlc.arg(provider)
  AND external_workspace_id = sqlc.arg(external_workspace_id)
  AND external_event_id = sqlc.arg(external_event_id);

-- name: ClaimInboundEvent :one
UPDATE messaging_inbound_events
SET status = 'processing',
    attempt_count = attempt_count + 1,
    last_error = NULL,
    recovery_enqueued_at = NULL,
    updated_at = NOW()
WHERE provider = sqlc.arg(provider)
  AND external_workspace_id = sqlc.arg(external_workspace_id)
  AND external_event_id = sqlc.arg(external_event_id)
  AND (
      status IN ('pending', 'failed')
      OR (
          status = 'processing'
          AND updated_at < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
      )
  )
RETURNING id,
          workspace_id,
          installation_generation,
          external_workspace_id,
          external_event_id,
          status,
          attempt_count,
          recovery_generation,
          processed_at,
          payload_encrypted;

-- name: CompleteInboundEvent :exec
UPDATE messaging_inbound_events
SET status = sqlc.arg(status),
    last_error = NULLIF(CAST(sqlc.arg(last_error) AS text), ''),
    processed_at = COALESCE(sqlc.narg(processed_at), processed_at),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'processing';
