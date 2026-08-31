CREATE TABLE public.user_onboarding_tour_progress (
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    tour_key text NOT NULL,
    tour_version text NOT NULL,
    completed_step_ids text[] NOT NULL DEFAULT ARRAY[]::text[],
    completed_action_ids text[] NOT NULL DEFAULT ARRAY[]::text[],
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_onboarding_tour_progress_pkey
        PRIMARY KEY (user_id, workspace_id, tour_key, tour_version),
    CONSTRAINT user_onboarding_tour_progress_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT user_onboarding_tour_progress_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT user_onboarding_tour_progress_status_check
        CHECK (status IN ('active', 'completed', 'skipped'))
);

CREATE INDEX idx_user_onboarding_tour_progress_workspace_id
    ON public.user_onboarding_tour_progress (workspace_id);
