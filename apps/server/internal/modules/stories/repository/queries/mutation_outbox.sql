-- name: ClaimStoryMutationEvents :many
WITH candidates AS (
    SELECT event.event_id
    FROM public.story_mutation_events AS event
    WHERE (
        event.status = 'pending'
        AND event.next_attempt_at <= CAST(sqlc.arg(now) AS timestamptz)
    ) OR (
        event.status = 'processing'
        AND event.claimed_at <= CAST(sqlc.arg(stale_before) AS timestamptz)
    )
    ORDER BY event.next_attempt_at, event.created_at, event.event_id
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
UPDATE public.story_mutation_events AS event
SET
    status = 'processing',
    attempt_count = event.attempt_count + 1,
    claim_token = gen_random_uuid(),
    claimed_at = CAST(sqlc.arg(now) AS timestamptz),
    updated_at = CAST(sqlc.arg(now) AS timestamptz),
    last_error = NULL
FROM candidates
WHERE event.event_id = candidates.event_id
RETURNING
    event.event_id,
    event.workspace_id,
    event.story_id,
    event.event_type,
    event.actor_kind,
    event.actor_id,
    event.actor_credential_id,
    event.payload,
    event.occurred_at,
    event.attempt_count,
    event.claim_token,
    event.claimed_at,
    event.created_at;

-- name: CompleteStoryMutationEvent :execrows
UPDATE public.story_mutation_events
SET
    status = 'completed',
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = CAST(sqlc.arg(completed_at) AS timestamptz),
    updated_at = CAST(sqlc.arg(completed_at) AS timestamptz),
    last_error = NULL
WHERE event_id = sqlc.arg(event_id)
  AND status = 'processing'
  AND claim_token = CAST(sqlc.arg(claim_token) AS uuid);

-- name: RetryStoryMutationEvent :execrows
UPDATE public.story_mutation_events
SET
    status = 'pending',
    claim_token = NULL,
    claimed_at = NULL,
    next_attempt_at = CAST(sqlc.arg(next_attempt_at) AS timestamptz),
    updated_at = CAST(sqlc.arg(updated_at) AS timestamptz),
    last_error = CAST(sqlc.arg(last_error) AS text)
WHERE event_id = sqlc.arg(event_id)
  AND status = 'processing'
  AND claim_token = CAST(sqlc.arg(claim_token) AS uuid);
