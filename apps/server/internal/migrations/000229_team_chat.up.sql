CREATE TABLE public.chat_conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('team', 'dm')),
    team_id uuid REFERENCES public.teams(team_id) ON DELETE CASCADE,
    dm_user_low uuid REFERENCES public.users(user_id) ON DELETE CASCADE,
    dm_user_high uuid REFERENCES public.users(user_id) ON DELETE CASCADE,
    next_sequence bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_message_at timestamptz,
    UNIQUE (id, workspace_id),
    CHECK ((kind = 'team' AND team_id IS NOT NULL AND dm_user_low IS NULL AND dm_user_high IS NULL)
        OR (kind = 'dm' AND team_id IS NULL AND dm_user_low IS NOT NULL AND dm_user_high IS NOT NULL AND dm_user_low < dm_user_high))
);
CREATE UNIQUE INDEX chat_team_room ON public.chat_conversations(workspace_id, team_id) WHERE kind = 'team';
CREATE UNIQUE INDEX chat_dm_pair ON public.chat_conversations(workspace_id, dm_user_low, dm_user_high) WHERE kind = 'dm';
CREATE INDEX chat_conversation_activity ON public.chat_conversations(workspace_id, last_message_at DESC, id DESC);

CREATE FUNCTION public.provision_team_chat() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO public.chat_conversations(workspace_id, kind, team_id) VALUES (NEW.workspace_id, 'team', NEW.team_id)
    ON CONFLICT (workspace_id, team_id) WHERE kind = 'team' DO NOTHING;
    RETURN NEW;
END;
$$;
CREATE TRIGGER teams_provision_chat AFTER INSERT ON public.teams FOR EACH ROW EXECUTE FUNCTION public.provision_team_chat();
INSERT INTO public.chat_conversations(workspace_id, kind, team_id) SELECT workspace_id, 'team', team_id FROM public.teams;

CREATE TABLE public.chat_conversation_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    sequence bigint NOT NULL,
    thread_root_id uuid,
    author_id uuid NOT NULL REFERENCES public.users(user_id),
    creation_id uuid NOT NULL,
    creation_hash bytea NOT NULL,
    text text NOT NULL,
    rich_content jsonb,
    agent_mentions text[] NOT NULL DEFAULT '{}',
    version integer NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    edited_at timestamptz,
    deleted_at timestamptz,
    UNIQUE (id, conversation_id, workspace_id),
    UNIQUE (conversation_id, sequence),
    UNIQUE (conversation_id, author_id, creation_id),
    FOREIGN KEY (conversation_id, workspace_id) REFERENCES public.chat_conversations(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (thread_root_id, conversation_id, workspace_id) REFERENCES public.chat_conversation_messages(id, conversation_id, workspace_id) ON DELETE CASCADE,
    CHECK (thread_root_id IS DISTINCT FROM id),
    CHECK (char_length(text) <= 4000),
    CHECK (octet_length(creation_hash) = 32),
    CHECK (agent_mentions <@ ARRAY['maya'])
);
CREATE INDEX chat_roots_sequence ON public.chat_conversation_messages(conversation_id, sequence) WHERE thread_root_id IS NULL;
CREATE INDEX chat_replies_sequence ON public.chat_conversation_messages(thread_root_id, sequence) WHERE deleted_at IS NULL;
CREATE INDEX chat_search ON public.chat_conversation_messages USING gin(to_tsvector('simple', text)) WHERE deleted_at IS NULL;

