-- name: ChatUnreadSummary :many
SELECT c.id AS conversation_id,
    count(*) AS unread_count,
    count(*) FILTER (WHERE unread.thread_root_id IS NULL) AS unread_root_count,
    count(*) FILTER (WHERE unread.thread_root_id IS NOT NULL) AS unread_reply_count,
    count(DISTINCT unread.thread_root_id) AS unread_threads
FROM public.chat_conversations c
CROSS JOIN LATERAL public.chat_unread_messages(c.id, sqlc.arg(user_id)) AS unread(message_id, thread_root_id, created_at, sequence)
WHERE c.workspace_id = sqlc.arg(workspace_id)
  AND public.can_read_chat_conversation(c.workspace_id, sqlc.arg(user_id), c.id)
  AND (CAST(sqlc.arg(unrestricted) AS boolean) OR c.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
GROUP BY c.id
ORDER BY c.id;

-- name: ListUnreadChatThreads :many
WITH unread_threads AS (
    SELECT unread.thread_root_id, max(unread.created_at) AS activity_at,
        min(unread.sequence) AS first_unread_sequence
    FROM public.chat_conversations c
    CROSS JOIN LATERAL public.chat_unread_messages(c.id, sqlc.arg(user_id)) AS unread(message_id, thread_root_id, created_at, sequence)
    WHERE c.workspace_id = sqlc.arg(workspace_id)
      AND public.can_read_chat_conversation(c.workspace_id, sqlc.arg(user_id), c.id)
      AND (CAST(sqlc.arg(unrestricted) AS boolean) OR c.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (CAST(sqlc.narg(conversation_id) AS uuid) IS NULL OR c.id = CAST(sqlc.narg(conversation_id) AS uuid))
      AND unread.thread_root_id IS NOT NULL
    GROUP BY unread.thread_root_id
)
SELECT root.id, CAST(unread.activity_at AS timestamptz) AS activity_at,
    CAST(public.chat_message_json(root.id, sqlc.arg(user_id)) || jsonb_build_object(
        'conversationName', CASE WHEN conversation.kind = 'team' THEN team.name ELSE COALESCE(peer.full_name, '') END,
        'conversationKind', conversation.kind,
        'firstUnreadReplyId', first_unread.id
    ) AS jsonb) AS value
FROM unread_threads unread
JOIN public.chat_conversation_messages root ON root.id = unread.thread_root_id
JOIN public.chat_conversation_messages first_unread
    ON first_unread.conversation_id = root.conversation_id AND first_unread.sequence = unread.first_unread_sequence
JOIN public.chat_conversations conversation ON conversation.id = root.conversation_id
LEFT JOIN public.teams team ON team.team_id = conversation.team_id
LEFT JOIN public.users peer ON peer.user_id = CASE WHEN conversation.dm_user_low = sqlc.arg(user_id) THEN conversation.dm_user_high ELSE conversation.dm_user_low END
WHERE CAST(sqlc.narg(cursor_time) AS timestamptz) IS NULL
   OR (unread.activity_at, root.id) < (CAST(sqlc.narg(cursor_time) AS timestamptz), CAST(sqlc.narg(cursor_id) AS uuid))
ORDER BY unread.activity_at DESC, root.id DESC
LIMIT sqlc.arg(page_limit);
