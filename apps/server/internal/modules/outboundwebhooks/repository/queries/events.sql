-- name: CreateOutboundWebhookEvent :one
INSERT INTO public.outbound_webhook_events (
    event_id,
    workspace_id,
    event_type,
    payload_version,
    subject_type,
    subject_id,
    actor_kind,
    actor_id,
    actor_credential_id,
    payload,
    occurred_at,
    created_at
) VALUES (
    sqlc.arg(event_id),
    sqlc.arg(workspace_id),
    sqlc.arg(event_type),
    sqlc.arg(payload_version),
    sqlc.arg(subject_type),
    sqlc.arg(subject_id),
    sqlc.arg(actor_kind),
    sqlc.arg(actor_id),
    sqlc.narg(actor_credential_id),
    sqlc.arg(payload),
    sqlc.arg(occurred_at),
    sqlc.arg(created_at)
)
RETURNING
    event_id,
    workspace_id,
    event_type,
    payload_version,
    subject_type,
    subject_id,
    actor_kind,
    actor_id,
    actor_credential_id,
    payload,
    occurred_at,
    created_at;

-- name: GetOutboundWebhookEvent :one
SELECT
    event_id,
    workspace_id,
    event_type,
    payload_version,
    subject_type,
    subject_id,
    actor_kind,
    actor_id,
    actor_credential_id,
    payload = CAST(sqlc.arg(expected_payload) AS jsonb) AS payload_matches,
    occurred_at,
    created_at
FROM public.outbound_webhook_events
WHERE event_id = sqlc.arg(event_id);

-- name: CreateOutboundWebhookDeliveries :many
INSERT INTO public.outbound_webhook_deliveries (
    delivery_id,
    workspace_id,
    event_id,
    endpoint_id,
    subscription_generation,
    payload_body,
    status,
    attempt_count,
    available_at,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    endpoint.workspace_id,
    sqlc.arg(event_id),
    endpoint.endpoint_id,
    endpoint.subscription_generation,
    sqlc.arg(payload_body),
    'pending',
    0,
    sqlc.arg(created_at),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
FROM public.outbound_webhook_endpoints AS endpoint
INNER JOIN public.outbound_webhook_subscriptions AS subscription
    ON subscription.endpoint_id = endpoint.endpoint_id
    AND subscription.workspace_id = endpoint.workspace_id
INNER JOIN public.principals AS principal
    ON principal.principal_id = endpoint.owner_principal_id
    AND principal.workspace_id = endpoint.workspace_id
LEFT JOIN public.users AS account
    ON account.user_id = principal.subject_user_id
LEFT JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
WHERE endpoint.workspace_id = sqlc.arg(workspace_id)
  AND endpoint.status = 'active'
  AND subscription.event_type = sqlc.arg(event_type)
  AND principal.status = 'active'
  AND principal.kind = 'human_user'
  AND account.is_active = TRUE
  AND membership.user_id IS NOT NULL
ON CONFLICT (event_id, endpoint_id) DO NOTHING
RETURNING delivery_id;
