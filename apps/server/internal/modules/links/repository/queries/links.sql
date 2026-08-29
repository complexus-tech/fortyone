-- name: CreateLinkForWorkspace :one
WITH authorized_actor AS MATERIALIZED (
    SELECT member.workspace_id
    FROM public.workspace_members AS member
    INNER JOIN public.users AS actor
        ON actor.user_id = member.user_id
       AND actor.is_active = TRUE
    WHERE member.workspace_id = sqlc.arg(workspace_id)
      AND member.user_id = sqlc.arg(actor_id)
      AND member.role IN ('member', 'admin')
    FOR UPDATE OF member, actor
)
INSERT INTO public.story_links (title, url, story_id)
SELECT
    CAST(sqlc.narg(title) AS text),
    CAST(sqlc.arg(url) AS text),
    story.id
FROM public.stories AS story
INNER JOIN authorized_actor
    ON authorized_actor.workspace_id = story.workspace_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
RETURNING
    link_id,
    title,
    url,
    story_id,
    created_at,
    updated_at;

-- name: UpdateLinkForWorkspace :execrows
WITH authorized_actor AS MATERIALIZED (
    SELECT member.workspace_id
    FROM public.workspace_members AS member
    INNER JOIN public.users AS actor
        ON actor.user_id = member.user_id
       AND actor.is_active = TRUE
    WHERE member.workspace_id = sqlc.arg(workspace_id)
      AND member.user_id = sqlc.arg(actor_id)
      AND member.role IN ('member', 'admin')
    FOR UPDATE OF member, actor
)
UPDATE public.story_links AS link
SET
    title = COALESCE(CAST(sqlc.narg(title) AS text), link.title),
    url = COALESCE(CAST(sqlc.narg(url) AS text), link.url),
    updated_at = CURRENT_TIMESTAMP
FROM public.stories AS story, authorized_actor
WHERE link.link_id = sqlc.arg(link_id)
  AND story.id = link.story_id
  AND story.workspace_id = authorized_actor.workspace_id;

-- name: DeleteLinkForWorkspace :execrows
WITH authorized_actor AS MATERIALIZED (
    SELECT member.workspace_id
    FROM public.workspace_members AS member
    INNER JOIN public.users AS actor
        ON actor.user_id = member.user_id
       AND actor.is_active = TRUE
    WHERE member.workspace_id = sqlc.arg(workspace_id)
      AND member.user_id = sqlc.arg(actor_id)
      AND member.role IN ('member', 'admin')
    FOR UPDATE OF member, actor
)
DELETE FROM public.story_links AS link
USING public.stories AS story, authorized_actor
WHERE link.link_id = sqlc.arg(link_id)
  AND story.id = link.story_id
  AND story.workspace_id = authorized_actor.workspace_id;
