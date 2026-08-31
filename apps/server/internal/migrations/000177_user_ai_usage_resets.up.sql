CREATE TABLE public.user_ai_usage_resets (
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    period_start timestamptz NOT NULL,
    baseline_message_count bigint NOT NULL,
    reset_at timestamptz NOT NULL,
    reset_by_user_id uuid NOT NULL,
    CONSTRAINT user_ai_usage_resets_pkey
        PRIMARY KEY (user_id, workspace_id, period_start),
    CONSTRAINT user_ai_usage_resets_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT user_ai_usage_resets_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT user_ai_usage_resets_reset_by_user_id_fkey
        FOREIGN KEY (reset_by_user_id) REFERENCES public.users(user_id),
    CONSTRAINT user_ai_usage_resets_baseline_message_count_check
        CHECK (baseline_message_count >= 0)
);

CREATE INDEX user_ai_usage_resets_reset_at_idx
    ON public.user_ai_usage_resets (reset_at DESC);
