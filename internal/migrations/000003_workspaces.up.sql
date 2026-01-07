-- 000003_workspaces.up.sql
CREATE TABLE "public"."workspaces" (
    "workspace_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "slug" varchar(255) NOT NULL,
    "color" varchar(50) NOT NULL DEFAULT '#EA6060'::character varying,
    "team_size" varchar(20) NOT NULL DEFAULT '2-10'::character varying,
    "trial_ends_on" timestamptz,
    "avatar_url" varchar(255),
    "deleted_at" timestamp,
    "deleted_by" uuid,
    "last_accessed_at" timestamp DEFAULT now(),
    "inactivity_warning_sent_at" timestamp,
    "created_by" uuid,
    CONSTRAINT "workspaces_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users"("user_id") ON DELETE SET NULL,
    CONSTRAINT "workspaces_deleted_by_fkey" FOREIGN KEY ("deleted_by") REFERENCES "public"."users"("user_id"),
    PRIMARY KEY ("workspace_id")
);


-- Indices
CREATE INDEX idx_workspaces_created_at ON public.workspaces USING btree (created_at);
CREATE UNIQUE INDEX workspaces_slug_key ON public.workspaces USING btree (slug);
