-- name: LockChatWorkspace :one
SELECT w.workspace_id FROM public.workspaces w
LEFT JOIN public.workspace_settings s ON s.workspace_id=w.workspace_id
WHERE w.workspace_id=sqlc.arg(workspace_id) AND w.deleted_at IS NULL
AND COALESCE(s.chat_enabled,TRUE)
FOR SHARE OF w;

-- name: AuthorizeConversation :one
SELECT c.id, c.kind, c.team_id
FROM public.chat_conversations c
WHERE c.workspace_id = sqlc.arg(workspace_id) AND c.id = sqlc.arg(conversation_id)
  AND public.can_read_chat_conversation(c.workspace_id, sqlc.arg(user_id), c.id);

-- name: LockConversation :one
SELECT c.id, c.kind, c.team_id
FROM public.chat_conversations c
WHERE c.workspace_id = sqlc.arg(workspace_id) AND c.id = sqlc.arg(conversation_id)
  AND public.can_read_chat_conversation(c.workspace_id, sqlc.arg(user_id), c.id)
FOR UPDATE;

-- name: GetConversation :one
SELECT CAST(public.chat_conversation_json(c.id, sqlc.arg(user_id), sqlc.arg(maya_id)) AS jsonb) AS value
FROM public.chat_conversations c WHERE c.workspace_id = sqlc.arg(workspace_id) AND c.id = sqlc.arg(conversation_id);

-- name: ListConversations :many
SELECT c.id, COALESCE(c.last_message_at, CAST('epoch' AS timestamptz)) AS activity_at,
    CAST(public.chat_conversation_json(c.id, sqlc.arg(user_id), sqlc.arg(maya_id)) AS jsonb) AS value
