-- 000043_team_story_automation_settings.up.sql
CREATE TABLE "public"."team_story_automation_settings" (
    "team_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "auto_close_inactive_enabled" bool NOT NULL DEFAULT true,
    "auto_close_inactive_months" int4 NOT NULL DEFAULT 3 CHECK ((auto_close_inactive_months >= 1) AND (auto_close_inactive_months <= 24)),
    "auto_archive_enabled" bool NOT NULL DEFAULT true,
    "auto_archive_months" int4 NOT NULL DEFAULT 3 CHECK ((auto_archive_months >= 1) AND (auto_archive_months <= 24)),
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    CONSTRAINT "team_story_automation_settings_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "team_story_automation_settings_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    PRIMARY KEY ("team_id","workspace_id")
);


-- Indices
CREATE INDEX idx_team_story_automation_settings_auto_close ON public.team_story_automation_settings USING btree (auto_close_inactive_enabled) WHERE (auto_close_inactive_enabled = true);
CREATE INDEX idx_team_story_automation_settings_auto_archive ON public.team_story_automation_settings USING btree (auto_archive_enabled) WHERE (auto_archive_enabled = true);
