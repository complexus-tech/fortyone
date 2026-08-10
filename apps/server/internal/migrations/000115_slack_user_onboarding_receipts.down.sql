-- Losing the durable receipts would cause existing users to receive the
-- first-use guide again after a reapply. Require an explicit data decision
-- instead of silently weakening the only-once contract during rollback.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.slack_user_onboarding_receipts
    ) OR EXISTS (
        SELECT 1
        FROM public.messaging_outbound_deliveries
        WHERE purpose = 'onboarding'
    ) THEN
        RAISE EXCEPTION 'migration 000115 cannot be rolled back while Slack onboarding history exists';
    END IF;
END
$$;

DROP TRIGGER IF EXISTS messaging_outbound_slack_user_onboarding_receipt
    ON public.messaging_outbound_deliveries;

DROP FUNCTION IF EXISTS public.record_slack_user_onboarding_receipt();

DROP TABLE IF EXISTS public.slack_user_onboarding_receipts;

ALTER TABLE public.messaging_outbound_deliveries
    DROP CONSTRAINT IF EXISTS messaging_outbound_deliveries_onboarding_recipient_check,
    DROP CONSTRAINT messaging_outbound_deliveries_purpose_check,
    ADD CONSTRAINT messaging_outbound_deliveries_purpose_check
    CHECK (purpose IN ('provider_message', 'assistant', 'account_link', 'access', 'creation_confirmation'));