CREATE TABLE public.chat_message_mentions (
    message_id uuid NOT NULL REFERENCES public.chat_conversation_messages(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (message_id, user_id)
);
CREATE TABLE public.chat_message_reactions (
    message_id uuid NOT NULL REFERENCES public.chat_conversation_messages(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    emoji text NOT NULL CHECK (char_length(emoji) BETWEEN 1 AND 16),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (message_id, user_id, emoji)
);
CREATE TABLE public.chat_conversation_reads (
    conversation_id uuid NOT NULL REFERENCES public.chat_conversations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    last_read_sequence bigint NOT NULL DEFAULT 0 CHECK (last_read_sequence >= 0),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (conversation_id, user_id)
);
CREATE TABLE public.chat_thread_reads (
    thread_root_id uuid NOT NULL REFERENCES public.chat_conversation_messages(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    last_read_sequence bigint NOT NULL DEFAULT 0 CHECK (last_read_sequence >= 0),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (thread_root_id, user_id)
);
CREATE TABLE public.chat_attachments (
    attachment_id uuid PRIMARY KEY REFERENCES public.attachments(attachment_id) ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
    workspace_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    uploaded_by uuid NOT NULL REFERENCES public.users(user_id),
    message_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (conversation_id, workspace_id) REFERENCES public.chat_conversations(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (message_id, conversation_id, workspace_id) REFERENCES public.chat_conversation_messages(id, conversation_id, workspace_id) ON DELETE CASCADE
);
CREATE INDEX chat_attachment_message ON public.chat_attachments(message_id) WHERE message_id IS NOT NULL;
CREATE INDEX chat_attachment_pending ON public.chat_attachments(created_at) WHERE message_id IS NULL;
CREATE TABLE public.chat_message_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id uuid NOT NULL REFERENCES public.chat_conversation_messages(id) ON DELETE CASCADE,
    url text NOT NULL,
    title text NOT NULL,
    UNIQUE (message_id, url)
);
CREATE TABLE public.chat_outbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type text NOT NULL,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    conversation_id uuid NOT NULL REFERENCES public.chat_conversations(id) ON DELETE CASCADE,
    message_id uuid NOT NULL REFERENCES public.chat_conversation_messages(id) ON DELETE CASCADE,
    actor_id uuid NOT NULL REFERENCES public.users(user_id),
    thread_root_id uuid,
    sequence bigint NOT NULL,
    version integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    available_at timestamptz NOT NULL DEFAULT now(),
    lease_token uuid,
    leased_until timestamptz,
    attempts integer NOT NULL DEFAULT 0,
    delivered_at timestamptz
);
CREATE INDEX chat_outbox_pending ON public.chat_outbox(available_at, created_at) WHERE delivered_at IS NULL;
CREATE TABLE public.chat_outbox_recipients (
    event_id uuid NOT NULL REFERENCES public.chat_outbox(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    reason text NOT NULL CHECK (reason IN ('dm','mention','reply')),
    PRIMARY KEY(event_id,user_id)
);

CREATE FUNCTION public.can_read_chat_conversation(p_workspace_id uuid, p_user_id uuid, p_conversation_id uuid)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS (
        SELECT 1 FROM public.chat_conversations c
        JOIN public.workspaces w ON w.workspace_id = c.workspace_id AND w.deleted_at IS NULL
        JOIN public.workspace_members wm ON wm.workspace_id = w.workspace_id AND wm.user_id = p_user_id
        JOIN public.users u ON u.user_id = wm.user_id AND u.is_active AND NOT u.is_system
        WHERE c.id = p_conversation_id AND c.workspace_id = p_workspace_id
          AND ((c.kind = 'team' AND EXISTS (SELECT 1 FROM public.team_members tm WHERE tm.team_id = c.team_id AND tm.user_id = p_user_id))
            OR (c.kind = 'dm' AND p_user_id IN (c.dm_user_low, c.dm_user_high)))
    );
$$;
CREATE FUNCTION public.can_read_chat_message(p_workspace_id uuid, p_user_id uuid, p_message_id uuid)
RETURNS boolean LANGUAGE sql STABLE AS $$
    SELECT EXISTS (SELECT 1 FROM public.chat_conversation_messages m WHERE m.id = p_message_id AND m.workspace_id = p_workspace_id
        AND m.deleted_at IS NULL AND public.can_read_chat_conversation(p_workspace_id, p_user_id, m.conversation_id));
$$;

-- JSON projections remain internal to the SQLC adapter. Public responses are decoded into typed domain DTOs.
CREATE FUNCTION public.chat_message_json(p_message_id uuid, p_viewer_id uuid) RETURNS jsonb LANGUAGE sql STABLE AS $$
SELECT jsonb_build_object(
    'id', m.id, 'conversationId', m.conversation_id, 'threadRootId', m.thread_root_id,
    'creationId', m.creation_id, 'sequence', m.sequence, 'version', m.version,
    'author', jsonb_build_object('id', u.user_id, 'fullName', COALESCE(u.full_name, ''), 'avatarUrl', u.avatar_url, 'isSystem', u.is_system),
    'text', CASE WHEN m.deleted_at IS NULL THEN m.text ELSE '' END, 'richContent', CASE WHEN m.deleted_at IS NULL THEN m.rich_content ELSE NULL END, 'createdAt', m.created_at, 'editedAt', m.edited_at, 'deletedAt', m.deleted_at,
    'mentionedUserIds', COALESCE((SELECT jsonb_agg(mm.user_id ORDER BY mm.user_id) FROM public.chat_message_mentions mm WHERE mm.message_id = m.id AND m.deleted_at IS NULL), '[]'),
    'agentMentions', CASE WHEN m.deleted_at IS NULL THEN to_jsonb(m.agent_mentions) ELSE '[]' END,
    'attachmentIds', COALESCE((SELECT jsonb_agg(a.attachment_id ORDER BY a.attachment_id) FROM public.chat_attachments a WHERE a.message_id = m.id AND m.deleted_at IS NULL), '[]'),
    'attachments', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', a.attachment_id, 'fileName', a.filename, 'mediaType', a.mime_type, 'sizeBytes', a.size) ORDER BY a.attachment_id) FROM public.chat_attachments ca JOIN public.attachments a ON a.attachment_id = ca.attachment_id WHERE ca.message_id = m.id AND m.deleted_at IS NULL), '[]'),
    'reactions', COALESCE((SELECT jsonb_agg(jsonb_build_object('emoji', r.emoji, 'count', r.total, 'reacted', r.reacted) ORDER BY r.emoji)
        FROM (SELECT emoji, count(*) AS total, bool_or(user_id = p_viewer_id) AS reacted FROM public.chat_message_reactions WHERE message_id = m.id AND m.deleted_at IS NULL GROUP BY emoji) r), '[]'),
    'replyCount', (SELECT count(*) FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL),
    'lastReplyAt', (SELECT max(reply.created_at) FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL),
    'replyAuthors', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', a.user_id, 'fullName', COALESCE(a.full_name, ''), 'avatarUrl', a.avatar_url, 'isSystem', a.is_system) ORDER BY latest.last_at DESC, a.user_id)
        FROM (SELECT reply.author_id, max(reply.created_at) AS last_at FROM public.chat_conversation_messages reply WHERE reply.thread_root_id = m.id AND reply.deleted_at IS NULL GROUP BY reply.author_id ORDER BY last_at DESC, reply.author_id LIMIT 3) latest
        JOIN public.users a ON a.user_id = latest.author_id), '[]')
) FROM public.chat_conversation_messages m JOIN public.users u ON u.user_id = m.author_id WHERE m.id = p_message_id;
$$;

CREATE FUNCTION public.chat_conversation_json(p_conversation_id uuid, p_viewer_id uuid, p_maya_id uuid) RETURNS jsonb LANGUAGE sql STABLE AS $$
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
