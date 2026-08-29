-- name: IsOAuthApplicationWorkspaceAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS account
        ON account.user_id = membership.user_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_user_id)
      AND membership.role = 'admin'
      AND account.is_active = TRUE
);

-- name: CreateManagedOAuthApplication :one
INSERT INTO public.oauth_applications (
    application_id,
    client_id,
    name,
    registration_kind,
    status,
    owner_workspace_id,
    owner_user_id,
    expires_at,
    created_at,
    updated_at
)
SELECT
    sqlc.arg(application_id),
    sqlc.arg(client_id),
    sqlc.arg(name),
    'confidential',
    'active',
    sqlc.arg(owner_workspace_id),
    sqlc.arg(owner_user_id),
    sqlc.arg(expires_at),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
FROM public.workspace_members AS membership
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE membership.workspace_id = sqlc.arg(owner_workspace_id)
  AND membership.user_id = sqlc.arg(owner_user_id)
  AND membership.role = 'admin'
  AND account.is_active = TRUE
RETURNING
    application_id,
    client_id,
    name,
    registration_kind,
    status,
    owner_workspace_id,
    owner_user_id,
    expires_at,
    created_at,
    updated_at,
    revoked_at;

-- name: InsertOAuthClientSecret :one
INSERT INTO public.oauth_client_secrets (
    secret_id,
    application_id,
    lookup_prefix,
    secret_digest,
    digest_key_id,
    expires_at,
    rotated_from_id,
    created_by_user_id,
    created_at
) VALUES (
    sqlc.arg(secret_id),
    sqlc.arg(application_id),
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(digest_key_id),
    sqlc.arg(expires_at),
    sqlc.narg(rotated_from_id),
    sqlc.arg(created_by_user_id),
    sqlc.arg(created_at)
)
RETURNING
    secret_id,
    application_id,
    lookup_prefix,
    expires_at,
    last_used_at,
    rotated_from_id,
    overlap_expires_at,
    revoked_at,
    created_at;

-- name: ListManagedOAuthApplications :many
SELECT
    application.application_id,
    application.client_id,
    application.name,
    application.registration_kind,
    application.status,
    application.owner_workspace_id,
    application.owner_user_id,
    application.expires_at,
    application.created_at,
    application.updated_at,
    application.revoked_at,
    CAST(ARRAY(
        SELECT redirect.redirect_uri
        FROM public.oauth_application_redirect_uris AS redirect
        WHERE redirect.application_id = application.application_id
        ORDER BY redirect.redirect_uri
    ) AS text[]) AS redirect_uris
FROM public.oauth_applications AS application
WHERE application.owner_workspace_id = sqlc.arg(owner_workspace_id)
  AND application.registration_kind = 'confidential'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = application.owner_workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
ORDER BY application.created_at DESC, application.application_id DESC;

-- name: ListOAuthClientSecrets :many
SELECT
    secret.secret_id,
    secret.application_id,
    secret.lookup_prefix,
    secret.expires_at,
    secret.last_used_at,
    secret.rotated_from_id,
    secret.overlap_expires_at,
    secret.revoked_at,
    secret.created_at
FROM public.oauth_client_secrets AS secret
INNER JOIN public.oauth_applications AS application
    ON application.application_id = secret.application_id
WHERE secret.application_id = sqlc.arg(application_id)
  AND application.owner_workspace_id = sqlc.arg(owner_workspace_id)
  AND application.registration_kind = 'confidential'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = application.owner_workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
ORDER BY secret.created_at DESC, secret.secret_id DESC;

-- name: LockManagedOAuthApplication :one
SELECT
    application.application_id,
    application.client_id,
    application.name,
    application.expires_at
FROM public.oauth_applications AS application
WHERE application.application_id = sqlc.arg(application_id)
  AND application.owner_workspace_id = sqlc.arg(owner_workspace_id)
  AND application.registration_kind = 'confidential'
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = application.owner_workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
FOR UPDATE OF application;

