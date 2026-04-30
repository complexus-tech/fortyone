CREATE TABLE public.integration_requests (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    team_id uuid NOT NULL,
    provider text NOT NULL,
    source_type text NOT NULL,
    source_external_id text NOT NULL,
    source_number integer,
    source_url text,
    title text NOT NULL,
    description text,
    status text NOT NULL DEFAULT 'pending',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    accepted_story_id uuid,
    accepted_by_user_id uuid,
    accepted_at timestamptz,
    declined_by_user_id uuid,
    declined_at timestamptz,
    created_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integration_requests_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT integration_requests_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT integration_requests_accepted_story_id_fkey FOREIGN KEY (accepted_story_id) REFERENCES public.stories(id) ON DELETE SET NULL,
    CONSTRAINT integration_requests_accepted_by_user_id_fkey FOREIGN KEY (accepted_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT integration_requests_declined_by_user_id_fkey FOREIGN KEY (declined_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT integration_requests_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT integration_requests_status_check CHECK (status IN ('pending', 'accepted', 'declined')),
    CONSTRAINT integration_requests_provider_check CHECK (provider IN ('github', 'slack', 'intercom')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX integration_requests_external_key
    ON public.integration_requests USING btree (workspace_id, provider, source_type, source_external_id);

CREATE INDEX idx_integration_requests_team_status
    ON public.integration_requests USING btree (workspace_id, team_id, status, created_at DESC);

CREATE INDEX idx_integration_requests_status
    ON public.integration_requests USING btree (workspace_id, status, created_at DESC);
