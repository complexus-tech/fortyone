-- Figma delivers the webhook passcode inside its JSON envelope. Older
-- application versions persisted that verified bearer credential with the
-- event audit payload. Remove every historical value before enforcing the
-- invariant that future payloads are credential-free.
UPDATE public.figma_webhook_events
SET payload = payload - 'passcode'
WHERE payload ? 'passcode';

ALTER TABLE public.figma_webhook_events
    ADD CONSTRAINT figma_webhook_events_payload_no_passcode_check
    CHECK (NOT (payload ? 'passcode'));
