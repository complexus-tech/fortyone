-- name: GetAuthorizedWebhookInstallation :one
SELECT
    installation.id,
    installation.workspace_id,
    installation.github_installation_id,
    installation.installation_generation,
    installation.is_active,
    installation.suspended_at,
    installation.disconnected_at,
    repository.id AS repository_id,
    repository.github_repository_id
FROM public.github_installations AS installation
INNER JOIN public.github_repositories AS repository
    ON repository.installation_id = installation.id
   AND repository.workspace_id = installation.workspace_id
WHERE installation.github_installation_id = sqlc.arg(github_installation_id)
  AND repository.github_repository_id = sqlc.arg(github_repository_id)
  AND installation.is_active = TRUE
  AND installation.suspended_at IS NULL
  AND installation.disconnected_at IS NULL
  AND repository.is_active = TRUE;

-- name: GetCurrentWebhookInstallation :one
SELECT
    installation.id,
    installation.workspace_id,
    installation.github_installation_id,
    installation.installation_generation,
    installation.is_active,
    installation.suspended_at,
    installation.disconnected_at,
    repository.id AS repository_id,
    repository.github_repository_id
FROM public.github_installations AS installation
INNER JOIN public.github_repositories AS repository
    ON repository.installation_id = installation.id
   AND repository.workspace_id = installation.workspace_id
WHERE installation.id = sqlc.arg(installation_id)
  AND installation.installation_generation = sqlc.arg(installation_generation)
  AND repository.github_repository_id = sqlc.arg(github_repository_id)
  AND installation.is_active = TRUE
  AND installation.suspended_at IS NULL
  AND installation.disconnected_at IS NULL
  AND repository.is_active = TRUE;
