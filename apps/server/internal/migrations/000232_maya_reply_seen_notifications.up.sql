CREATE TABLE public.story_comment_seen (
    comment_id uuid NOT NULL REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    seen_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (comment_id, user_id)
);

CREATE TABLE public.maya_reply_notification_deliveries (
    dedupe_key text PRIMARY KEY,
    comment_id uuid NOT NULL REFERENCES public.story_comments(comment_id) ON DELETE CASCADE,
    recipient_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    payload jsonb NOT NULL,
    occurred_at timestamptz NOT NULL,
    available_at timestamptz NOT NULL,
    completed_at timestamptz,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    UNIQUE (comment_id, recipient_id)
);

CREATE INDEX maya_reply_notification_deliveries_pending
    ON public.maya_reply_notification_deliveries (available_at, dedupe_key)
    WHERE completed_at IS NULL;

COMMENT ON TABLE public.story_comment_seen IS
    'Exact visible-comment acknowledgments; independent of broad story/thread read watermarks.';
COMMENT ON TABLE public.maya_reply_notification_deliveries IS
    'Durable Maya reply decisions after a five-second visibility grace period. Explicit mentions remain immediate.';
