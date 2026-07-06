CREATE TABLE public.slack_request_logs (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    request_type text NOT NULL,
    endpoint text NOT NULL,
    workspace_id uuid,
    slack_team_id text,
    slack_user_id text,
    slack_channel_id text,
    command text,
    trigger_id text,
    request_body text,
    headers jsonb NOT NULL DEFAULT '{}'::jsonb,
    response_code integer NOT NULL,
    outcome text NOT NULL,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_request_logs_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_slack_request_logs_created_at ON public.slack_request_logs USING btree (created_at DESC);
CREATE INDEX idx_slack_request_logs_workspace_id ON public.slack_request_logs USING btree (workspace_id, created_at DESC);
CREATE INDEX idx_slack_request_logs_team_id ON public.slack_request_logs USING btree (slack_team_id, created_at DESC);

DROP TABLE IF EXISTS public.slack_team_channel_links;
DROP TABLE IF EXISTS public.slack_workspace_settings;
