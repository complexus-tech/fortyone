CREATE OR REPLACE FUNCTION public.chat_conversation_json(p_conversation_id uuid, p_viewer_id uuid, p_maya_id uuid) RETURNS jsonb LANGUAGE sql STABLE AS $$
SELECT jsonb_build_object(
    'id', c.id, 'kind', c.kind, 'teamId', c.team_id,
    'name', CASE WHEN c.kind = 'team' THEN t.name ELSE COALESCE(peer.full_name, '') END,
    'description', CASE WHEN c.kind = 'team' THEN 'Your team conversation' ELSE 'Direct conversation' END,
    'color', t.color,
    'members', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', member.user_id, 'fullName', COALESCE(member.full_name, ''), 'avatarUrl', member.avatar_url, 'isSystem', member.is_system) ORDER BY member.is_system DESC, member.full_name, member.user_id)
        FROM (
            SELECT member.user_id, member.full_name, member.avatar_url, member.is_system FROM public.users member WHERE member.is_active AND (
            (c.kind = 'team' AND (member.user_id = p_maya_id OR (NOT member.is_system AND EXISTS (SELECT 1 FROM public.team_members tm JOIN public.workspace_members wm ON wm.user_id = tm.user_id AND wm.workspace_id = c.workspace_id WHERE tm.team_id = c.team_id AND tm.user_id = member.user_id))))
            OR (c.kind = 'dm' AND member.user_id IN (c.dm_user_low, c.dm_user_high))
        )
            ORDER BY member.is_system DESC, member.full_name, member.user_id LIMIT 50
        ) member), '[]'),
    'memberCount', (SELECT count(*) FROM public.users member WHERE member.is_active AND (
        (c.kind = 'team' AND (member.user_id = p_maya_id OR (NOT member.is_system AND EXISTS (SELECT 1 FROM public.team_members tm JOIN public.workspace_members wm ON wm.user_id = tm.user_id AND wm.workspace_id = c.workspace_id WHERE tm.team_id = c.team_id AND tm.user_id = member.user_id))))
        OR (c.kind = 'dm' AND member.user_id IN (c.dm_user_low, c.dm_user_high)))),
    'lastMessageAt', c.last_message_at,
    'lastMessageText', COALESCE((SELECT left(m.text, 160) FROM public.chat_conversation_messages m WHERE m.conversation_id = c.id AND m.deleted_at IS NULL ORDER BY m.sequence DESC LIMIT 1), ''),
    'unreadCount', (SELECT count(*) FROM public.chat_conversation_messages m WHERE m.conversation_id = c.id AND m.thread_root_id IS NULL AND m.deleted_at IS NULL AND m.author_id <> p_viewer_id AND m.sequence > COALESCE(read.last_read_sequence, 0)),
    'lastReadSequence', COALESCE(read.last_read_sequence, 0)
) FROM public.chat_conversations c
LEFT JOIN public.teams t ON t.team_id = c.team_id
LEFT JOIN public.users peer ON peer.user_id = CASE WHEN c.dm_user_low = p_viewer_id THEN c.dm_user_high ELSE c.dm_user_low END
LEFT JOIN public.chat_conversation_reads read ON read.conversation_id = c.id AND read.user_id = p_viewer_id
WHERE c.id = p_conversation_id;
$$;

CREATE OR REPLACE FUNCTION public.chat_message_json(p_message_id uuid, p_viewer_id uuid) RETURNS jsonb LANGUAGE sql STABLE AS $$
SELECT jsonb_build_object(
    'id', m.id, 'conversationId', m.conversation_id, 'threadRootId', m.thread_root_id,
    'creationId', m.creation_id, 'sequence', m.sequence, 'version', m.version,
    'author', jsonb_build_object('id', u.user_id, 'fullName', COALESCE(u.full_name, ''), 'avatarUrl', u.avatar_url, 'isSystem', u.is_system),
    'text', CASE WHEN m.deleted_at IS NULL THEN m.text ELSE '' END, 'richContent', CASE WHEN m.deleted_at IS NULL THEN m.rich_content ELSE NULL END, 'createdAt', m.created_at, 'editedAt', m.edited_at, 'deletedAt', m.deleted_at,
    'mentionedUserIds', COALESCE((SELECT jsonb_agg(mm.user_id ORDER BY mm.user_id) FROM public.chat_message_mentions mm WHERE mm.message_id = m.id AND m.deleted_at IS NULL), '[]'),
    'agentMentions', CASE WHEN m.deleted_at IS NULL THEN to_jsonb(m.agent_mentions) ELSE '[]' END,
    'attachmentIds', COALESCE((SELECT jsonb_agg(a.attachment_id ORDER BY a.attachment_id) FROM public.chat_attachments a WHERE a.message_id = m.id AND m.deleted_at IS NULL), '[]'),
    'attachments', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', a.attachment_id, 'fileName', a.filename, 'mediaType', a.mime_type, 'sizeBytes', a.size, 'kind', ca.kind, 'durationMs', ca.duration_ms) ORDER BY a.attachment_id) FROM public.chat_attachments ca JOIN public.attachments a ON a.attachment_id = ca.attachment_id WHERE ca.message_id = m.id AND m.deleted_at IS NULL), '[]'),
    'reactions', COALESCE((SELECT jsonb_agg(jsonb_build_object('emoji', r.emoji, 'count', r.total, 'reacted', r.reacted) ORDER BY r.emoji)
        FROM (SELECT emoji, count(*) AS total, bool_or(user_id = p_viewer_id) AS reacted FROM public.chat_message_reactions WHERE message_id = m.id AND m.deleted_at IS NULL GROUP BY emoji) r), '[]'),
    'replyCount', (SELECT count(*) FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL),
    'lastReplyAt', (SELECT max(reply.created_at) FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL),
    'replyAuthors', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', a.user_id, 'fullName', COALESCE(a.full_name, ''), 'avatarUrl', a.avatar_url, 'isSystem', a.is_system) ORDER BY latest.last_at DESC, a.user_id)
        FROM (SELECT reply.author_id, max(reply.created_at) AS last_at FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL GROUP BY reply.author_id ORDER BY last_at DESC, reply.author_id LIMIT 3) latest
        JOIN public.users a ON a.user_id = latest.author_id), '[]')
) FROM public.chat_conversation_messages m JOIN public.users u ON u.user_id = m.author_id WHERE m.id = p_message_id;
$$;

DROP FUNCTION public.chat_unread_messages(uuid, uuid);
UPDATE public.chat_outbox_recipients SET reason = 'reply' WHERE reason = 'message';
ALTER TABLE public.chat_outbox_recipients DROP CONSTRAINT chat_outbox_recipients_reason_check;
ALTER TABLE public.chat_outbox_recipients ADD CONSTRAINT chat_outbox_recipients_reason_check
    CHECK (reason IN ('dm', 'mention', 'reply'));
