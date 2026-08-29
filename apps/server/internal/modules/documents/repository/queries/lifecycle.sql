-- name: GetAccessibleDocumentForDuplicate :one
SELECT
    document.document_id,
    document.workspace_id,
    document.title,
    document.content_html,
    document.content_text,
    document.visibility,
    document.created_by,
    document.updated_by,
    document.created_at,
    document.updated_at,
    document.archived_at,
    TRUE AS can_edit
FROM public.documents AS document
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.archived_at IS NULL
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE document.visibility = 'restricted'
            AND reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
FOR SHARE OF document;

-- name: CopyWorkspaceDocumentMedia :exec
INSERT INTO public.document_attachments (document_id, attachment_id, created_by)
SELECT target.document_id, relation.attachment_id, sqlc.arg(actor_id)
FROM public.document_attachments AS relation
INNER JOIN public.documents AS source ON source.document_id = relation.document_id
INNER JOIN public.documents AS target ON target.document_id = sqlc.arg(target_document_id)
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = relation.attachment_id
   AND attachment.workspace_id = source.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = source.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
   AND membership.role <> CAST('guest' AS public.user_role)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = source.workspace_id
   AND workspace.deleted_at IS NULL
WHERE source.document_id = sqlc.arg(source_document_id)
  AND source.workspace_id = sqlc.arg(workspace_id)
  AND target.workspace_id = source.workspace_id
  AND target.created_by = sqlc.arg(actor_id)
  AND source.archived_at IS NULL
  AND (
      source.visibility = 'workspace'
      OR source.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE source.visibility = 'restricted'
            AND reader.document_id = source.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
ON CONFLICT (document_id, attachment_id) DO NOTHING;

-- name: ArchiveOwnedDocument :one
UPDATE public.documents AS document
SET archived_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP,
    updated_by = sqlc.arg(actor_id)
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.created_by = sqlc.arg(actor_id)
  AND document.archived_at IS NULL
  AND membership.workspace_id = document.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
RETURNING document.document_id;

-- name: LockOwnedDocumentForDelete :one
SELECT document.document_id
FROM public.documents AS document
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.created_by = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
FOR UPDATE OF document;

-- name: ListOrphanedDocumentMediaCandidates :many
SELECT attachment.attachment_id
FROM public.document_attachments AS relation
INNER JOIN public.documents AS document ON document.document_id = relation.document_id
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = relation.attachment_id
   AND attachment.workspace_id = document.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
   AND membership.role <> CAST('guest' AS public.user_role)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.created_by = sqlc.arg(actor_id)
  AND NOT EXISTS (
      SELECT 1
      FROM public.document_attachments AS other_document
      WHERE other_document.attachment_id = relation.attachment_id
        AND other_document.document_id <> document.document_id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_attachments AS story_file
      WHERE story_file.attachment_id = relation.attachment_id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_inline_attachments AS story_media
      WHERE story_media.attachment_id = relation.attachment_id
  )
ORDER BY attachment.attachment_id
FOR UPDATE OF attachment;

-- name: DeleteOwnedDocument :one
DELETE FROM public.documents AS document
USING public.workspace_members AS membership, public.users AS actor, public.workspaces AS workspace
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.created_by = sqlc.arg(actor_id)
  AND membership.workspace_id = document.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
  AND workspace.workspace_id = document.workspace_id
  AND workspace.deleted_at IS NULL
RETURNING document.document_id;
