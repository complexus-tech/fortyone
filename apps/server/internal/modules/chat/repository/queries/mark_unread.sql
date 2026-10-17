-- name: MarkConversationUnread :one
UPDATE public.chat_conversation_reads
SET last_read_sequence = sqlc.arg(sequence), updated_at = now()
WHERE conversation_id = sqlc.arg(conversation_id) AND user_id = sqlc.arg(user_id)
  AND last_read_sequence > sqlc.arg(sequence)
RETURNING last_read_sequence;

-- name: FirstUnreadThreadReply :one
SELECT id FROM public.chat_conversation_messages
WHERE workspace_id = sqlc.arg(workspace_id) AND conversation_id = sqlc.arg(conversation_id)
  AND thread_root_id = sqlc.narg(thread_root_id) AND sequence > sqlc.arg(sequence)
  AND author_id <> sqlc.arg(user_id) AND deleted_at IS NULL
ORDER BY sequence LIMIT 1;

-- name: MarkThreadUnread :one
UPDATE public.chat_thread_reads
SET last_read_sequence = sqlc.arg(sequence), updated_at = now()
WHERE thread_root_id = sqlc.arg(thread_root_id) AND user_id = sqlc.arg(user_id)
  AND last_read_sequence > sqlc.arg(sequence)
RETURNING last_read_sequence;
