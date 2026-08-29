-- Interactive hard deletion captures every attachment related to the locked
-- target stories before cascading relation deletion. The application passes a
-- bounded look-ahead limit (maximum supported attachments plus one) so an
-- unexpectedly large request fails before any story or attachment is removed.

-- name: ListInteractiveHardDeleteAttachmentCandidates :many
WITH relation AS MATERIALIZED (
    SELECT story_attachment.attachment_id
    FROM public.story_attachments AS story_attachment
    INNER JOIN public.stories AS story
        ON story.id = story_attachment.story_id
       AND story.workspace_id = sqlc.arg(workspace_id)
    INNER JOIN public.attachments AS attachment
        ON attachment.attachment_id = story_attachment.attachment_id
       AND attachment.workspace_id = story.workspace_id
    WHERE story_attachment.story_id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
    UNION
    SELECT inline_attachment.attachment_id
    FROM public.story_inline_attachments AS inline_attachment
    INNER JOIN public.stories AS story
        ON story.id = inline_attachment.story_id
       AND story.workspace_id = sqlc.arg(workspace_id)
    INNER JOIN public.attachments AS attachment
        ON attachment.attachment_id = inline_attachment.attachment_id
       AND attachment.workspace_id = story.workspace_id
    WHERE inline_attachment.story_id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
)
SELECT relation.attachment_id
FROM relation
ORDER BY relation.attachment_id
LIMIT CAST(sqlc.arg(maximum_attachment_count) AS integer);
