-- name: GetActiveBrowserSessionVersion :one
SELECT account.auth_session_version
FROM public.users AS account
WHERE account.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE;
