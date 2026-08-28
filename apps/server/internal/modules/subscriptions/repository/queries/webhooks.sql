-- name: ClaimStripeWebhookEvent :one
INSERT INTO stripe_webhook_events (
    event_id,
    event_type,
    processing_state,
    attempts,
    first_received_at,
    last_attempted_at,
    lease_expires_at,
    lease_token
) VALUES (
    sqlc.arg(event_id),
    sqlc.arg(event_type),
    'processing',
    1,
    sqlc.arg(attempted_at),
    sqlc.arg(attempted_at),
    sqlc.arg(lease_expires_at),
    sqlc.arg(lease_token)
)
ON CONFLICT (event_id) DO UPDATE
SET processing_state = 'processing',
    processing_result = NULL,
    attempts = stripe_webhook_events.attempts + 1,
    last_attempted_at = EXCLUDED.last_attempted_at,
    lease_expires_at = EXCLUDED.lease_expires_at,
    lease_token = EXCLUDED.lease_token,
    processed_at = NULL,
    failed_at = NULL,
    last_error_code = NULL
WHERE stripe_webhook_events.event_type = EXCLUDED.event_type
  AND (
      stripe_webhook_events.processing_state = 'failed'
      OR (
          stripe_webhook_events.processing_state = 'processing'
          AND stripe_webhook_events.lease_expires_at <= EXCLUDED.last_attempted_at
      )
  )
RETURNING lease_token, attempts;

-- name: GetStripeWebhookClaimState :one
SELECT event_type, processing_state, attempts
FROM stripe_webhook_events
WHERE event_id = sqlc.arg(event_id);

-- name: CompleteStripeWebhookEvent :execrows
UPDATE stripe_webhook_events
SET processing_state = 'processed',
    processing_result = sqlc.arg(processing_result),
    workspace_id = sqlc.narg(workspace_id),
    processed_at = sqlc.arg(processed_at),
    lease_expires_at = NULL,
    lease_token = NULL,
    failed_at = NULL,
    last_error_code = NULL
WHERE event_id = sqlc.arg(event_id)
  AND processing_state = 'processing'
  AND lease_token = sqlc.arg(lease_token);

-- name: FailStripeWebhookEvent :execrows
UPDATE stripe_webhook_events
SET processing_state = 'failed',
    processing_result = NULL,
    processed_at = NULL,
    lease_expires_at = NULL,
    lease_token = NULL,
    failed_at = sqlc.arg(failed_at),
    last_error_code = sqlc.arg(last_error_code)
WHERE event_id = sqlc.arg(event_id)
  AND processing_state = 'processing'
  AND lease_token = sqlc.arg(lease_token);