-- name: GetOAuthClientSecretRotationHeadForUpdate :one
SELECT
    secret_id,
    application_id,
    lookup_prefix,
    expires_at,
    last_used_at,
    rotated_from_id,
    overlap_expires_at,
    revoked_at,
    created_at
FROM public.oauth_client_secrets
WHERE application_id = sqlc.arg(application_id)
  AND revoked_at IS NULL
  AND overlap_expires_at IS NULL
ORDER BY created_at DESC, secret_id DESC
LIMIT 1
FOR UPDATE;

-- name: SetOAuthClientSecretOverlap :one
UPDATE public.oauth_client_secrets
SET overlap_expires_at = sqlc.arg(overlap_expires_at)
WHERE secret_id = sqlc.arg(secret_id)
  AND application_id = sqlc.arg(application_id)
  AND revoked_at IS NULL
  AND overlap_expires_at IS NULL
RETURNING secret_id;

-- name: GetManagedOAuthClientSecretForUpdate :one
SELECT
    secret.secret_id,
    secret.application_id,
    secret.lookup_prefix,
    secret.expires_at,
    secret.last_used_at,
    secret.rotated_from_id,
    secret.overlap_expires_at,
    secret.revoked_at,
    secret.created_at
FROM public.oauth_client_secrets AS secret
INNER JOIN public.oauth_applications AS application
    ON application.application_id = secret.application_id
WHERE secret.secret_id = sqlc.arg(secret_id)
  AND secret.application_id = sqlc.arg(application_id)
  AND application.owner_workspace_id = sqlc.arg(owner_workspace_id)
  AND application.registration_kind = 'confidential'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = application.owner_workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
FOR UPDATE OF secret;

-- name: SetOAuthClientSecretRevoked :one
UPDATE public.oauth_client_secrets
SET
    revoked_at = sqlc.arg(revoked_at),
    revoked_by_user_id = sqlc.arg(actor_user_id),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE secret_id = sqlc.arg(secret_id)
  AND application_id = sqlc.arg(application_id)
  AND revoked_at IS NULL
RETURNING secret_id;

-- name: LockManagedOAuthApplicationForInstallation :one
SELECT
    application.application_id,
    application.client_id,
    application.name,
    application.registration_kind,
    application.status,
    application.owner_workspace_id,
    application.owner_user_id,
    application.expires_at,
    application.created_at,
    application.updated_at,
    application.revoked_at
FROM public.oauth_applications AS application
WHERE application.client_id = sqlc.arg(client_id)
  AND application.registration_kind = 'confidential'
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND application.owner_workspace_id IS NOT NULL
  AND application.owner_user_id IS NOT NULL
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = sqlc.arg(workspace_id)
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
FOR KEY SHARE OF application;

-- name: InsertOAuthApplicationPrincipal :one
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
    'oauth_application',
    sqlc.arg(name),
    'member',
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
RETURNING principal_id;

-- name: InsertOAuthApplicationInstallation :one
INSERT INTO public.oauth_application_installations (
    installation_id,
    application_id,
    workspace_id,
    principal_id,
    resource,
    status,
    installed_by_user_id,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(installation_id),
    sqlc.arg(application_id),
    sqlc.arg(workspace_id),
    sqlc.arg(principal_id),
    sqlc.arg(resource),
    'active',
    sqlc.arg(installed_by_user_id),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
)
RETURNING
    installation_id,
    application_id,
    workspace_id,
    principal_id,
    resource,
    status,
    installed_by_user_id,
    created_at,
    updated_at,
    last_used_at,
    revoked_at,
    revoked_by_user_id,
    revoked_reason;

-- name: InsertOAuthApplicationInstallationScopes :exec
INSERT INTO public.oauth_application_installation_scopes (
    installation_id,
    scope
)
SELECT
    sqlc.arg(installation_id),
    requested.scope
