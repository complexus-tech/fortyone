-- Schedule-transition events are committed with the story state that produced
-- them, then dispatched asynchronously. This outbox intentionally has no
-- foreign keys to mutable story, actor, or workspace rows: the complete
-- events.Event JSON snapshot must survive later aggregate deletion until the
-- publisher has delivered it and bounded retention cleanup removes it.
CREATE TABLE public.story_schedule_transition_outbox (
    schedule_transition_event_id uuid NOT NULL DEFAULT gen_random_uuid(),
    actor_id uuid NOT NULL,
    story_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    event_type text NOT NULL DEFAULT 'story.updated',
    event_payload jsonb NOT NULL,
    semantic_fingerprint text NOT NULL,
    transition_sequence bigint NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz DEFAULT now(),
    claim_token uuid,
    claimed_at timestamptz,
    completed_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT story_schedule_transition_outbox_event_type_check
        CHECK (event_type = 'story.updated'),
    CONSTRAINT story_schedule_transition_outbox_payload_check
        CHECK (
            jsonb_typeof(event_payload) = 'object'
            AND event_payload ? 'type'
            AND event_payload ? 'payload'
            AND event_payload ? 'timestamp'
            AND event_payload ? 'actor_id'
            AND jsonb_typeof(event_payload -> 'payload') = 'object'
            AND event_payload ->> 'type' = event_type
            AND event_payload ->> 'actor_id' = CAST(actor_id AS text)
            AND event_payload #>> '{payload,story_id}' = CAST(story_id AS text)
            AND event_payload #>> '{payload,workspace_id}' = CAST(workspace_id AS text)
        ),
    CONSTRAINT story_schedule_transition_outbox_fingerprint_check
        CHECK (
            length(btrim(semantic_fingerprint)) BETWEEN 1 AND 256
        ),
    CONSTRAINT story_schedule_transition_outbox_sequence_check
        CHECK (transition_sequence > 0),
    CONSTRAINT story_schedule_transition_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'retrying', 'completed', 'failed')),
    CONSTRAINT story_schedule_transition_outbox_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT story_schedule_transition_outbox_claimed_at_check
        CHECK (claimed_at IS NULL OR claimed_at >= created_at),
    CONSTRAINT story_schedule_transition_outbox_completed_at_check
        CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT story_schedule_transition_outbox_updated_at_check
        CHECK (updated_at >= created_at),
    CONSTRAINT story_schedule_transition_outbox_last_error_check
        CHECK (
            last_error IS NULL
            OR length(btrim(last_error)) BETWEEN 1 AND 4000
        ),
    CONSTRAINT story_schedule_transition_outbox_lifecycle_check
        CHECK (
            (
                status = 'pending'
                AND attempt_count = 0
                AND next_attempt_at IS NOT NULL
                AND claim_token IS NULL
                AND claimed_at IS NULL
                AND completed_at IS NULL
                AND last_error IS NULL
            )
            OR (
                status = 'processing'
                AND attempt_count > 0
                AND next_attempt_at IS NULL
                AND claim_token IS NOT NULL
                AND claimed_at IS NOT NULL
                AND completed_at IS NULL
                AND last_error IS NULL
            )
            OR (
                status = 'retrying'
                AND attempt_count > 0
                AND next_attempt_at IS NOT NULL
                AND claim_token IS NULL
                AND claimed_at IS NULL
                AND completed_at IS NULL
                AND last_error IS NOT NULL
            )
            OR (
                status = 'completed'
                AND attempt_count > 0
                AND next_attempt_at IS NULL
                AND claim_token IS NULL
                AND claimed_at IS NULL
                AND completed_at IS NOT NULL
                AND last_error IS NULL
            )
            OR (
                -- Failed is reserved for a permanently malformed immutable
                -- payload. Transient publication failures remain retrying.
                status = 'failed'
                AND attempt_count > 0
                AND next_attempt_at IS NULL
                AND claim_token IS NULL
                AND claimed_at IS NULL
                AND completed_at IS NULL
                AND last_error IS NOT NULL
            )
        ),
    PRIMARY KEY (schedule_transition_event_id),
    CONSTRAINT story_schedule_transition_outbox_story_sequence_key
        UNIQUE (workspace_id, story_id, transition_sequence)
);

CREATE UNIQUE INDEX story_schedule_transition_outbox_claim_token_key
    ON public.story_schedule_transition_outbox (claim_token)
    WHERE claim_token IS NOT NULL;

CREATE INDEX idx_story_schedule_transition_outbox_ready
    ON public.story_schedule_transition_outbox (
        next_attempt_at,
        created_at,
        schedule_transition_event_id
    )
    WHERE status IN ('pending', 'retrying');

CREATE INDEX idx_story_schedule_transition_outbox_stale_claim
    ON public.story_schedule_transition_outbox (
        claimed_at,
        schedule_transition_event_id
    )
    WHERE status = 'processing';

CREATE INDEX idx_story_schedule_transition_outbox_story_latest
    ON public.story_schedule_transition_outbox (
        workspace_id,
        story_id,
        transition_sequence DESC,
        schedule_transition_event_id DESC
    );

-- This fingerprint is deliberately non-unique. The same meaningful schedule
-- state can recur after intervening transitions; consumers use recent history
-- for semantic suppression without discarding a later committed event.
CREATE INDEX idx_story_schedule_transition_outbox_semantic_fingerprint
    ON public.story_schedule_transition_outbox (
        workspace_id,
        story_id,
        semantic_fingerprint,
        transition_sequence DESC
    );

-- Supports bounded DELETE batches ordered by completed_at and id. Permanently
-- malformed rows are excluded so operators retain their diagnostic payload.
CREATE INDEX idx_story_schedule_transition_outbox_retention
    ON public.story_schedule_transition_outbox (
        completed_at,
        schedule_transition_event_id
    )
    WHERE status = 'completed';
