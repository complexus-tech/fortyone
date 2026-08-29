DROP TRIGGER IF EXISTS messaging_outbound_comment_delivery_status
    ON public.messaging_outbound_deliveries;

DROP FUNCTION IF EXISTS public.sync_integration_request_comment_delivery_status();

DROP INDEX IF EXISTS public.integration_request_comments_client_idempotency_key;

ALTER TABLE public.integration_request_comments
    DROP CONSTRAINT IF EXISTS integration_request_comments_delivery_status_check,
    DROP CONSTRAINT IF EXISTS integration_request_comments_client_idempotency_check,
    DROP COLUMN IF EXISTS delivery_status,
    DROP COLUMN IF EXISTS client_idempotency_key;
