-- name: CreateOAuthRefreshTokenFamily :exec
INSERT INTO public.oauth_refresh_token_families (
    family_id,
    grant_id,
    resource,
    created_at,
    expires_at
) VALUES (
    sqlc.arg(family_id),
    sqlc.arg(grant_id),
    sqlc.arg(resource),
    sqlc.arg(created_at),
    sqlc.arg(expires_at)
);

-- name: CreateOAuthRefreshToken :exec
INSERT INTO public.oauth_refresh_tokens (
    refresh_token_id,
    family_id,
    parent_token_id,
    lookup_prefix,
    secret_digest,
    digest_key_id,
    created_at,
    expires_at
) VALUES (
    sqlc.arg(refresh_token_id),
    sqlc.arg(family_id),
    sqlc.narg(parent_token_id),
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(digest_key_id),
    sqlc.arg(created_at),
    sqlc.arg(expires_at)
);

-- name: GetOAuthRefreshTokenForUpdate :one
SELECT
    token.refresh_token_id,
    token.family_id,
    token.parent_token_id,
    token.lookup_prefix,
    token.secret_digest,
    token.digest_key_id,
    token.created_at,
    token.expires_at,
    token.consumed_at,
    family.grant_id,
    family.resource,
    family.expires_at AS family_expires_at,
    family.revoked_at AS family_revoked_at,
    oauth_grant.application_id,
    application.client_id,
    oauth_grant.user_id,
    oauth_grant.actor_kind,
    CAST(ARRAY(
        SELECT granted_scope.scope
        FROM public.oauth_grant_scopes AS granted_scope
        WHERE granted_scope.grant_id = oauth_grant.grant_id
        ORDER BY granted_scope.scope
    ) AS text[]) AS scopes
FROM public.oauth_refresh_tokens AS token
INNER JOIN public.oauth_refresh_token_families AS family
    ON family.family_id = token.family_id
INNER JOIN public.oauth_grants AS oauth_grant
    ON oauth_grant.grant_id = family.grant_id
    AND oauth_grant.resource = family.resource
INNER JOIN public.oauth_applications AS application
    ON application.application_id = oauth_grant.application_id
INNER JOIN public.users AS account
    ON account.user_id = oauth_grant.user_id
WHERE token.lookup_prefix = sqlc.arg(lookup_prefix)
  AND token.expires_at > sqlc.arg(active_at)
  AND family.expires_at > sqlc.arg(active_at)
  AND family.revoked_at IS NULL
  AND oauth_grant.status = 'active'
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND account.is_active = TRUE
FOR UPDATE OF token, family, oauth_grant;

-- name: RevokeActiveOAuthRefreshTokenFamiliesForGrant :exec
UPDATE public.oauth_refresh_token_families
SET revoked_at = sqlc.arg(revoked_at),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE grant_id = sqlc.arg(grant_id)
  AND revoked_at IS NULL;

-- name: ConsumeOAuthRefreshToken :one
UPDATE public.oauth_refresh_tokens
SET consumed_at = sqlc.arg(consumed_at)
WHERE refresh_token_id = sqlc.arg(refresh_token_id)
  AND consumed_at IS NULL
  AND expires_at > sqlc.arg(consumed_at)
RETURNING refresh_token_id;

-- name: RevokeOAuthRefreshTokenFamily :execrows
UPDATE public.oauth_refresh_token_families
SET revoked_at = sqlc.arg(revoked_at),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE family_id = sqlc.arg(family_id)
  AND revoked_at IS NULL;
