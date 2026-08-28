-- Secret redaction is intentionally forward-only once an event exists. A
-- rollback must never make an older application able to persist passcodes
-- beside retained webhook audit events.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.figma_webhook_events) THEN
        RAISE EXCEPTION '000165 cannot be reversed after Figma webhook events exist; deploy a forward fix';
    END IF;
END
$$;

ALTER TABLE public.figma_webhook_events
    DROP CONSTRAINT IF EXISTS figma_webhook_events_payload_no_passcode_check;
