-- name: MarkStoryCommentSeen :one
INSERT INTO public.story_comment_seen (comment_id, user_id)
SELECT comment.comment_id, actor.user_id
FROM public.story_comments AS comment
JOIN public.stories AS story ON story.id = comment.story_id
JOIN public.workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
JOIN public.users AS actor ON actor.user_id = sqlc.arg(actor_id) AND actor.is_active = TRUE AND actor.is_system = FALSE
JOIN public.workspace_members AS member ON member.workspace_id = workspace.workspace_id AND member.user_id = actor.user_id
JOIN public.team_members AS team_member ON team_member.team_id = story.team_id AND team_member.user_id = actor.user_id
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND story.id = sqlc.arg(story_id) AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean)
       OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
ON CONFLICT (comment_id, user_id) DO UPDATE SET seen_at = story_comment_seen.seen_at
RETURNING seen_at;
