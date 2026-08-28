-- name: FindIntegrationRequestCommentByClientKey :one
SELECT
    request_comment.id,
    request_comment.workspace_id,
    request_comment.thread_id,
    request_comment.direction,
    request_comment.author_user_id,
    COALESCE(actor_user.full_name, actor_user.email, NULLIF(request_comment.external_author_id, ''), 'Slack user') AS author_name,
    actor_user.avatar_url AS author_avatar,
    request_comment.external_author_id,
    request_comment.external_message_id,
    request_comment.client_idempotency_key,
    request_comment.outbound_idempotency_key,
    request_comment.delivery_status,
    request_comment.body,
    request_comment.created_at,
    request_comment.updated_at
FROM public.integration_request_comments AS request_comment
LEFT JOIN public.users AS actor_user
    ON actor_user.user_id = request_comment.author_user_id
WHERE request_comment.workspace_id = sqlc.arg(workspace_id)
  AND request_comment.client_idempotency_key = sqlc.arg(client_idempotency_key);

-- name: CreateOutboundIntegrationRequestComment :one
INSERT INTO public.integration_request_comments (
    id,
    workspace_id,
    thread_id,
    direction,
    author_user_id,
    client_idempotency_key,
    outbound_idempotency_key,
    delivery_status,
    body
) VALUES (
    sqlc.arg(comment_id),
    sqlc.arg(workspace_id),
    sqlc.arg(thread_id),
    'outbound',
    sqlc.arg(actor_id),
    sqlc.arg(client_idempotency_key),
    sqlc.arg(outbound_idempotency_key),
    'sending',
    sqlc.arg(body)
)
ON CONFLICT (workspace_id, client_idempotency_key)
WHERE client_idempotency_key IS NOT NULL
DO NOTHING
RETURNING
    id,
    workspace_id,
    thread_id,
    direction,
    author_user_id,
    CAST(COALESCE((SELECT COALESCE(full_name, email) FROM public.users WHERE user_id = sqlc.arg(actor_id)), 'FortyOne user') AS text) AS author_name,
    (SELECT avatar_url FROM public.users WHERE user_id = sqlc.arg(actor_id)) AS author_avatar,
    external_author_id,
    external_message_id,
    client_idempotency_key,
    outbound_idempotency_key,
    delivery_status,
    body,
    created_at,
    updated_at;

-- name: CreateIntegrationRequestCommentDelivery :exec
INSERT INTO public.messaging_outbound_deliveries (
    provider,
    workspace_id,
    user_id,
    installation_generation,
    external_workspace_id,
    external_recipient_user_id,
    idempotency_key,
    external_channel_id,
    external_thread_id,
    content,
    provider_payload,
    purpose,
    status,
    attempt_count
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(workspace_id),
    sqlc.arg(actor_id),
    sqlc.narg(installation_generation),
    sqlc.arg(external_workspace_id),
    NULLIF(CAST(sqlc.arg(external_recipient_user_id) AS text), ''),
    sqlc.arg(idempotency_key),
    sqlc.arg(external_channel_id),
    sqlc.arg(external_thread_id),
    sqlc.arg(content),
    CAST(NULLIF(CAST(sqlc.arg(provider_payload) AS text), '') AS jsonb),
    'provider_message',
    'pending',
    0
);

-- name: IngestInboundIntegrationRequestComment :execrows
INSERT INTO public.integration_request_comments (
    workspace_id,
    thread_id,
    direction,
    author_user_id,
    external_author_id,
    external_message_id,
    body,
    created_at,
    updated_at
)
SELECT
    request_thread.workspace_id,
    request_thread.id,
    'inbound',
    sqlc.narg(author_user_id),
    NULLIF(CAST(sqlc.arg(external_author_id) AS text), ''),
    sqlc.arg(external_message_id),
    sqlc.arg(body),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
FROM public.integration_request_threads AS request_thread
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
WHERE request_thread.provider = sqlc.arg(provider)
  AND request_thread.external_workspace_id = sqlc.arg(external_workspace_id)
  AND request_thread.external_channel_id = sqlc.arg(external_channel_id)
  AND request_thread.external_thread_id = sqlc.arg(external_thread_id)
  AND request_thread.installation_generation = sqlc.arg(installation_generation)
ON CONFLICT (thread_id, external_message_id)
WHERE external_message_id IS NOT NULL
DO NOTHING;

-- name: GetAuthorizedIntegrationRequestComment :one
SELECT
    request_comment.id,
    request_comment.workspace_id,
    request_comment.thread_id,
    request_comment.direction,
    request_comment.author_user_id,
    COALESCE(actor_user.full_name, actor_user.email, NULLIF(request_comment.external_author_id, ''), 'Slack user') AS author_name,
    actor_user.avatar_url AS author_avatar,
    request_comment.external_author_id,
    request_comment.external_message_id,
    request_comment.client_idempotency_key,
    request_comment.outbound_idempotency_key,
    request_comment.delivery_status,
    request_comment.body,
    request_comment.created_at,
    request_comment.updated_at
FROM public.integration_request_comments AS request_comment
LEFT JOIN public.users AS actor_user
    ON actor_user.user_id = request_comment.author_user_id
INNER JOIN public.integration_request_threads AS request_thread
    ON request_thread.id = request_comment.thread_id
INNER JOIN public.integration_requests AS request
    ON request.id = request_thread.integration_request_id
WHERE request_comment.workspace_id = sqlc.arg(workspace_id)
  AND request_comment.id = sqlc.arg(comment_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  );

-- name: ListIntegrationRequestThreadComments :many
SELECT
    request_comment.id,
    request_comment.workspace_id,
    request_comment.thread_id,
    request_comment.direction,
    request_comment.author_user_id,
    COALESCE(actor_user.full_name, actor_user.email, NULLIF(request_comment.external_author_id, ''), 'Slack user') AS author_name,
    actor_user.avatar_url AS author_avatar,
    request_comment.external_author_id,
    request_comment.external_message_id,
    request_comment.client_idempotency_key,
    request_comment.outbound_idempotency_key,
    request_comment.delivery_status,
    request_comment.body,
    request_comment.created_at,
    request_comment.updated_at
FROM public.integration_request_comments AS request_comment
LEFT JOIN public.users AS actor_user
    ON actor_user.user_id = request_comment.author_user_id
WHERE request_comment.thread_id = sqlc.arg(thread_id)
ORDER BY request_comment.created_at ASC, request_comment.id ASC;
