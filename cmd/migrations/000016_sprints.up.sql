-- 000016_sprints.up.sql
CREATE TABLE public.sprints (
    sprint_id uuid DEFAULT gen_random_uuid() NOT NULL,
    objective_id uuid,
    team_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    goal text,
    status character varying(255),
    backlog_status character varying(255),
    completed_story_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.sprints ADD CONSTRAINT sprints_pkey PRIMARY KEY (sprint_id);

ALTER TABLE ONLY public.sprints 
    ADD CONSTRAINT sprints_objective_id_fkey FOREIGN KEY (objective_id) REFERENCES public.objectives(objective_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.sprints 
    ADD CONSTRAINT sprints_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.sprints 
    ADD CONSTRAINT sprints_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;
