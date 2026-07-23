ALTER TABLE public.feedback_comments
    ADD COLUMN parent_id uuid,
    ADD CONSTRAINT feedback_comments_parent_id_fkey
        FOREIGN KEY (parent_id) REFERENCES public.feedback_comments(id) ON DELETE CASCADE,
    ADD CONSTRAINT feedback_comments_parent_not_self_check
        CHECK (parent_id IS NULL OR parent_id <> id);

CREATE INDEX idx_feedback_comments_item_parent_created
    ON public.feedback_comments (item_id, parent_id, created_at DESC, id DESC);
