-- name: UpsertActorConversation :one
INSERT INTO messaging_conversations (
    provider,
    workspace_id,
    external_workspace_id,
    external_channel_id,
    external_thread_id,
    user_id,
    audience_scope,
    audience_fingerprint
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(workspace_id),
    sqlc.arg(external_workspace_id),
    sqlc.arg(external_channel_id),
    sqlc.arg(external_thread_id),
    sqlc.arg(user_id),
    'actor',
    ''
)
ON CONFLICT (
    provider,
    workspace_id,
    external_workspace_id,
    external_channel_id,
    external_thread_id,
    user_id
) WHERE audience_scope = 'actor'
DO UPDATE SET updated_at = NOW()
RETURNING id;

-- name: UpsertChannelConversation :one
INSERT INTO messaging_conversations (
    provider,
    workspace_id,
    external_workspace_id,
    external_channel_id,
    external_thread_id,
    user_id,
    audience_scope,
    audience_fingerprint
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(workspace_id),
    sqlc.arg(external_workspace_id),
    sqlc.arg(external_channel_id),
    sqlc.arg(external_thread_id),
    sqlc.arg(user_id),
    'channel',
    sqlc.arg(audience_fingerprint)
)
ON CONFLICT (
    provider,
    workspace_id,
    external_workspace_id,
    external_channel_id,
    external_thread_id,
    audience_fingerprint
) WHERE audience_scope = 'channel'
DO UPDATE SET updated_at = NOW()
RETURNING id;

-- name: FindActorConversation :one
SELECT id, updated_at
FROM messaging_conversations
WHERE provider = sqlc.arg(provider)
  AND workspace_id = sqlc.arg(workspace_id)
  AND external_workspace_id = sqlc.arg(external_workspace_id)
  AND external_channel_id = sqlc.arg(external_channel_id)
  AND external_thread_id = sqlc.arg(external_thread_id)
  AND audience_scope = 'actor'
  AND user_id = sqlc.arg(user_id)
ORDER BY updated_at DESC
LIMIT 1;

-- name: FindChannelConversationByFingerprint :one
SELECT id, updated_at
FROM messaging_conversations
WHERE provider = sqlc.arg(provider)
  AND workspace_id = sqlc.arg(workspace_id)
  AND external_workspace_id = sqlc.arg(external_workspace_id)
  AND external_channel_id = sqlc.arg(external_channel_id)
  AND external_thread_id = sqlc.arg(external_thread_id)
  AND audience_scope = 'channel'
  AND audience_fingerprint = sqlc.arg(audience_fingerprint)
ORDER BY updated_at DESC
LIMIT 1;

-- name: FindLatestChannelConversation :one
SELECT id, updated_at
FROM messaging_conversations
WHERE provider = sqlc.arg(provider)
  AND workspace_id = sqlc.arg(workspace_id)
  AND external_workspace_id = sqlc.arg(external_workspace_id)
  AND external_channel_id = sqlc.arg(external_channel_id)
  AND external_thread_id = sqlc.arg(external_thread_id)
  AND audience_scope = 'channel'
ORDER BY updated_at DESC
LIMIT 1;

-- name: AppendConversationMessage :exec
INSERT INTO messaging_messages (
    conversation_id,
    external_message_id,
    role,
    content
) VALUES (
    sqlc.arg(conversation_id),
    NULLIF(CAST(sqlc.arg(external_message_id) AS text), ''),
    sqlc.arg(role),
    sqlc.arg(content)
)
ON CONFLICT (conversation_id, external_message_id, role)
WHERE external_message_id IS NOT NULL
DO NOTHING;

-- name: ListRecentConversationMessages :many
SELECT external_message_id, role, content, created_at
FROM (
    SELECT external_message_id, role, content, created_at
    FROM messaging_messages
    WHERE conversation_id = sqlc.arg(conversation_id)
    ORDER BY created_at DESC
    LIMIT sqlc.arg(row_limit)
) AS recent
ORDER BY created_at;
