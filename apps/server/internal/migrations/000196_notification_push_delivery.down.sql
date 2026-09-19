DROP INDEX IF EXISTS public.idx_notifications_pending_push;

ALTER TABLE public.notifications
    DROP COLUMN IF EXISTS push_sent_at;

DROP TABLE IF EXISTS public.notification_push_devices;
