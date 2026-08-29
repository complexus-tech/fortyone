-- name: GetSlackAgentSettingsTrusted :one
SELECT
    workspace.workspace_id,
    COALESCE(settings.guidance, '') AS guidance,
    COALESCE(settings.created_at, workspace.created_at) AS created_at,
    COALESCE(settings.updated_at, workspace.updated_at) AS updated_at
FROM public.workspaces AS workspace
LEFT JOIN public.slack_agent_settings AS settings
  ON settings.workspace_id = workspace.workspace_id
WHERE workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND workspace.deleted_at IS NULL;

-- name: GetSlackAgentSettingsForAdmin :one
SELECT
    workspace.workspace_id,
    COALESCE(settings.guidance, '') AS guidance,
    COALESCE(settings.created_at, workspace.created_at) AS created_at,
    COALESCE(settings.updated_at, workspace.updated_at) AS updated_at
FROM public.workspaces AS workspace
JOIN public.workspace_members AS membership
  ON membership.workspace_id = workspace.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role = 'admin'
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
LEFT JOIN public.slack_agent_settings AS settings
  ON settings.workspace_id = workspace.workspace_id
WHERE workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND workspace.deleted_at IS NULL;

-- name: UpsertSlackAgentSettingsForAdmin :one
INSERT INTO public.slack_agent_settings (workspace_id, guidance)
SELECT workspace.workspace_id, CAST(sqlc.arg(guidance) AS text)
FROM public.workspaces AS workspace
JOIN public.workspace_members AS membership
  ON membership.workspace_id = workspace.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role = 'admin'
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
WHERE workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND workspace.deleted_at IS NULL
ON CONFLICT (workspace_id) DO UPDATE SET
    guidance = EXCLUDED.guidance,
    updated_at = NOW()
RETURNING workspace_id, guidance, created_at, updated_at;
