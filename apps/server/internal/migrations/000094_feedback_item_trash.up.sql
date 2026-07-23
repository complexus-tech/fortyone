ALTER TABLE public.feedback_items
    ADD COLUMN deleted_at timestamptz;

CREATE INDEX idx_feedback_items_deleted_at
    ON public.feedback_items (deleted_at)
    WHERE deleted_at IS NOT NULL;
