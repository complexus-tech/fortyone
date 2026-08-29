CREATE TABLE public.maya_realtime_voice_tool_calls (
    tool_call_id uuid NOT NULL DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL,
    call_id text NOT NULL,
    tool_name text NOT NULL,
    request_hash text NOT NULL,
    response jsonb,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT maya_realtime_voice_tool_calls_pkey PRIMARY KEY (tool_call_id),
    CONSTRAINT maya_realtime_voice_tool_calls_session_id_fkey
        FOREIGN KEY (session_id) REFERENCES public.maya_realtime_voice_sessions(session_id) ON DELETE CASCADE,
    CONSTRAINT maya_realtime_voice_tool_calls_session_call_unique UNIQUE (session_id, call_id)
);

CREATE INDEX idx_maya_realtime_voice_tool_calls_session_created
    ON public.maya_realtime_voice_tool_calls (session_id, created_at DESC);
