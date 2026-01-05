-- 000045_user_automation_preferences.up.sql
CREATE TABLE public.user_automation_preferences (
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    auto_move_story_on_status_change boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.user_automation_preferences ADD CONSTRAINT user_automation_preferences_pkey PRIMARY KEY (user_id, workspace_id);

ALTER TABLE ONLY public.user_automation_preferences 
    ADD CONSTRAINT user_automation_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_automation_preferences 
    ADD CONSTRAINT user_automation_preferences_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;
