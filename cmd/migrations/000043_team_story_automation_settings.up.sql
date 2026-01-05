-- 000043_team_story_automation_settings.up.sql
CREATE TABLE public.team_story_automation_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_id uuid NOT NULL,
    move_completed_to_next_sprint boolean DEFAULT true,
    move_incomplete_to_next_sprint boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.team_story_automation_settings ADD CONSTRAINT team_story_automation_settings_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.team_story_automation_settings ADD CONSTRAINT team_story_automation_settings_team_id_key UNIQUE (team_id);

ALTER TABLE ONLY public.team_story_automation_settings 
    ADD CONSTRAINT team_story_automation_settings_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;
