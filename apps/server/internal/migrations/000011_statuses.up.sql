-- 000011_statuses.up.sql
CREATE TABLE "public"."statuses" (
    "status_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "category" varchar(255),
    "order_index" int4,
    "workspace_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "team_id" uuid NOT NULL,
    "is_default" bool NOT NULL DEFAULT false,
    "color" varchar(16) DEFAULT '#6b665c'::character varying,
    CONSTRAINT "statuses_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "fk_statuses_team" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    PRIMARY KEY ("status_id")
);


-- Indices
CREATE INDEX idx_statuses_category ON public.statuses USING btree (category);
CREATE INDEX idx_statuses_team ON public.statuses USING btree (team_id);
CREATE INDEX idx_statuses_workspace_order ON public.statuses USING btree (workspace_id, order_index);
CREATE UNIQUE INDEX unique_default_status_per_team ON public.statuses USING btree (team_id) WHERE (is_default = true);
