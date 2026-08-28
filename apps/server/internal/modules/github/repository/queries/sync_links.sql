-- name: CreateGitHubIssueSyncLink :one
INSERT INTO public.github_issue_sync_links (
    workspace_id,
    repository_id,
    team_id,
    sync_direction,
    created_by_user_id
)
SELECT sqlc.arg(workspace_id),
       gr.id,
       t.team_id,
       sqlc.arg(sync_direction),
       sqlc.arg(created_by_user_id)
FROM public.github_repositories AS gr
INNER JOIN public.teams AS t
    ON t.team_id = sqlc.arg(team_id)
   AND t.workspace_id = sqlc.arg(workspace_id)
WHERE gr.id = sqlc.arg(repository_id)
  AND gr.workspace_id = sqlc.arg(workspace_id)
RETURNING id;

-- name: GetGitHubIssueSyncLink :one
SELECT l.id,
       l.repository_id,
       gr.full_name AS repository_name,
       l.team_id,
       t.name AS team_name,
       t.color AS team_color,
       l.sync_direction,
       l.is_active,
       l.created_at,
       l.updated_at
FROM public.github_issue_sync_links AS l
INNER JOIN public.github_repositories AS gr ON gr.id = l.repository_id
INNER JOIN public.teams AS t ON t.team_id = l.team_id
WHERE l.workspace_id = sqlc.arg(workspace_id)
  AND l.id = sqlc.arg(link_id);

-- name: UpdateGitHubIssueSyncLink :one
UPDATE public.github_issue_sync_links
SET sync_direction = COALESCE(sqlc.narg(sync_direction), sync_direction),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    updated_at = NOW()
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(link_id)
RETURNING id;

-- name: DeleteGitHubIssueSyncLink :exec
DELETE FROM public.github_issue_sync_links
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(link_id);

-- name: FindGitHubIssueSyncLinkByRepositoryID :one
SELECT l.id,
       l.repository_id,
       gr.full_name AS repository_name,
       l.team_id,
       t.name AS team_name,
       t.color AS team_color,
       l.sync_direction,
       l.is_active,
       l.created_at,
       l.updated_at
FROM public.github_issue_sync_links AS l
INNER JOIN public.github_repositories AS gr ON gr.id = l.repository_id
INNER JOIN public.teams AS t ON t.team_id = l.team_id
WHERE l.repository_id = sqlc.arg(repository_id)
  AND l.is_active = TRUE
ORDER BY l.created_at ASC, l.id ASC
LIMIT 1;

-- name: FindBidirectionalGitHubIssueSyncLinkByTeamID :one
SELECT l.id,
       l.repository_id,
       l.team_id,
       l.sync_direction,
       gr.full_name AS repository_name,
       gr.owner_login,
       gr.name AS repository_slug,
       gr.html_url AS repository_html_url,
       gi.github_installation_id
FROM public.github_issue_sync_links AS l
INNER JOIN public.github_repositories AS gr ON gr.id = l.repository_id
INNER JOIN public.github_installations AS gi ON gi.id = gr.installation_id
WHERE l.workspace_id = sqlc.arg(workspace_id)
  AND l.team_id = sqlc.arg(team_id)
  AND l.is_active = TRUE
  AND l.sync_direction = 'bidirectional'
  AND gr.is_active = TRUE
ORDER BY l.created_at ASC, l.id ASC
LIMIT 1;

-- name: ListGitHubTeamWorkflowRules :many
SELECT id,
       event_key,
       target_status_id,
       base_branch_pattern,
       is_active,
       created_at,
       updated_at
FROM public.github_team_workflow_rules
WHERE workspace_id = sqlc.arg(workspace_id)
  AND team_id = sqlc.arg(team_id)
ORDER BY created_at ASC, id ASC;

-- name: LockGitHubTeam :one
SELECT team_id
FROM public.teams
WHERE workspace_id = sqlc.arg(workspace_id)
  AND team_id = sqlc.arg(team_id)
FOR UPDATE;

-- name: DeleteGitHubTeamWorkflowRules :exec
DELETE FROM public.github_team_workflow_rules
WHERE workspace_id = sqlc.arg(workspace_id)
  AND team_id = sqlc.arg(team_id);

-- name: InsertGitHubTeamWorkflowRule :one
INSERT INTO public.github_team_workflow_rules (
    workspace_id,
    team_id,
    event_key,
    target_status_id,
    base_branch_pattern,
    is_active
) SELECT
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    sqlc.arg(event_key),
    sqlc.narg(target_status_id),
    sqlc.narg(base_branch_pattern),
    sqlc.arg(is_active)
WHERE CAST(sqlc.narg(target_status_id) AS uuid) IS NULL
   OR EXISTS (
       SELECT 1
       FROM public.statuses
       WHERE status_id = sqlc.narg(target_status_id)
         AND workspace_id = sqlc.arg(workspace_id)
         AND team_id = sqlc.arg(team_id)
   )
RETURNING id;

-- name: ListGitHubTeamStatuses :many
SELECT status_id,
       name,
       COALESCE(category, '') AS category,
       COALESCE(color, '') AS color
FROM public.statuses
WHERE team_id = sqlc.arg(team_id)
ORDER BY order_index ASC, status_id ASC;
