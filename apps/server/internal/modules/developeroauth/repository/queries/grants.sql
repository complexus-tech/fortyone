-- name: UpsertOAuthUserGrant :one
INSERT INTO public.oauth_grants (
    grant_id,
    application_id,
    user_id,
    actor_kind,
    resource,
    status,
    created_at,
    updated_at
) SELECT
    sqlc.arg(grant_id),
    sqlc.arg(application_id),
    sqlc.arg(user_id),
    'oauth_user',
    sqlc.arg(resource),
    'active',
    sqlc.arg(granted_at),
    sqlc.arg(granted_at)
FROM public.oauth_applications AS application
INNER JOIN public.users AS account
    ON account.user_id = sqlc.arg(user_id)
WHERE application.application_id = sqlc.arg(application_id)
  AND application.client_id = sqlc.arg(client_id)
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(granted_at)
  AND account.is_active = TRUE
ON CONFLICT (application_id, user_id, resource) DO UPDATE
SET status = 'active',
    updated_at = EXCLUDED.updated_at,
    revoked_at = NULL,
    revoked_reason = NULL
RETURNING grant_id, application_id, user_id, actor_kind, resource, status,
    created_at, updated_at, last_used_at, revoked_at, revoked_reason;

-- name: DeleteOAuthGrantScopes :exec
DELETE FROM public.oauth_grant_scopes
WHERE grant_id = sqlc.arg(grant_id);

-- name: CreateOAuthGrantScope :exec
INSERT INTO public.oauth_grant_scopes (grant_id, scope)
VALUES (sqlc.arg(grant_id), sqlc.arg(scope));

-- name: GetActiveOAuthGrant :one
SELECT
    oauth_grant.grant_id,
    oauth_grant.application_id,
    application.client_id,
    oauth_grant.user_id,
    oauth_grant.actor_kind,
    oauth_grant.resource,
    oauth_grant.status,
    oauth_grant.created_at,
    oauth_grant.updated_at,
    oauth_grant.last_used_at,
    oauth_grant.revoked_at,
    oauth_grant.revoked_reason,
    CAST(ARRAY(
        SELECT granted_scope.scope
        FROM public.oauth_grant_scopes AS granted_scope
        WHERE granted_scope.grant_id = oauth_grant.grant_id
        ORDER BY granted_scope.scope
    ) AS text[]) AS scopes
FROM public.oauth_grants AS oauth_grant
INNER JOIN public.oauth_applications AS application
    ON application.application_id = oauth_grant.application_id
INNER JOIN public.users AS account
    ON account.user_id = oauth_grant.user_id
WHERE oauth_grant.grant_id = sqlc.arg(grant_id)
  AND oauth_grant.application_id = sqlc.arg(application_id)
  AND oauth_grant.resource = sqlc.arg(resource)
  AND oauth_grant.status = 'active'
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND account.is_active = TRUE
  AND EXISTS (
      SELECT 1
      FROM public.oauth_grant_scopes AS required_scope
      WHERE required_scope.grant_id = oauth_grant.grant_id
  );

-- name: TouchOAuthGrant :exec
UPDATE public.oauth_grants
SET last_used_at = sqlc.arg(used_at),
    updated_at = GREATEST(updated_at, sqlc.arg(used_at))
WHERE grant_id = sqlc.arg(grant_id)
  AND status = 'active'
  AND (last_used_at IS NULL OR last_used_at < sqlc.arg(touch_before));

-- name: RevokeOAuthGrant :execrows
UPDATE public.oauth_grants
SET status = 'revoked',
    revoked_at = sqlc.arg(revoked_at),
    revoked_reason = sqlc.arg(revoked_reason),
    updated_at = sqlc.arg(revoked_at)
WHERE grant_id = sqlc.arg(grant_id)
  AND status = 'active';
