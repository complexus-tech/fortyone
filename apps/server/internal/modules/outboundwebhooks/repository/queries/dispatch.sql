-- name: LockNextOutboundWebhookEndpointForDelivery :one
SELECT endpoint.endpoint_id
FROM public.outbound_webhook_endpoints AS endpoint
WHERE endpoint.status = 'active'
  AND NOT EXISTS (
      SELECT 1
      FROM public.outbound_webhook_deliveries AS in_flight
      WHERE in_flight.endpoint_id = endpoint.endpoint_id
        AND in_flight.workspace_id = endpoint.workspace_id
        AND in_flight.status = 'delivering'
  )
  AND EXISTS (
      SELECT 1
      FROM public.outbound_webhook_deliveries AS delivery
      INNER JOIN public.outbound_webhook_events AS event
          ON event.event_id = delivery.event_id
          AND event.workspace_id = delivery.workspace_id
      INNER JOIN public.outbound_webhook_subscriptions AS subscription
          ON subscription.endpoint_id = delivery.endpoint_id
          AND subscription.workspace_id = delivery.workspace_id
          AND subscription.event_type = event.event_type
      WHERE delivery.endpoint_id = endpoint.endpoint_id
        AND delivery.workspace_id = endpoint.workspace_id
        AND delivery.subscription_generation = endpoint.subscription_generation
        AND delivery.status IN ('pending', 'retry_scheduled')
        AND delivery.available_at <= sqlc.arg(claimed_at)
  )
  AND EXISTS (
      SELECT 1
      FROM public.principals AS principal
      LEFT JOIN public.users AS account
          ON account.user_id = principal.subject_user_id
      LEFT JOIN public.workspace_members AS membership
          ON membership.workspace_id = principal.workspace_id
          AND membership.user_id = principal.subject_user_id
      WHERE principal.principal_id = endpoint.owner_principal_id
        AND principal.workspace_id = endpoint.workspace_id
        AND principal.status = 'active'
        AND principal.kind = 'human_user'
        AND account.is_active = TRUE
        AND membership.user_id IS NOT NULL
  )
ORDER BY
    COALESCE(endpoint.last_delivery_claimed_at, CAST('epoch' AS timestamptz)),
    endpoint.endpoint_id
FOR UPDATE OF endpoint SKIP LOCKED
LIMIT 1;

-- name: ClaimNextOutboundWebhookDelivery :one
WITH candidate AS (
    SELECT delivery.delivery_id
    FROM public.outbound_webhook_deliveries AS delivery
    INNER JOIN public.outbound_webhook_events AS event
        ON event.event_id = delivery.event_id
        AND event.workspace_id = delivery.workspace_id
    INNER JOIN public.outbound_webhook_endpoints AS endpoint
        ON endpoint.endpoint_id = delivery.endpoint_id
        AND endpoint.workspace_id = delivery.workspace_id
    INNER JOIN public.outbound_webhook_subscriptions AS subscription
        ON subscription.endpoint_id = delivery.endpoint_id
        AND subscription.workspace_id = delivery.workspace_id
        AND subscription.event_type = event.event_type
    WHERE delivery.endpoint_id = sqlc.arg(endpoint_id)
      AND delivery.status IN ('pending', 'retry_scheduled')
      AND delivery.available_at <= sqlc.arg(claimed_at)
      AND delivery.subscription_generation = endpoint.subscription_generation
      AND endpoint.status = 'active'
    ORDER BY delivery.available_at, delivery.created_at, delivery.delivery_id
    FOR UPDATE OF delivery SKIP LOCKED
    LIMIT 1
), claimed AS (
    UPDATE public.outbound_webhook_deliveries AS delivery
    SET
        status = 'delivering',
        attempt_count = delivery.attempt_count + 1,
        lease_token = sqlc.arg(lease_token),
        lease_expires_at = sqlc.arg(lease_expires_at),
        updated_at = sqlc.arg(claimed_at)
    WHERE delivery.delivery_id IN (SELECT candidate.delivery_id FROM candidate)
    RETURNING delivery.*
)
SELECT
    claimed.delivery_id,
    claimed.workspace_id,
    claimed.event_id,
    event.event_type,
    claimed.endpoint_id,
    endpoint.endpoint_url,
    endpoint.signing_secret_envelope,
    endpoint.secret_generation,
    endpoint.previous_secret_envelope,
    endpoint.previous_secret_generation,
    endpoint.previous_secret_expires_at,
    claimed.subscription_generation,
    claimed.payload_body,
    claimed.attempt_count,
    claimed.lease_token,
    claimed.lease_expires_at,
    claimed.created_at
