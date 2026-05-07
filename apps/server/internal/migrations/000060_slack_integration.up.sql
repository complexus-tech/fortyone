CREATE TABLE public.slack_workspace_settings (
    workspace_id uuid NOT NULL,
    default_create_mode text NOT NULL DEFAULT 'create_task_now',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_workspace_settings_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_workspace_settings_default_create_mode_check CHECK (default_create_mode IN ('create_task_now', 'send_to_requests')),
    PRIMARY KEY (workspace_id)
);

CREATE TABLE public.slack_workspaces (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    slack_team_id text NOT NULL,
    slack_team_name text NOT NULL,
    slack_team_domain text NOT NULL,
    bot_user_id text,
    bot_access_token text NOT NULL,
    scope text,
    is_active bool NOT NULL DEFAULT true,
    installed_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_workspaces_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_workspaces_installed_by_user_id_fkey FOREIGN KEY (installed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX slack_workspaces_workspace_id_key
    ON public.slack_workspaces USING btree (workspace_id);
CREATE UNIQUE INDEX slack_workspaces_slack_team_id_key
    ON public.slack_workspaces USING btree (slack_team_id);

CREATE TABLE public.slack_channels (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    slack_workspace_id uuid NOT NULL,
    slack_channel_id text NOT NULL,
    name text NOT NULL,
    is_private bool NOT NULL DEFAULT false,
    is_archived bool NOT NULL DEFAULT false,
    is_member bool NOT NULL DEFAULT false,
    is_active bool NOT NULL DEFAULT true,
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_channels_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_channels_slack_workspace_id_fkey FOREIGN KEY (slack_workspace_id) REFERENCES public.slack_workspaces(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX slack_channels_workspace_channel_id_key
    ON public.slack_channels USING btree (workspace_id, slack_channel_id);
CREATE INDEX idx_slack_channels_workspace_active
    ON public.slack_channels USING btree (workspace_id, is_active, name);

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
