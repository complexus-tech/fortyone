CREATE INDEX idx_feedback_items_author_created
    ON public.feedback_items (author_id, created_at DESC, id DESC)
    INCLUDE (portal_id)
    WHERE author_id IS NOT NULL;
