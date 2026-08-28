-- name: CreateOutboundWebhookEndpoint :one
INSERT INTO public.outbound_webhook_endpoints (
    endpoint_id,
    workspace_id,
    owner_principal_id,
    name,
    endpoint_url,
    signing_secret_envelope,
    secret_generation,
    subscription_generation,
    created_by_user_id,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(endpoint_id),
    sqlc.arg(workspace_id),
    sqlc.arg(owner_principal_id),
    sqlc.arg(name),
    sqlc.arg(endpoint_url),
    sqlc.arg(signing_secret_envelope),
    1,
    1,
    sqlc.narg(created_by_user_id),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
)
RETURNING
    endpoint_id,
    workspace_id,
    owner_principal_id,
    name,
    endpoint_url,
    status,
    secret_generation,
    subscription_generation,
    consecutive_failures,
    last_success_at,
    disabled_at,
    disabled_reason,
    created_at,
    updated_at;

-- name: AddOutboundWebhookSubscription :exec
INSERT INTO public.outbound_webhook_subscriptions (
    endpoint_id,
    workspace_id,
    event_type,
    created_at
) VALUES (
    sqlc.arg(endpoint_id),
    sqlc.arg(workspace_id),
    sqlc.arg(event_type),
    sqlc.arg(created_at)
);

-- name: DeleteOutboundWebhookSubscriptions :exec
DELETE FROM public.outbound_webhook_subscriptions
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id);

-- name: ListOutboundWebhookSubscriptions :many
SELECT event_type
FROM public.outbound_webhook_subscriptions
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id)
ORDER BY event_type;

-- name: GetOutboundWebhookEndpoint :one
SELECT
    endpoint_id,
    workspace_id,
    owner_principal_id,
    name,
    endpoint_url,
    status,
    secret_generation,
    subscription_generation,
    consecutive_failures,
    last_success_at,
    disabled_at,
    disabled_reason,
    created_at,
    updated_at,
    CAST(ARRAY(
        SELECT subscription.event_type
        FROM public.outbound_webhook_subscriptions AS subscription
        WHERE subscription.endpoint_id = endpoint.endpoint_id
          AND subscription.workspace_id = endpoint.workspace_id
        ORDER BY subscription.event_type
    ) AS text[]) AS subscriptions
FROM public.outbound_webhook_endpoints AS endpoint
WHERE endpoint.endpoint_id = sqlc.arg(endpoint_id)
  AND endpoint.workspace_id = sqlc.arg(workspace_id);

-- name: ListOutboundWebhookEndpoints :many
SELECT
    endpoint_id,
    workspace_id,
    owner_principal_id,
    name,
    endpoint_url,
    status,
    secret_generation,
    subscription_generation,
    consecutive_failures,
    last_success_at,
    disabled_at,
    disabled_reason,
    created_at,
    updated_at,
    CAST(ARRAY(
        SELECT subscription.event_type
        FROM public.outbound_webhook_subscriptions AS subscription
        WHERE subscription.endpoint_id = endpoint.endpoint_id
          AND subscription.workspace_id = endpoint.workspace_id
        ORDER BY subscription.event_type
    ) AS text[]) AS subscriptions
FROM public.outbound_webhook_endpoints AS endpoint
WHERE endpoint.workspace_id = sqlc.arg(workspace_id)
  AND (
      CAST(sqlc.narg(cursor_created_at) AS timestamptz) IS NULL
      OR (endpoint.created_at, endpoint.endpoint_id) < (
          CAST(sqlc.narg(cursor_created_at) AS timestamptz),
          CAST(sqlc.narg(cursor_endpoint_id) AS uuid)
      )
  )
ORDER BY endpoint.created_at DESC, endpoint.endpoint_id DESC
LIMIT sqlc.arg(page_size);

-- name: LockOutboundWebhookEndpoint :one
SELECT
    endpoint_id,
    workspace_id,
    owner_principal_id,
    status,
    secret_generation,
    subscription_generation
FROM public.outbound_webhook_endpoints
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id)
FOR UPDATE;

-- name: ReplaceOutboundWebhookSubscriptionGeneration :one
UPDATE public.outbound_webhook_endpoints
SET
    subscription_generation = subscription_generation + 1,
    updated_at = sqlc.arg(updated_at)
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status = 'active'
RETURNING subscription_generation;

-- name: RotateOutboundWebhookSigningSecret :one
UPDATE public.outbound_webhook_endpoints
SET
    previous_secret_envelope = signing_secret_envelope,
    previous_secret_generation = secret_generation,
    previous_secret_expires_at = sqlc.arg(previous_secret_expires_at),
    signing_secret_envelope = sqlc.arg(signing_secret_envelope),
    secret_generation = secret_generation + 1,
    updated_at = sqlc.arg(rotated_at)
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status = 'active'
  AND secret_generation = sqlc.arg(expected_secret_generation)
RETURNING secret_generation;

-- name: DisableOutboundWebhookEndpoint :one
UPDATE public.outbound_webhook_endpoints
SET
    status = 'disabled',
    disabled_at = sqlc.arg(disabled_at),
    disabled_reason = sqlc.arg(disabled_reason),
    updated_at = sqlc.arg(disabled_at)
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status = 'active'
RETURNING endpoint_id;

-- name: CancelPendingOutboundWebhookDeliveries :execrows
UPDATE public.outbound_webhook_deliveries
SET
    status = 'cancelled',
    lease_token = NULL,
    lease_expires_at = NULL,
    last_error_code = sqlc.arg(error_code),
    completed_at = sqlc.arg(cancelled_at),
    updated_at = sqlc.arg(cancelled_at)
WHERE endpoint_id = sqlc.arg(endpoint_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status IN ('pending', 'retry_scheduled');

-- name: EnsureOutboundWebhookOwnerActive :one
SELECT principal.principal_id
FROM public.principals AS principal
LEFT JOIN public.users AS account
    ON account.user_id = principal.subject_user_id
LEFT JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
WHERE principal.principal_id = sqlc.arg(principal_id)
  AND principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.status = 'active'
  AND principal.kind = 'human_user'
  AND account.is_active = TRUE
  AND membership.user_id IS NOT NULL;
