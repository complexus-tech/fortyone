-- Story comment reads repeat the live user/workspace/team fence used by story
-- reads. Credential team restrictions only narrow current product membership.

-- name: ListVisibleStoryCommentRoots :many
SELECT
    comment.comment_id,
    comment.story_id,
    comment.parent_id,
    comment.commenter_id,
    comment.content,
    comment.created_at,
    comment.updated_at
FROM public.story_comments AS comment
INNER JOIN public.stories AS story
    ON story.id = comment.story_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE comment.story_id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND comment.parent_id IS NULL
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY comment.created_at DESC, comment.comment_id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: ListVisibleStoryCommentReplies :many
SELECT
    reply.comment_id,
    reply.story_id,
    reply.parent_id,
    reply.commenter_id,
    reply.content,
    reply.created_at,
    reply.updated_at
FROM public.story_comments AS reply
INNER JOIN public.stories AS story
    ON story.id = reply.story_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE reply.story_id = sqlc.arg(story_id)
  AND reply.parent_id = ANY(CAST(sqlc.arg(parent_ids) AS uuid[]))
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY reply.parent_id, reply.created_at, reply.comment_id;

-- name: GetVisibleStoryComment :one
SELECT
    comment.comment_id,
    comment.story_id,
    comment.parent_id,
    comment.commenter_id,
    comment.content,
    comment.created_at,
    comment.updated_at
FROM public.story_comments AS comment
INNER JOIN public.stories AS story
    ON story.id = comment.story_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND comment.story_id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  );
