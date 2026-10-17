-- name: InsertEvent :exec
WITH event AS (
    INSERT INTO public.chat_outbox(event_type,workspace_id,conversation_id,message_id,actor_id,thread_root_id,sequence,version)
    SELECT sqlc.arg(event_type),m.workspace_id,m.conversation_id,m.id,CAST(sqlc.arg(actor_id) AS uuid),m.thread_root_id,m.sequence,m.version
    FROM public.chat_conversation_messages m WHERE m.id=sqlc.arg(message_id) AND m.workspace_id=sqlc.arg(workspace_id)
    RETURNING id,message_id,event_type
)
INSERT INTO public.chat_outbox_recipients(event_id,user_id,reason)
SELECT event.id,wm.user_id,CASE
    WHEN c.kind='dm' THEN 'dm'
    WHEN EXISTS (SELECT 1 FROM public.chat_message_mentions mm WHERE mm.message_id=m.id AND mm.user_id=wm.user_id) THEN 'mention'
    WHEN m.thread_root_id IS NOT NULL AND EXISTS (
        SELECT 1 FROM public.chat_conversation_messages participant
        WHERE (participant.id=m.thread_root_id OR participant.thread_root_id=m.thread_root_id)
          AND participant.author_id=wm.user_id AND participant.deleted_at IS NULL
    ) THEN 'reply'
    ELSE 'message'
END
FROM event JOIN public.chat_conversation_messages m ON m.id=event.message_id JOIN public.chat_conversations c ON c.id=m.conversation_id
JOIN public.workspace_members wm ON wm.workspace_id=m.workspace_id
WHERE event.event_type='message.created' AND wm.user_id<>m.author_id
AND public.can_read_chat_conversation(m.workspace_id,wm.user_id,m.conversation_id);

-- name: InsertReactionEvent :exec
-- The reaction and event share transaction-stable now(). Eligibility compares
-- that timestamp so a later remove/re-add cannot revive an older add event.
WITH event AS (
    INSERT INTO public.chat_outbox(event_type,workspace_id,conversation_id,message_id,actor_id,thread_root_id,sequence,version,reaction_emoji)
    SELECT 'reaction.updated',m.workspace_id,m.conversation_id,m.id,CAST(sqlc.arg(actor_id) AS uuid),m.thread_root_id,m.sequence,m.version,CAST(sqlc.narg(reaction_emoji) AS text)
    FROM public.chat_conversation_messages m WHERE m.id=sqlc.arg(message_id) AND m.workspace_id=sqlc.arg(workspace_id)
    RETURNING id,message_id,actor_id,reaction_emoji
)
INSERT INTO public.chat_outbox_recipients(event_id,user_id,reason)
SELECT event.id,message.author_id,'reaction'
FROM event JOIN public.chat_conversation_messages message ON message.id=event.message_id
WHERE event.reaction_emoji IS NOT NULL AND event.actor_id<>message.author_id
AND public.can_read_chat_message(message.workspace_id,message.author_id,message.id);

-- name: ClaimEvents :many
WITH candidates AS (SELECT id FROM public.chat_outbox WHERE EXISTS (SELECT 1 FROM public.workspaces w LEFT JOIN public.workspace_settings s ON s.workspace_id=w.workspace_id WHERE w.workspace_id=public.chat_outbox.workspace_id AND w.deleted_at IS NULL AND COALESCE(s.chat_enabled,TRUE)) AND delivered_at IS NULL AND available_at<=now() AND (leased_until IS NULL OR leased_until<now()) ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT sqlc.arg(page_limit))
UPDATE public.chat_outbox e SET lease_token=gen_random_uuid(),leased_until=now()+interval '60 seconds',attempts=attempts+1
FROM candidates c WHERE e.id=c.id
RETURNING e.id,e.event_type,e.workspace_id,e.conversation_id,e.message_id,e.actor_id,e.thread_root_id,e.sequence,e.version,e.created_at,e.lease_token;

-- name: CompleteEvent :execrows
UPDATE public.chat_outbox SET delivered_at=now(),leased_until=NULL WHERE id=sqlc.arg(event_id) AND lease_token=sqlc.arg(lease_token) AND delivered_at IS NULL AND leased_until>now()
AND EXISTS (SELECT 1 FROM public.workspaces w LEFT JOIN public.workspace_settings s ON s.workspace_id=w.workspace_id WHERE w.workspace_id=public.chat_outbox.workspace_id AND w.deleted_at IS NULL AND COALESCE(s.chat_enabled,TRUE));

-- name: ReleaseEvent :execrows
UPDATE public.chat_outbox SET lease_token=NULL,leased_until=NULL,available_at=now()+LEAST(attempts,60)*interval '1 second' WHERE id=sqlc.arg(event_id) AND lease_token=sqlc.arg(lease_token) AND delivered_at IS NULL;

-- name: RecipientIDs :many
SELECT wm.user_id FROM public.workspace_members wm WHERE wm.workspace_id=sqlc.arg(workspace_id)
AND public.can_read_chat_conversation(wm.workspace_id,wm.user_id,sqlc.arg(conversation_id)) ORDER BY wm.user_id;

