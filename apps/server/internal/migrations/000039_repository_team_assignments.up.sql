-- 000039_repository_team_assignments.up.sql
CREATE TABLE "public"."repository_team_assignments" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "repository_id" uuid NOT NULL,
    "team_id" uuid NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "repository_team_assignments_repository_id_fkey" FOREIGN KEY ("repository_id") REFERENCES "public"."github_repositories"("id") ON DELETE CASCADE,
    CONSTRAINT "repository_team_assignments_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE UNIQUE INDEX repository_team_assignments_repository_id_team_id_key ON public.repository_team_assignments USING btree (repository_id, team_id);
