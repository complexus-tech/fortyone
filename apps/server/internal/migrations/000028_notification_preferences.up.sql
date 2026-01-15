-- 000028_notification_preferences.up.sql
CREATE TABLE "public"."notification_preferences" (
    "preference_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "preferences" jsonb NOT NULL DEFAULT '{"mention": {"email": true, "in_app": true}, "story_update": {"email": true, "in_app": true}, "comment_reply": {"email": true, "in_app": true}, "story_comment": {"email": true, "in_app": true}, "overdue_stories": {"email": true, "in_app": true}, "objective_update": {"email": true, "in_app": true}, "key_result_update": {"email": true, "in_app": true}}'::jsonb,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT "notification_preferences_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "notification_preferences_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    PRIMARY KEY ("preference_id")
);


-- Indices
CREATE UNIQUE INDEX notification_preferences_user_id_workspace_id_key ON public.notification_preferences USING btree (user_id, workspace_id);
