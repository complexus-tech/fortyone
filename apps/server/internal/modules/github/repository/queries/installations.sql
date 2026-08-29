-- name: UpsertGitHubInstallation :one
INSERT INTO public.github_installations (
    workspace_id,
    github_app_id,
    github_installation_id,
    account_id,
    account_login,
    account_type,
    account_avatar_url,
    repository_selection,
    permissions,
    events,
    installed_by_user_id,
    installed_by_github_user_id,
    is_active,
    disconnected_at,
    updated_at
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(github_app_id),
    sqlc.arg(github_installation_id),
    sqlc.arg(account_id),
    sqlc.arg(account_login),
    sqlc.arg(account_type),
    sqlc.narg(account_avatar_url),
    sqlc.arg(repository_selection),
    sqlc.arg(permissions),
    sqlc.arg(events),
    sqlc.arg(installed_by_user_id),
    sqlc.arg(installed_by_github_user_id),
    TRUE,
    NULL,
    NOW()
)
ON CONFLICT (github_installation_id) DO UPDATE SET
    account_id = EXCLUDED.account_id,
    account_login = EXCLUDED.account_login,
    account_type = EXCLUDED.account_type,
    account_avatar_url = EXCLUDED.account_avatar_url,
    repository_selection = EXCLUDED.repository_selection,
    permissions = EXCLUDED.permissions,
    events = EXCLUDED.events,
    is_active = TRUE,
    disconnected_at = NULL,
    updated_at = NOW()
WHERE github_installations.workspace_id = EXCLUDED.workspace_id
RETURNING id;

-- name: UpsertGitHubRepository :exec
INSERT INTO public.github_repositories (
    workspace_id,
    installation_id,
    github_repository_id,
    owner_id,
    owner_login,
    name,
    full_name,
    description,
    html_url,
    clone_url,
    ssh_url,
    default_branch,
    is_private,
    is_archived,
    is_disabled,
    is_active,
    last_synced_at,
    updated_at
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(installation_id),
    sqlc.arg(github_repository_id),
    sqlc.arg(owner_id),
    sqlc.arg(owner_login),
    sqlc.arg(name),
    sqlc.arg(full_name),
    sqlc.narg(description),
    sqlc.arg(html_url),
    sqlc.arg(clone_url),
    sqlc.arg(ssh_url),
    sqlc.arg(default_branch),
    sqlc.arg(is_private),
    sqlc.arg(is_archived),
    sqlc.arg(is_disabled),
    TRUE,
    NOW(),
    NOW()
)
ON CONFLICT (installation_id, github_repository_id) DO UPDATE SET
    owner_id = EXCLUDED.owner_id,
    owner_login = EXCLUDED.owner_login,
    name = EXCLUDED.name,
    full_name = EXCLUDED.full_name,
    description = EXCLUDED.description,
    html_url = EXCLUDED.html_url,
    clone_url = EXCLUDED.clone_url,
    ssh_url = EXCLUDED.ssh_url,
    default_branch = EXCLUDED.default_branch,
    is_private = EXCLUDED.is_private,
    is_archived = EXCLUDED.is_archived,
    is_disabled = EXCLUDED.is_disabled,
    is_active = TRUE,
    last_synced_at = NOW(),
    updated_at = NOW();

-- name: DeactivateAllGitHubRepositories :exec
UPDATE public.github_repositories
SET is_active = FALSE,
    updated_at = NOW()
WHERE installation_id = sqlc.arg(installation_id);

-- name: DeactivateMissingGitHubRepositories :exec
UPDATE public.github_repositories
SET is_active = FALSE,
    updated_at = NOW()
WHERE installation_id = sqlc.arg(installation_id)
  AND NOT (github_repository_id = ANY(CAST(sqlc.arg(github_repository_ids) AS bigint[])));
