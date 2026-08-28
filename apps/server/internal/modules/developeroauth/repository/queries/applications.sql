-- name: CreateOAuthApplication :one
INSERT INTO public.oauth_applications (
    application_id,
    client_id,
    name,
    registration_kind,
    status,
    expires_at,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(application_id),
    sqlc.arg(client_id),
    sqlc.arg(name),
    sqlc.arg(registration_kind),
    'active',
    sqlc.arg(expires_at),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
)
RETURNING application_id, client_id, name, registration_kind, status,
    expires_at, created_at, updated_at, revoked_at;

-- name: CreateOAuthApplicationRedirectURI :exec
INSERT INTO public.oauth_application_redirect_uris (
    application_id,
    redirect_uri,
    created_at
) VALUES (
    sqlc.arg(application_id),
    sqlc.arg(redirect_uri),
    sqlc.arg(created_at)
);

-- name: GetActiveOAuthApplicationByClientID :one
SELECT
    application.application_id,
    application.client_id,
    application.name,
    application.registration_kind,
    application.status,
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
WHERE application.client_id = sqlc.arg(client_id)
  AND application.status = 'active'
  AND application.expires_at > sqlc.arg(active_at);
