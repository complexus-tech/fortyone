-- 000008_stripe_webhook_events.up.sql
CREATE TABLE "public"."stripe_webhook_events" (
    "event_id" varchar(255) NOT NULL,
    "event_type" varchar(255) NOT NULL,
    "workspace_id" uuid,
    "processed_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "payload" jsonb,
    CONSTRAINT "fk_webhook_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id"),
    PRIMARY KEY ("event_id")
);


-- Indices
CREATE INDEX idx_webhook_workspace ON public.stripe_webhook_events USING btree (workspace_id);
CREATE INDEX idx_webhook_event_type ON public.stripe_webhook_events USING btree (event_type);
