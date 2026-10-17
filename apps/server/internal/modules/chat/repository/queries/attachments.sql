-- name: LinkPendingAttachment :execrows
INSERT INTO public.chat_attachments(attachment_id,workspace_id,conversation_id,uploaded_by,kind,duration_ms)
SELECT a.attachment_id,a.workspace_id,CAST(sqlc.arg(conversation_id) AS uuid),a.uploaded_by,NULLIF(CAST(sqlc.arg(kind) AS text),''),CAST(sqlc.narg(duration_ms) AS integer) FROM public.attachments a
WHERE a.attachment_id=sqlc.arg(attachment_id) AND a.workspace_id=sqlc.arg(workspace_id) AND a.uploaded_by=sqlc.arg(user_id) AND a.scan_status IN ('unscanned','clean')
ON CONFLICT DO NOTHING;

-- name: AuthorizeAttachment :one
SELECT EXISTS (SELECT 1 FROM public.chat_attachments ca JOIN public.attachments a ON a.attachment_id=ca.attachment_id AND a.workspace_id=ca.workspace_id AND a.scan_status IN ('unscanned','clean') WHERE ca.attachment_id=sqlc.arg(attachment_id) AND ca.workspace_id=sqlc.arg(workspace_id) AND ca.conversation_id=sqlc.arg(conversation_id)
AND ((ca.message_id IS NULL AND ca.uploaded_by=sqlc.arg(user_id)) OR EXISTS (SELECT 1 FROM public.chat_conversation_messages m WHERE m.id=ca.message_id AND m.deleted_at IS NULL))) AS allowed;

-- name: UnlinkPendingAttachment :many
DELETE FROM public.chat_attachments WHERE attachment_id=sqlc.arg(attachment_id) AND workspace_id=sqlc.arg(workspace_id) AND conversation_id=sqlc.arg(conversation_id) AND uploaded_by=sqlc.arg(user_id) AND message_id IS NULL
RETURNING attachment_id;

-- name: ClaimAttachment :execrows
UPDATE public.chat_attachments ca SET message_id=sqlc.arg(message_id)
WHERE ca.attachment_id=sqlc.arg(attachment_id) AND ca.workspace_id=sqlc.arg(workspace_id) AND ca.conversation_id=sqlc.arg(conversation_id)
AND ca.uploaded_by=sqlc.arg(user_id) AND (ca.message_id IS NULL OR ca.message_id=sqlc.arg(message_id))
AND EXISTS (SELECT 1 FROM public.attachments a WHERE a.attachment_id=ca.attachment_id AND a.workspace_id=ca.workspace_id AND a.scan_status IN ('unscanned','clean'));

-- name: RemoveMessageAttachments :many
DELETE FROM public.chat_attachments WHERE message_id=sqlc.arg(message_id) AND NOT (attachment_id=ANY(CAST(sqlc.arg(keep_ids) AS uuid[])))
RETURNING attachment_id;

-- name: ListResources :many
WITH resources AS (
    SELECT ca.attachment_id AS id, m.created_at, 'file' AS kind, a.mime_type AS media_type, a.filename || ' ' || a.mime_type AS search_text,
    jsonb_build_object('id',ca.attachment_id,'conversationId',m.conversation_id,'messageId',m.id,'kind','file','title',a.filename,'description','','createdAt',m.created_at,
      'author',jsonb_build_object('id',u.user_id,'fullName',COALESCE(u.full_name,''),'avatarUrl',u.avatar_url,'isSystem',u.is_system),
      'attachmentId',a.attachment_id,'fileName',a.filename,'mediaType',a.mime_type,'sizeBytes',a.size,'attachmentKind',ca.kind,'durationMs',ca.duration_ms) AS value
    FROM public.chat_attachments ca JOIN public.attachments a ON a.attachment_id=ca.attachment_id JOIN public.chat_conversation_messages m ON m.id=ca.message_id JOIN public.users u ON u.user_id=m.author_id
    WHERE m.conversation_id=sqlc.arg(conversation_id) AND m.workspace_id=sqlc.arg(workspace_id) AND m.deleted_at IS NULL
    UNION ALL
    SELECT l.id, m.created_at, 'link', CAST(NULL AS text), l.title || ' ' || l.url,
    jsonb_build_object('id',l.id,'conversationId',m.conversation_id,'messageId',m.id,'kind','link','title',l.title,'description','','createdAt',m.created_at,
      'author',jsonb_build_object('id',u.user_id,'fullName',COALESCE(u.full_name,''),'avatarUrl',u.avatar_url,'isSystem',u.is_system),'url',l.url)
    FROM public.chat_message_links l JOIN public.chat_conversation_messages m ON m.id=l.message_id JOIN public.users u ON u.user_id=m.author_id
    WHERE m.conversation_id=sqlc.arg(conversation_id) AND m.workspace_id=sqlc.arg(workspace_id) AND m.deleted_at IS NULL
)
SELECT r.id, r.created_at, CAST(r.value AS jsonb) AS value FROM resources r
WHERE (CAST(sqlc.arg(kind) AS text)='all' OR r.kind=sqlc.arg(kind)
    OR (sqlc.arg(kind)='media' AND (LOWER(r.media_type) LIKE 'image/%' OR LOWER(r.media_type) LIKE 'video/%')))
AND r.search_text ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
AND (CAST(sqlc.narg(cursor_time) AS timestamptz) IS NULL
    OR (CAST(sqlc.arg(oldest_first) AS boolean) AND (r.created_at, r.id) > (CAST(sqlc.narg(cursor_time) AS timestamptz), CAST(sqlc.narg(cursor_id) AS uuid)))
    OR (NOT CAST(sqlc.arg(oldest_first) AS boolean) AND (r.created_at, r.id) < (CAST(sqlc.narg(cursor_time) AS timestamptz), CAST(sqlc.narg(cursor_id) AS uuid))))
ORDER BY
    CASE WHEN CAST(sqlc.arg(oldest_first) AS boolean) THEN r.created_at END ASC,
    CASE WHEN CAST(sqlc.arg(oldest_first) AS boolean) THEN r.id END ASC,
    CASE WHEN NOT CAST(sqlc.arg(oldest_first) AS boolean) THEN r.created_at END DESC,
    CASE WHEN NOT CAST(sqlc.arg(oldest_first) AS boolean) THEN r.id END DESC
LIMIT sqlc.arg(page_limit);