-- name: NotificationRecipients :many
SELECT recipient.user_id, recipient.reason
FROM public.chat_outbox_recipients recipient JOIN public.chat_outbox event ON event.id=recipient.event_id
JOIN public.chat_conversation_messages message ON message.id=event.message_id
LEFT JOIN public.chat_conversation_reads conversation_read ON conversation_read.conversation_id=message.conversation_id AND conversation_read.user_id=recipient.user_id
LEFT JOIN public.chat_thread_reads thread_read ON thread_read.thread_root_id=message.thread_root_id AND thread_read.user_id=recipient.user_id
WHERE event.id=sqlc.arg(event_id) AND event.workspace_id=sqlc.arg(workspace_id)
AND public.can_read_chat_message(event.workspace_id,recipient.user_id,event.message_id)
AND (
    (event.event_type='message.created' AND recipient.user_id<>message.author_id
     AND message.sequence>CASE WHEN message.thread_root_id IS NULL THEN COALESCE(conversation_read.last_read_sequence,0) ELSE COALESCE(thread_read.last_read_sequence,0) END)
    OR
    (event.event_type='reaction.updated' AND event.reaction_emoji IS NOT NULL AND recipient.reason='reaction'
     AND recipient.user_id=message.author_id AND recipient.user_id<>event.actor_id
     AND public.can_read_chat_message(event.workspace_id,event.actor_id,event.message_id)
     AND EXISTS (SELECT 1 FROM public.chat_message_reactions reaction
         WHERE reaction.message_id=event.message_id AND reaction.user_id=event.actor_id AND reaction.emoji=event.reaction_emoji
         AND reaction.created_at=event.created_at))
)
ORDER BY recipient.user_id;

-- name: NotificationContent :one
SELECT jsonb_build_object(
    'conversationKind', conversation.kind,
    'conversationName', COALESCE(team.name, ''),
    'authorName', COALESCE(NULLIF(author.full_name, ''), author.username),
    'actorName', COALESCE(NULLIF(actor.full_name, ''), actor.username),
    'reactionEmoji', COALESCE(event.reaction_emoji, ''),
    'text', message.text,
    'richContent', message.rich_content,
    'attachments', COALESCE((
        SELECT jsonb_agg(jsonb_build_object(
            'fileName', attachment.filename, 'mediaType', attachment.mime_type, 'kind', COALESCE(chat_attachment.kind, '')
        ) ORDER BY attachment.attachment_id)
        FROM public.chat_attachments chat_attachment
        JOIN public.attachments attachment ON attachment.attachment_id=chat_attachment.attachment_id
        WHERE chat_attachment.message_id=message.id
    ), '[]')
)
FROM public.chat_outbox event
JOIN public.chat_conversation_messages message ON message.id=event.message_id AND message.workspace_id=event.workspace_id
JOIN public.chat_conversations conversation ON conversation.id=message.conversation_id
JOIN public.users author ON author.user_id=message.author_id
JOIN public.users actor ON actor.user_id=event.actor_id
LEFT JOIN public.teams team ON team.team_id=conversation.team_id
WHERE event.id=sqlc.arg(event_id) AND event.workspace_id=sqlc.arg(workspace_id)
AND message.deleted_at IS NULL
AND EXISTS (
    SELECT 1 FROM public.chat_outbox_recipients recipient
    WHERE recipient.event_id=event.id AND public.can_read_chat_message(event.workspace_id,recipient.user_id,message.id)
    AND (
        event.event_type='message.created'
        OR
        (event.event_type='reaction.updated' AND event.reaction_emoji IS NOT NULL AND recipient.reason='reaction'
         AND recipient.user_id=message.author_id AND recipient.user_id<>event.actor_id
         AND public.can_read_chat_message(event.workspace_id,event.actor_id,event.message_id)
         AND EXISTS (SELECT 1 FROM public.chat_message_reactions reaction
             WHERE reaction.message_id=event.message_id AND reaction.user_id=event.actor_id AND reaction.emoji=event.reaction_emoji
             AND reaction.created_at=event.created_at))
    )
);

-- name: ReadRecipientIDs :many
SELECT u.user_id FROM public.users u WHERE u.user_id=sqlc.arg(actor_id)
AND public.can_read_chat_conversation(sqlc.arg(workspace_id),u.user_id,sqlc.arg(conversation_id));

-- name: RenewEventLease :execrows
UPDATE public.chat_outbox SET leased_until=now()+interval '60 seconds'
WHERE id=sqlc.arg(event_id) AND lease_token=sqlc.arg(lease_token) AND delivered_at IS NULL AND leased_until>now()
AND EXISTS (SELECT 1 FROM public.workspaces w LEFT JOIN public.workspace_settings s ON s.workspace_id=w.workspace_id WHERE w.workspace_id=public.chat_outbox.workspace_id AND w.deleted_at IS NULL AND COALESCE(s.chat_enabled,TRUE));
