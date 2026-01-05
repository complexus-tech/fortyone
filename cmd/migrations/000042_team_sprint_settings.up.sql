-- 000042_team_sprint_settings.up.sql
CREATE TABLE public.team_sprint_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    sprint_duration_weeks integer DEFAULT 2,
    auto_create_next_sprint boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    sprint_start_day integer DEFAULT 1
);

ALTER TABLE ONLY public.team_sprint_settings ADD CONSTRAINT team_sprint_settings_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.team_sprint_settings ADD CONSTRAINT team_sprint_settings_team_id_key UNIQUE (team_id);

ALTER TABLE ONLY public.team_sprint_settings 
    ADD CONSTRAINT team_sprint_settings_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;
