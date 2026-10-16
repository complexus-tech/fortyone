CREATE TABLE public.agent_invocation_assistant_usage_events (
    invocation_id uuid NOT NULL REFERENCES public.agent_invocations(id) ON DELETE CASCADE,
    attempt_count integer NOT NULL CHECK (attempt_count > 0),
    usage_date date NOT NULL,
    input_tokens bigint NOT NULL CHECK (input_tokens >= 0),
    output_tokens bigint NOT NULL CHECK (output_tokens >= 0),
    total_tokens bigint NOT NULL CHECK (total_tokens >= 0 AND total_tokens = input_tokens + output_tokens),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (invocation_id, attempt_count)
);

CREATE INDEX agent_invocation_assistant_usage_events_created_idx
    ON public.agent_invocation_assistant_usage_events (created_at);
