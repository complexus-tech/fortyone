CREATE TABLE public.chat_mutation_approval_executions (
    session_id text NOT NULL,
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    tool_call_id text NOT NULL,
    fingerprint text NOT NULL,
    status text NOT NULL DEFAULT 'in_progress',
    output jsonb,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at timestamptz,
    CONSTRAINT chat_mutation_approval_executions_pkey
        PRIMARY KEY (session_id, user_id, workspace_id, tool_call_id),
    CONSTRAINT chat_mutation_approval_executions_session_fkey
        FOREIGN KEY (session_id)
        REFERENCES public.chat_sessions(id)
        ON DELETE CASCADE,
    CONSTRAINT chat_mutation_approval_executions_tool_call_id_check
        CHECK (char_length(tool_call_id) BETWEEN 1 AND 255),
    CONSTRAINT chat_mutation_approval_executions_fingerprint_check
        CHECK (fingerprint ~ '^[0-9a-f]{64}$'),
    CONSTRAINT chat_mutation_approval_executions_status_check
        CHECK (status IN ('in_progress', 'completed')),
    CONSTRAINT chat_mutation_approval_executions_completion_check
        CHECK (
            (status = 'in_progress' AND output IS NULL AND completed_at IS NULL)
            OR
            (status = 'completed' AND output IS NOT NULL AND completed_at IS NOT NULL)
        )
);

-- A tool call belongs to exactly one owner-scoped chat session. This secondary
-- key also prevents a differently scoped row from being inserted for the same
-- session/tool-call pair if application ownership checks ever regress.
CREATE UNIQUE INDEX chat_mutation_approval_executions_session_tool_call_key
    ON public.chat_mutation_approval_executions (session_id, tool_call_id);
