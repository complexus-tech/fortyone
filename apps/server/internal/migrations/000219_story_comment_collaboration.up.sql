CREATE TABLE public.story_comment_reactions (
    comment_id uuid NOT NULL REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    emoji text NOT NULL CHECK (emoji IN ('👍', '✅', '👀')),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (comment_id, user_id, emoji)
);

CREATE TABLE public.story_comment_resolutions (
    comment_id uuid PRIMARY KEY REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    resolved_at timestamptz,
    resolved_by uuid REFERENCES public.users(user_id) ON DELETE SET NULL
);

CREATE TABLE public.story_comment_reads (
    story_id uuid NOT NULL REFERENCES public.stories(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    last_read_at timestamptz NOT NULL,
    PRIMARY KEY (story_id, user_id)
);

CREATE INDEX idx_story_comments_unread ON public.story_comments (story_id, created_at, commenter_id);
