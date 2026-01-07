-- 000016_sprints.up.sql
CREATE TABLE "public"."sprints" (
    "sprint_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "objective_id" uuid,
    "team_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "name" varchar(255) NOT NULL,
    "start_date" date NOT NULL,
    "end_date" date NOT NULL,
    "goal" text,
    "status" varchar(255),
    "backlog_status" varchar(255),
    "completed_story_count" int4 NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "sprints_objective_id_fkey" FOREIGN KEY ("objective_id") REFERENCES "public"."objectives"("objective_id") ON DELETE SET NULL,
    CONSTRAINT "sprints_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    CONSTRAINT "sprints_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("sprint_id")
);
