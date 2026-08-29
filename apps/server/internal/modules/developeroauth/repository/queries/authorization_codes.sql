-- name: CreateOAuthAuthorizationCode :exec
INSERT INTO public.oauth_authorization_codes (
    authorization_code_id,
    application_id,
    grant_id,
    lookup_prefix,
    secret_digest,
    digest_key_id,
    redirect_uri,
    resource,
    code_challenge,
    created_at,
    expires_at
) VALUES (
    sqlc.arg(authorization_code_id),
    sqlc.arg(application_id),
    sqlc.arg(grant_id),
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(digest_key_id),
    sqlc.arg(redirect_uri),
    sqlc.arg(resource),
    sqlc.arg(code_challenge),
    sqlc.arg(created_at),
    sqlc.arg(expires_at)
);

-- name: InvalidateUnconsumedOAuthAuthorizationCodesForSubject :exec
UPDATE public.oauth_authorization_codes AS code
SET consumed_at = sqlc.arg(invalidated_at)
FROM public.oauth_grants AS oauth_grant
WHERE code.grant_id = oauth_grant.grant_id
  AND code.application_id = sqlc.arg(application_id)
  AND code.resource = sqlc.arg(resource)
  AND oauth_grant.application_id = code.application_id
  AND oauth_grant.resource = code.resource
  AND oauth_grant.user_id = sqlc.arg(user_id)
  AND code.consumed_at IS NULL;

-- name: GetOAuthAuthorizationCodeForUpdate :one
SELECT
    code.authorization_code_id,
    code.application_id,
    application.client_id,
    code.grant_id,
    oauth_grant.user_id,
    oauth_grant.actor_kind,
    code.lookup_prefix,
    code.secret_digest,
    code.digest_key_id,
    code.redirect_uri,
    code.resource,
    code.code_challenge,
    code.created_at,
    code.expires_at,
    code.consumed_at,
    CAST(ARRAY(
        SELECT granted_scope.scope
        FROM public.oauth_grant_scopes AS granted_scope
        WHERE granted_scope.grant_id = oauth_grant.grant_id
        ORDER BY granted_scope.scope
    ) AS text[]) AS scopes
FROM public.oauth_authorization_codes AS code
INNER JOIN public.oauth_applications AS application
    ON application.application_id = code.application_id
INNER JOIN public.oauth_grants AS oauth_grant
    ON oauth_grant.grant_id = code.grant_id
    AND oauth_grant.application_id = code.application_id
    AND oauth_grant.resource = code.resource
INNER JOIN public.users AS account
    ON account.user_id = oauth_grant.user_id
WHERE code.lookup_prefix = sqlc.arg(lookup_prefix)
  AND code.expires_at > sqlc.arg(active_at)
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at)
  AND oauth_grant.status = 'active'
  AND account.is_active = TRUE
FOR UPDATE OF code;

-- name: ConsumeOAuthAuthorizationCode :one
UPDATE public.oauth_authorization_codes
SET consumed_at = sqlc.arg(consumed_at)
WHERE authorization_code_id = sqlc.arg(authorization_code_id)
  AND consumed_at IS NULL
  AND expires_at > sqlc.arg(consumed_at)
RETURNING authorization_code_id;
