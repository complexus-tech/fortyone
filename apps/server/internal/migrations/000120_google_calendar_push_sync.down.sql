DROP INDEX IF EXISTS public.calendar_connections_notification_channel_unique;

ALTER TABLE public.calendar_connections
    DROP COLUMN IF EXISTS notification_expires_at,
    DROP COLUMN IF EXISTS notification_resource_id,
    DROP COLUMN IF EXISTS notification_channel_id,
    DROP COLUMN IF EXISTS sync_token;
