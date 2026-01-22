-- 000031_workspace_settings.up.sql
CREATE TABLE "public"."workspace_settings" (
    "workspace_id" uuid NOT NULL,
    "story_term" varchar(50) NOT NULL DEFAULT 'task'::character varying,
    "sprint_term" varchar(50) NOT NULL DEFAULT 'sprint'::character varying,
    "objective_term" varchar(50) NOT NULL DEFAULT 'objective'::character varying,
    "key_result_term" varchar(50) NOT NULL DEFAULT 'key result'::character varying,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "objective_enabled" bool NOT NULL DEFAULT true,
    "key_result_enabled" bool NOT NULL DEFAULT true,
    CONSTRAINT "workspace_terminology_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("workspace_id")
);


-- Indices
CREATE UNIQUE INDEX workspace_terminology_pkey ON public.workspace_settings USING btree (workspace_id);
