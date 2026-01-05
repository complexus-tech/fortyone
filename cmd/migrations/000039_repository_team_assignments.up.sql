-- 000039_repository_team_assignments.up.sql
CREATE TABLE public.repository_team_assignments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    repository_id uuid NOT NULL,
    team_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.repository_team_assignments ADD CONSTRAINT repository_team_assignments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.repository_team_assignments ADD CONSTRAINT repository_team_assignments_repository_id_team_id_key UNIQUE (repository_id, team_id);

ALTER TABLE ONLY public.repository_team_assignments 
    ADD CONSTRAINT repository_team_assignments_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.repository_team_assignments 
    ADD CONSTRAINT repository_team_assignments_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;
