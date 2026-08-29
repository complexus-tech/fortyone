CREATE INDEX idx_feedback_items_portal_author_created
    ON public.feedback_items (portal_id, author_id, created_at DESC, id DESC)
    WHERE author_id IS NOT NULL;

CREATE INDEX idx_feedback_comments_author_created
    ON public.feedback_comments (author_id, created_at DESC, id DESC)
    INCLUDE (item_id)
    WHERE author_id IS NOT NULL;
