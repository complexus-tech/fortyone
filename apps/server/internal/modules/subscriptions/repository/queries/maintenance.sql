-- name: PurgeTerminalStripeWebhookEvents :execrows
WITH candidates AS MATERIALIZED (
    SELECT webhook.event_id
    FROM public.stripe_webhook_events AS webhook
    WHERE (
        webhook.processing_state = 'processed'
        AND webhook.processed_at IS NOT NULL
        AND webhook.processed_at AT TIME ZONE 'UTC' < sqlc.arg(terminal_before)
    ) OR (
        webhook.processing_state = 'failed'
        AND webhook.failed_at IS NOT NULL
        AND webhook.failed_at < sqlc.arg(terminal_before)
    )
    ORDER BY
        CASE webhook.processing_state
            WHEN 'processed' THEN webhook.processed_at AT TIME ZONE 'UTC'
            ELSE webhook.failed_at
        END,
        webhook.event_id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF webhook SKIP LOCKED
)
DELETE FROM public.stripe_webhook_events AS webhook
USING candidates
WHERE webhook.event_id = candidates.event_id
  AND webhook.processing_state IN ('processed', 'failed');
