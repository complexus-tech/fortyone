-- name: AuthorizeAgentReply :one
SELECT parent.comment_id
FROM public.story_comments AS parent
JOIN public.stories AS story ON story.id = parent.story_id AND story.deleted_at IS NULL
JOIN public.workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
JOIN public.users AS requester ON requester.user_id = sqlc.arg(requester_id) AND requester.is_active AND NOT requester.is_system
JOIN public.workspace_members AS member ON member.workspace_id = story.workspace_id AND member.user_id = requester.user_id AND member.role IN ('admin', 'member')
JOIN public.team_members AS team_member ON team_member.team_id = story.team_id AND team_member.user_id = requester.user_id
JOIN public.users AS author ON author.user_id = sqlc.arg(author_id) AND author.is_active AND author.is_system
WHERE parent.comment_id = sqlc.arg(parent_id)
  AND parent.commenter_id = requester.user_id
  AND story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean) OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
  AND (CAST(sqlc.arg(author_team_access_unrestricted) AS boolean) OR story.team_id = ANY(CAST(sqlc.arg(author_allowed_team_ids) AS uuid[])));

-- name: ResolveAgentMentionTargets :many
SELECT recipient.user_id, CAST(COALESCE(NULLIF(TRIM(recipient.full_name), ''), recipient.username) AS text) AS name
FROM public.stories AS story
JOIN public.workspace_members AS member ON member.workspace_id = story.workspace_id
JOIN public.team_members AS team_member ON team_member.team_id = story.team_id AND team_member.user_id = member.user_id
JOIN public.users AS recipient ON recipient.user_id = member.user_id AND recipient.is_active AND NOT recipient.is_system
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND recipient.user_id = ANY(CAST(sqlc.arg(mentioned_user_ids) AS uuid[]))
ORDER BY recipient.user_id;

-- name: InsertAgentCommentMentions :execrows
INSERT INTO public.comment_mentions (comment_id, mentioned_user_id)
SELECT comment.comment_id, recipient.user_id
FROM public.story_comments AS comment
JOIN public.story_comments AS parent ON parent.comment_id = comment.parent_id AND parent.commenter_id = sqlc.arg(requester_id)
JOIN public.stories AS story ON story.id = comment.story_id AND story.id = parent.story_id AND story.deleted_at IS NULL
JOIN public.workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
JOIN public.users AS requester ON requester.user_id = parent.commenter_id AND requester.is_active AND NOT requester.is_system
JOIN public.workspace_members AS actor_member ON actor_member.workspace_id = story.workspace_id AND actor_member.user_id = requester.user_id AND actor_member.role IN ('admin', 'member')
JOIN public.team_members AS actor_team_member ON actor_team_member.team_id = story.team_id AND actor_team_member.user_id = requester.user_id
JOIN public.users AS author ON author.user_id = comment.commenter_id AND author.is_active AND author.is_system
JOIN public.workspace_members AS member ON member.workspace_id = story.workspace_id
JOIN public.team_members AS team_member ON team_member.team_id = story.team_id AND team_member.user_id = member.user_id
JOIN public.users AS recipient ON recipient.user_id = member.user_id AND recipient.is_active AND NOT recipient.is_system
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND comment.commenter_id = sqlc.arg(author_id)
  AND parent.comment_id = sqlc.arg(parent_id)
  AND story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND (CAST(sqlc.arg(team_access_unrestricted) AS boolean) OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[])))
  AND (CAST(sqlc.arg(author_team_access_unrestricted) AS boolean) OR story.team_id = ANY(CAST(sqlc.arg(author_allowed_team_ids) AS uuid[])))
  AND recipient.user_id = ANY(CAST(sqlc.arg(mentioned_user_ids) AS uuid[]));
