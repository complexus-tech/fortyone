DROP TRIGGER IF EXISTS outbound_webhook_audit_events_immutable
    ON public.outbound_webhook_audit_events;
DROP TRIGGER IF EXISTS outbound_webhook_delivery_attempts_immutable
    ON public.outbound_webhook_delivery_attempts;
DROP FUNCTION IF EXISTS public.reject_outbound_webhook_immutable_mutation();

DROP TABLE IF EXISTS public.outbound_webhook_audit_events;
DROP TABLE IF EXISTS public.outbound_webhook_delivery_attempts;
DROP TABLE IF EXISTS public.outbound_webhook_deliveries;
DROP TABLE IF EXISTS public.outbound_webhook_events;
DROP TABLE IF EXISTS public.outbound_webhook_subscriptions;
DROP TABLE IF EXISTS public.outbound_webhook_endpoints;
