-- name: LockWorkspaceCommentForMentions :one
SELECT comment.comment_id
FROM public.story_comments AS comment
INNER JOIN public.stories AS story ON story.id = comment.story_id
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE OF comment;

-- name: GetWorkspaceCommentForMentions :one
SELECT comment.comment_id
FROM public.story_comments AS comment
INNER JOIN public.stories AS story ON story.id = comment.story_id
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND story.workspace_id = sqlc.arg(workspace_id);

-- name: DeleteCommentMentions :exec
DELETE FROM public.comment_mentions
WHERE comment_id = sqlc.arg(comment_id);

-- name: InsertActiveWorkspaceCommentMentions :execrows
WITH requested_user AS (
    SELECT requested.user_id
    FROM unnest(CAST(sqlc.arg(mentioned_user_ids) AS uuid[])) AS requested(user_id)
)
INSERT INTO public.comment_mentions (
    comment_id,
    mentioned_user_id
)
SELECT
    sqlc.arg(comment_id),
    requested_user.user_id
FROM requested_user
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = sqlc.arg(workspace_id)
    AND membership.user_id = requested_user.user_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
    AND account.is_active = TRUE
ORDER BY requested_user.user_id;

-- name: ListWorkspaceCommentMentions :many
SELECT mention.mentioned_user_id
FROM public.comment_mentions AS mention
INNER JOIN public.story_comments AS comment ON comment.comment_id = mention.comment_id
INNER JOIN public.stories AS story ON story.id = comment.story_id
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
ORDER BY mention.mentioned_user_id;
