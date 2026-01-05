-- 000036_github_automation_preferences.up.sql
CREATE TABLE public.github_automation_preferences (
    user_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    auto_create_branch boolean DEFAULT false,
    auto_create_pr boolean DEFAULT false,
    auto_move_story_on_pr_merge boolean DEFAULT true,
    auto_assign_pr_reviewer boolean DEFAULT false,
    branch_naming_pattern text DEFAULT 'story/{story-id}-{title}'::text,
    pr_template text,
    default_reviewer_ids uuid[] DEFAULT '{}'::uuid[],
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.github_automation_preferences ADD CONSTRAINT github_automation_preferences_pkey PRIMARY KEY (user_id, workspace_id);

ALTER TABLE ONLY public.github_automation_preferences 
    ADD CONSTRAINT github_automation_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.github_automation_preferences 
    ADD CONSTRAINT github_automation_preferences_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_user_automation_preferences_user_id ON public.github_automation_preferences USING btree (user_id);
CREATE INDEX idx_user_automation_preferences_user_workspace ON public.github_automation_preferences USING btree (user_id, workspace_id);
CREATE INDEX idx_user_automation_preferences_workspace_id ON public.github_automation_preferences USING btree (workspace_id);
