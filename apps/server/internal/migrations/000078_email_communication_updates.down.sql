DROP INDEX IF EXISTS public.idx_notifications_pending_email_digest;

ALTER TABLE public.notifications
    DROP COLUMN IF EXISTS email_sent_at;

UPDATE public.notification_preferences
SET preferences = preferences - 'weekly_digest'
WHERE preferences ? 'weekly_digest';
