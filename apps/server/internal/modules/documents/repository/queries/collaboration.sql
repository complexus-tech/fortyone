-- Access is checked with GetAccessibleDocument in the same serializable transaction.
-- name: GetDocumentEditingMetadata :one
SELECT revision, collaboration_epoch, CAST(collaboration_state IS NOT NULL AS boolean) AS collaborative,
       CAST(COALESCE(CASE WHEN created_by = sqlc.arg(actor_id) THEN public_token ELSE NULL END, '') AS text) AS public_token
FROM public.documents WHERE document_id = sqlc.arg(document_id);

-- name: ListDocumentRevisions :many
SELECT revision, title, edited_by, created_at
FROM public.document_revisions
WHERE document_id = sqlc.arg(document_id)
  AND (CAST(sqlc.arg(before_revision) AS bigint) = 0 OR revision < sqlc.arg(before_revision))
ORDER BY revision DESC LIMIT 10;

-- name: GetDocumentRevision :one
SELECT revision, title, content_html, content_text, edited_by, created_at
FROM public.document_revisions
WHERE document_id = sqlc.arg(document_id) AND revision = sqlc.arg(revision);

-- name: RestoreDocumentRevision :one
UPDATE public.documents AS document
SET title = snapshot.title, content_html = snapshot.content_html, content_text = snapshot.content_text,
    collaboration_state = NULL,
    collaboration_epoch = document.collaboration_epoch + 1, updated_by = sqlc.arg(actor_id)
FROM public.document_revisions AS snapshot
WHERE document.document_id = sqlc.arg(document_id)
  AND document.revision = sqlc.arg(expected_revision)
  AND snapshot.document_id = document.document_id AND snapshot.revision = sqlc.arg(revision)
RETURNING document.document_id;

-- name: SetDocumentPublicToken :exec
UPDATE public.documents SET public_token = sqlc.narg(public_token)
WHERE document_id = sqlc.arg(document_id);

-- name: GetPublicDocument :one
SELECT document.document_id, document.workspace_id, document.title, document.content_html,
       document.content_text, document.updated_at
FROM public.documents AS document
JOIN public.workspaces AS workspace ON workspace.workspace_id = document.workspace_id
WHERE document.public_token = sqlc.arg(public_token)
  AND document.archived_at IS NULL AND workspace.deleted_at IS NULL;

-- name: AuthorizePublicDocumentMedia :one
SELECT media.attachment_id
FROM public.document_attachments AS media
JOIN public.documents AS document ON document.document_id = media.document_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = document.workspace_id
WHERE document.public_token = sqlc.arg(public_token)
  AND document.archived_at IS NULL AND workspace.deleted_at IS NULL
  AND media.attachment_id = sqlc.arg(attachment_id)
  AND position(CAST(media.attachment_id AS text) in document.content_html) > 0;

-- name: CreateDocumentCollaborationSession :exec
INSERT INTO public.document_collaboration_sessions
    (token_hash, document_id, user_id, session_version, epoch, expires_at)
SELECT sqlc.arg(token_hash), document.document_id, actor.user_id,
       actor.auth_session_version, document.collaboration_epoch, CURRENT_TIMESTAMP + interval '1 hour'
FROM public.documents AS document, public.users AS actor
WHERE document.document_id = sqlc.arg(document_id) AND actor.user_id = sqlc.arg(actor_id);

-- name: DeleteExpiredDocumentCollaborationSessions :exec
DELETE FROM public.document_collaboration_sessions WHERE expires_at < CURRENT_TIMESTAMP;

-- name: DocumentMediaHasHistory :one
SELECT EXISTS (SELECT 1 FROM public.document_revisions
    WHERE document_id = sqlc.arg(document_id)
      AND position(CAST(sqlc.arg(attachment_id) AS text) in content_html) > 0) AS retained;
