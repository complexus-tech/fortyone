-- 000009_teams.up.sql
CREATE TABLE public.teams (
    team_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    workspace_id uuid NOT NULL,
    code character varying(255) NOT NULL,
    color character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    is_private boolean DEFAULT false NOT NULL
);

ALTER TABLE ONLY public.teams ADD CONSTRAINT teams_pkey PRIMARY KEY (team_id);
ALTER TABLE ONLY public.teams ADD CONSTRAINT teams_workspace_id_code_key UNIQUE (workspace_id, code);

ALTER TABLE ONLY public.teams 
    ADD CONSTRAINT teams_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_teams_created ON public.teams USING btree (created_at DESC);
CREATE INDEX idx_teams_workspace_created ON public.teams USING btree (workspace_id, created_at DESC);
