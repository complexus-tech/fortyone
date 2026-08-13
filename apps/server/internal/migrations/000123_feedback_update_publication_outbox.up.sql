ALTER TABLE public.feedback_updates
    ADD COLUMN publication_sequence bigint NOT NULL DEFAULT 0;

-- Rows published before the durable outbox existed establish sequence one but
-- are not re-emitted. A later unpublish/publish transition advances to two.
UPDATE public.feedback_updates
SET publication_sequence = 1
WHERE status = 'published';

ALTER TABLE public.feedback_updates
    ADD CONSTRAINT feedback_updates_publication_sequence_check
        CHECK (
            publication_sequence >= 0
            AND (status <> 'published' OR publication_sequence > 0)
        );

-- This table intentionally has no foreign keys to mutable aggregate rows. An
-- Update, portal, user, or workspace deletion must not cascade a committed
-- publication event before its worker has dispatched it. The immutable ids and
-- event_payload are the delivery snapshot; retention cleanup owns deletion.
CREATE TABLE public.feedback_update_publication_outbox (
    publication_event_id uuid NOT NULL DEFAULT gen_random_uuid(),
    update_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    portal_id uuid NOT NULL,
    published_by_user_id uuid,
    publication_sequence bigint NOT NULL,
    published_at timestamptz NOT NULL,
    event_type text NOT NULL DEFAULT 'feedback.update.published',
    event_payload jsonb NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz DEFAULT now(),
    claim_token uuid,
    claimed_at timestamptz,
    completed_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_update_publication_outbox_event_type_check
        CHECK (event_type = 'feedback.update.published'),
    CONSTRAINT feedback_update_publication_outbox_sequence_check
        CHECK (publication_sequence > 0),
    CONSTRAINT feedback_update_publication_outbox_payload_check
        CHECK (jsonb_typeof(event_payload) = 'object'),
    CONSTRAINT feedback_update_publication_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'retrying', 'completed', 'failed')),
    CONSTRAINT feedback_update_publication_outbox_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT feedback_update_publication_outbox_published_at_check
        CHECK (published_at <= created_at),
    CONSTRAINT feedback_update_publication_outbox_claimed_at_check
        CHECK (claimed_at IS NULL OR claimed_at >= created_at),
    CONSTRAINT feedback_update_publication_outbox_completed_at_check
        CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT feedback_update_publication_outbox_updated_at_check
        CHECK (updated_at >= created_at),
    CONSTRAINT feedback_update_publication_outbox_last_error_check
        CHECK (
            last_error IS NULL
            OR length(btrim(last_error)) BETWEEN 1 AND 4000
        ),
    CONSTRAINT feedback_update_publication_outbox_lifecycle_check
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
                status = 'failed'
                AND attempt_count > 0
                AND next_attempt_at IS NULL
                AND claim_token IS NULL
                AND claimed_at IS NULL
                AND completed_at IS NULL
                AND last_error IS NOT NULL
            )
        ),
    PRIMARY KEY (publication_event_id),
    CONSTRAINT feedback_update_publication_outbox_update_sequence_key
        UNIQUE (update_id, publication_sequence)
);

CREATE UNIQUE INDEX feedback_update_publication_outbox_claim_token_key
    ON public.feedback_update_publication_outbox (claim_token)
    WHERE claim_token IS NOT NULL;

CREATE INDEX idx_feedback_update_publication_outbox_ready
    ON public.feedback_update_publication_outbox (
        next_attempt_at,
        created_at,
        publication_event_id
    )
    WHERE status IN ('pending', 'retrying');

CREATE INDEX idx_feedback_update_publication_outbox_stale_claim
    ON public.feedback_update_publication_outbox (claimed_at, publication_event_id)
    WHERE status = 'processing';

CREATE INDEX idx_feedback_update_publication_outbox_update
    ON public.feedback_update_publication_outbox (
        update_id,
        publication_sequence DESC
    );

CREATE INDEX idx_feedback_update_publication_outbox_retention
    ON public.feedback_update_publication_outbox (updated_at, publication_event_id)
    WHERE status IN ('completed', 'failed');
