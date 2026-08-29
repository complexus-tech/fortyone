-- name: GetOrCreateGitHubWorkspaceSettings :one
INSERT INTO public.github_workspace_settings (workspace_id)
VALUES (sqlc.arg(workspace_id))
ON CONFLICT (workspace_id) DO UPDATE
SET workspace_id = EXCLUDED.workspace_id
RETURNING workspace_id,
          branch_format,
          link_commits_by_magic_words,
          sync_assignees,
          sync_labels,
          auto_populate_pr_body,
          close_on_commit_keywords,
          created_at,
          updated_at;

-- name: GetGitHubWorkspaceSettings :one
SELECT workspace_id,
       branch_format,
       link_commits_by_magic_words,
       sync_assignees,
       sync_labels,
       auto_populate_pr_body,
       close_on_commit_keywords,
       created_at,
       updated_at
FROM public.github_workspace_settings
WHERE workspace_id = sqlc.arg(workspace_id);

-- name: UpdateGitHubWorkspaceSettings :one
UPDATE public.github_workspace_settings
SET branch_format = COALESCE(sqlc.narg(branch_format), branch_format),
    link_commits_by_magic_words = COALESCE(sqlc.narg(link_commits_by_magic_words), link_commits_by_magic_words),
    sync_assignees = COALESCE(sqlc.narg(sync_assignees), sync_assignees),
    sync_labels = COALESCE(sqlc.narg(sync_labels), sync_labels),
    auto_populate_pr_body = COALESCE(sqlc.narg(auto_populate_pr_body), auto_populate_pr_body),
    close_on_commit_keywords = COALESCE(sqlc.narg(close_on_commit_keywords), close_on_commit_keywords),
    updated_at = NOW()
WHERE workspace_id = sqlc.arg(workspace_id)
RETURNING workspace_id,
          branch_format,
          link_commits_by_magic_words,
          sync_assignees,
          sync_labels,
          auto_populate_pr_body,
          close_on_commit_keywords,
          created_at,
          updated_at;

-- name: ListGitHubInstallations :many
SELECT id,
       github_installation_id,
       account_id,
       account_login,
       account_type,
       account_avatar_url,
       repository_selection,
       is_active,
       suspended_at,
       disconnected_at,
       created_at,
       updated_at
FROM public.github_installations
WHERE workspace_id = sqlc.arg(workspace_id)
ORDER BY created_at DESC, id ASC;

-- name: ListGitHubRepositories :many
SELECT id,
       installation_id,
       github_repository_id,
       owner_login,
       name,
       full_name,
       description,
       html_url,
       default_branch,
       is_private,
       is_archived,
       is_disabled,
       is_active,
       last_synced_at,
       created_at,
       updated_at
FROM public.github_repositories
WHERE workspace_id = sqlc.arg(workspace_id)
ORDER BY full_name ASC, id ASC;

-- name: ListGitHubIssueSyncLinks :many
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
ORDER BY gr.full_name ASC, l.id ASC;
