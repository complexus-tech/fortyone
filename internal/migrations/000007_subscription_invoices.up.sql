-- 000007_subscription_invoices.up.sql
-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS subscription_invoices_invoice_id_seq;

-- Table Definition
CREATE TABLE "public"."subscription_invoices" (
    "invoice_id" int4 NOT NULL DEFAULT nextval('subscription_invoices_invoice_id_seq'::regclass),
    "workspace_id" uuid NOT NULL,
    "stripe_invoice_id" text NOT NULL,
    "amount_paid" numeric(10,2) NOT NULL,
    "invoice_date" timestamptz NOT NULL,
    "status" text NOT NULL,
    "seats_count" int4 NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "hosted_url" varchar DEFAULT ''::character varying,
    "customer_name" varchar,
    CONSTRAINT "subscription_invoices_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces"("workspace_id"),
    PRIMARY KEY ("invoice_id")
);


-- Indices
CREATE INDEX idx_subscription_invoices_workspace ON public.subscription_invoices USING btree (workspace_id);
