-- 000040_team_github_automation_settings.up.sql
CREATE TABLE "public"."team_github_automation_settings" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "team_id" uuid NOT NULL,
    "event_type" varchar(50) NOT NULL,
    "target_status_id" uuid,
    "enabled" bool DEFAULT true,
    "branch_pattern" varchar(255),
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "team_github_automation_settings_target_status_id_fkey" FOREIGN KEY ("target_status_id") REFERENCES "public"."statuses"("status_id"),
    CONSTRAINT "team_github_automation_settings_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id"),
    PRIMARY KEY ("id")
);
