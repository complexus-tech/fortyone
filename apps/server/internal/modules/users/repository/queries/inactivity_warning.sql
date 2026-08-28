-- name: ListUserInactivityWarningCandidates :many
SELECT
    account.user_id,
    CAST(BTRIM(account.email) AS text) AS email,
    CAST(COALESCE(account.full_name, '') AS text) AS full_name,
    account.last_login_at
FROM public.users AS account
WHERE account.last_login_at < sqlc.arg(inactive_before)
  AND account.is_active = TRUE
  AND account.is_system = FALSE
  AND NULLIF(BTRIM(account.email), '') IS NOT NULL
  AND account.inactivity_warning_sent_at IS NULL
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR account.last_login_at > sqlc.arg(after_last_login_at)
      OR (
          account.last_login_at = sqlc.arg(after_last_login_at)
          AND account.user_id > sqlc.arg(after_user_id)
      )
  )
ORDER BY account.last_login_at, account.user_id
LIMIT CAST(sqlc.arg(batch_size) AS integer);

-- name: GetEligibleUserInactivityWarningCandidate :one
SELECT
    account.user_id,
    CAST(BTRIM(account.email) AS text) AS email,
    CAST(COALESCE(account.full_name, '') AS text) AS full_name,
    account.last_login_at
FROM public.users AS account
WHERE account.user_id = sqlc.arg(user_id)
  AND account.last_login_at < sqlc.arg(inactive_before)
  AND account.is_active = TRUE
  AND account.is_system = FALSE
  AND NULLIF(BTRIM(account.email), '') IS NOT NULL
  AND account.inactivity_warning_sent_at IS NULL;

-- name: MarkUserInactivityWarningSent :execrows
UPDATE public.users AS account
SET inactivity_warning_sent_at = (
    CAST(sqlc.arg(warning_sent_at) AS timestamptz) AT TIME ZONE 'UTC'
)
WHERE account.user_id = sqlc.arg(user_id)
  AND account.last_login_at < sqlc.arg(inactive_before)
  AND account.is_active = TRUE
  AND account.is_system = FALSE
  AND account.inactivity_warning_sent_at IS NULL;
