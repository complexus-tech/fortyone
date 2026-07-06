CREATE TABLE public.workspace_analytics_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    team_id uuid NULL,
    story_id uuid NULL,
    objective_id uuid NULL,
    sprint_id uuid NULL,
    key_result_id uuid NULL,
    event_name varchar(120) NOT NULL,
    surface varchar(80) NOT NULL DEFAULT 'unknown',
    properties jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT workspace_analytics_events_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT workspace_analytics_events_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT workspace_analytics_events_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE SET NULL,
    CONSTRAINT workspace_analytics_events_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE SET NULL,
    CONSTRAINT workspace_analytics_events_objective_id_fkey
        FOREIGN KEY (objective_id) REFERENCES public.objectives(objective_id) ON DELETE SET NULL,
    CONSTRAINT workspace_analytics_events_sprint_id_fkey
        FOREIGN KEY (sprint_id) REFERENCES public.sprints(sprint_id) ON DELETE SET NULL,
    CONSTRAINT workspace_analytics_events_key_result_id_fkey
        FOREIGN KEY (key_result_id) REFERENCES public.key_results(id) ON DELETE SET NULL
);

CREATE INDEX idx_workspace_analytics_events_workspace_time
    ON public.workspace_analytics_events USING btree (workspace_id, occurred_at DESC);

CREATE INDEX idx_workspace_analytics_events_workspace_event_time
    ON public.workspace_analytics_events USING btree (workspace_id, event_name, occurred_at DESC);

CREATE INDEX idx_workspace_analytics_events_workspace_surface_time
    ON public.workspace_analytics_events USING btree (workspace_id, surface, occurred_at DESC);

CREATE INDEX idx_workspace_analytics_events_workspace_user_time
    ON public.workspace_analytics_events USING btree (workspace_id, user_id, occurred_at DESC);

CREATE INDEX idx_workspace_analytics_events_properties
    ON public.workspace_analytics_events USING gin (properties);
