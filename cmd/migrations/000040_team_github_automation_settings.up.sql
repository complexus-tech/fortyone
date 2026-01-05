-- 000040_team_github_automation_settings.up.sql
CREATE TABLE public.team_github_automation_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    event_type character varying(50) NOT NULL,
    target_status_id uuid,
    enabled boolean DEFAULT true,
    branch_pattern character varying(255),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.team_github_automation_settings ADD CONSTRAINT team_github_automation_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.team_github_automation_settings 
    ADD CONSTRAINT team_github_automation_settings_target_status_id_fkey FOREIGN KEY (target_status_id) REFERENCES public.statuses(status_id);

ALTER TABLE ONLY public.team_github_automation_settings 
    ADD CONSTRAINT team_github_automation_settings_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);
