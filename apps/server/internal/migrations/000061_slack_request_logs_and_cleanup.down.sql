CREATE TABLE public.slack_workspace_settings (
    workspace_id uuid NOT NULL,
    default_create_mode text NOT NULL DEFAULT 'create_task_now',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_workspace_settings_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_workspace_settings_default_create_mode_check CHECK (default_create_mode IN ('create_task_now', 'send_to_requests')),
    PRIMARY KEY (workspace_id)
);

CREATE TABLE public.slack_team_channel_links (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    slack_channel_id text NOT NULL,
    team_id uuid NOT NULL,
    is_active bool NOT NULL DEFAULT true,
    created_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_team_channel_links_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_team_channel_links_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT slack_team_channel_links_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX slack_team_channel_links_active_channel_key
    ON public.slack_team_channel_links USING btree (workspace_id, slack_channel_id)
    WHERE is_active = true;

CREATE UNIQUE INDEX slack_team_channel_links_active_team_key
    ON public.slack_team_channel_links USING btree (workspace_id, team_id)
    WHERE is_active = true;

DROP TABLE IF EXISTS public.slack_request_logs;
