ALTER TABLE public.messaging_conversations
    ADD COLUMN audience_scope text NOT NULL DEFAULT 'actor';

ALTER TABLE public.messaging_outbound_deliveries
    ADD COLUMN provider_payload jsonb;

ALTER TABLE public.messaging_conversations
    ADD CONSTRAINT messaging_conversations_audience_scope_check
    CHECK (audience_scope IN ('actor', 'channel'));

DROP INDEX public.messaging_conversations_external_key;

CREATE UNIQUE INDEX messaging_conversations_actor_key
    ON public.messaging_conversations USING btree (
        provider,
        workspace_id,
        external_workspace_id,
        external_channel_id,
        external_thread_id,
        user_id
    )
    WHERE audience_scope = 'actor';

CREATE UNIQUE INDEX messaging_conversations_channel_key
    ON public.messaging_conversations USING btree (
        provider,
        workspace_id,
        external_workspace_id,
        external_channel_id,
        external_thread_id
    )
    WHERE audience_scope = 'channel';

CREATE TABLE public.slack_channel_team_access (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    slack_workspace_id uuid NOT NULL,
    slack_channel_id text NOT NULL,
    team_id uuid NOT NULL,
    created_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_channel_team_access_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_channel_team_access_slack_workspace_id_fkey
        FOREIGN KEY (slack_workspace_id) REFERENCES public.slack_workspaces(id) ON DELETE CASCADE,
    CONSTRAINT slack_channel_team_access_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT slack_channel_team_access_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX slack_channel_team_access_channel_team_key
    ON public.slack_channel_team_access USING btree (
        workspace_id,
        slack_workspace_id,
        slack_channel_id,
        team_id
    );

CREATE INDEX slack_channel_team_access_channel_idx
    ON public.slack_channel_team_access USING btree (
        workspace_id,
        slack_workspace_id,
        slack_channel_id
    );

CREATE TABLE public.slack_agent_settings (
    workspace_id uuid NOT NULL,
    assistant_enabled boolean NOT NULL DEFAULT true,
    workflow_actions_enabled boolean NOT NULL DEFAULT true,
    guidance text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_agent_settings_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_agent_settings_guidance_length_check
        CHECK (char_length(guidance) <= 4000),
    PRIMARY KEY (workspace_id)
);
