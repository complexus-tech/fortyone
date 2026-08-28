-- name: LockStoryForScheduleTransition :one
SELECT
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_locked
FROM public.stories AS story
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.updated_at = sqlc.arg(expected_updated_at)
FOR UPDATE;

-- name: GetLatestStoryScheduleTransition :one
SELECT
    outbox.semantic_fingerprint,
    outbox.transition_sequence
FROM public.story_schedule_transition_outbox AS outbox
WHERE outbox.story_id = sqlc.arg(story_id)
  AND outbox.workspace_id = sqlc.arg(workspace_id)
ORDER BY outbox.transition_sequence DESC, outbox.schedule_transition_event_id DESC
LIMIT 1;

-- name: UpdateStoryScheduleTransitionState :execrows
UPDATE public.stories AS story
SET
    auto_scheduling_status = sqlc.arg(auto_scheduling_status),
    auto_scheduling_reason = sqlc.narg(auto_scheduling_reason),
    auto_scheduling_updated_at = sqlc.arg(auto_scheduling_updated_at),
    auto_scheduling_locked = CASE
        WHEN CAST(sqlc.arg(set_auto_scheduling_locked) AS boolean)
        THEN CAST(sqlc.arg(auto_scheduling_locked) AS boolean)
        ELSE story.auto_scheduling_locked
    END
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.updated_at = sqlc.arg(expected_updated_at);

-- name: InsertStoryScheduleTransition :exec
INSERT INTO public.story_schedule_transition_outbox (
    schedule_transition_event_id,
    actor_id,
    story_id,
    workspace_id,
    event_type,
    event_payload,
    semantic_fingerprint,
    transition_sequence,
    status,
    attempt_count,
    next_attempt_at,
    claim_token,
    claimed_at
) VALUES (
    sqlc.arg(event_id),
    sqlc.arg(actor_id),
    sqlc.arg(story_id),
    sqlc.arg(workspace_id),
    'story.updated',
    sqlc.arg(event_payload),
    sqlc.arg(semantic_fingerprint),
    sqlc.arg(transition_sequence),
    CASE WHEN CAST(sqlc.arg(claim_immediately) AS boolean) THEN 'processing' ELSE 'pending' END,
    CASE WHEN CAST(sqlc.arg(claim_immediately) AS boolean) THEN 1 ELSE 0 END,
    CASE WHEN CAST(sqlc.arg(claim_immediately) AS boolean) THEN NULL ELSE CURRENT_TIMESTAMP END,
    CASE
        WHEN CAST(sqlc.arg(claim_immediately) AS boolean)
        THEN CAST(sqlc.narg(claim_token) AS uuid)
        ELSE NULL
    END,
    CASE WHEN CAST(sqlc.arg(claim_immediately) AS boolean) THEN CURRENT_TIMESTAMP ELSE NULL END
);

-- name: ClaimStoryScheduleTransitions :many
WITH candidates AS (
    SELECT outbox.schedule_transition_event_id
    FROM public.story_schedule_transition_outbox AS outbox
    WHERE (
        outbox.status IN ('pending', 'retrying')
        AND outbox.next_attempt_at <= CURRENT_TIMESTAMP
    ) OR (
        outbox.status = 'processing'
        AND outbox.claimed_at <= CURRENT_TIMESTAMP - make_interval(
            secs => CAST(sqlc.arg(stale_after_seconds) AS double precision)
        )
    )
    ORDER BY COALESCE(outbox.next_attempt_at, outbox.claimed_at), outbox.created_at, outbox.schedule_transition_event_id
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
UPDATE public.story_schedule_transition_outbox AS outbox
SET
    status = 'processing',
    attempt_count = outbox.attempt_count + 1,
    next_attempt_at = NULL,
    claim_token = gen_random_uuid(),
    claimed_at = CURRENT_TIMESTAMP,
    completed_at = NULL,
    last_error = NULL,
    updated_at = CURRENT_TIMESTAMP
FROM candidates
WHERE outbox.schedule_transition_event_id = candidates.schedule_transition_event_id
RETURNING
    outbox.schedule_transition_event_id,
    outbox.actor_id,
    outbox.story_id,
    outbox.workspace_id,
    outbox.semantic_fingerprint,
    outbox.transition_sequence,
    outbox.claim_token,
    outbox.attempt_count,
    outbox.event_payload,
    outbox.created_at;

-- name: CompleteStoryScheduleTransition :execrows
UPDATE public.story_schedule_transition_outbox
SET
    status = 'completed',
    next_attempt_at = NULL,
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = CURRENT_TIMESTAMP,
    last_error = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE schedule_transition_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';

-- name: RetryStoryScheduleTransition :execrows
UPDATE public.story_schedule_transition_outbox
SET
    status = 'retrying',
    next_attempt_at = sqlc.arg(retry_at),
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    last_error = LEFT(CAST(sqlc.arg(last_error) AS text), 4000),
    updated_at = CURRENT_TIMESTAMP
WHERE schedule_transition_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';

-- name: FailStoryScheduleTransition :execrows
UPDATE public.story_schedule_transition_outbox
SET
    status = 'failed',
    next_attempt_at = NULL,
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    last_error = LEFT(CAST(sqlc.arg(last_error) AS text), 4000),
    updated_at = CURRENT_TIMESTAMP
WHERE schedule_transition_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';

-- name: DeleteCompletedStoryScheduleTransitions :execrows
WITH expired AS (
    SELECT outbox.schedule_transition_event_id
    FROM public.story_schedule_transition_outbox AS outbox
    WHERE outbox.status = 'completed'
      AND outbox.completed_at < sqlc.arg(completed_before)
    ORDER BY outbox.completed_at, outbox.schedule_transition_event_id
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
DELETE FROM public.story_schedule_transition_outbox AS outbox
USING expired
WHERE outbox.schedule_transition_event_id = expired.schedule_transition_event_id;