FROM public.chat_conversations c
WHERE c.workspace_id = sqlc.arg(workspace_id)
  AND public.can_read_chat_conversation(c.workspace_id, sqlc.arg(user_id), c.id)
  AND (CAST(sqlc.arg(unrestricted) AS boolean) OR c.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (CAST(sqlc.arg(search) AS text) = '' OR EXISTS (SELECT 1 FROM public.teams t WHERE t.team_id=c.team_id AND t.name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%')
    OR EXISTS (SELECT 1 FROM public.users peer WHERE c.kind='dm' AND peer.user_id=CASE WHEN c.dm_user_low=sqlc.arg(user_id) THEN c.dm_user_high ELSE c.dm_user_low END AND peer.full_name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'))
  AND (CAST(sqlc.narg(cursor_time) AS timestamptz) IS NULL OR (COALESCE(c.last_message_at, CAST('epoch' AS timestamptz)), c.id) < (CAST(sqlc.narg(cursor_time) AS timestamptz), CAST(sqlc.narg(cursor_id) AS uuid)))
ORDER BY COALESCE(c.last_message_at, CAST('epoch' AS timestamptz)) DESC, c.id DESC LIMIT sqlc.arg(page_limit);

-- name: CreateDirectConversation :one
INSERT INTO public.chat_conversations(workspace_id, kind, dm_user_low, dm_user_high)
SELECT CAST(sqlc.arg(workspace_id) AS uuid), 'dm', LEAST(CAST(sqlc.arg(user_id) AS uuid), CAST(sqlc.arg(peer_id) AS uuid)), GREATEST(CAST(sqlc.arg(user_id) AS uuid), CAST(sqlc.arg(peer_id) AS uuid))
WHERE sqlc.arg(user_id) <> sqlc.arg(peer_id) AND EXISTS (
    SELECT 1 FROM public.workspaces w JOIN public.workspace_members wm ON wm.workspace_id = w.workspace_id
    JOIN public.users u ON u.user_id = wm.user_id AND u.is_active AND NOT u.is_system
    WHERE w.workspace_id = sqlc.arg(workspace_id) AND w.deleted_at IS NULL AND u.user_id IN (sqlc.arg(user_id), sqlc.arg(peer_id))
    GROUP BY w.workspace_id HAVING count(*) = 2
)
ON CONFLICT (workspace_id, dm_user_low, dm_user_high) WHERE kind = 'dm' DO UPDATE SET kind = EXCLUDED.kind
RETURNING id;

-- name: GetMessageIdentity :one
SELECT id, conversation_id, thread_root_id, author_id, creation_id, creation_hash, sequence, version, deleted_at
FROM public.chat_conversation_messages WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(message_id);

-- name: FindMessageCreation :one
SELECT id, creation_hash FROM public.chat_conversation_messages
WHERE workspace_id = sqlc.arg(workspace_id) AND conversation_id = sqlc.arg(conversation_id) AND author_id = sqlc.arg(author_id) AND creation_id = sqlc.arg(creation_id);

-- name: GetMessage :one
SELECT CAST(public.chat_message_json(m.id, sqlc.arg(user_id)) AS jsonb) AS value
FROM public.chat_conversation_messages m WHERE m.workspace_id = sqlc.arg(workspace_id) AND m.id = sqlc.arg(message_id);

-- name: SearchMessages :many
SELECT m.sequence, CAST(public.chat_message_json(m.id, sqlc.arg(user_id)) AS jsonb) AS value
FROM public.chat_conversation_messages m
WHERE m.workspace_id = sqlc.arg(workspace_id) AND m.conversation_id = sqlc.arg(conversation_id) AND m.deleted_at IS NULL
  AND m.sequence < sqlc.arg(before_sequence)
  AND to_tsvector('simple', m.text) @@ websearch_to_tsquery('simple', CAST(sqlc.arg(search) AS text))
ORDER BY m.sequence DESC LIMIT sqlc.arg(page_limit);

-- name: AllocateMessageSequence :one
UPDATE public.chat_conversations SET next_sequence = next_sequence + 1, last_message_at = now()
WHERE id = sqlc.arg(conversation_id) AND workspace_id = sqlc.arg(workspace_id) RETURNING next_sequence;

-- name: InsertMessage :one
INSERT INTO public.chat_conversation_messages(workspace_id,conversation_id,sequence,thread_root_id,author_id,creation_id,creation_hash,text,rich_content,agent_mentions)
VALUES (sqlc.arg(workspace_id),sqlc.arg(conversation_id),sqlc.arg(sequence),sqlc.narg(thread_root_id),sqlc.arg(author_id),sqlc.arg(creation_id),sqlc.arg(creation_hash),sqlc.arg(text),sqlc.narg(rich_content),sqlc.arg(agent_mentions))
RETURNING id;

-- name: UpdateMessage :execrows
UPDATE public.chat_conversation_messages SET text = sqlc.arg(text), rich_content = sqlc.narg(rich_content), agent_mentions = sqlc.arg(agent_mentions), version = version + 1, edited_at = now()
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(message_id) AND author_id = sqlc.arg(user_id) AND version = sqlc.arg(version) AND deleted_at IS NULL;

-- name: DeleteMessage :execrows
UPDATE public.chat_conversation_messages SET text = '', rich_content = NULL, agent_mentions = '{}', deleted_at = now(), version = version + 1
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(message_id) AND author_id = sqlc.arg(user_id) AND version = sqlc.arg(version) AND deleted_at IS NULL;

-- name: ReplaceMessageMentions :exec
DELETE FROM public.chat_message_mentions WHERE message_id = sqlc.arg(message_id);

-- name: InsertMessageMention :execrows
INSERT INTO public.chat_message_mentions(message_id, user_id)
SELECT CAST(sqlc.arg(message_id) AS uuid), CAST(sqlc.arg(user_id) AS uuid)
WHERE public.can_read_chat_conversation(sqlc.arg(workspace_id), sqlc.arg(user_id), sqlc.arg(conversation_id))
ON CONFLICT DO NOTHING;

-- name: DeleteMessageLinks :exec
DELETE FROM public.chat_message_links WHERE message_id = sqlc.arg(message_id);

-- name: InsertMessageLink :exec
INSERT INTO public.chat_message_links(message_id, url, title) VALUES(sqlc.arg(message_id),sqlc.arg(url),sqlc.arg(title)) ON CONFLICT DO NOTHING;

-- name: AddReaction :execrows
INSERT INTO public.chat_message_reactions(message_id,user_id,emoji) VALUES(sqlc.arg(message_id),sqlc.arg(user_id),sqlc.arg(emoji)) ON CONFLICT DO NOTHING;

-- name: RemoveReaction :execrows
DELETE FROM public.chat_message_reactions WHERE message_id=sqlc.arg(message_id) AND user_id=sqlc.arg(user_id) AND emoji=sqlc.arg(emoji);

-- name: DeleteMessageReactions :exec
DELETE FROM public.chat_message_reactions WHERE message_id=sqlc.arg(message_id);

-- name: MarkConversationRead :one
INSERT INTO public.chat_conversation_reads(conversation_id,user_id,last_read_sequence) VALUES(sqlc.arg(conversation_id),sqlc.arg(user_id),sqlc.arg(sequence))
ON CONFLICT (conversation_id,user_id) DO UPDATE SET last_read_sequence = GREATEST(chat_conversation_reads.last_read_sequence, EXCLUDED.last_read_sequence), updated_at=now()
WHERE chat_conversation_reads.last_read_sequence < EXCLUDED.last_read_sequence
RETURNING last_read_sequence;

-- name: MarkThreadRead :one
INSERT INTO public.chat_thread_reads(thread_root_id,user_id,last_read_sequence) VALUES(sqlc.arg(thread_root_id),sqlc.arg(user_id),sqlc.arg(sequence))
ON CONFLICT (thread_root_id,user_id) DO UPDATE SET last_read_sequence = GREATEST(chat_thread_reads.last_read_sequence, EXCLUDED.last_read_sequence), updated_at=now()
WHERE chat_thread_reads.last_read_sequence < EXCLUDED.last_read_sequence
RETURNING last_read_sequence;

-- name: CountUnreadMessages :one
SELECT count(*) FROM public.chat_conversation_messages WHERE conversation_id=sqlc.arg(conversation_id) AND workspace_id=sqlc.arg(workspace_id)
AND thread_root_id IS NOT DISTINCT FROM CAST(sqlc.narg(thread_root_id) AS uuid) AND sequence > sqlc.arg(sequence) AND author_id <> sqlc.arg(user_id) AND deleted_at IS NULL;

-- name: GetConversationRead :one
SELECT last_read_sequence FROM public.chat_conversation_reads WHERE conversation_id=sqlc.arg(conversation_id) AND user_id=sqlc.arg(user_id);

-- name: GetThreadRead :one
SELECT last_read_sequence FROM public.chat_thread_reads WHERE thread_root_id=sqlc.arg(thread_root_id) AND user_id=sqlc.arg(user_id);
