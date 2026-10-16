CREATE TABLE public.story_comment_thread_reads (
    thread_id uuid NOT NULL REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    last_read_at timestamptz NOT NULL,
    PRIMARY KEY (thread_id, user_id)
);

CREATE INDEX idx_story_comment_thread_reads_user
    ON public.story_comment_thread_reads (user_id, thread_id);

COMMENT ON TABLE public.story_comment_thread_reads IS
    'Per-user reply-thread read cutoffs. The existing story marker applies only to root comments.';
