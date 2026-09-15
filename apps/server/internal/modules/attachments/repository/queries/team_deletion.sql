-- Capture and lock metadata before cascading relation deletion. Locks protect
-- shared-file checks against concurrent attachment links (their FK key locks).
-- name: LockTeamDeletionAttachments :many
WITH candidates AS (
    SELECT relation.attachment_id
    FROM public.story_attachments AS relation
    INNER JOIN public.stories AS story ON story.id = relation.story_id
    WHERE story.team_id = sqlc.arg(team_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
    UNION
    SELECT relation.attachment_id
    FROM public.story_inline_attachments AS relation
    INNER JOIN public.stories AS story ON story.id = relation.story_id
    WHERE story.team_id = sqlc.arg(team_id)
      AND story.workspace_id = sqlc.arg(workspace_id)
    UNION
    SELECT relation.attachment_id
    FROM public.feedback_item_attachments AS relation
    INNER JOIN public.feedback_items AS item ON item.id = relation.item_id
    INNER JOIN public.feedback_boards AS board ON board.id = item.board_id
    WHERE board.team_id = sqlc.arg(team_id)
      AND board.workspace_id = sqlc.arg(workspace_id)
      AND item.workspace_id = board.workspace_id
)
SELECT attachment.attachment_id
FROM public.attachments AS attachment
INNER JOIN candidates ON candidates.attachment_id = attachment.attachment_id
WHERE attachment.workspace_id = sqlc.arg(workspace_id)
ORDER BY attachment.attachment_id
FOR UPDATE OF attachment;

-- Retire only files with no surviving consumer and atomically enqueue physical
-- deletion. Batches avoid one database round trip per uploaded file.
-- name: RetireTeamDeletionAttachments :execrows
WITH retired AS (
    DELETE FROM public.attachments AS attachment
    WHERE attachment.attachment_id = ANY(CAST(sqlc.arg(attachment_ids) AS uuid[]))
      AND attachment.workspace_id = sqlc.arg(workspace_id)
      AND NOT EXISTS (
          SELECT 1 FROM public.story_attachments AS relation
          WHERE relation.attachment_id = attachment.attachment_id
      )
      AND NOT EXISTS (
          SELECT 1 FROM public.story_inline_attachments AS relation
          WHERE relation.attachment_id = attachment.attachment_id
      )
      AND NOT EXISTS (
          SELECT 1 FROM public.document_attachments AS relation
          WHERE relation.attachment_id = attachment.attachment_id
      )
      AND NOT EXISTS (
          SELECT 1 FROM public.feedback_item_attachments AS relation
          WHERE relation.attachment_id = attachment.attachment_id
      )
    RETURNING attachment.attachment_id, attachment.workspace_id, attachment.blob_name
)
INSERT INTO public.attachment_object_deletion_outbox (
    attachment_id, workspace_id, storage_provider, container_name, blob_name,
    status, attempt_count, next_attempt_at, created_at, updated_at
)
SELECT
    retired.attachment_id, retired.workspace_id, sqlc.arg(storage_provider),
    sqlc.arg(container_name), retired.blob_name, 'pending', 0,
    sqlc.arg(deleted_at), sqlc.arg(deleted_at), sqlc.arg(deleted_at)
FROM retired;
