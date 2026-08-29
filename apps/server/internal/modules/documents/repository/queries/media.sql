-- name: LinkEditableDocumentMedia :one
INSERT INTO public.document_attachments (document_id, attachment_id, created_by)
SELECT document.document_id, attachment.attachment_id, sqlc.arg(actor_id)
FROM public.documents AS document
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = sqlc.arg(attachment_id)
   AND attachment.workspace_id = document.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.archived_at IS NULL
  AND membership.role <> CAST('guest' AS public.user_role)
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS editor
          WHERE document.visibility = 'restricted'
            AND editor.document_id = document.document_id
            AND editor.user_id = sqlc.arg(actor_id)
            AND editor.role = 'editor'
      )
  )
ON CONFLICT (document_id, attachment_id) DO UPDATE
SET created_by = document_attachments.created_by
RETURNING attachment_id;

-- name: AuthorizeAccessibleDocumentMedia :one
SELECT attachment.attachment_id
FROM public.document_attachments AS media
INNER JOIN public.documents AS document ON document.document_id = media.document_id
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = media.attachment_id
   AND attachment.workspace_id = document.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND media.attachment_id = sqlc.arg(attachment_id)
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
  );

-- name: UnlinkEditableDocumentMedia :one
DELETE FROM public.document_attachments AS media
USING public.documents AS document, public.attachments AS attachment,
      public.workspace_members AS membership, public.users AS actor,
      public.workspaces AS workspace
WHERE media.document_id = document.document_id
  AND media.attachment_id = attachment.attachment_id
  AND media.document_id = sqlc.arg(document_id)
  AND media.attachment_id = sqlc.arg(attachment_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND attachment.workspace_id = document.workspace_id
  AND document.archived_at IS NULL
  AND membership.workspace_id = document.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
  AND actor.user_id = membership.user_id
  AND actor.is_active = TRUE
  AND workspace.workspace_id = document.workspace_id
  AND workspace.deleted_at IS NULL
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS editor
          WHERE document.visibility = 'restricted'
            AND editor.document_id = document.document_id
            AND editor.user_id = sqlc.arg(actor_id)
            AND editor.role = 'editor'
      )
  )
RETURNING media.attachment_id;

-- name: IsWorkspaceAttachmentUnreferenced :one
SELECT NOT EXISTS (
           SELECT 1
           FROM public.document_attachments AS document_media
           WHERE document_media.attachment_id = attachment.attachment_id
       )
       AND NOT EXISTS (
           SELECT 1
           FROM public.story_attachments AS story_file
           WHERE story_file.attachment_id = attachment.attachment_id
       )
       AND NOT EXISTS (
           SELECT 1
           FROM public.story_inline_attachments AS story_media
           WHERE story_media.attachment_id = attachment.attachment_id
       ) AS is_unreferenced
FROM public.attachments AS attachment
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = attachment.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
   AND membership.role <> CAST('guest' AS public.user_role)
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = attachment.workspace_id
   AND workspace.deleted_at IS NULL
WHERE attachment.attachment_id = sqlc.arg(attachment_id)
  AND attachment.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE OF attachment;
