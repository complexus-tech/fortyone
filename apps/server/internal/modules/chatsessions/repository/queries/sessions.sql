-- name: UpsertChatSession :one
INSERT INTO public.chat_sessions AS session (
    id,
    user_id,
    workspace_id,
    title
)
VALUES (
    sqlc.arg(session_id),
    sqlc.arg(user_id),
    sqlc.arg(workspace_id),
    sqlc.arg(title)
)
ON CONFLICT (id) DO UPDATE
SET updated_at = CURRENT_TIMESTAMP
WHERE session.user_id = EXCLUDED.user_id
  AND session.workspace_id = EXCLUDED.workspace_id
  AND session.deleted_at IS NULL
RETURNING session.*;

-- name: InsertInitialChatMessages :execrows
INSERT INTO public.chat_messages (
    session_id,
    messages
)
VALUES (
    sqlc.arg(session_id),
    sqlc.arg(messages)
)
ON CONFLICT (session_id) DO NOTHING;

-- name: UpdateChatSessionTitle :execrows
UPDATE public.chat_sessions
SET
    title = sqlc.arg(title),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND deleted_at IS NULL;

-- name: SoftDeleteChatSession :execrows
UPDATE public.chat_sessions
SET
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND deleted_at IS NULL;

-- name: InitializeLegacyChatMessages :execrows
INSERT INTO public.chat_messages (
    session_id,
    messages
)
SELECT
    session.id,
    sqlc.arg(messages)
FROM public.chat_sessions AS session
WHERE session.id = sqlc.arg(session_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
ON CONFLICT (session_id) DO NOTHING;

-- name: ChatSessionExists :one
SELECT EXISTS (
    SELECT 1
    FROM public.chat_sessions AS session
    WHERE session.id = sqlc.arg(session_id)
      AND session.user_id = sqlc.arg(user_id)
      AND session.workspace_id = sqlc.arg(workspace_id)
      AND session.deleted_at IS NULL
);

-- name: GetChatSession :one
SELECT session.*
FROM public.chat_sessions AS session
WHERE session.id = sqlc.arg(session_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL;

-- name: ListChatSessions :many
SELECT session.*
FROM public.chat_sessions AS session
WHERE session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
ORDER BY session.updated_at DESC
LIMIT 25;

-- name: GetChatMessages :one
SELECT COALESCE(messages.messages, CAST('[]' AS jsonb)) AS messages
FROM public.chat_sessions AS session
LEFT JOIN public.chat_messages AS messages
    ON messages.session_id = session.id
WHERE session.id = sqlc.arg(session_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL;

-- name: GetLatestAssistantMessage :one
SELECT CAST((
    SELECT messages.messages -> message_index
    FROM generate_series(
        jsonb_array_length(COALESCE(messages.messages, CAST('[]' AS jsonb))) - 1,
        0,
        -1
    ) AS indexes(message_index)
    WHERE (messages.messages -> message_index) ->> 'role' = 'assistant'
    LIMIT 1
) AS jsonb) AS message
FROM public.chat_sessions AS session
LEFT JOIN public.chat_messages AS messages
    ON messages.session_id = session.id
WHERE session.id = sqlc.arg(session_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL;

-- name: CountUserMessages :one
SELECT COUNT(*)
FROM public.chat_sessions AS session
INNER JOIN public.chat_messages AS messages
    ON messages.session_id = session.id
CROSS JOIN LATERAL jsonb_array_elements(messages.messages) AS message
WHERE session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.created_at >= sqlc.arg(start_date)
  AND session.created_at < sqlc.arg(end_date)
  AND message ->> 'role' = 'user';
