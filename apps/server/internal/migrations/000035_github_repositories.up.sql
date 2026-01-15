-- 000035_github_repositories.up.sql
CREATE TABLE "public"."github_repositories" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "installation_id" int8 NOT NULL,
    "workspace_id" uuid NOT NULL,
    "github_repo_id" int8 NOT NULL,
    "name" text NOT NULL,
    "full_name" text NOT NULL,
    "description" text,
    "private" bool NOT NULL DEFAULT false,
    "default_branch" text NOT NULL DEFAULT 'main'::text,
    "clone_url" text NOT NULL,
    "ssh_url" text NOT NULL,
    "webhook_id" int8,
    "webhook_secret" text,
    "is_active" bool DEFAULT true,
    "last_synced_at" timestamp,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "github_repositories_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE INDEX idx_github_repositories_workspace ON public.github_repositories USING btree (workspace_id);
CREATE UNIQUE INDEX github_repositories_workspace_id_github_repo_id_key ON public.github_repositories USING btree (workspace_id, github_repo_id);
CREATE INDEX idx_github_repositories_integration ON public.github_repositories USING btree (installation_id);
