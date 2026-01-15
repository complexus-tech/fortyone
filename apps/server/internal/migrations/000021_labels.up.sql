-- 000021_labels.up.sql
CREATE TABLE "public"."labels" (
    "label_id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(100) NOT NULL,
    "team_id" uuid,
    "workspace_id" uuid,
    "color" varchar(56) DEFAULT 'red'::character varying,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "labels_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "labels_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id"),
    PRIMARY KEY ("label_id")
);


-- Indices
CREATE INDEX idx_labels_label_id ON public.labels USING btree (label_id);
CREATE INDEX idx_labels_team_created ON public.labels USING btree (team_id, created_at DESC);
CREATE INDEX idx_labels_team_workspace ON public.labels USING btree (team_id, workspace_id);
CREATE INDEX idx_labels_workspace_created ON public.labels USING btree (workspace_id, created_at DESC);
