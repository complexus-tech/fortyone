-- name: ListReadyGoogleDriveRevocations :many
SELECT revocation_id, user_id, google_subject
FROM public.google_drive_revocation_outbox
WHERE (
        status = 'pending'
        AND available_at <= sqlc.arg(ready_at)
    ) OR (
        status = 'processing'
        AND lease_expires_at <= sqlc.arg(ready_at)
    )
ORDER BY available_at, created_at, revocation_id
LIMIT sqlc.arg(row_limit);

-- name: EnqueueGoogleDriveRevocation :one
INSERT INTO public.google_drive_revocation_outbox (
    source_account_id,
    user_id,
    google_subject,
    installation_generation,
    credential_payload,
    credential_key_version
) VALUES (
    sqlc.narg(source_account_id),
    sqlc.arg(user_id),
    sqlc.arg(google_subject),
    sqlc.arg(installation_generation),
    sqlc.arg(credential_payload),
    sqlc.arg(credential_key_version)
)
ON CONFLICT (google_subject, installation_generation)
DO UPDATE SET updated_at = google_drive_revocation_outbox.updated_at
RETURNING revocation_id, user_id, google_subject;

-- name: GetGoogleDriveRevocationIdentityForUpdate :one
SELECT user_id, google_subject
FROM public.google_drive_revocation_outbox
WHERE revocation_id = sqlc.arg(revocation_id)
FOR UPDATE;

-- name: SupersedeGoogleDriveRevocationIfSubjectActive :execrows
UPDATE public.google_drive_revocation_outbox AS revocation
SET status = 'superseded',
    credential_payload = NULL,
    claim_token = NULL,
    lease_expires_at = NULL,
    last_error = NULL,
    terminal_at = sqlc.arg(superseded_at),
    updated_at = sqlc.arg(superseded_at)
WHERE revocation.revocation_id = sqlc.arg(revocation_id)
  AND revocation.status IN ('pending', 'processing', 'failed')
  AND EXISTS (
      SELECT 1
      FROM public.google_drive_accounts AS account
      WHERE account.google_subject = revocation.google_subject
        AND account.revoked_at IS NULL
  );

-- name: ClaimGoogleDriveRevocation :one
UPDATE public.google_drive_revocation_outbox AS revocation
SET status = 'processing',
    attempt_count = revocation.attempt_count + 1,
    claim_token = sqlc.arg(claim_token),
    lease_expires_at = sqlc.arg(lease_expires_at),
    last_error = NULL,
    terminal_at = NULL,
    updated_at = sqlc.arg(claimed_at)
WHERE revocation.revocation_id = sqlc.arg(revocation_id)
  AND (
        (revocation.status = 'pending' AND revocation.available_at <= sqlc.arg(claimed_at))
        OR (revocation.status = 'processing' AND revocation.lease_expires_at <= sqlc.arg(claimed_at))
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.google_drive_accounts AS account
      WHERE account.google_subject = revocation.google_subject
        AND account.revoked_at IS NULL
  )
RETURNING revocation_id, source_account_id, user_id, google_subject,
    installation_generation, credential_payload, credential_key_version,
    attempt_count, claim_token, lease_expires_at, created_at, updated_at;

-- name: CompleteGoogleDriveRevocation :execrows
UPDATE public.google_drive_revocation_outbox
SET status = 'completed',
    credential_payload = NULL,
    claim_token = NULL,
    lease_expires_at = NULL,
    last_error = NULL,
    terminal_at = sqlc.arg(completed_at),
    updated_at = sqlc.arg(completed_at)
WHERE revocation_id = sqlc.arg(revocation_id)
  AND status = 'processing'
  AND claim_token = sqlc.arg(claim_token);

-- name: RetryGoogleDriveRevocation :execrows
UPDATE public.google_drive_revocation_outbox
SET status = CASE WHEN CAST(sqlc.arg(terminal) AS boolean) THEN 'failed' ELSE 'pending' END,
    claim_token = NULL,
    lease_expires_at = NULL,
    last_error = LEFT(sqlc.arg(last_error), 2000),
    available_at = sqlc.arg(available_at),
    terminal_at = CASE WHEN CAST(sqlc.arg(terminal) AS boolean) THEN sqlc.arg(released_at) ELSE NULL END,
    updated_at = sqlc.arg(released_at)
WHERE revocation_id = sqlc.arg(revocation_id)
  AND status = 'processing'
  AND claim_token = sqlc.arg(claim_token);

-- name: SupersedeGoogleDriveRevocationsForGeneration :execrows
UPDATE public.google_drive_revocation_outbox
SET status = 'superseded',
    credential_payload = NULL,
    claim_token = NULL,
    lease_expires_at = NULL,
    last_error = NULL,
    terminal_at = sqlc.arg(superseded_at),
    updated_at = sqlc.arg(superseded_at)
WHERE google_subject = sqlc.arg(google_subject)
  AND installation_generation <> sqlc.arg(active_generation)
  AND status IN ('pending', 'processing', 'failed');
