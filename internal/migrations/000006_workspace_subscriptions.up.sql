-- 000006_workspace_subscriptions.up.sql
DROP TYPE IF EXISTS "public"."subscription_tier_enum";
CREATE TYPE "public"."subscription_tier_enum" AS ENUM ('free', 'pro', 'business', 'enterprise');
DROP TYPE IF EXISTS "public"."billing_interval_enum";
CREATE TYPE "public"."billing_interval_enum" AS ENUM ('day', 'week', 'month', 'year');

-- Table Definition
CREATE TABLE "public"."workspace_subscriptions" (
    "workspace_id" uuid NOT NULL,
    "stripe_customer_id" text NOT NULL,
    "stripe_subscription_id" text NOT NULL,
    "subscription_status" text,
    "seat_count" int4 NOT NULL DEFAULT 0,
    "trial_end_date" timestamptz,
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now(),
    "stripe_subscription_item_id" varchar(255),
    "subscription_tier" "public"."subscription_tier_enum" DEFAULT 'free'::subscription_tier_enum,
    "billing_interval" "public"."billing_interval_enum",
    "billing_ends_at" timestamptz,
    CONSTRAINT "workspace_subscriptions_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id") ON DELETE CASCADE,
    PRIMARY KEY ("stripe_subscription_id")
);


-- Indices
CREATE INDEX idx_workspace_subscriptions_customer_id ON public.workspace_subscriptions USING btree (stripe_customer_id);
