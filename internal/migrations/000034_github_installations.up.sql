-- 000034_github_installations.up.sql
CREATE TABLE "public"."github_installations" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "workspace_id" uuid NOT NULL,
    "github_app_id" int8 NOT NULL,
    "installation_id" int8 NOT NULL,
    "account_type" varchar(20) NOT NULL,
    "account_login" varchar(255) NOT NULL,
    "account_id" int8 NOT NULL,
    "repository_selection" varchar(10) NOT NULL,
    "is_active" bool NOT NULL DEFAULT true,
    "suspended_at" timestamptz,
    "suspended_by" varchar(255),
    "permissions" jsonb,
    "events" _text,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT "github_installations_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);

-- Indices
CREATE UNIQUE INDEX github_installations_installation_id_key ON public.github_installations USING btree (installation_id);
CREATE UNIQUE INDEX github_installations_workspace_id_key ON public.github_installations USING btree (workspace_id);
CREATE INDEX idx_github_installations_workspace_id ON public.github_installations USING btree (workspace_id);
CREATE INDEX idx_github_installations_installation_id ON public.github_installations USING btree (installation_id);
CREATE INDEX idx_github_installations_account ON public.github_installations USING btree (account_type, account_login);
