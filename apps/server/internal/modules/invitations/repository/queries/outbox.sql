-- name: InsertInvitationOutboxEvent :one
INSERT INTO public.workspace_invitation_outbox (
    outbox_id,
    invitation_id,
    workspace_id,
    actor_id,
    event_type,
    event_payload,
    idempotency_key,
    status,
    attempt_count,
    next_attempt_at,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(outbox_id),
    sqlc.arg(invitation_id),
    sqlc.arg(workspace_id),
    sqlc.arg(actor_id),
    sqlc.arg(event_type),
    CAST(sqlc.arg(event_payload) AS jsonb),
    sqlc.arg(idempotency_key),
    'pending',
    0,
    sqlc.arg(ready_at),
    sqlc.arg(ready_at),
    sqlc.arg(ready_at)
)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING outbox_id;

-- name: ClaimInvitationOutboxEvents :many
WITH candidates AS (
    SELECT outbox.outbox_id
    FROM public.workspace_invitation_outbox AS outbox
    WHERE (
            outbox.status IN ('pending', 'retrying')
            AND outbox.next_attempt_at <= sqlc.arg(claimed_at)
        )
        OR (
            outbox.status = 'processing'
            AND outbox.claimed_at <= sqlc.arg(stale_before)
        )
    ORDER BY COALESCE(outbox.next_attempt_at, outbox.claimed_at), outbox.created_at, outbox.outbox_id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(batch_size)
), claimed AS (
    UPDATE public.workspace_invitation_outbox AS outbox
    SET
        status = 'processing',
        attempt_count = outbox.attempt_count + 1,
        next_attempt_at = NULL,
        claim_token = gen_random_uuid(),
        claimed_at = sqlc.arg(claimed_at),
        completed_at = NULL,
        last_error = NULL,
        updated_at = sqlc.arg(claimed_at)
    FROM candidates
    WHERE outbox.outbox_id = candidates.outbox_id
    RETURNING
        outbox.outbox_id,
        outbox.invitation_id,
        outbox.workspace_id,
        outbox.actor_id,
        outbox.event_type,
        outbox.event_payload,
        outbox.idempotency_key,
        outbox.claim_token,
        outbox.attempt_count,
        outbox.created_at
)
SELECT
    claimed.outbox_id,
    claimed.invitation_id,
    claimed.workspace_id,
    claimed.actor_id,
    claimed.event_type,
    claimed.event_payload,
    claimed.idempotency_key,
    claimed.claim_token,
    claimed.attempt_count,
    claimed.created_at,
    invitation.token_digest,
    invitation.token_nonce,
    invitation.token_key_id,
    invitation.token_version,
    invitation.expires_at AS invitation_expires_at,
    invitation.used_at AS invitation_used_at
FROM claimed
LEFT JOIN public.workspace_invitations AS invitation
    ON invitation.invitation_id = claimed.invitation_id
ORDER BY claimed.created_at, claimed.outbox_id;

-- name: CompleteInvitationOutboxEvent :execrows
UPDATE public.workspace_invitation_outbox AS outbox
SET
    status = 'completed',
    next_attempt_at = NULL,
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = sqlc.arg(completed_at),
    last_error = NULL,
    updated_at = sqlc.arg(completed_at)
WHERE outbox.outbox_id = sqlc.arg(outbox_id)
  AND outbox.claim_token = sqlc.arg(claim_token)
  AND outbox.status = 'processing';

-- name: RetryInvitationOutboxEvent :execrows
UPDATE public.workspace_invitation_outbox AS outbox
SET
    status = 'retrying',
    next_attempt_at = sqlc.arg(retry_at),
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    last_error = LEFT(sqlc.arg(last_error), 4000),
    updated_at = sqlc.arg(updated_at)
WHERE outbox.outbox_id = sqlc.arg(outbox_id)
  AND outbox.claim_token = sqlc.arg(claim_token)
  AND outbox.status = 'processing';

-- name: FailInvitationOutboxEvent :execrows
UPDATE public.workspace_invitation_outbox AS outbox
SET
    status = 'failed',
    next_attempt_at = NULL,
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    last_error = LEFT(sqlc.arg(last_error), 4000),
    updated_at = sqlc.arg(updated_at)
WHERE outbox.outbox_id = sqlc.arg(outbox_id)
  AND outbox.claim_token = sqlc.arg(claim_token)
  AND outbox.status = 'processing';
