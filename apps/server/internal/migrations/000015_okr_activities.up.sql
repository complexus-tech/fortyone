-- 000015_okr_activities.up.sql
DROP TYPE IF EXISTS "public"."okr_activity_type";
CREATE TYPE "public"."okr_activity_type" AS ENUM ('create', 'update', 'delete');
DROP TYPE IF EXISTS "public"."okr_update_type";
CREATE TYPE "public"."okr_update_type" AS ENUM ('objective', 'key_result');

-- Table Definition
CREATE TABLE "public"."okr_activities" (
    "activity_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "objective_id" uuid NOT NULL,
    "key_result_id" uuid,
    "user_id" uuid NOT NULL,
    "activity_type" "public"."okr_activity_type" NOT NULL,
    "update_type" "public"."okr_update_type" NOT NULL,
    "field_changed" varchar(100),
    "current_value" text,
    "comment" text,
    "created_at" timestamp DEFAULT now(),
    "workspace_id" uuid NOT NULL,
    CONSTRAINT "okr_activities_objective_id_fkey" FOREIGN KEY ("objective_id") REFERENCES "public"."objectives"("objective_id") ON DELETE CASCADE,
    CONSTRAINT "okr_activities_key_result_id_fkey" FOREIGN KEY ("key_result_id") REFERENCES "public"."key_results"("id") ON DELETE SET NULL,
    CONSTRAINT "okr_activities_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id"),
    CONSTRAINT "okr_activities_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id"),
    PRIMARY KEY ("activity_id")
);
