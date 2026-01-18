-- 000045_user_automation_preferences.up.sql
CREATE TABLE "public"."user_automation_preferences" (
    "user_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "auto_assign_self" bool DEFAULT false,
    "assign_self_on_branch_copy" bool DEFAULT false,
    "move_story_to_started_on_branch" bool DEFAULT false,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP,
    "open_story_in_dialog" bool DEFAULT true,
    CONSTRAINT "user_automation_preferences_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "user_automation_preferences_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("user_id","workspace_id")
);


-- Indices
CREATE INDEX idx_user_automation_preferences_user_id ON public.user_automation_preferences USING btree (user_id);
CREATE INDEX idx_user_automation_preferences_user_workspace ON public.user_automation_preferences USING btree (user_id, workspace_id);
CREATE INDEX idx_user_automation_preferences_workspace_id ON public.user_automation_preferences USING btree (workspace_id);
