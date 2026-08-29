-- Rollback is data-preserving only before adoption or after every adopting
-- route has been disabled and the bounded retention window has drained.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.api_idempotency_receipts) THEN
        RAISE EXCEPTION '000156 cannot be reversed while idempotency receipts exist; disable adopting routes, wait for retention, purge expired rows, and retry';
    END IF;
END
$$;

DROP TABLE public.api_idempotency_receipts;
