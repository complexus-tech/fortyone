-- name: CreateStoryLink :exec
INSERT INTO public.story_links (title, url, story_id, external_source_key)
VALUES (
    CAST(sqlc.arg(title) AS text),
    CAST(sqlc.arg(url) AS text),
    CAST(sqlc.arg(story_id) AS uuid),
    NULLIF(CAST(sqlc.arg(source_key) AS text), '')
)
ON CONFLICT (external_source_key)
WHERE external_source_key IS NOT NULL
DO UPDATE SET
    title = EXCLUDED.title,
    url = EXCLUDED.url,
    story_id = EXCLUDED.story_id,
    updated_at = NOW();

-- name: ListWorkspaceMembersForSlackLinking :many
SELECT actor.user_id, actor.email
FROM public.workspace_members AS membership
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
 AND actor.is_system = FALSE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = membership.workspace_id
 AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND TRIM(COALESCE(actor.email, '')) <> ''
ORDER BY LOWER(actor.email), actor.email, actor.user_id;

-- name: UpsertSlackUserLink :execrows
INSERT INTO public.slack_user_links (
    workspace_id,
    slack_workspace_id,
    slack_team_id,
    slack_user_id,
    user_id,
    linked_via,
    linked_at
)
SELECT
    CAST(sqlc.arg(workspace_id) AS uuid),
    CAST(sqlc.arg(slack_workspace_id) AS uuid),
    CAST(sqlc.arg(slack_team_id) AS text),
    CAST(sqlc.arg(slack_user_id) AS text),
    CAST(sqlc.arg(user_id) AS uuid),
    CAST(sqlc.arg(linked_via) AS text),
    NOW()
FROM public.slack_workspaces AS installation
JOIN public.workspace_members AS membership
  ON membership.workspace_id = installation.workspace_id
 AND membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = membership.workspace_id
 AND workspace.deleted_at IS NULL
WHERE installation.id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND installation.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND installation.slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation.is_active = TRUE
ON CONFLICT (workspace_id, slack_team_id, slack_user_id) DO UPDATE SET
    slack_workspace_id = EXCLUDED.slack_workspace_id,
    user_id = EXCLUDED.user_id,
    linked_via = EXCLUDED.linked_via,
    linked_at = NOW(),
    updated_at = NOW();

-- name: FindLinkedUserIDBySlackUser :one
SELECT link.user_id
FROM public.slack_user_links AS link
JOIN public.users AS actor
  ON actor.user_id = link.user_id
 AND actor.is_active = TRUE
JOIN public.workspace_members AS membership
  ON membership.workspace_id = link.workspace_id
 AND membership.user_id = link.user_id
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = link.workspace_id
 AND workspace.deleted_at IS NULL
JOIN public.slack_workspaces AS installation
  ON installation.id = link.slack_workspace_id
 AND installation.workspace_id = link.workspace_id
 AND installation.slack_team_id = link.slack_team_id
 AND installation.is_active = TRUE
WHERE link.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND link.slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND link.slack_user_id = CAST(sqlc.arg(slack_user_id) AS text)
LIMIT 1;

-- name: FindSlackUserLinkByUser :one
SELECT link.slack_user_id, link.user_id, link.linked_via, link.linked_at
FROM public.slack_user_links AS link
JOIN public.users AS actor
  ON actor.user_id = link.user_id
 AND actor.is_active = TRUE
JOIN public.workspace_members AS membership
  ON membership.workspace_id = link.workspace_id
 AND membership.user_id = link.user_id
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = link.workspace_id
 AND workspace.deleted_at IS NULL
JOIN public.slack_workspaces AS installation
  ON installation.id = link.slack_workspace_id
 AND installation.workspace_id = link.workspace_id
 AND installation.slack_team_id = link.slack_team_id
 AND installation.is_active = TRUE
WHERE link.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND link.slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND link.user_id = CAST(sqlc.arg(user_id) AS uuid)
LIMIT 1;

-- name: DeleteSlackUserLink :execrows
DELETE FROM public.slack_user_links
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND slack_user_id = CAST(sqlc.arg(slack_user_id) AS text)
  AND user_id = CAST(sqlc.arg(user_id) AS uuid);

