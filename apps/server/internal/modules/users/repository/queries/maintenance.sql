-- name: PurgeExpiredVerificationTokens :execrows
WITH candidates AS MATERIALIZED (
    SELECT verification_token.id
    FROM public.verification_tokens AS verification_token
    WHERE verification_token.expires_at < sqlc.arg(retained_before)
    ORDER BY verification_token.expires_at, verification_token.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF verification_token SKIP LOCKED
)
DELETE FROM public.verification_tokens AS verification_token
USING candidates
WHERE verification_token.id = candidates.id;

-- name: DeactivateInactiveUsers :execrows
WITH candidates AS MATERIALIZED (
    SELECT account.user_id
    FROM public.users AS account
    WHERE account.last_login_at < sqlc.arg(inactive_before)
      AND account.inactivity_warning_sent_at IS NOT NULL
      AND account.inactivity_warning_sent_at <= (
          CAST(sqlc.arg(warning_sent_before) AS timestamptz) AT TIME ZONE 'UTC'
      )
      AND account.is_active = TRUE
      AND account.is_system = FALSE
    ORDER BY account.inactivity_warning_sent_at, account.last_login_at, account.user_id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF account SKIP LOCKED
)
UPDATE public.users AS account
SET
    is_active = FALSE,
    login_reactivation_policy = 'verified_sign_in',
    auth_session_version = account.auth_session_version + 1,
    updated_at = sqlc.arg(deactivated_at)
FROM candidates
WHERE account.user_id = candidates.user_id
  AND account.last_login_at < sqlc.arg(inactive_before)
  AND account.inactivity_warning_sent_at IS NOT NULL
  AND account.inactivity_warning_sent_at <= (
      CAST(sqlc.arg(warning_sent_before) AS timestamptz) AT TIME ZONE 'UTC'
  )
  AND account.is_active = TRUE
  AND account.is_system = FALSE;
