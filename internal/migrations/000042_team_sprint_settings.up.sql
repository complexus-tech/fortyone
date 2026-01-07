-- 000042_team_sprint_settings.up.sql
CREATE TABLE "public"."team_sprint_settings" (
    "team_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "auto_create_sprints" bool NOT NULL DEFAULT false,
    "upcoming_sprints_count" int4 NOT NULL DEFAULT 2 CHECK ((upcoming_sprints_count >= 0) AND (upcoming_sprints_count <= 10)),
    "sprint_duration_weeks" int4 NOT NULL DEFAULT 2 CHECK ((sprint_duration_weeks >= 1) AND (sprint_duration_weeks <= 8)),
    "sprint_start_day" varchar(10) NOT NULL DEFAULT 'Monday'::character varying CHECK ((sprint_start_day)::text = ANY ((ARRAY['Monday'::character varying, 'Tuesday'::character varying, 'Wednesday'::character varying, 'Thursday'::character varying, 'Friday'::character varying, 'Saturday'::character varying, 'Sunday'::character varying])::text[])),
    "move_incomplete_stories_enabled" bool NOT NULL DEFAULT true,
    "last_auto_sprint_number" int4 NOT NULL DEFAULT 0 CHECK (last_auto_sprint_number >= 0),
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    CONSTRAINT "team_sprint_settings_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    CONSTRAINT "team_sprint_settings_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("team_id","workspace_id")
);
