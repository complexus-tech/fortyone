-- 000028_notification_preferences.up.sql
CREATE TABLE public.notification_preferences (
    preference_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    preferences jsonb DEFAULT '{"mention": {"email": true, "in_app": true}, "story_update": {"email": true, "in_app": true}, "comment_reply": {"email": true, "in_app": true}, "story_comment": {"email": true, "in_app": true}, "overdue_stories": {"email": true, "in_app": true}, "objective_update": {"email": true, "in_app": true}, "key_result_update": {"email": true, "in_app": true}}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.notification_preferences ADD CONSTRAINT notification_preferences_pkey PRIMARY KEY (preference_id);
ALTER TABLE ONLY public.notification_preferences ADD CONSTRAINT notification_preferences_user_id_workspace_id_key UNIQUE (user_id, workspace_id);

ALTER TABLE ONLY public.notification_preferences 
    ADD CONSTRAINT notification_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.notification_preferences 
    ADD CONSTRAINT notification_preferences_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;
