-- Stripe can deliver the same event concurrently and retry it after transient
-- failures. Persist a short-lived ownership lease before executing domain
-- effects so only one worker handles a delivery at a time, while crashed
-- workers can be replaced after the lease expires.
ALTER TABLE public.stripe_webhook_events
    ALTER COLUMN processed_at DROP DEFAULT,
    ALTER COLUMN processed_at DROP NOT NULL,
    ADD COLUMN processing_state varchar(16),
    ADD COLUMN processing_result varchar(16),
    ADD COLUMN attempts integer,
    ADD COLUMN first_received_at timestamptz,
    ADD COLUMN last_attempted_at timestamptz,
    ADD COLUMN lease_expires_at timestamptz,
    ADD COLUMN lease_token uuid,
    ADD COLUMN failed_at timestamptz,
    ADD COLUMN last_error_code varchar(64);

-- Rows written by the previous implementation represented successfully
-- terminal events, even though they did not carry explicit state metadata.
UPDATE public.stripe_webhook_events
SET processing_state = 'processed',
    processing_result = 'handled',
    attempts = 1,
    first_received_at = processed_at AT TIME ZONE 'UTC',
    last_attempted_at = processed_at AT TIME ZONE 'UTC';

ALTER TABLE public.stripe_webhook_events
    ALTER COLUMN processing_state SET NOT NULL,
    ALTER COLUMN attempts SET NOT NULL,
    ALTER COLUMN first_received_at SET NOT NULL,
    ALTER COLUMN last_attempted_at SET NOT NULL,
    DROP COLUMN payload,
    ADD CONSTRAINT stripe_webhook_events_processing_state_check
        CHECK (processing_state IN ('processing', 'processed', 'failed')),
    ADD CONSTRAINT stripe_webhook_events_processing_result_check
        CHECK (processing_result IS NULL OR processing_result IN ('handled', 'ignored')),
    ADD CONSTRAINT stripe_webhook_events_attempts_check
        CHECK (attempts > 0),
    ADD CONSTRAINT stripe_webhook_events_state_metadata_check
        CHECK (
            (
                processing_state = 'processing'
                AND processed_at IS NULL
                AND processing_result IS NULL
                AND lease_expires_at IS NOT NULL
                AND lease_token IS NOT NULL
                AND failed_at IS NULL
                AND last_error_code IS NULL
            )
            OR (
                processing_state = 'processed'
                AND processed_at IS NOT NULL
                AND processing_result IS NOT NULL
                AND lease_expires_at IS NULL
                AND lease_token IS NULL
                AND failed_at IS NULL
                AND last_error_code IS NULL
            )
            OR (
                processing_state = 'failed'
                AND processed_at IS NULL
                AND processing_result IS NULL
                AND lease_expires_at IS NULL
                AND lease_token IS NULL
                AND failed_at IS NOT NULL
                AND last_error_code IS NOT NULL
            )
        );

CREATE INDEX idx_stripe_webhook_events_retryable
    ON public.stripe_webhook_events (processing_state, lease_expires_at, last_attempted_at);

-- Crash recovery can replay domain handling after its database write committed
-- but before the webhook event reached its terminal state. Stripe invoice IDs
-- are globally unique, so collapse any historical duplicate rows and enforce
-- idempotency for future replays.
DELETE FROM public.subscription_invoices AS older
USING public.subscription_invoices AS newer
WHERE older.stripe_invoice_id = newer.stripe_invoice_id
  AND older.invoice_id < newer.invoice_id;

CREATE UNIQUE INDEX subscription_invoices_stripe_invoice_id_key
    ON public.subscription_invoices (stripe_invoice_id);
