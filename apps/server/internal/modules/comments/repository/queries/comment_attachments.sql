-- name: ReplaceCommentAttachmentReferences :one
WITH removed AS (
    DELETE FROM public.story_comment_attachments AS reference
    WHERE reference.comment_id = sqlc.arg(comment_id)
      AND NOT (reference.attachment_id = ANY(CAST(sqlc.arg(attachment_ids) AS uuid[])))
), valid AS (
    SELECT attachment.attachment_id
    FROM public.story_attachments AS link
    INNER JOIN public.attachments AS attachment ON attachment.attachment_id = link.attachment_id
    INNER JOIN public.story_comments AS comment ON comment.story_id = link.story_id
    WHERE comment.comment_id = sqlc.arg(comment_id)
      AND attachment.workspace_id = sqlc.arg(workspace_id)
      AND attachment.attachment_id = ANY(CAST(sqlc.arg(attachment_ids) AS uuid[]))
), inserted AS (
    INSERT INTO public.story_comment_attachments (comment_id, attachment_id)
    SELECT sqlc.arg(comment_id), valid.attachment_id FROM valid
    ON CONFLICT (comment_id, attachment_id) DO NOTHING
)
SELECT COUNT(*) FROM valid;
