CREATE TABLE public.maya_realtime_voice_sessions (
    session_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    started_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT maya_realtime_voice_sessions_pkey PRIMARY KEY (session_id),
    CONSTRAINT maya_realtime_voice_sessions_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT maya_realtime_voice_sessions_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT maya_realtime_voice_sessions_end_after_start_check
        CHECK (ended_at IS NULL OR ended_at >= started_at)
);

CREATE INDEX idx_maya_realtime_voice_sessions_workspace_started
    ON public.maya_realtime_voice_sessions (workspace_id, started_at DESC);

CREATE INDEX idx_maya_realtime_voice_sessions_user_started
    ON public.maya_realtime_voice_sessions (user_id, started_at DESC);

CREATE INDEX idx_maya_realtime_voice_sessions_open
    ON public.maya_realtime_voice_sessions (workspace_id, started_at DESC)
    WHERE ended_at IS NULL;
