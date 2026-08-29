DROP INDEX IF EXISTS public.idx_notifications_recipient_entity_unread;
DROP INDEX IF EXISTS public.idx_notifications_recipient_entity_created;
DROP INDEX IF EXISTS public.idx_notifications_dedupe_key;

ALTER TABLE public.notifications
    DROP COLUMN IF EXISTS dedupe_key;

-- PostgreSQL enum values cannot be removed without rebuilding the enum and all
-- dependent columns. Keeping those values is backward compatible on rollback.
