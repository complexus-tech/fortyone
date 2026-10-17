-- name: ListRootMessages :many
WITH earlier AS (
    SELECT m.id, m.sequence FROM public.chat_conversation_messages m
    WHERE m.workspace_id = sqlc.arg(workspace_id) AND m.conversation_id = sqlc.arg(conversation_id)
      AND m.sequence > sqlc.arg(after_sequence) AND m.sequence < sqlc.arg(before_sequence) AND m.thread_root_id IS NULL AND (m.deleted_at IS NULL OR EXISTS (SELECT 1 FROM public.chat_conversation_messages reply WHERE reply.thread_root_id=m.id AND reply.deleted_at IS NULL)) AND NOT CAST(sqlc.arg(ascending) AS boolean)
    ORDER BY m.sequence DESC LIMIT sqlc.arg(page_limit)
), later AS (
    SELECT m.id, m.sequence FROM public.chat_conversation_messages m
    WHERE m.workspace_id = sqlc.arg(workspace_id) AND m.conversation_id = sqlc.arg(conversation_id)
      AND m.sequence > sqlc.arg(after_sequence) AND m.sequence < sqlc.arg(before_sequence) AND m.thread_root_id IS NULL AND (m.deleted_at IS NULL OR EXISTS (SELECT 1 FROM public.chat_conversation_messages reply WHERE reply.thread_root_id=m.id AND reply.deleted_at IS NULL)) AND CAST(sqlc.arg(ascending) AS boolean)
    ORDER BY m.sequence ASC LIMIT sqlc.arg(page_limit)
), message_window AS (
    SELECT id,sequence FROM earlier UNION ALL SELECT id,sequence FROM later
)
SELECT message_window.sequence, CAST(public.chat_message_json(message_window.id,sqlc.arg(user_id)) AS jsonb) AS value FROM message_window ORDER BY message_window.sequence;

-- name: RootPageBounds :one
SELECT EXISTS (SELECT 1 FROM public.chat_conversation_messages m WHERE m.workspace_id=sqlc.arg(workspace_id) AND m.conversation_id=sqlc.arg(conversation_id) AND m.thread_root_id IS NULL AND (m.deleted_at IS NULL OR EXISTS (SELECT 1 FROM public.chat_conversation_messages reply WHERE reply.thread_root_id=m.id AND reply.deleted_at IS NULL)) AND m.sequence<sqlc.arg(first_sequence)) AS has_older,
EXISTS (SELECT 1 FROM public.chat_conversation_messages m WHERE m.workspace_id=sqlc.arg(workspace_id) AND m.conversation_id=sqlc.arg(conversation_id) AND m.thread_root_id IS NULL AND (m.deleted_at IS NULL OR EXISTS (SELECT 1 FROM public.chat_conversation_messages reply WHERE reply.thread_root_id=m.id AND reply.deleted_at IS NULL)) AND m.sequence>sqlc.arg(last_sequence)) AS has_newer;

-- name: ListReplyMessages :many
WITH earlier AS (
    SELECT m.id, m.sequence FROM public.chat_conversation_messages m
    WHERE m.workspace_id = sqlc.arg(workspace_id) AND m.conversation_id = sqlc.arg(conversation_id)
      AND m.sequence > sqlc.arg(after_sequence) AND m.sequence < sqlc.arg(before_sequence) AND m.thread_root_id = sqlc.arg(thread_root_id) AND m.deleted_at IS NULL AND NOT CAST(sqlc.arg(ascending) AS boolean)
    ORDER BY m.sequence DESC LIMIT sqlc.arg(page_limit)
), later AS (
    SELECT m.id, m.sequence FROM public.chat_conversation_messages m
    WHERE m.workspace_id = sqlc.arg(workspace_id) AND m.conversation_id = sqlc.arg(conversation_id)
      AND m.sequence > sqlc.arg(after_sequence) AND m.sequence < sqlc.arg(before_sequence) AND m.thread_root_id = sqlc.arg(thread_root_id) AND m.deleted_at IS NULL AND CAST(sqlc.arg(ascending) AS boolean)
    ORDER BY m.sequence ASC LIMIT sqlc.arg(page_limit)
), message_window AS (
    SELECT id,sequence FROM earlier UNION ALL SELECT id,sequence FROM later
)
SELECT message_window.sequence, CAST(public.chat_message_json(message_window.id,sqlc.arg(user_id)) AS jsonb) AS value FROM message_window ORDER BY message_window.sequence;

-- name: ReplyPageBounds :one
SELECT EXISTS (SELECT 1 FROM public.chat_conversation_messages m WHERE m.workspace_id=sqlc.arg(workspace_id) AND m.conversation_id=sqlc.arg(conversation_id) AND m.thread_root_id = sqlc.arg(thread_root_id) AND m.deleted_at IS NULL AND m.sequence<sqlc.arg(first_sequence)) AS has_older,
EXISTS (SELECT 1 FROM public.chat_conversation_messages m WHERE m.workspace_id=sqlc.arg(workspace_id) AND m.conversation_id=sqlc.arg(conversation_id) AND m.thread_root_id = sqlc.arg(thread_root_id) AND m.deleted_at IS NULL AND m.sequence>sqlc.arg(last_sequence)) AS has_newer;
