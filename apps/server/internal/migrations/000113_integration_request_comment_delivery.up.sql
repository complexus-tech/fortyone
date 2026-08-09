ALTER TABLE public.integration_request_comments
    ADD COLUMN client_idempotency_key uuid,
    ADD COLUMN delivery_status text;

UPDATE public.integration_request_comments
SET client_idempotency_key = id
WHERE direction = 'outbound';

UPDATE public.integration_request_comments irc
SET delivery_status = CASE
    WHEN mod.status = 'delivered' THEN 'sent'
    WHEN mod.status IN ('pending', 'delivering') THEN 'sending'
    WHEN mod.status = 'failed' AND mod.attempt_count < 20 THEN 'retrying'
    WHEN mod.status = 'failed' THEN 'failed'
    WHEN mod.status = 'cancelled' THEN 'not-sent'
    ELSE 'not-sent'
END
FROM public.messaging_outbound_deliveries mod
WHERE irc.direction = 'outbound'
  AND mod.workspace_id = irc.workspace_id
  AND mod.idempotency_key = irc.outbound_idempotency_key;

UPDATE public.integration_request_comments
SET delivery_status = 'not-sent'
WHERE direction = 'outbound'
  AND delivery_status IS NULL;

ALTER TABLE public.integration_request_comments
    ADD CONSTRAINT integration_request_comments_client_idempotency_check
    CHECK (
        (direction = 'outbound' AND client_idempotency_key IS NOT NULL)
        OR (direction = 'inbound' AND client_idempotency_key IS NULL)
    ),
    ADD CONSTRAINT integration_request_comments_delivery_status_check
    CHECK (
        (direction = 'outbound' AND delivery_status IN ('sent', 'sending', 'retrying', 'failed', 'not-sent'))
        OR (direction = 'inbound' AND delivery_status IS NULL)
    );

CREATE UNIQUE INDEX integration_request_comments_client_idempotency_key
    ON public.integration_request_comments USING btree (workspace_id, client_idempotency_key)
    WHERE client_idempotency_key IS NOT NULL;

CREATE OR REPLACE FUNCTION public.sync_integration_request_comment_delivery_status()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE public.integration_request_comments
    SET delivery_status = CASE
        WHEN NEW.status = 'delivered' THEN 'sent'
        WHEN NEW.status IN ('pending', 'delivering') THEN 'sending'
        WHEN NEW.status = 'failed' AND NEW.attempt_count < 20 THEN 'retrying'
        WHEN NEW.status = 'failed' THEN 'failed'
        WHEN NEW.status = 'cancelled' THEN 'not-sent'
        ELSE delivery_status
    END,
        updated_at = NOW()
    WHERE workspace_id = NEW.workspace_id
      AND outbound_idempotency_key = NEW.idempotency_key;

    RETURN NEW;
END;
$$;

CREATE TRIGGER messaging_outbound_comment_delivery_status
AFTER INSERT OR UPDATE OF status, attempt_count
ON public.messaging_outbound_deliveries
FOR EACH ROW
EXECUTE FUNCTION public.sync_integration_request_comment_delivery_status();
