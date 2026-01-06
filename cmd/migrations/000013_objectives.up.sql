-- 000013_objectives.up.sql
DROP TYPE IF EXISTS "public"."objective_health_status";
CREATE TYPE "public"."objective_health_status" AS ENUM ('At Risk', 'On Track', 'Off Track');

-- Table Definition
CREATE TABLE "public"."objectives" (
    "objective_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "description" text,
    "lead_user_id" uuid,
    "team_id" uuid,
    "workspace_id" uuid,
    "start_date" date,
    "end_date" date,
    "is_private" bool NOT NULL DEFAULT false,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "status_id" uuid,
    "priority" varchar(100),
    "health" "public"."objective_health_status",
    "created_by" uuid,
    "search_vector" tsvector,
    CONSTRAINT "objectives_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id"),
    CONSTRAINT "objectives_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "objectives_status_id_fkey" FOREIGN KEY ("status_id") REFERENCES "public"."objective_statuses"("status_id") ON DELETE SET NULL,
    CONSTRAINT "objectives_lead_user_id_fkey" FOREIGN KEY ("lead_user_id") REFERENCES "public"."users"("user_id"),
    CONSTRAINT "objectives_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users"("user_id"),
    PRIMARY KEY ("objective_id")
);


-- Indices
CREATE INDEX idx_objectives_lead_user_id ON public.objectives USING btree (lead_user_id);
CREATE INDEX idx_objectives_name_trigram ON public.objectives USING gin (name gin_trgm_ops);
CREATE INDEX idx_objectives_search ON public.objectives USING gin (search_vector);
CREATE INDEX idx_objectives_search_workspace_team ON public.objectives USING btree (workspace_id, team_id);
CREATE UNIQUE INDEX objectives_name_team_unique ON public.objectives USING btree (name, team_id);
