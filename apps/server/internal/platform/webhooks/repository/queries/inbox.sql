-- name: InsertDelivery :one
INSERT INTO public.messaging_inbound_events (
    envelope_version,
    provider,
    workspace_id,
    installation_id,
    installation_generation,
    external_workspace_id,
    external_event_id,
    event_type,
    trace_id,
    payload_encrypted,
    payload_expires_at,
    received_at
) VALUES (
    sqlc.arg(envelope_version),
    sqlc.arg(provider),
    CAST(sqlc.narg(workspace_id) AS uuid),
    CAST(sqlc.narg(installation_id) AS uuid),
    CAST(sqlc.narg(installation_generation) AS uuid),
    sqlc.arg(external_account_id),
    sqlc.arg(delivery_id),
    sqlc.arg(event_type),
    CAST(sqlc.narg(trace_id) AS text),
    CAST(sqlc.arg(payload_encrypted) AS text),
    CAST(sqlc.narg(payload_expires_at) AS timestamptz),
    sqlc.arg(received_at)
)
ON CONFLICT (provider, external_workspace_id, external_event_id) DO NOTHING
RETURNING messaging_inbound_events.*;

-- name: ReadDuplicateDelivery :one
UPDATE public.messaging_inbound_events AS inbox
SET payload_encrypted = CASE
        WHEN inbox.status IN ('pending', 'processing', 'failed')
            THEN COALESCE(inbox.payload_encrypted, CAST(sqlc.arg(payload_encrypted) AS text))
        ELSE inbox.payload_encrypted
    END,
    payload_expires_at = CASE
        WHEN inbox.status IN ('pending', 'processing', 'failed') AND inbox.payload_encrypted IS NULL
            THEN CAST(sqlc.narg(payload_expires_at) AS timestamptz)
        ELSE inbox.payload_expires_at
    END
WHERE inbox.provider = sqlc.arg(provider)
  AND inbox.external_workspace_id = sqlc.arg(external_account_id)
  AND inbox.external_event_id = sqlc.arg(delivery_id)
RETURNING inbox.*;

-- name: GetDeliveryByExternalKey :one
SELECT inbox.*
FROM public.messaging_inbound_events AS inbox
WHERE inbox.provider = sqlc.arg(provider)
  AND inbox.external_workspace_id = sqlc.arg(external_account_id)
  AND inbox.external_event_id = sqlc.arg(delivery_id);

-- name: GetDeliveryByID :one
SELECT inbox.*
FROM public.messaging_inbound_events AS inbox
WHERE inbox.id = sqlc.arg(id);

-- name: MarkDeliveryQueued :execrows
UPDATE public.messaging_inbound_events AS inbox
SET recovery_enqueued_at = CAST(sqlc.arg(queued_at) AS timestamptz),
    updated_at = CAST(sqlc.arg(queued_at) AS timestamptz)
WHERE inbox.id = sqlc.arg(id)
  AND inbox.status IN ('pending', 'processing', 'failed');

-- name: ClaimRecoverableDeliveries :many
WITH candidates AS (
    SELECT id
    FROM public.messaging_inbound_events AS inbox
    WHERE inbox.provider = sqlc.arg(provider)
      AND inbox.payload_encrypted IS NOT NULL
      AND inbox.attempt_count < sqlc.arg(max_attempts)
      AND (
          (inbox.status = 'pending' AND inbox.updated_at < CAST(sqlc.arg(now) AS timestamptz) - (CAST(sqlc.arg(pending_age_seconds) AS bigint) * INTERVAL '1 second'))
          OR (inbox.status = 'failed' AND inbox.updated_at < CAST(sqlc.arg(now) AS timestamptz) - (CAST(sqlc.arg(failed_age_seconds) AS bigint) * INTERVAL '1 second'))
          OR (inbox.status = 'processing' AND inbox.updated_at < CAST(sqlc.arg(now) AS timestamptz) - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second'))
      )
      AND (
          inbox.recovery_enqueued_at IS NULL
          OR inbox.recovery_enqueued_at < CAST(sqlc.arg(now) AS timestamptz) - (
              (CAST(sqlc.arg(recovery_base_seconds) AS bigint) * POWER(2, LEAST(inbox.recovery_generation, CAST(sqlc.arg(recovery_max_shift) AS integer)))) * INTERVAL '1 second'
          )
      )
    ORDER BY COALESCE(inbox.recovery_enqueued_at, inbox.received_at), inbox.received_at, inbox.id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(claim_limit)
)
UPDATE public.messaging_inbound_events AS delivery
SET recovery_generation = delivery.recovery_generation + 1,
    recovery_enqueued_at = CAST(sqlc.arg(now) AS timestamptz),
    updated_at = CAST(sqlc.arg(now) AS timestamptz)
FROM candidates
WHERE delivery.id = candidates.id
RETURNING delivery.*;

-- name: ReleaseDeliveryRecovery :execrows
UPDATE public.messaging_inbound_events AS inbox
SET recovery_enqueued_at = NULL,
    updated_at = sqlc.arg(released_at)
WHERE inbox.id = sqlc.arg(id)
  AND inbox.recovery_generation = sqlc.arg(recovery_generation);

-- name: TryStartDelivery :one
UPDATE public.messaging_inbound_events AS inbox
SET status = 'processing',
    attempt_count = inbox.attempt_count + 1,
    last_error = NULL,
    recovery_enqueued_at = NULL,
    updated_at = CAST(sqlc.arg(now) AS timestamptz)
WHERE inbox.id = sqlc.arg(id)
  AND (
      inbox.status IN ('pending', 'failed')
      OR (inbox.status = 'processing' AND inbox.updated_at < CAST(sqlc.arg(now) AS timestamptz) - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second'))
  )
RETURNING inbox.*;

-- name: CompleteDelivery :execrows
UPDATE public.messaging_inbound_events AS inbox
SET status = sqlc.arg(status),
    last_error = NULLIF(CAST(sqlc.arg(safe_message) AS text), ''),
    processed_at = CASE
        WHEN sqlc.arg(status) IN ('completed', 'ignored', 'cancelled') THEN sqlc.arg(completed_at)
        ELSE inbox.processed_at
    END,
    updated_at = sqlc.arg(completed_at)
WHERE inbox.id = sqlc.arg(id)
  AND inbox.status = 'processing';

-- name: ExpireDeliveryPayloads :many
UPDATE public.messaging_inbound_events AS delivery
SET payload_encrypted = NULL,
    updated_at = sqlc.arg(now)
WHERE id IN (
    SELECT id
    FROM public.messaging_inbound_events AS candidate
    WHERE payload_encrypted IS NOT NULL
      AND payload_expires_at IS NOT NULL
      AND payload_expires_at <= sqlc.arg(now)
    ORDER BY payload_expires_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(expiry_limit)
)
RETURNING id;
