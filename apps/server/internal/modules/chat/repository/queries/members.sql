-- name: ListConversationMembers :many
SELECT CAST(jsonb_build_object('id',u.user_id,'fullName',COALESCE(u.full_name,''),'avatarUrl',u.avatar_url,'isSystem',u.is_system) AS jsonb) AS value
FROM public.chat_conversations c JOIN public.users u ON u.is_active
WHERE c.id=sqlc.arg(conversation_id) AND c.workspace_id=sqlc.arg(workspace_id)
  AND public.can_read_chat_conversation(c.workspace_id,sqlc.arg(user_id),c.id)
  AND (CAST(sqlc.arg(search) AS text)='' OR COALESCE(u.full_name,'') ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%')
  AND ((c.kind='team' AND (u.user_id=sqlc.arg(maya_id) OR (NOT u.is_system AND EXISTS (
    SELECT 1 FROM public.team_members tm JOIN public.workspace_members wm ON wm.user_id=tm.user_id AND wm.workspace_id=c.workspace_id WHERE tm.team_id=c.team_id AND tm.user_id=u.user_id))))
    OR (c.kind='dm' AND u.user_id IN (c.dm_user_low,c.dm_user_high) AND EXISTS (SELECT 1 FROM public.workspace_members wm WHERE wm.workspace_id=c.workspace_id AND wm.user_id=u.user_id)))
ORDER BY u.is_system DESC,u.full_name,u.user_id LIMIT sqlc.arg(page_limit);
