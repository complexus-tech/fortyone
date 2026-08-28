-- name: LinkGitHubUser :exec
UPDATE public.users
SET github_user_id = sqlc.arg(github_user_id),
    github_username = sqlc.arg(github_username),
    github_access_token = sqlc.arg(credential_payload),
    github_access_token_envelope_version = sqlc.arg(envelope_version),
    github_access_token_generation = sqlc.arg(generation),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id);

-- name: UnlinkGitHubUser :exec
UPDATE public.users
SET github_user_id = NULL,
    github_username = NULL,
    github_access_token = NULL,
    github_access_token_envelope_version = 0,
    github_access_token_generation = NULL,
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id);

-- name: GetGitHubUserCredential :one
SELECT user_id,
       github_access_token AS credential_payload,
       CAST(github_access_token_envelope_version AS integer) AS envelope_version,
       github_access_token_generation AS generation
FROM public.users
WHERE user_id = sqlc.arg(user_id)
  AND NULLIF(github_access_token, '') IS NOT NULL
  AND github_access_token_generation IS NOT NULL;

-- name: ListGitHubUserCredentialsForRewrap :many
SELECT user_id,
       github_access_token AS credential_payload,
       CAST(github_access_token_envelope_version AS integer) AS envelope_version,
       github_access_token_generation AS generation
FROM public.users
WHERE github_access_token_envelope_version = sqlc.arg(envelope_version)
  AND github_access_token_generation IS NOT NULL
  AND NULLIF(github_access_token, '') IS NOT NULL
ORDER BY user_id
LIMIT sqlc.arg(page_limit);

-- name: ListGitHubUserCredentialsForRewrapAfter :many
SELECT user_id,
       github_access_token AS credential_payload,
       CAST(github_access_token_envelope_version AS integer) AS envelope_version,
       github_access_token_generation AS generation
FROM public.users
WHERE github_access_token_envelope_version = sqlc.arg(envelope_version)
  AND github_access_token_generation IS NOT NULL
  AND NULLIF(github_access_token, '') IS NOT NULL
  AND user_id > sqlc.arg(after_user_id)
ORDER BY user_id
LIMIT sqlc.arg(page_limit);

-- name: RewrapGitHubUserCredential :execrows
UPDATE public.users
SET github_access_token = sqlc.arg(rewrapped_payload),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND github_access_token_generation = sqlc.arg(generation)
  AND github_access_token_envelope_version = sqlc.arg(envelope_version)
  AND github_access_token = sqlc.arg(expected_payload);

-- name: ListLegacyGitHubUserCredentials :many
SELECT user_id,
       github_access_token AS credential_payload
FROM public.users
WHERE github_access_token_envelope_version = 0
  AND github_access_token_generation IS NULL
  AND NULLIF(github_access_token, '') IS NOT NULL
ORDER BY user_id
LIMIT sqlc.arg(page_limit);

-- name: UpgradeLegacyGitHubUserCredential :execrows
UPDATE public.users
SET github_access_token = sqlc.arg(encrypted_payload),
    github_access_token_envelope_version = sqlc.arg(envelope_version),
    github_access_token_generation = sqlc.arg(generation),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND github_access_token = sqlc.arg(expected_plaintext)
  AND github_access_token_envelope_version = 0
  AND github_access_token_generation IS NULL;