FROM unnest(CAST(sqlc.arg(scopes) AS text[])) AS requested(scope);

-- name: ListOAuthApplicationInstallations :many
SELECT
    installation.installation_id,
    installation.application_id,
    application.client_id,
    installation.workspace_id,
    installation.principal_id,
    installation.resource,
    installation.status,
    installation.installed_by_user_id,
    installation.created_at,
    installation.updated_at,
    installation.last_used_at,
    installation.revoked_at,
    installation.revoked_by_user_id,
    installation.revoked_reason,
    CAST(ARRAY(
        SELECT installation_scope.scope
        FROM public.oauth_application_installation_scopes AS installation_scope
        WHERE installation_scope.installation_id = installation.installation_id
        ORDER BY installation_scope.scope
    ) AS text[]) AS scopes
FROM public.oauth_application_installations AS installation
INNER JOIN public.oauth_applications AS application
    ON application.application_id = installation.application_id
WHERE installation.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = installation.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
ORDER BY installation.created_at DESC, installation.installation_id DESC;

-- name: GetOAuthApplicationInstallationForUpdate :one
SELECT
    installation.installation_id,
    installation.application_id,
    application.client_id,
    installation.workspace_id,
    installation.principal_id,
    installation.resource,
    installation.status,
    installation.installed_by_user_id,
    installation.created_at,
    installation.updated_at,
    installation.last_used_at,
    installation.revoked_at,
    installation.revoked_by_user_id,
    installation.revoked_reason,
    CAST(ARRAY(
        SELECT installation_scope.scope
        FROM public.oauth_application_installation_scopes AS installation_scope
        WHERE installation_scope.installation_id = installation.installation_id
        ORDER BY installation_scope.scope
    ) AS text[]) AS scopes
FROM public.oauth_application_installations AS installation
INNER JOIN public.oauth_applications AS application
    ON application.application_id = installation.application_id
INNER JOIN public.principals AS principal
    ON principal.principal_id = installation.principal_id
   AND principal.workspace_id = installation.workspace_id
WHERE installation.installation_id = sqlc.arg(installation_id)
  AND installation.workspace_id = sqlc.arg(workspace_id)
  AND principal.kind = 'oauth_application'
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account
          ON account.user_id = membership.user_id
      WHERE membership.workspace_id = installation.workspace_id
        AND membership.user_id = sqlc.arg(actor_user_id)
        AND membership.role = 'admin'
        AND account.is_active = TRUE
  )
FOR UPDATE OF installation, principal;

-- name: DeleteOAuthApplicationInstallationScopes :exec
DELETE FROM public.oauth_application_installation_scopes
WHERE installation_id = sqlc.arg(installation_id);

-- name: UpdateOAuthApplicationInstallation :one
UPDATE public.oauth_application_installations
SET
    resource = sqlc.arg(resource),
    updated_at = sqlc.arg(updated_at)
WHERE installation_id = sqlc.arg(installation_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status = 'active'
RETURNING installation_id;

-- name: RevokeOAuthApplicationInstallation :one
UPDATE public.oauth_application_installations
SET
    status = 'revoked',
    updated_at = sqlc.arg(revoked_at),
    revoked_at = sqlc.arg(revoked_at),
    revoked_by_user_id = sqlc.arg(actor_user_id),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE installation_id = sqlc.arg(installation_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status = 'active'
RETURNING principal_id;

-- name: DisableOAuthApplicationPrincipal :one
UPDATE public.principals
SET
    status = 'disabled',
    updated_at = sqlc.arg(disabled_at),
    disabled_at = sqlc.arg(disabled_at),
    disabled_by_user_id = sqlc.arg(actor_user_id),
    disabled_reason = sqlc.arg(disabled_reason)
WHERE principal_id = sqlc.arg(principal_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND kind = 'oauth_application'
  AND status = 'active'
RETURNING principal_id;
