ALTER TABLE public.workspace_subscriptions
    DROP CONSTRAINT IF EXISTS workspace_subscriptions_stripe_event_cursor_check,
    DROP COLUMN IF EXISTS last_stripe_event_id,
    DROP COLUMN IF EXISTS last_stripe_event_priority,
    DROP COLUMN IF EXISTS last_stripe_event_created_at;
