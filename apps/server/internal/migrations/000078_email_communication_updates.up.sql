ALTER TABLE public.notifications
    ADD COLUMN IF NOT EXISTS email_sent_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_notifications_pending_email_digest
    ON public.notifications (recipient_id, workspace_id, created_at)
    WHERE read_at IS NULL AND email_sent_at IS NULL;

UPDATE public.notification_preferences
SET preferences = jsonb_set(
    jsonb_set(
        preferences,
        '{reminders}',
        COALESCE(
            preferences -> 'reminders',
            preferences -> 'overdue_stories',
            '{"email": true, "in_app": true}'::jsonb
        ),
        true
    ),
    '{weekly_digest}',
    COALESCE(
        preferences -> 'weekly_digest',
        '{"email": true, "in_app": true}'::jsonb
    ),
    true
)
WHERE NOT preferences ? 'reminders'
    OR NOT preferences ? 'weekly_digest';
