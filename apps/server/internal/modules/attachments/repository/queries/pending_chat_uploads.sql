-- Only unclaimed uploads expire. A sender holding the association row lock
-- wins over cleanup; overlapping cleanup workers skip each other's candidates.
-- name: ExpirePendingChatUploads :many
WITH candidates AS (
    SELECT pending.attachment_id
    FROM public.chat_attachments AS pending
    WHERE pending.message_id IS NULL AND pending.created_at < sqlc.arg(created_before)
    ORDER BY pending.created_at, pending.attachment_id
    FOR UPDATE OF pending SKIP LOCKED
    LIMIT sqlc.arg(batch_limit)
)
DELETE FROM public.chat_attachments AS attachment
USING candidates
WHERE attachment.attachment_id = candidates.attachment_id AND attachment.message_id IS NULL
RETURNING attachment.attachment_id, attachment.workspace_id;
