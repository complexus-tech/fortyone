-- 000047_user_team_orders.up.sql
CREATE TABLE "public"."user_team_orders" (
    "user_id" uuid NOT NULL,
    "team_id" uuid NOT NULL,
    "workspace_id" uuid NOT NULL,
    "order_index" int4 NOT NULL,
    "created_at" timestamp DEFAULT now(),
    "updated_at" timestamp DEFAULT now(),
    CONSTRAINT "user_team_orders_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("user_id") ON DELETE CASCADE,
    CONSTRAINT "user_team_orders_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    CONSTRAINT "user_team_orders_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    PRIMARY KEY ("user_id","team_id","workspace_id")
);


-- Indices
CREATE INDEX idx_user_team_orders_user_workspace ON public.user_team_orders USING btree (user_id, workspace_id, order_index);
CREATE INDEX idx_user_team_orders_team ON public.user_team_orders USING btree (team_id);
