-- Stripe does not guarantee webhook delivery order. Persist the last applied
-- provider cursor on each subscription so an older delivery cannot overwrite a
-- newer billing state. Priority is used only for events created in the same
-- second; terminal deletion outranks snapshot updates.
ALTER TABLE public.workspace_subscriptions
    ADD COLUMN last_stripe_event_created_at timestamptz,
    ADD COLUMN last_stripe_event_priority smallint,
    ADD COLUMN last_stripe_event_id varchar(255),
    ADD CONSTRAINT workspace_subscriptions_stripe_event_cursor_check
        CHECK (
            (
                last_stripe_event_created_at IS NULL
                AND last_stripe_event_priority IS NULL
                AND last_stripe_event_id IS NULL
            )
            OR (
                last_stripe_event_created_at IS NOT NULL
                AND last_stripe_event_priority IS NOT NULL
                AND last_stripe_event_priority >= 0
                AND last_stripe_event_id IS NOT NULL
                AND length(last_stripe_event_id) > 0
            )
        );
