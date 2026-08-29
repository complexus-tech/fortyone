-- Rolling back this migration destroys Maya's durable conversation history and
-- restores a purpose constraint that rejects retained email reply deliveries.
-- Require an explicit retention decision whenever either form of history exists.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.messaging_outbound_deliveries
        WHERE purpose = 'email_reply'
    ) OR EXISTS (
        SELECT 1
        FROM public.messaging_email_threads
    ) OR EXISTS (
        SELECT 1
        FROM public.messaging_email_reply_tokens
    ) OR EXISTS (
        SELECT 1
        FROM public.messaging_email_messages
    ) OR EXISTS (
        SELECT 1
        FROM public.messaging_email_action_proposals
    ) THEN
        RAISE EXCEPTION 'migration 000118 cannot be rolled back while Maya email conversation history exists';
    END IF;
END
$$;

DROP TABLE IF EXISTS public.messaging_email_action_proposals;
DROP TABLE IF EXISTS public.messaging_email_messages;
DROP TABLE IF EXISTS public.messaging_email_reply_tokens;
DROP TABLE IF EXISTS public.messaging_email_threads;

ALTER TABLE public.messaging_outbound_deliveries
    DROP CONSTRAINT messaging_outbound_deliveries_purpose_check,
    ADD CONSTRAINT messaging_outbound_deliveries_purpose_check
    CHECK (
        purpose IN (
            'provider_message',
            'assistant',
            'account_link',
            'access',
            'creation_confirmation',
            'onboarding'
        )
    );
