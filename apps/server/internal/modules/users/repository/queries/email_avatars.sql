-- name: EnsureEmailAvatarHandle :one
INSERT INTO email_avatar_handles (user_id)
SELECT account.user_id FROM users AS account
WHERE account.user_id = sqlc.arg(user_id)
    AND account.is_active AND NOT account.is_system
    AND NULLIF(TRIM(account.avatar_url), '') IS NOT NULL
ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING handle;

-- name: GetEmailAvatar :one
SELECT account.avatar_url
FROM email_avatar_handles AS image
JOIN users AS account ON account.user_id = image.user_id
WHERE image.handle = sqlc.arg(handle)
    AND account.is_active AND NOT account.is_system
    AND NULLIF(TRIM(account.avatar_url), '') IS NOT NULL;
