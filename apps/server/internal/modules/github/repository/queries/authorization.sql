-- name: GetGitHubWorkspaceRole :one
SELECT CAST(wm.role AS text) AS role
FROM public.workspace_members AS wm
INNER JOIN public.users AS u ON u.user_id = wm.user_id
INNER JOIN public.workspaces AS w ON w.workspace_id = wm.workspace_id
WHERE wm.workspace_id = sqlc.arg(workspace_id)
  AND wm.user_id = sqlc.arg(actor_id)
  AND u.is_active = TRUE
  AND w.deleted_at IS NULL;

