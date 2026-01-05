-- 000006_workspace_subscriptions.up.sql
CREATE TABLE public.workspace_subscriptions (
    workspace_id uuid NOT NULL,
    stripe_customer_id text NOT NULL,
    stripe_subscription_id text NOT NULL,
    subscription_status text,
    seat_count integer DEFAULT 0 NOT NULL,
    trial_end_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    stripe_subscription_item_id character varying(255),
    subscription_tier public.subscription_tier_enum DEFAULT 'free'::public.subscription_tier_enum,
    billing_interval public.billing_interval_enum,
    billing_ends_at timestamp with time zone
);

ALTER TABLE ONLY public.workspace_subscriptions ADD CONSTRAINT workspace_subscriptions_pkey PRIMARY KEY (stripe_subscription_id);

ALTER TABLE ONLY public.workspace_subscriptions 
    ADD CONSTRAINT workspace_subscriptions_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_workspace_subscriptions_customer_id ON public.workspace_subscriptions USING btree (stripe_customer_id);
