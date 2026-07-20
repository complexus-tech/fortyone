CREATE TABLE public.feedback_item_reads (
    item_id uuid NOT NULL,
    user_id uuid NOT NULL,
    read_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_item_reads_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.feedback_items(id) ON DELETE CASCADE,
    CONSTRAINT feedback_item_reads_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, user_id)
);

CREATE INDEX idx_feedback_item_reads_user
    ON public.feedback_item_reads (user_id, item_id);
