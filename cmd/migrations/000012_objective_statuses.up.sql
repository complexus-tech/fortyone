-- 000012_objective_statuses.up.sql
CREATE TABLE "public"."objective_statuses" (
    "status_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "category" varchar(255),
    "order_index" int4,
    "workspace_id" uuid NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP,
    "is_default" bool NOT NULL DEFAULT false,
    "color" varchar(16) DEFAULT '#6b665c'::character varying,
    CONSTRAINT "objective_statuses_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("status_id")
);


-- Indices
CREATE INDEX idx_objective_statuses_status_id ON public.objective_statuses USING btree (status_id);
CREATE INDEX idx_objective_statuses_workspace_name ON public.objective_statuses USING btree (workspace_id, name);
CREATE INDEX idx_objective_statuses_workspace_order ON public.objective_statuses USING btree (workspace_id, order_index);
CREATE UNIQUE INDEX unique_default_status_per_workspace ON public.objective_statuses USING btree (workspace_id) WHERE (is_default = true);