FROM claimed
INNER JOIN public.outbound_webhook_events AS event
    ON event.event_id = claimed.event_id
    AND event.workspace_id = claimed.workspace_id
INNER JOIN public.outbound_webhook_endpoints AS endpoint
    ON endpoint.endpoint_id = claimed.endpoint_id
    AND endpoint.workspace_id = claimed.workspace_id;

-- name: TouchOutboundWebhookEndpointClaim :execrows
UPDATE public.outbound_webhook_endpoints
SET
    last_delivery_claimed_at = sqlc.arg(claimed_at),
    updated_at = GREATEST(updated_at, sqlc.arg(claimed_at))
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND status = 'active';

-- name: RecoverExpiredOutboundWebhookLeases :execrows
UPDATE public.outbound_webhook_deliveries
SET
    status = 'retry_scheduled',
    lease_token = NULL,
    lease_expires_at = NULL,
    last_error_code = 'lease_expired',
    available_at = sqlc.arg(recovered_at),
    updated_at = sqlc.arg(recovered_at)
WHERE status = 'delivering'
  AND lease_expires_at <= sqlc.arg(recovered_at);

-- name: CompleteOutboundWebhookDelivery :execrows
UPDATE public.outbound_webhook_deliveries
SET
    status = sqlc.arg(status),
    lease_token = NULL,
    lease_expires_at = NULL,
    last_http_status = sqlc.narg(http_status),
    last_error_code = sqlc.narg(error_code),
    available_at = COALESCE(sqlc.narg(next_attempt_at), available_at),
    completed_at = sqlc.narg(completed_at),
    updated_at = sqlc.arg(finished_at)
WHERE delivery_id = sqlc.arg(delivery_id)
  AND status = 'delivering'
  AND lease_token = sqlc.arg(lease_token)
  AND attempt_count = sqlc.arg(attempt_number);

-- name: RecordOutboundWebhookDeliveryAttempt :exec
INSERT INTO public.outbound_webhook_delivery_attempts (
    attempt_id,
    delivery_id,
    attempt_number,
    outcome,
    resolved_ip,
    http_status,
    response_bytes,
    response_digest,
    error_code,
    duration_ms,
    started_at,
    finished_at
) VALUES (
    sqlc.arg(attempt_id),
    sqlc.arg(delivery_id),
    sqlc.arg(attempt_number),
    sqlc.arg(outcome),
    sqlc.narg(resolved_ip),
    sqlc.narg(http_status),
    sqlc.narg(response_bytes),
    sqlc.narg(response_digest),
    sqlc.narg(error_code),
    sqlc.arg(duration_ms),
    sqlc.arg(started_at),
    sqlc.arg(finished_at)
);

-- name: RecordOutboundWebhookEndpointSuccess :execrows
UPDATE public.outbound_webhook_endpoints
SET
    consecutive_failures = 0,
    last_success_at = sqlc.arg(succeeded_at),
    updated_at = GREATEST(updated_at, sqlc.arg(succeeded_at))
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id);

-- name: RecordOutboundWebhookEndpointFailure :execrows
UPDATE public.outbound_webhook_endpoints
SET
    consecutive_failures = consecutive_failures + 1,
    status = CASE
        WHEN CAST(sqlc.arg(disable_endpoint) AS boolean)
          OR consecutive_failures + 1 >= sqlc.arg(disable_after_failures)
        THEN 'disabled'
        ELSE status
    END,
    disabled_at = CASE
        WHEN CAST(sqlc.arg(disable_endpoint) AS boolean)
          OR consecutive_failures + 1 >= sqlc.arg(disable_after_failures)
        THEN sqlc.arg(failed_at)
        ELSE disabled_at
    END,
    disabled_reason = CASE
        WHEN CAST(sqlc.arg(disable_endpoint) AS boolean)
          OR consecutive_failures + 1 >= sqlc.arg(disable_after_failures)
        THEN sqlc.arg(disabled_reason)
        ELSE disabled_reason
    END,
    updated_at = GREATEST(updated_at, sqlc.arg(failed_at))
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id);
