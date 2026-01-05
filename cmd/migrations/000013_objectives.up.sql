-- 000013_objectives.up.sql
CREATE TABLE public.objectives (
    objective_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    lead_user_id uuid,
    team_id uuid,
    workspace_id uuid,
    start_date date,
    end_date date,
    is_private boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status_id uuid,
    priority character varying(100),
    health public.objective_health_status,
    created_by uuid,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, (((COALESCE(name, ''::character varying))::text || ' '::text) || COALESCE(description, ''::text)))) STORED
);

ALTER TABLE ONLY public.objectives ADD CONSTRAINT objectives_pkey PRIMARY KEY (objective_id);

ALTER TABLE ONLY public.objectives 
    ADD CONSTRAINT objectives_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(user_id);

ALTER TABLE ONLY public.objectives 
    ADD CONSTRAINT objectives_lead_user_id_fkey FOREIGN KEY (lead_user_id) REFERENCES public.users(user_id);

ALTER TABLE ONLY public.objectives 
    ADD CONSTRAINT objectives_status_id_fkey FOREIGN KEY (status_id) REFERENCES public.objective_statuses(status_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.objectives 
    ADD CONSTRAINT objectives_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);

ALTER TABLE ONLY public.objectives 
    ADD CONSTRAINT objectives_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_objectives_lead_user_id ON public.objectives USING btree (lead_user_id);
CREATE INDEX idx_objectives_name_trigram ON public.objectives USING gin (name public.gin_trgm_ops);
CREATE INDEX idx_objectives_search ON public.objectives USING gin (search_vector);
CREATE INDEX idx_objectives_search_workspace_team ON public.objectives USING btree (workspace_id, team_id);
CREATE UNIQUE INDEX objectives_name_team_unique ON public.objectives USING btree (name, team_id);
