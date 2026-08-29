-- name: FindActiveSystemActorByEmail :one
SELECT user_id
FROM users
WHERE email = sqlc.arg(email)
  AND is_system = TRUE
  AND is_active = TRUE;
