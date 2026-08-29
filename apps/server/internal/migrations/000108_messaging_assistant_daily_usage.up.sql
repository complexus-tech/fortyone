CREATE TABLE public.messaging_assistant_daily_usage (
    workspace_id uuid NOT NULL,
    usage_date date NOT NULL,
    input_tokens bigint NOT NULL DEFAULT 0,
    output_tokens bigint NOT NULL DEFAULT 0,
    total_tokens bigint NOT NULL DEFAULT 0,
    request_count bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_assistant_daily_usage_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_assistant_daily_usage_input_tokens_check CHECK (input_tokens >= 0),
    CONSTRAINT messaging_assistant_daily_usage_output_tokens_check CHECK (output_tokens >= 0),
    CONSTRAINT messaging_assistant_daily_usage_total_tokens_check CHECK (
        total_tokens >= 0 AND total_tokens = input_tokens + output_tokens
    ),
    CONSTRAINT messaging_assistant_daily_usage_request_count_check CHECK (request_count >= 0),
    PRIMARY KEY (workspace_id, usage_date)
);

CREATE INDEX messaging_assistant_daily_usage_date_idx
    ON public.messaging_assistant_daily_usage USING btree (usage_date);

CREATE TABLE public.messaging_assistant_usage_events (
    inbound_event_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    provider text NOT NULL,
    external_workspace_id text NOT NULL,
    external_event_id text NOT NULL,
    attempt_count integer NOT NULL,
    usage_date date NOT NULL,
    input_tokens bigint NOT NULL,
    output_tokens bigint NOT NULL,
    total_tokens bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_assistant_usage_events_inbound_event_id_fkey
        FOREIGN KEY (inbound_event_id) REFERENCES public.messaging_inbound_events(id) ON DELETE CASCADE,
    CONSTRAINT messaging_assistant_usage_events_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_assistant_usage_events_attempt_count_check CHECK (attempt_count > 0),
    CONSTRAINT messaging_assistant_usage_events_input_tokens_check CHECK (input_tokens >= 0),
    CONSTRAINT messaging_assistant_usage_events_output_tokens_check CHECK (output_tokens >= 0),
    CONSTRAINT messaging_assistant_usage_events_total_tokens_check CHECK (
        total_tokens >= 0 AND total_tokens = input_tokens + output_tokens
    ),
    PRIMARY KEY (inbound_event_id, attempt_count),
    CONSTRAINT messaging_assistant_usage_events_external_attempt_key UNIQUE (
        workspace_id,
        provider,
        external_workspace_id,
        external_event_id,
        attempt_count
    )
);

CREATE INDEX messaging_assistant_usage_events_created_idx
    ON public.messaging_assistant_usage_events USING btree (created_at);
