-- name: ListDocumentCommentMentions :many
SELECT account.user_id, COALESCE(NULLIF(account.full_name,''),account.username) AS name,
 account.username, account.avatar_url
FROM public.workspace_members AS member
JOIN public.users AS account ON account.user_id=member.user_id AND account.is_active=TRUE AND account.is_system=FALSE
WHERE member.workspace_id=sqlc.arg(workspace_id)
 AND public.can_read_document(sqlc.arg(workspace_id),sqlc.arg(actor_id),sqlc.arg(document_id))
 AND public.can_read_document(sqlc.arg(workspace_id),account.user_id,sqlc.arg(document_id))
ORDER BY name,account.user_id;

-- name: GetDocumentCommentCreation :one
SELECT comment.comment_id, comment.thread_id, thread.document_id, comment.creation_hash
FROM public.document_comments AS comment
JOIN public.document_comment_threads AS thread ON thread.thread_id=comment.thread_id
WHERE comment.created_by=sqlc.arg(author_id) AND comment.creation_id=sqlc.arg(creation_id)
 AND thread.workspace_id=sqlc.arg(workspace_id)
 AND public.can_read_document(thread.workspace_id,sqlc.arg(actor_id),thread.document_id);

-- name: RecordDocumentCommentMentions :exec
INSERT INTO public.document_comment_mentions(comment_id,user_id)
SELECT sqlc.arg(comment_id),candidate.user_id
FROM unnest(CAST(sqlc.arg(user_ids) AS uuid[])) AS candidate(user_id)
WHERE public.can_read_document(sqlc.arg(workspace_id),candidate.user_id,sqlc.arg(document_id))
ON CONFLICT DO NOTHING;

-- name: ListDocumentCommentMentionIDs :many
SELECT user_id FROM public.document_comment_mentions WHERE comment_id=sqlc.arg(comment_id) ORDER BY user_id;

-- name: NotifyDocumentComment :exec
WITH source AS (
 SELECT comment.*, thread.document_id,thread.workspace_id,thread.created_by AS thread_author,
 document.title,COALESCE(NULLIF(author.full_name,''),author.username) AS author_name,
 parent.created_by AS parent_author
 FROM public.document_comments AS comment
 JOIN public.document_comment_threads AS thread ON thread.thread_id=comment.thread_id
 JOIN public.documents AS document ON document.document_id=thread.document_id AND document.workspace_id=thread.workspace_id AND document.archived_at IS NULL
 JOIN public.users AS author ON author.user_id=comment.created_by AND author.is_active=TRUE
 LEFT JOIN public.document_comments AS parent ON parent.comment_id=comment.reply_to_comment_id AND parent.thread_id=thread.thread_id
 WHERE comment.comment_id=sqlc.arg(comment_id)
), recipients AS (
 SELECT mention.user_id,TRUE AS mentioned FROM public.document_comment_mentions AS mention JOIN source ON source.comment_id=mention.comment_id
 UNION ALL SELECT source.thread_author,FALSE FROM source WHERE CAST(sqlc.arg(is_reply) AS boolean)
 UNION ALL SELECT source.parent_author,FALSE FROM source WHERE CAST(sqlc.arg(is_reply) AS boolean) AND source.parent_author IS NOT NULL
), eligible AS (
 SELECT recipients.user_id, bool_or(recipients.mentioned) AS mentioned FROM recipients CROSS JOIN source
 WHERE recipients.user_id<>source.created_by AND public.can_read_document(source.workspace_id,recipients.user_id,source.document_id)
 GROUP BY recipients.user_id
), inserted AS (
 INSERT INTO public.notifications(recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message,dedupe_key,in_app_enabled,coalescing_key)
 SELECT eligible.user_id,source.workspace_id,
 CAST(CASE WHEN eligible.mentioned THEN 'mention' ELSE 'comment_reply' END AS public.notification_type),
 CAST('document' AS public.entity_type),source.document_id,source.created_by,
 CASE WHEN eligible.mentioned THEN 'Mentioned you in a document' ELSE 'Replied in a document' END,
 jsonb_build_object('template',CASE WHEN eligible.mentioned THEN '{actor} mentioned you in {document_title}' ELSE '{actor} replied in {document_title}' END,
 'variables',jsonb_build_object('actor',jsonb_build_object('value',source.author_name,'type','text'),
 'document_title',jsonb_build_object('value',source.title,'type','text'),'document_id',jsonb_build_object('value',CAST(source.document_id AS text),'type','text'),
 'thread_id',jsonb_build_object('value',CAST(source.thread_id AS text),'type','text'),'comment_id',jsonb_build_object('value',CAST(source.comment_id AS text),'type','text')),
 'identityReferences',jsonb_build_object('actor',source.created_by)),
 'document-comment:'||CAST(source.comment_id AS text)||':'||CAST(eligible.user_id AS text),
 CASE WHEN jsonb_typeof(preference.preferences->(CASE WHEN eligible.mentioned THEN 'mention' ELSE 'comment_reply' END)->'in_app')='boolean'
 THEN CAST(preference.preferences->(CASE WHEN eligible.mentioned THEN 'mention' ELSE 'comment_reply' END)->>'in_app' AS boolean) ELSE TRUE END,
 'document_comment'
 FROM eligible CROSS JOIN source
 LEFT JOIN public.notification_preferences AS preference ON preference.user_id=eligible.user_id AND preference.workspace_id=source.workspace_id
 WHERE NOT EXISTS(SELECT 1 FROM public.document_comment_notification_receipts AS receipt WHERE receipt.comment_id=source.comment_id AND receipt.recipient_id=eligible.user_id)
 ON CONFLICT(dedupe_key) DO NOTHING RETURNING notification_id,recipient_id
)
INSERT INTO public.document_comment_notification_receipts(comment_id,recipient_id,notification_id)
SELECT sqlc.arg(comment_id),recipient_id,notification_id FROM inserted ON CONFLICT DO NOTHING;

-- name: EnqueueDocumentAgentEdit :one
INSERT INTO public.document_agent_edits(invocation_id,workspace_id,document_id,requester_id,expected_revision,expected_epoch,operations)
SELECT invocation.id,invocation.workspace_id,invocation.entity_id,invocation.requester_id,sqlc.arg(expected_revision),sqlc.arg(expected_epoch),sqlc.arg(operations)
FROM public.agent_invocations AS invocation
WHERE invocation.id=sqlc.arg(invocation_id) AND invocation.workspace_id=sqlc.arg(workspace_id)
 AND invocation.entity_id=sqlc.arg(document_id) AND invocation.requester_id=sqlc.arg(requester_id)
 AND invocation.surface='document_comment' AND invocation.team_id IS NULL AND invocation.status='running'
 AND public.can_read_document(invocation.workspace_id,invocation.requester_id,invocation.entity_id)
ON CONFLICT(invocation_id) DO NOTHING RETURNING *;

-- name: GetDocumentAgentEdit :one
SELECT * FROM public.document_agent_edits WHERE invocation_id=sqlc.arg(invocation_id) AND workspace_id=sqlc.arg(workspace_id);

-- name: DocumentAgentEditsAvailable :one
SELECT EXISTS(SELECT 1 FROM public.document_agent_consumer_health WHERE expires_at>clock_timestamp()) AS available;

-- name: ExpireDocumentAgentEdit :exec
UPDATE public.document_agent_edits SET status='failed',error_code='expired',updated_at=clock_timestamp()
WHERE workspace_id=sqlc.arg(workspace_id) AND invocation_id=sqlc.arg(invocation_id)
 AND status='pending' AND created_at<clock_timestamp()-INTERVAL '15 minutes';
