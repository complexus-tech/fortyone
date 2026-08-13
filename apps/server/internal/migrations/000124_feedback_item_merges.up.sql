ALTER TABLE public.feedback_items
    ADD COLUMN merged_into_item_id uuid,
    ADD COLUMN merged_at timestamptz,
    ADD COLUMN merged_by_user_id uuid,
    ADD CONSTRAINT feedback_items_workspace_portal_id_key
        UNIQUE (workspace_id, portal_id, id),
    ADD CONSTRAINT feedback_items_merged_into_item_fkey
        FOREIGN KEY (workspace_id, portal_id, merged_into_item_id)
        REFERENCES public.feedback_items(workspace_id, portal_id, id)
        ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
    ADD CONSTRAINT feedback_items_merged_by_user_id_fkey
        FOREIGN KEY (merged_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    ADD CONSTRAINT feedback_items_merge_not_self_check
        CHECK (merged_into_item_id IS NULL OR merged_into_item_id <> id),
    ADD CONSTRAINT feedback_items_merge_lifecycle_check
        CHECK (
            (
                merged_into_item_id IS NULL
                AND merged_at IS NULL
                AND merged_by_user_id IS NULL
            )
            OR (
                merged_into_item_id IS NOT NULL
                AND merged_at IS NOT NULL
                AND merged_at >= created_at
            )
        );

CREATE INDEX idx_feedback_items_merged_into_item
    ON public.feedback_items (
        workspace_id,
        portal_id,
        merged_into_item_id,
        merged_at DESC,
        id
    )
    WHERE merged_into_item_id IS NOT NULL;

CREATE INDEX idx_feedback_items_public_active
    ON public.feedback_items (portal_id, status, created_at DESC, id DESC)
    WHERE deleted_at IS NULL AND merged_into_item_id IS NULL;

-- Merge delivery has different aggregate and recipient semantics from Update
-- publication, so it has a focused outbox. It intentionally has no foreign
-- keys to mutable item, user, portal, or workspace rows: a committed merge
-- event and its follower snapshot must survive later aggregate deletion.
CREATE TABLE public.feedback_item_merge_outbox (
    merge_event_id uuid NOT NULL DEFAULT gen_random_uuid(),
    source_item_id uuid NOT NULL,
    target_item_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    portal_id uuid NOT NULL,
    merged_by_user_id uuid NOT NULL,
    merged_at timestamptz NOT NULL,
    event_type text NOT NULL DEFAULT 'feedback.item.merged',
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
    CONSTRAINT feedback_item_merge_outbox_distinct_items_check
        CHECK (source_item_id <> target_item_id),
    CONSTRAINT feedback_item_merge_outbox_event_type_check
        CHECK (event_type = 'feedback.item.merged'),
    CONSTRAINT feedback_item_merge_outbox_payload_check
        CHECK (jsonb_typeof(event_payload) = 'object'),
    CONSTRAINT feedback_item_merge_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'retrying', 'completed', 'failed')),
    CONSTRAINT feedback_item_merge_outbox_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT feedback_item_merge_outbox_merged_at_check
        CHECK (merged_at <= created_at),
    CONSTRAINT feedback_item_merge_outbox_claimed_at_check
        CHECK (claimed_at IS NULL OR claimed_at >= created_at),
    CONSTRAINT feedback_item_merge_outbox_completed_at_check
        CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT feedback_item_merge_outbox_updated_at_check
        CHECK (updated_at >= created_at),
    CONSTRAINT feedback_item_merge_outbox_last_error_check
        CHECK (
            last_error IS NULL
            OR length(btrim(last_error)) BETWEEN 1 AND 4000
        ),
    CONSTRAINT feedback_item_merge_outbox_lifecycle_check
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
    PRIMARY KEY (merge_event_id),
    CONSTRAINT feedback_item_merge_outbox_source_item_key UNIQUE (source_item_id)
);

CREATE UNIQUE INDEX feedback_item_merge_outbox_claim_token_key
    ON public.feedback_item_merge_outbox (claim_token)
    WHERE claim_token IS NOT NULL;

CREATE INDEX idx_feedback_item_merge_outbox_ready
    ON public.feedback_item_merge_outbox (
        next_attempt_at,
        created_at,
        merge_event_id
    )
    WHERE status IN ('pending', 'retrying');

CREATE INDEX idx_feedback_item_merge_outbox_stale_claim
    ON public.feedback_item_merge_outbox (claimed_at, merge_event_id)
    WHERE status = 'processing';

CREATE INDEX idx_feedback_item_merge_outbox_target
    ON public.feedback_item_merge_outbox (target_item_id, merged_at DESC, merge_event_id);

CREATE INDEX idx_feedback_item_merge_outbox_retention
    ON public.feedback_item_merge_outbox (updated_at, merge_event_id)
    WHERE status IN ('completed', 'failed');
