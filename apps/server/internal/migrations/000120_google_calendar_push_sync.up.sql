ALTER TABLE public.calendar_connections
    ADD COLUMN sync_token text,
    ADD COLUMN notification_channel_id text,
    ADD COLUMN notification_resource_id text,
    ADD COLUMN notification_expires_at timestamptz;

CREATE UNIQUE INDEX calendar_connections_notification_channel_unique
    ON public.calendar_connections (notification_channel_id)
    WHERE notification_channel_id IS NOT NULL
      AND revoked_at IS NULL;

