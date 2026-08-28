DROP INDEX IF EXISTS public.idx_stripe_webhook_events_retryable;
DROP INDEX IF EXISTS public.subscription_invoices_stripe_invoice_id_key;

ALTER TABLE public.stripe_webhook_events
    DROP CONSTRAINT IF EXISTS stripe_webhook_events_state_metadata_check,
    DROP CONSTRAINT IF EXISTS stripe_webhook_events_attempts_check,
    DROP CONSTRAINT IF EXISTS stripe_webhook_events_processing_result_check,
    DROP CONSTRAINT IF EXISTS stripe_webhook_events_processing_state_check,
    ADD COLUMN payload jsonb;

-- The legacy schema represented every row as processed. Preserve downgrade
-- compatibility by assigning the latest known attempt time to retryable rows.
UPDATE public.stripe_webhook_events
SET processed_at = COALESCE(
    processed_at,
    failed_at AT TIME ZONE 'UTC',
    last_attempted_at AT TIME ZONE 'UTC',
    first_received_at AT TIME ZONE 'UTC',
    CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
);

ALTER TABLE public.stripe_webhook_events
    ALTER COLUMN processed_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN processed_at SET NOT NULL,
    DROP COLUMN last_error_code,
    DROP COLUMN failed_at,
    DROP COLUMN lease_token,
    DROP COLUMN lease_expires_at,
    DROP COLUMN last_attempted_at,
    DROP COLUMN first_received_at,
    DROP COLUMN attempts,
    DROP COLUMN processing_result,
    DROP COLUMN processing_state;
