-- 000037_github_webhook_events.up.sql
CREATE TABLE "public"."github_webhook_events" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "repository_id" uuid,
    "event_type" text NOT NULL,
    "github_delivery_id" text NOT NULL,
    "payload" jsonb NOT NULL,
    "processed" bool DEFAULT false,
    "processed_at" timestamp,
    "error_message" text,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "github_webhook_events_repository_id_fkey" FOREIGN KEY ("repository_id") REFERENCES "public"."github_repositories"("id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);


-- Indices
CREATE UNIQUE INDEX github_webhook_events_github_delivery_id_key ON public.github_webhook_events USING btree (github_delivery_id);
CREATE INDEX idx_github_webhook_events_processed ON public.github_webhook_events USING btree (processed, created_at);
