-- 000009_teams.up.sql
CREATE TABLE "public"."teams" (
    "team_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "workspace_id" uuid NOT NULL,
    "code" varchar(255) NOT NULL,
    "color" varchar(100) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "is_private" bool NOT NULL DEFAULT false,
    CONSTRAINT "teams_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("team_id")
);


-- Indices
CREATE UNIQUE INDEX teams_workspace_id_code_key ON public.teams USING btree (workspace_id, code);
CREATE INDEX idx_teams_created ON public.teams USING btree (created_at DESC);
CREATE INDEX idx_teams_workspace_created ON public.teams USING btree (workspace_id, created_at DESC);
