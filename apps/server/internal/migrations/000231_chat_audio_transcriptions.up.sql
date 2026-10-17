ALTER TABLE public.chat_attachments
    ADD COLUMN kind text CHECK (kind IS NULL OR kind = 'voice-note'),
    ADD COLUMN duration_ms integer CHECK (duration_ms IS NULL OR duration_ms BETWEEN 1 AND 300000),
    ADD CONSTRAINT chat_voice_note_metadata CHECK ((kind IS NULL) = (duration_ms IS NULL)),
    ADD CONSTRAINT chat_attachment_scope UNIQUE (attachment_id, conversation_id, workspace_id);

CREATE TABLE public.chat_audio_transcriptions (
    attachment_id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    requested_by uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','ready','failed')),
    transcript text CHECK (transcript IS NULL OR octet_length(transcript) <= 1048576),
    error_code text,
    model text,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    lease_token uuid,
    leased_until timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (attachment_id, conversation_id, workspace_id) REFERENCES public.chat_attachments(attachment_id, conversation_id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (conversation_id, workspace_id) REFERENCES public.chat_conversations(id, workspace_id) ON DELETE CASCADE,
    CHECK ((status = 'ready') = (transcript IS NOT NULL)),
    CHECK ((status = 'processing') = (lease_token IS NOT NULL AND leased_until IS NOT NULL))
);
CREATE INDEX chat_audio_transcription_pending ON public.chat_audio_transcriptions(created_at) WHERE status = 'pending';
CREATE INDEX chat_audio_transcription_leases ON public.chat_audio_transcriptions(leased_until) WHERE status = 'processing';

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

