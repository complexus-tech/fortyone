ALTER TABLE public.feedback_items
    ADD COLUMN submission_source text NOT NULL DEFAULT 'internal',
    ADD CONSTRAINT feedback_items_submission_source_check
        CHECK (submission_source IN ('internal', 'portal', 'widget', 'integration'));

CREATE INDEX idx_feedback_items_board_external_created
    ON public.feedback_items (board_id, created_at DESC)
    WHERE submission_source IN ('portal', 'widget', 'integration');

CREATE TABLE public.feedback_board_subscriptions (
    board_id uuid NOT NULL,
    user_id uuid NOT NULL,
    email_frequency text NOT NULL DEFAULT 'daily',
    last_digest_sent_at timestamptz,
    last_digest_cursor_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_board_subscriptions_board_id_fkey
        FOREIGN KEY (board_id) REFERENCES public.feedback_boards(id) ON DELETE CASCADE,
    CONSTRAINT feedback_board_subscriptions_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT feedback_board_subscriptions_email_frequency_check
        CHECK (email_frequency IN ('daily', 'weekly')),
    PRIMARY KEY (board_id, user_id)
);

CREATE INDEX idx_feedback_board_subscriptions_user_frequency
    ON public.feedback_board_subscriptions (user_id, email_frequency);

CREATE INDEX idx_feedback_board_subscriptions_frequency_digest
    ON public.feedback_board_subscriptions (email_frequency, last_digest_sent_at);

INSERT INTO public.feedback_board_subscriptions (board_id, user_id, email_frequency)
SELECT fb.id, wm.user_id, 'daily'
FROM public.feedback_boards fb
INNER JOIN public.workspace_members wm
    ON wm.workspace_id = fb.workspace_id
    AND wm.role = 'admin'
INNER JOIN public.team_members tm
    ON tm.team_id = fb.team_id
    AND tm.user_id = wm.user_id
INNER JOIN public.users u
    ON u.user_id = wm.user_id
    AND u.is_active = true
    AND u.is_system = false
ON CONFLICT (board_id, user_id) DO NOTHING;

CREATE TABLE public.feedback_digest_deliveries (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    recipient_id uuid NOT NULL,
    local_date date NOT NULL,
    status text NOT NULL DEFAULT 'processing',
    window_start timestamptz NOT NULL,
    window_end timestamptz NOT NULL,
    item_count integer NOT NULL DEFAULT 0,
    sent_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT feedback_digest_deliveries_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT feedback_digest_deliveries_recipient_id_fkey
        FOREIGN KEY (recipient_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT feedback_digest_deliveries_status_check
        CHECK (status IN ('processing', 'sent', 'skipped', 'failed')),
    CONSTRAINT feedback_digest_deliveries_item_count_check
        CHECK (item_count >= 0),
    CONSTRAINT feedback_digest_deliveries_window_check
        CHECK (window_end > window_start),
    PRIMARY KEY (id),
    UNIQUE (workspace_id, recipient_id, local_date)
);

CREATE INDEX idx_feedback_digest_deliveries_status_date
    ON public.feedback_digest_deliveries (status, local_date);

CREATE INDEX idx_feedback_digest_deliveries_recipient_date
    ON public.feedback_digest_deliveries (recipient_id, local_date DESC);
