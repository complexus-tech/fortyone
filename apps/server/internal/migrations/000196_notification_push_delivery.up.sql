CREATE TABLE public.notification_push_devices (
    device_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    expo_push_token text NOT NULL,
    platform text NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    disabled_at timestamptz,
    CONSTRAINT notification_push_devices_pkey PRIMARY KEY (device_id),
    CONSTRAINT notification_push_devices_user_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT notification_push_devices_token_key UNIQUE (expo_push_token),
    CONSTRAINT notification_push_devices_token_length
        CHECK (char_length(expo_push_token) BETWEEN 20 AND 512),
    CONSTRAINT notification_push_devices_platform
        CHECK (platform IN ('ios', 'android'))
);

CREATE INDEX idx_notification_push_devices_active_user
    ON public.notification_push_devices (user_id, device_id)
    WHERE disabled_at IS NULL;

ALTER TABLE public.notifications
    ADD COLUMN push_sent_at timestamptz;

CREATE INDEX idx_notifications_pending_push
    ON public.notifications (notification_id)
    WHERE in_app_enabled = TRUE AND push_sent_at IS NULL;

COMMENT ON TABLE public.notification_push_devices IS
    'Expo push tokens registered by authenticated native app installations. A token can belong to only one account at a time.';

COMMENT ON COLUMN public.notifications.push_sent_at IS
    'Time the push payload was accepted by Expo or deliberately covered because no active device remained.';
