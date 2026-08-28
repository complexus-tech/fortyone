-- name: ResolveUserByGitHubID :one
SELECT user_id
FROM public.users
WHERE github_user_id = sqlc.arg(github_user_id)
LIMIT 1;

-- name: ResolveFortyOneUsersByGitHubIDs :many
SELECT github_user_id,
       username,
       full_name,
       avatar_url
FROM public.users
WHERE github_user_id = ANY(CAST(sqlc.arg(github_user_ids) AS bigint[]))
  AND is_active = TRUE;

-- name: ResolveFortyOneUserByFullName :one
SELECT username,
       full_name,
       avatar_url
FROM public.users
WHERE full_name = sqlc.arg(full_name)
  AND is_active = TRUE
ORDER BY user_id
LIMIT 1;

-- name: ResolveFortyOneUserByEmail :one
SELECT username,
       full_name,
       avatar_url
FROM public.users
WHERE email = sqlc.arg(email)
  AND is_system = TRUE
ORDER BY user_id
LIMIT 1;
