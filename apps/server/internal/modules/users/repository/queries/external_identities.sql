-- name: AcquireExternalIdentityLock :exec
SELECT pg_advisory_xact_lock(hashtextextended(CAST(sqlc.arg(lock_identity) AS text), 0));

-- name: GetUserByExternalIdentity :one
SELECT
    account.user_id,
    account.username,
    account.email,
    account.full_name,
    account.avatar_url,
    account.is_active,
    account.is_system,
    account.is_internal,
    account.has_seen_walkthrough,
    account.timezone,
    account.working_days,
    account.working_start_minute,
    account.working_end_minute,
    account.last_login_at,
    account.last_used_workspace_id,
    account.github_username,
    account.created_at,
    account.updated_at
FROM public.user_external_identities AS identity
INNER JOIN public.users AS account ON account.user_id = identity.user_id
WHERE identity.provider = CAST(sqlc.arg(provider) AS text)
  AND identity.issuer = CAST(sqlc.arg(issuer) AS text)
  AND identity.subject = CAST(sqlc.arg(subject) AS text);

-- name: TouchExternalIdentity :execrows
UPDATE public.user_external_identities
SET
    email_at_link = CAST(sqlc.arg(email) AS text),
    last_authenticated_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE provider = CAST(sqlc.arg(provider) AS text)
  AND issuer = CAST(sqlc.arg(issuer) AS text)
  AND subject = CAST(sqlc.arg(subject) AS text);

-- name: CreateExternalIdentityUser :one
INSERT INTO public.users (
    username,
    email,
    full_name,
    avatar_url,
    timezone,
    last_login_at
)
VALUES (
    CAST(sqlc.arg(username) AS text),
    CAST(sqlc.arg(email) AS text),
    CAST(sqlc.arg(full_name) AS text),
    CAST(sqlc.arg(avatar_url) AS text),
    CAST(sqlc.arg(timezone) AS text),
    CURRENT_TIMESTAMP
)
RETURNING
    user_id,
    username,
    email,
    full_name,
    avatar_url,
    is_active,
    is_system,
    is_internal,
    has_seen_walkthrough,
    timezone,
    working_days,
    working_start_minute,
    working_end_minute,
    last_login_at,
    last_used_workspace_id,
    github_username,
    created_at,
    updated_at;

-- name: LinkExternalIdentity :exec
INSERT INTO public.user_external_identities (
    user_id,
    provider,
    issuer,
    subject,
    email_at_link,
    last_authenticated_at
)
VALUES (
    sqlc.arg(user_id),
    CAST(sqlc.arg(provider) AS text),
    CAST(sqlc.arg(issuer) AS text),
    CAST(sqlc.arg(subject) AS text),
    CAST(sqlc.arg(email) AS text),
    CURRENT_TIMESTAMP
);
