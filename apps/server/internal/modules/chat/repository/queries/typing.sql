-- name: AuthorizeTyping :one
SELECT CAST(jsonb_build_object('id',u.user_id,'fullName',COALESCE(u.full_name,''),'avatarUrl',u.avatar_url,'isSystem',false) AS jsonb) AS value
FROM public.users u
WHERE u.user_id=sqlc.arg(user_id) AND NOT u.is_system
AND public.can_read_chat_conversation(sqlc.arg(workspace_id),u.user_id,sqlc.arg(conversation_id))
AND (CAST(sqlc.narg(thread_root_id) AS uuid) IS NULL OR EXISTS (
  SELECT 1 FROM public.chat_conversation_messages root
  WHERE root.id=sqlc.narg(thread_root_id) AND root.workspace_id=sqlc.arg(workspace_id)
    AND root.conversation_id=sqlc.arg(conversation_id) AND root.thread_root_id IS NULL
    AND (root.deleted_at IS NULL OR EXISTS (SELECT 1 FROM public.chat_conversation_messages reply WHERE reply.thread_root_id=root.id AND reply.deleted_at IS NULL))
));

-- name: TypingRecipientIDs :many
SELECT wm.user_id FROM public.workspace_members wm
WHERE wm.workspace_id=sqlc.arg(workspace_id) AND wm.user_id<>sqlc.arg(user_id)
AND public.can_read_chat_conversation(wm.workspace_id,sqlc.arg(user_id),sqlc.arg(conversation_id))
AND public.can_read_chat_conversation(wm.workspace_id,wm.user_id,sqlc.arg(conversation_id))
AND (CAST(sqlc.narg(thread_root_id) AS uuid) IS NULL OR EXISTS (
  SELECT 1 FROM public.chat_conversation_messages root
  WHERE root.id=sqlc.narg(thread_root_id) AND root.workspace_id=wm.workspace_id
    AND root.conversation_id=sqlc.arg(conversation_id) AND root.thread_root_id IS NULL
    AND (root.deleted_at IS NULL OR EXISTS (SELECT 1 FROM public.chat_conversation_messages reply WHERE reply.thread_root_id=root.id AND reply.deleted_at IS NULL))
)) ORDER BY wm.user_id;
