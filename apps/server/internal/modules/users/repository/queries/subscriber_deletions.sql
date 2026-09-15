-- name: ClaimAccountSubscriberDeletion :one
WITH candidate AS (
    SELECT id FROM public.account_subscriber_deletions
    WHERE next_attempt_at <= sqlc.arg(claimed_at)
      AND (lease_expires_at IS NULL OR lease_expires_at <= sqlc.arg(claimed_at))
    ORDER BY next_attempt_at, created_at, id
    LIMIT 1 FOR UPDATE SKIP LOCKED
)
UPDATE public.account_subscriber_deletions AS deletion
SET lease_token = sqlc.arg(lease_token), lease_expires_at = sqlc.arg(lease_expires_at),
    updated_at = sqlc.arg(claimed_at), attempt_count = deletion.attempt_count + 1
FROM candidate WHERE deletion.id = candidate.id
RETURNING deletion.id, deletion.email, deletion.attempt_count;

-- Revalidate after waiting for the email lifecycle lock; an expired worker
-- must never delete a newly recreated contact after another worker completed.
-- name: RenewAccountSubscriberDeletion :execrows
UPDATE public.account_subscriber_deletions SET lease_expires_at = sqlc.arg(lease_expires_at)
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token)
  AND lease_expires_at > sqlc.arg(checked_at);

-- name: CompleteAccountSubscriberDeletion :execrows
DELETE FROM public.account_subscriber_deletions
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token);

-- name: RetryAccountSubscriberDeletion :execrows
UPDATE public.account_subscriber_deletions
SET lease_token = NULL, lease_expires_at = NULL,
    next_attempt_at = sqlc.arg(next_attempt_at), updated_at = sqlc.arg(released_at)
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token);

-- Same key as the account deletion transaction; remote I/O uses a session lock.
-- name: AcquireAccountSubscriberLifecycle :exec
SELECT pg_advisory_lock(hashtextextended('subscriber:' || lower(btrim(CAST(sqlc.arg(email) AS text))), 0));

-- name: ReleaseAccountSubscriberLifecycle :one
SELECT pg_advisory_unlock(hashtextextended('subscriber:' || lower(btrim(CAST(sqlc.arg(email) AS text))), 0));

-- Never replay stale queued profile data after an account was erased. A new
-- account with the same email can subscribe after older cleanup is confirmed.
-- name: GetActiveAccountSubscriber :one
SELECT account.user_id, account.email, COALESCE(account.full_name, '') AS full_name,
    EXISTS (SELECT 1 FROM public.account_subscriber_deletions AS deletion
        WHERE lower(btrim(deletion.email)) = lower(btrim(account.email))) AS cleanup_pending
FROM public.users AS account
WHERE lower(btrim(account.email)) = lower(btrim(CAST(sqlc.arg(email) AS text)))
  AND account.is_active = TRUE AND account.is_system = FALSE;
