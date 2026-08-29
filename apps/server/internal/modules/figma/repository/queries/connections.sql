-- name: DeactivateConnectionWebhooks :execrows
UPDATE public.figma_webhooks AS webhook
SET is_active = FALSE,
    updated_at = sqlc.arg(updated_at)
WHERE webhook.connection_id IN (
    SELECT connection.id
    FROM public.figma_connections AS connection
    WHERE connection.workspace_id = sqlc.arg(workspace_id)
      AND connection.is_active
)
  AND webhook.is_active;

-- name: DeactivateWorkspaceConnection :execrows
UPDATE public.figma_connections
SET is_active = FALSE,
    disconnected_at = sqlc.arg(disconnected_at),
    updated_at = sqlc.arg(disconnected_at)
WHERE workspace_id = sqlc.arg(workspace_id)
  AND is_active;

-- name: CreateConnection :one
INSERT INTO public.figma_connections (
    id,
    workspace_id,
    figma_user_id,
    figma_email,
    figma_handle,
    token_payload,
    credential_key_version,
    installation_generation,
    scopes,
    expires_at,
    connected_by_user_id
)
SELECT
    sqlc.arg(id),
    workspace.workspace_id,
    sqlc.arg(figma_user_id),
    CAST(sqlc.narg(figma_email) AS text),
    CAST(sqlc.narg(figma_handle) AS text),
    sqlc.arg(token_payload),
    sqlc.arg(credential_key_version),
    sqlc.arg(installation_generation),
    sqlc.arg(scopes),
    sqlc.arg(expires_at),
    member.user_id
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS member
    ON member.workspace_id = workspace.workspace_id
   AND member.user_id = sqlc.arg(connected_by_user_id)
INNER JOIN public.users AS account
    ON account.user_id = member.user_id
   AND account.is_active = TRUE
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL
RETURNING id, workspace_id, figma_user_id, figma_email, figma_handle,
    token_payload, credential_key_version, installation_generation, scopes,
    expires_at, connected_by_user_id, is_active, created_at, updated_at;

-- name: GetActiveConnection :one
SELECT id, workspace_id, figma_user_id, figma_email, figma_handle,
    token_payload, credential_key_version, installation_generation, scopes,
    expires_at, connected_by_user_id, is_active, created_at, updated_at
FROM public.figma_connections
WHERE workspace_id = sqlc.arg(workspace_id)
  AND is_active;

-- name: CompareAndSwapConnectionCredential :execrows
UPDATE public.figma_connections
SET token_payload = sqlc.arg(next_payload),
    credential_key_version = sqlc.arg(credential_key_version),
    expires_at = sqlc.arg(expires_at),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
  AND installation_generation = sqlc.arg(installation_generation)
  AND token_payload = sqlc.arg(previous_payload)
  AND is_active;

-- name: ListLegacyConnections :many
SELECT id, workspace_id, token_payload, installation_generation
FROM public.figma_connections
WHERE credential_key_version = 1
  AND (
      CAST(sqlc.narg(after_id) AS uuid) IS NULL
      OR id > CAST(sqlc.narg(after_id) AS uuid)
  )
ORDER BY id
LIMIT sqlc.arg(page_limit);

-- name: UpgradeLegacyConnectionCredential :execrows
UPDATE public.figma_connections
SET token_payload = sqlc.arg(next_payload),
    credential_key_version = sqlc.arg(credential_key_version),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
  AND installation_generation = sqlc.arg(installation_generation)
  AND credential_key_version = 1
  AND token_payload = sqlc.arg(previous_payload);

-- name: ListConnectionsForCredentialRewrap :many
SELECT id, workspace_id, token_payload, credential_key_version, installation_generation
FROM public.figma_connections
WHERE credential_key_version = sqlc.arg(credential_key_version)
  AND (
      CAST(sqlc.narg(after_id) AS uuid) IS NULL
      OR id > CAST(sqlc.narg(after_id) AS uuid)
  )
ORDER BY id
LIMIT sqlc.arg(page_limit);

-- name: CompareAndSwapRewrappedCredential :execrows
UPDATE public.figma_connections
SET token_payload = sqlc.arg(next_payload),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
  AND installation_generation = sqlc.arg(installation_generation)
  AND credential_key_version = sqlc.arg(credential_key_version)
  AND token_payload = sqlc.arg(previous_payload);
