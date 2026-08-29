-- name: GetOAuthApplicationCredentialForUpdate :one
SELECT
    secret.secret_id,
    secret.application_id,
    secret.lookup_prefix,
    secret.secret_digest,
    secret.digest_key_id,
    secret.expires_at AS secret_expires_at,
    secret.last_used_at AS secret_last_used_at,
    secret.rotated_from_id,
    secret.overlap_expires_at,
    secret.revoked_at AS secret_revoked_at,
    secret.created_at AS secret_created_at,
    application.client_id,
    application.name AS application_name,
    application.registration_kind,
    application.expires_at AS application_expires_at,
    application.created_at AS application_created_at,
    CAST(ARRAY(
        SELECT redirect.redirect_uri
        FROM public.oauth_application_redirect_uris AS redirect
        WHERE redirect.application_id = application.application_id
        ORDER BY redirect.redirect_uri
    ) AS text[]) AS redirect_uris,
    installation.installation_id,
    installation.workspace_id,
    installation.principal_id,
    installation.resource,
    installation.status AS installation_status,
    installation.installed_by_user_id,
    installation.created_at AS installation_created_at,
    installation.updated_at AS installation_updated_at,
    installation.last_used_at AS installation_last_used_at,
    installation.revoked_at AS installation_revoked_at,
    installation.revoked_by_user_id,
    installation.revoked_reason,
    CAST(ARRAY(
        SELECT installation_scope.scope
        FROM public.oauth_application_installation_scopes AS installation_scope
        WHERE installation_scope.installation_id = installation.installation_id
        ORDER BY installation_scope.scope
    ) AS text[]) AS installation_scopes
FROM public.oauth_client_secrets AS secret
INNER JOIN public.oauth_applications AS application
    ON application.application_id = secret.application_id
INNER JOIN public.oauth_application_installations AS installation
    ON installation.application_id = application.application_id
INNER JOIN public.principals AS principal
    ON principal.principal_id = installation.principal_id
   AND principal.workspace_id = installation.workspace_id
WHERE secret.lookup_prefix = sqlc.arg(lookup_prefix)
  AND installation.installation_id = sqlc.arg(installation_id)
  AND secret.revoked_at IS NULL
  AND secret.expires_at > sqlc.arg(active_at)
  AND (
      secret.overlap_expires_at IS NULL
      OR secret.overlap_expires_at > sqlc.arg(active_at)
  )
  AND application.registration_kind = 'confidential'
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND application.owner_workspace_id IS NOT NULL
  AND application.owner_user_id IS NOT NULL
  AND installation.status = 'active'
  AND principal.kind = 'oauth_application'
  AND principal.status = 'active'
FOR UPDATE OF secret, installation;

-- name: TouchOAuthClientSecret :one
UPDATE public.oauth_client_secrets
SET last_used_at = GREATEST(
    COALESCE(last_used_at, sqlc.arg(used_at)),
    sqlc.arg(used_at)
)
WHERE secret_id = sqlc.arg(secret_id)
  AND revoked_at IS NULL
RETURNING secret_id;

-- name: TouchOAuthApplicationInstallationAuthentication :one
UPDATE public.oauth_application_installations
SET
    last_used_at = GREATEST(
        COALESCE(last_used_at, sqlc.arg(used_at)),
        sqlc.arg(used_at)
    ),
    updated_at = GREATEST(updated_at, sqlc.arg(used_at))
WHERE installation_id = sqlc.arg(installation_id)
  AND status = 'active'
RETURNING installation_id;

-- name: GetActiveOAuthApplicationInstallation :one
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
  AND installation.application_id = sqlc.arg(application_id)
  AND installation.resource = sqlc.arg(resource)
  AND installation.status = 'active'
  AND application.registration_kind = 'confidential'
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND application.owner_workspace_id IS NOT NULL
  AND application.owner_user_id IS NOT NULL
  AND principal.kind = 'oauth_application'
  AND principal.status = 'active';

-- name: TouchActiveOAuthApplicationInstallation :one
UPDATE public.oauth_application_installations
SET
    last_used_at = CASE
        WHEN last_used_at IS NULL OR last_used_at <= sqlc.arg(touch_before)
        THEN sqlc.arg(used_at)
        ELSE last_used_at
    END,
    updated_at = CASE
        WHEN last_used_at IS NULL OR last_used_at <= sqlc.arg(touch_before)
        THEN GREATEST(updated_at, sqlc.arg(used_at))
        ELSE updated_at
    END
WHERE installation_id = sqlc.arg(installation_id)
  AND status = 'active'
RETURNING installation_id;
