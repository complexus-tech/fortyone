-- name: InsertServiceAccount :one
INSERT INTO public.principals (
    principal_id,
    workspace_id,
    kind,
    name,
    workspace_role,
    status,
    created_by_user_id,
    created_at,
    updated_at
)
SELECT
    sqlc.arg(principal_id),
    sqlc.arg(workspace_id),
    'service_account',
    sqlc.arg(name),
    CAST(sqlc.arg(workspace_role) AS public.user_role),
    'active',
    sqlc.arg(actor_user_id),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
FROM public.workspace_members AS membership
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(actor_user_id)
  AND membership.role = 'admin'
  AND account.is_active = TRUE
RETURNING
    principal_id,
    workspace_id,
    kind,
    name,
    CAST(workspace_role AS text) AS workspace_role,
    status,
    created_by_user_id,
    created_at,
    updated_at,
    disabled_at,
    disabled_by_user_id,
    disabled_reason;

-- name: ListServiceAccounts :many
SELECT
    principal.principal_id,
    principal.workspace_id,
    principal.kind,
    principal.name,
    CAST(principal.workspace_role AS text) AS workspace_role,
    principal.status,
    principal.created_by_user_id,
    principal.created_at,
    principal.updated_at,
    principal.disabled_at,
    principal.disabled_by_user_id,
    principal.disabled_reason
FROM public.principals AS principal
WHERE principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.kind = 'service_account'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = principal.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
ORDER BY principal.created_at DESC, principal.principal_id DESC;

-- name: DisableServiceAccount :one
UPDATE public.principals AS principal
SET
    status = 'disabled',
    updated_at = sqlc.arg(disabled_at),
    disabled_at = sqlc.arg(disabled_at),
    disabled_by_user_id = sqlc.arg(actor_user_id),
    disabled_reason = sqlc.arg(disabled_reason)
WHERE principal.principal_id = sqlc.arg(principal_id)
  AND principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.kind = 'service_account'
  AND principal.status = 'active'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = principal.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
RETURNING principal.principal_id;

-- name: RevokeServiceAccountCredentials :exec
UPDATE public.api_credentials
SET
    revoked_at = sqlc.arg(revoked_at),
    revoked_by_user_id = sqlc.arg(actor_user_id),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE principal_id = sqlc.arg(principal_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND kind = 'service_account_key'
  AND revoked_at IS NULL;

-- name: InsertServiceAccountKey :one
INSERT INTO public.api_credentials (
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    secret_digest,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    created_by_user_id,
    created_at
)
SELECT
    sqlc.arg(credential_id),
    principal.workspace_id,
    principal.principal_id,
    'service_account_key',
    sqlc.arg(name),
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(token_version),
    sqlc.arg(digest_key_id),
    sqlc.arg(digest_key_version),
    sqlc.arg(expires_at),
    sqlc.arg(actor_user_id),
    sqlc.arg(created_at)
FROM public.principals AS principal
WHERE principal.principal_id = sqlc.arg(principal_id)
  AND principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.kind = 'service_account'
  AND principal.status = 'active'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = principal.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
RETURNING
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    last_used_at,
    rotated_from_id,
    rotated_at,
    revoked_at,
    revoked_by_user_id,
    revoked_reason,
    created_by_user_id,
    created_at;

-- name: ListServiceAccountKeys :many
SELECT
    credential.credential_id,
    credential.workspace_id,
    credential.principal_id,
    credential.kind,
    credential.name,
    credential.lookup_prefix,
    credential.token_version,
    credential.expires_at,
    credential.last_used_at,
    credential.rotated_from_id,
    credential.rotated_at,
    credential.revoked_at,
    credential.revoked_by_user_id,
    credential.revoked_reason,
    credential.created_by_user_id,
    credential.created_at,
    CAST(ARRAY(
        SELECT scope.scope
        FROM public.api_credential_scopes AS scope
        WHERE scope.credential_id = credential.credential_id
        ORDER BY scope.scope
    ) AS text[]) AS scopes,
    CAST(ARRAY(
        SELECT restriction.team_id
        FROM public.api_credential_team_restrictions AS restriction
        WHERE restriction.credential_id = credential.credential_id
        ORDER BY restriction.team_id
    ) AS uuid[]) AS team_restrictions
FROM public.api_credentials AS credential
INNER JOIN public.principals AS principal
    ON principal.principal_id = credential.principal_id
    AND principal.workspace_id = credential.workspace_id
WHERE credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.principal_id = sqlc.arg(principal_id)
  AND credential.kind = 'service_account_key'
  AND principal.kind = 'service_account'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = principal.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
ORDER BY credential.created_at DESC, credential.credential_id DESC;

-- name: LockServiceAccountKeyForRotation :one
SELECT credential.credential_id
FROM public.api_credentials AS credential
INNER JOIN public.principals AS principal
    ON principal.principal_id = credential.principal_id
    AND principal.workspace_id = credential.workspace_id
WHERE credential.credential_id = sqlc.arg(credential_id)
  AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.principal_id = sqlc.arg(principal_id)
  AND credential.kind = 'service_account_key'
  AND credential.revoked_at IS NULL
  AND credential.expires_at > sqlc.arg(rotated_at)
  AND credential.rotated_at IS NULL
  AND principal.kind = 'service_account'
  AND principal.status = 'active'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = principal.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
FOR UPDATE OF credential;

-- name: InsertRotatedServiceAccountKey :one
INSERT INTO public.api_credentials (
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    secret_digest,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    rotated_from_id,
    created_by_user_id,
    created_at
)
SELECT
    sqlc.arg(new_credential_id),
    credential.workspace_id,
    credential.principal_id,
    credential.kind,
    credential.name,
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(token_version),
    sqlc.arg(digest_key_id),
    sqlc.arg(digest_key_version),
    sqlc.arg(expires_at),
    credential.credential_id,
    sqlc.arg(actor_user_id),
    sqlc.arg(created_at)
FROM public.api_credentials AS credential
WHERE credential.credential_id = sqlc.arg(old_credential_id)
  AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.principal_id = sqlc.arg(principal_id)
  AND credential.kind = 'service_account_key'
  AND credential.revoked_at IS NULL
  AND credential.rotated_at IS NULL
RETURNING
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    last_used_at,
    rotated_from_id,
    rotated_at,
    revoked_at,
    revoked_by_user_id,
    revoked_reason,
    created_by_user_id,
    created_at;

-- name: RevokeServiceAccountKey :one
UPDATE public.api_credentials AS credential
SET
    revoked_at = sqlc.arg(revoked_at),
    revoked_by_user_id = sqlc.arg(actor_user_id),
    revoked_reason = sqlc.arg(revoked_reason)
FROM public.principals AS principal
WHERE credential.principal_id = principal.principal_id
  AND credential.workspace_id = principal.workspace_id
  AND credential.credential_id = sqlc.arg(credential_id)
  AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.principal_id = sqlc.arg(principal_id)
  AND credential.kind = 'service_account_key'
  AND credential.revoked_at IS NULL
  AND principal.kind = 'service_account'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = principal.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
RETURNING credential.credential_id;
