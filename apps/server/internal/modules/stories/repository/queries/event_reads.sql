-- Metadata for trusted asynchronous events. This is separate from interactive
-- reads: the event actor is explicit, and notification creation rechecks the
-- recipient independently. A system identity must be active in the database.
-- name: GetEventStoryTitle :one
SELECT story.title
FROM public.stories AS story
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = story.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      actor.is_system = TRUE
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS membership
          WHERE membership.workspace_id = story.workspace_id
            AND membership.user_id = actor.user_id
            AND membership.role IN ('admin', 'member', 'guest')
            AND (
                membership.role = 'admin'
                OR EXISTS (
                    SELECT 1 FROM public.team_members AS team_member
                    WHERE team_member.team_id = story.team_id
                      AND team_member.user_id = actor.user_id
                )
            )
      )
  );

-- name: GetSystemStoryStatusCategory :one
SELECT CAST(status.category AS text) AS category
FROM public.statuses AS status
INNER JOIN public.teams AS team ON team.team_id = status.team_id
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = team.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE AND actor.is_system = TRUE
WHERE status.status_id = sqlc.arg(status_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR team.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])));

-- name: GetSystemStoryComment :one
SELECT comment.comment_id, comment.story_id, comment.parent_id, comment.commenter_id,
       comment.content, comment.created_at, comment.updated_at
FROM public.story_comments AS comment
INNER JOIN public.stories AS story ON story.id = comment.story_id
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE AND actor.is_system = TRUE
WHERE comment.comment_id = sqlc.arg(comment_id)
  AND story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])));
