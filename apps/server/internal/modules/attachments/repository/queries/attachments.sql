-- name: CreateAttachment :one
INSERT INTO public.attachments (
    filename,
    blob_name,
    size,
    mime_type,
    uploaded_by,
    workspace_id,
    scan_status,
    optimization_status
)
VALUES (
    sqlc.arg(filename),
    sqlc.arg(blob_name),
    sqlc.arg(size),
    sqlc.arg(mime_type),
    sqlc.arg(uploaded_by),
    sqlc.arg(workspace_id),
    sqlc.arg(scan_status),
    sqlc.arg(optimization_status)
)
RETURNING *;

-- name: GetWorkspaceAttachment :one
SELECT *
FROM public.attachments
WHERE attachment_id = sqlc.arg(attachment_id)
  AND workspace_id = sqlc.arg(workspace_id);

-- name: ListStoryAttachments :many
SELECT attachment.*
FROM public.attachments AS attachment
INNER JOIN public.story_attachments AS relation
    ON relation.attachment_id = attachment.attachment_id
INNER JOIN public.stories AS story
    ON story.id = relation.story_id
WHERE relation.story_id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND attachment.workspace_id = story.workspace_id
ORDER BY attachment.created_at DESC, attachment.attachment_id DESC;

-- name: StoryExistsInWorkspace :one
SELECT EXISTS (
    SELECT 1
    FROM public.stories
    WHERE id = sqlc.arg(story_id)
      AND workspace_id = sqlc.arg(workspace_id)
);

-- name: DeleteWorkspaceAttachment :execrows
DELETE FROM public.attachments
WHERE attachment_id = sqlc.arg(attachment_id)
  AND workspace_id = sqlc.arg(workspace_id);

-- name: DeleteUnreferencedWorkspaceAttachment :one
DELETE FROM public.attachments AS attachment
WHERE attachment.attachment_id = sqlc.arg(attachment_id)
  AND attachment.workspace_id = sqlc.arg(workspace_id)
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_inline_attachments AS inline_relation
      WHERE inline_relation.attachment_id = attachment.attachment_id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_attachments AS story_relation
      WHERE story_relation.attachment_id = attachment.attachment_id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.document_attachments AS document_relation
      WHERE document_relation.attachment_id = attachment.attachment_id
  )
RETURNING attachment.*;

-- name: LinkWorkspaceStoryAttachment :one
INSERT INTO public.story_attachments (story_id, attachment_id)
SELECT story.id, attachment.attachment_id
FROM public.stories AS story
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = sqlc.arg(attachment_id)
   AND attachment.workspace_id = story.workspace_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
ON CONFLICT (story_id, attachment_id) DO UPDATE
SET attachment_id = EXCLUDED.attachment_id
RETURNING attachment_id;

-- name: LinkWorkspaceStoryMedia :one
INSERT INTO public.story_inline_attachments (story_id, attachment_id, created_by)
SELECT story.id, attachment.attachment_id, sqlc.arg(created_by)
FROM public.stories AS story
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = sqlc.arg(attachment_id)
   AND attachment.workspace_id = story.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = story.workspace_id
   AND membership.user_id = sqlc.arg(created_by)
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
ON CONFLICT (story_id, attachment_id) DO UPDATE
SET created_by = story_inline_attachments.created_by
RETURNING attachment_id;

-- name: AuthorizeWorkspaceStoryMedia :one
SELECT attachment.*
FROM public.story_inline_attachments AS media
INNER JOIN public.stories AS story ON story.id = media.story_id
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = media.attachment_id
   AND attachment.workspace_id = story.workspace_id
WHERE media.story_id = sqlc.arg(story_id)
  AND media.attachment_id = sqlc.arg(attachment_id)
  AND story.workspace_id = sqlc.arg(workspace_id);

-- name: AuthorizeWorkspaceStoryAttachment :one
SELECT attachment.*
FROM public.story_attachments AS relation
INNER JOIN public.stories AS story ON story.id = relation.story_id
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = relation.attachment_id
   AND attachment.workspace_id = story.workspace_id
WHERE relation.story_id = sqlc.arg(story_id)
  AND relation.attachment_id = sqlc.arg(attachment_id)
  AND story.workspace_id = sqlc.arg(workspace_id);

-- name: UnlinkWorkspaceStoryMedia :one
DELETE FROM public.story_inline_attachments AS media
USING public.stories AS story, public.attachments AS attachment
WHERE media.story_id = story.id
  AND media.attachment_id = attachment.attachment_id
  AND media.story_id = sqlc.arg(story_id)
  AND media.attachment_id = sqlc.arg(attachment_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND attachment.workspace_id = story.workspace_id
RETURNING media.attachment_id;

-- name: StartWorkspaceAttachmentOptimization :one
UPDATE public.attachments
SET optimization_status = 'processing',
    optimization_attempts = optimization_attempts + 1,
    optimization_started_at = CURRENT_TIMESTAMP,
    optimization_completed_at = NULL,
    optimization_lease_expires_at = CURRENT_TIMESTAMP + CAST(sqlc.arg(lease_seconds) AS integer) * INTERVAL '1 second',
    optimization_last_error = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE attachment_id = sqlc.arg(attachment_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND (
      optimization_status IN ('queued', 'failed')
      OR (
          optimization_status = 'processing'
          AND optimization_lease_expires_at < CURRENT_TIMESTAMP
      )
  )
RETURNING *;

-- name: CompleteWorkspaceAttachmentOptimization :execrows
UPDATE public.attachments
SET size = sqlc.arg(size),
    mime_type = sqlc.arg(mime_type),
    optimization_status = sqlc.arg(optimization_status),
    optimization_completed_at = CURRENT_TIMESTAMP,
    optimization_lease_expires_at = NULL,
    optimization_last_error = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE attachment_id = sqlc.arg(attachment_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND optimization_status = 'processing';

-- name: FailWorkspaceAttachmentOptimization :execrows
UPDATE public.attachments
SET optimization_status = 'failed',
    optimization_completed_at = CURRENT_TIMESTAMP,
    optimization_lease_expires_at = NULL,
    optimization_last_error = sqlc.arg(error_message),
    updated_at = CURRENT_TIMESTAMP
WHERE attachment_id = sqlc.arg(attachment_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND optimization_status = 'processing';

-- name: FailQueuedWorkspaceAttachmentOptimization :execrows
UPDATE public.attachments
SET optimization_status = 'failed',
    optimization_completed_at = CURRENT_TIMESTAMP,
    optimization_lease_expires_at = NULL,
    optimization_last_error = sqlc.arg(error_message),
    updated_at = CURRENT_TIMESTAMP
WHERE attachment_id = sqlc.arg(attachment_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND optimization_status = 'queued';
