-- name: EnqueueSlackUninstall :one
INSERT INTO public.slack_uninstall_outbox (
    slack_workspace_id,
    workspace_id,
    installation_generation,
    slack_team_id,
    uninstall_kind,
    credential_payload,
    credential_key_version,
    status,
    next_attempt_at
) VALUES (
    CAST(sqlc.arg(slack_workspace_id) AS uuid),
    CAST(sqlc.arg(workspace_id) AS uuid),
    CAST(sqlc.arg(installation_generation) AS uuid),
    CAST(sqlc.arg(slack_team_id) AS text),
    CAST(sqlc.arg(uninstall_kind) AS text),
    CAST(sqlc.arg(credential_payload) AS text),
    CAST(sqlc.arg(credential_key_version) AS smallint),
    'pending',
    NOW()
)
RETURNING
    id, slack_workspace_id, workspace_id, installation_generation,
    slack_team_id, uninstall_kind, credential_payload, credential_key_version,
    status, attempt_count, last_error, next_attempt_at,
    processing_started_at, completed_at, created_at, updated_at;

-- name: ClaimSlackUninstall :one
UPDATE public.slack_uninstall_outbox
SET status = 'processing',
    attempt_count = attempt_count + 1,
    last_error = NULL,
    next_attempt_at = NULL,
    processing_started_at = NOW(),
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(uninstall_id) AS uuid)
  AND attempt_count < CAST(sqlc.arg(max_attempts) AS integer)
  AND (
      (status IN ('pending', 'failed') AND COALESCE(next_attempt_at, NOW()) <= NOW())
      OR (
          status = 'processing'
          AND updated_at < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
      )
  )
RETURNING
    id, slack_workspace_id, workspace_id, installation_generation,
    slack_team_id, uninstall_kind, credential_payload, credential_key_version,
    status, attempt_count, last_error, next_attempt_at,
    processing_started_at, completed_at, created_at, updated_at;

-- name: DeadLetterExhaustedSlackUninstalls :execrows
UPDATE public.slack_uninstall_outbox
SET status = 'revocation_required',
    last_error = COALESCE(NULLIF(last_error, ''), 'Slack uninstall recovery lease expired after the final attempt'),
    next_attempt_at = NULL,
    processing_started_at = NULL,
    updated_at = NOW()
WHERE attempt_count >= CAST(sqlc.arg(max_attempts) AS integer)
  AND (
      status IN ('pending', 'failed')
      OR (
          status = 'processing'
          AND updated_at < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
      )
  );

-- name: ClaimRecoverableSlackUninstalls :many
WITH candidates AS (
    SELECT id
    FROM public.slack_uninstall_outbox
    WHERE attempt_count < CAST(sqlc.arg(max_attempts) AS integer)
      AND (
          (status IN ('pending', 'failed') AND COALESCE(next_attempt_at, NOW()) <= NOW())
          OR (
              status = 'processing'
              AND updated_at < NOW() - (CAST(sqlc.arg(lease_seconds) AS bigint) * INTERVAL '1 second')
          )
      )
    ORDER BY COALESCE(next_attempt_at, created_at), created_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT CAST(sqlc.arg(result_limit) AS integer)
)
UPDATE public.slack_uninstall_outbox AS uninstall
SET status = 'processing',
    attempt_count = uninstall.attempt_count + 1,
    last_error = NULL,
    next_attempt_at = NULL,
    processing_started_at = NOW(),
    updated_at = NOW()
FROM candidates
WHERE uninstall.id = candidates.id
RETURNING
    uninstall.id,
    uninstall.slack_workspace_id,
    uninstall.workspace_id,
    uninstall.installation_generation,
    uninstall.slack_team_id,
    uninstall.uninstall_kind,
    uninstall.credential_payload,
    uninstall.credential_key_version,
    uninstall.status,
    uninstall.attempt_count,
    uninstall.last_error,
    uninstall.next_attempt_at,
    uninstall.processing_started_at,
    uninstall.completed_at,
    uninstall.created_at,
    uninstall.updated_at;

-- name: CompleteSlackUninstall :execrows
UPDATE public.slack_uninstall_outbox
SET status = 'completed',
    credential_payload = NULL,
    last_error = NULLIF(CAST(sqlc.arg(message) AS text), ''),
    next_attempt_at = NULL,
    processing_started_at = NULL,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(uninstall_id) AS uuid)
  AND status = 'processing';

-- name: FailSlackUninstall :execrows
UPDATE public.slack_uninstall_outbox
SET status = CASE
        WHEN CAST(sqlc.narg(next_attempt_at) AS timestamptz) IS NULL THEN 'revocation_required'
        ELSE 'failed'
    END,
    last_error = CAST(sqlc.arg(message) AS text),
    next_attempt_at = CAST(sqlc.narg(next_attempt_at) AS timestamptz),
    processing_started_at = NULL,
    updated_at = NOW()
WHERE id = CAST(sqlc.arg(uninstall_id) AS uuid)
  AND status = 'processing';
