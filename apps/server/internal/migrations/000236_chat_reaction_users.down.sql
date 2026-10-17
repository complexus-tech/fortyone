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
    'unreadReplyCount', (SELECT count(*) FROM public.chat_conversation_messages reply
        LEFT JOIN public.chat_thread_reads thread_read ON thread_read.thread_root_id = m.id AND thread_read.user_id = p_viewer_id
        WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL AND reply.author_id <> p_viewer_id
          AND reply.sequence > COALESCE(thread_read.last_read_sequence, 0)),
    'lastReplyAt', (SELECT max(reply.created_at) FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL),
    'replyAuthors', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', a.user_id, 'fullName', COALESCE(a.full_name, ''), 'avatarUrl', a.avatar_url, 'isSystem', a.is_system) ORDER BY latest.last_at DESC, a.user_id)
        FROM (SELECT reply.author_id, max(reply.created_at) AS last_at FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL GROUP BY reply.author_id ORDER BY last_at DESC, reply.author_id LIMIT 3) latest
        JOIN public.users a ON a.user_id = latest.author_id), '[]')
) FROM public.chat_conversation_messages m JOIN public.users u ON u.user_id = m.author_id WHERE m.id = p_message_id;
$$;
