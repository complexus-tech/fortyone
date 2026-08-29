-- name: ListRecoverableOutboundDeliveries :many
SELECT id,
       workspace_id,
       user_id,
       installation_generation,
       external_workspace_id,
       external_recipient_user_id,
       inbound_event_id,
       idempotency_key,
       external_channel_id,
       external_thread_id,
       external_message_id,
       content,
       provider_payload,
       purpose,
       expires_at,
       status,
       attempt_count
FROM messaging_outbound_deliveries
WHERE provider = sqlc.arg(provider)
  AND content IS NOT NULL
  AND attempt_count < 20
  AND (
      (status IN ('pending', 'failed') AND updated_at < NOW() - INTERVAL '5 minutes')
      OR (
          status = 'delivering'
          AND updated_at < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
      )
  )
ORDER BY created_at
LIMIT sqlc.arg(row_limit);

-- name: ClaimOutboundDelivery :one
INSERT INTO messaging_outbound_deliveries (
    provider,
    workspace_id,
    user_id,
    installation_generation,
    external_workspace_id,
    external_recipient_user_id,
    inbound_event_id,
    idempotency_key,
    external_channel_id,
    external_thread_id,
    content,
    provider_payload,
    purpose,
    expires_at,
    status,
    attempt_count
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(workspace_id),
    sqlc.narg(user_id),
    sqlc.narg(installation_generation),
    sqlc.arg(external_workspace_id),
    NULLIF(CAST(sqlc.arg(external_recipient_user_id) AS text), ''),
    sqlc.narg(inbound_event_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(external_channel_id),
    NULLIF(CAST(sqlc.arg(external_thread_id) AS text), ''),
    NULLIF(CAST(sqlc.arg(content) AS text), ''),
    CAST(NULLIF(CAST(sqlc.arg(provider_payload) AS text), '') AS jsonb),
    sqlc.arg(purpose),
    sqlc.narg(expires_at),
    'delivering',
    1
)
ON CONFLICT (provider, workspace_id, idempotency_key) DO UPDATE
SET status = 'delivering',
    attempt_count = messaging_outbound_deliveries.attempt_count + 1,
    content = COALESCE(messaging_outbound_deliveries.content, EXCLUDED.content),
    provider_payload = COALESCE(messaging_outbound_deliveries.provider_payload, EXCLUDED.provider_payload),
    user_id = COALESCE(messaging_outbound_deliveries.user_id, EXCLUDED.user_id),
    installation_generation = COALESCE(
        messaging_outbound_deliveries.installation_generation,
        EXCLUDED.installation_generation
    ),
    external_recipient_user_id = COALESCE(
        messaging_outbound_deliveries.external_recipient_user_id,
        EXCLUDED.external_recipient_user_id
    ),
    expires_at = COALESCE(messaging_outbound_deliveries.expires_at, EXCLUDED.expires_at),
    last_error = NULL,
    updated_at = NOW()
WHERE (
    messaging_outbound_deliveries.status IN ('pending', 'failed')
    OR (
        messaging_outbound_deliveries.status = 'delivering'
        AND messaging_outbound_deliveries.updated_at
            < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
    )
)
AND messaging_outbound_deliveries.external_workspace_id = EXCLUDED.external_workspace_id
AND messaging_outbound_deliveries.purpose = EXCLUDED.purpose
AND messaging_outbound_deliveries.user_id IS NOT DISTINCT FROM EXCLUDED.user_id
AND messaging_outbound_deliveries.installation_generation
    IS NOT DISTINCT FROM EXCLUDED.installation_generation
AND messaging_outbound_deliveries.external_recipient_user_id
    IS NOT DISTINCT FROM EXCLUDED.external_recipient_user_id
RETURNING id,
          workspace_id,
          user_id,
          installation_generation,
          external_workspace_id,
          external_recipient_user_id,
          inbound_event_id,
          idempotency_key,
          external_channel_id,
          external_thread_id,
          external_message_id,
          content,
          provider_payload,
          purpose,
          expires_at,
          status,
          attempt_count;

-- name: GetOutboundDelivery :one
SELECT id,
       workspace_id,
       user_id,
       installation_generation,
       external_workspace_id,
       external_recipient_user_id,
       inbound_event_id,
       idempotency_key,
       external_channel_id,
       external_thread_id,
       external_message_id,
       content,
       provider_payload,
       purpose,
       expires_at,
       status,
       attempt_count
FROM messaging_outbound_deliveries
WHERE provider = sqlc.arg(provider)
  AND workspace_id = sqlc.arg(workspace_id)
  AND idempotency_key = sqlc.arg(idempotency_key);

-- name: SetOutboundDeliveryContent :execrows
UPDATE messaging_outbound_deliveries
SET content = sqlc.arg(content),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'delivering';

-- name: SetOutboundDeliveryContentAndProviderPayload :execrows
UPDATE messaging_outbound_deliveries
SET content = sqlc.arg(content),
    provider_payload = CAST(sqlc.arg(provider_payload) AS jsonb),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'delivering'
  AND external_message_id IS NULL;

-- name: SetOutboundDeliveryContentAndDestination :execrows
UPDATE messaging_outbound_deliveries
SET content = sqlc.arg(content),
    external_channel_id = sqlc.arg(external_channel_id),
    external_thread_id = NULLIF(CAST(sqlc.arg(external_thread_id) AS text), ''),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'delivering'
  AND external_message_id IS NULL;

-- name: CompleteOutboundDelivery :execrows
UPDATE messaging_outbound_deliveries
SET status = 'delivered',
    external_message_id = NULLIF(CAST(sqlc.arg(external_message_id) AS text), ''),
    delivered_at = NOW(),
    last_error = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'delivering';

-- name: FailOutboundDelivery :exec
UPDATE messaging_outbound_deliveries
SET status = 'failed',
    last_error = sqlc.arg(last_error),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'delivering';

-- name: CancelOutboundDelivery :exec
UPDATE messaging_outbound_deliveries
SET status = 'cancelled',
    content = NULL,
    last_error = sqlc.arg(last_error),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status IN ('pending', 'delivering', 'failed');
