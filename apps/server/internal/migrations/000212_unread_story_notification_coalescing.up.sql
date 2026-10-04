ALTER TABLE public.notifications
    ADD COLUMN event_at timestamptz,
    ADD COLUMN coalescing_key text NOT NULL DEFAULT 'story_update';

UPDATE public.notifications SET event_at = COALESCE(created_at, CURRENT_TIMESTAMP);
ALTER TABLE public.notifications
    ALTER COLUMN event_at SET NOT NULL,
    ALTER COLUMN event_at SET DEFAULT CURRENT_TIMESTAMP;

-- Event identity and original payload survive inbox edits and deletion.
CREATE TABLE public.notification_event_receipts (
    dedupe_key text PRIMARY KEY,
    recipient_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    actor_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    entity_id uuid NOT NULL,
    notification_type public.notification_type NOT NULL,
    entity_type public.entity_type NOT NULL,
    coalescing_key text NOT NULL,
    occurred_at timestamptz NOT NULL,
    payload jsonb NOT NULL,
    notification_id uuid REFERENCES public.notifications(notification_id) ON DELETE SET NULL
);

INSERT INTO public.notification_event_receipts (
    dedupe_key, recipient_id, workspace_id, actor_id, entity_id, notification_type,
    entity_type, coalescing_key, occurred_at, payload, notification_id
)
SELECT dedupe_key, recipient_id, workspace_id, actor_id, entity_id, type, entity_type,
       coalescing_key, event_at,
       jsonb_build_array(type, entity_type, entity_id, actor_id, title, message),
       notification_id
FROM public.notifications;

CREATE INDEX idx_notification_event_receipts_story_latest
    ON public.notification_event_receipts (recipient_id, workspace_id, entity_id, coalescing_key, occurred_at DESC)
    WHERE notification_type = 'story_update' AND entity_type = 'story';

CREATE INDEX idx_notification_event_receipts_recipient ON public.notification_event_receipts (recipient_id);
CREATE INDEX idx_notification_event_receipts_actor ON public.notification_event_receipts (actor_id);

-- Consolidate redundant unread routine rows, retaining every original payload
-- in the immutable receipts. Read history and discussion notifications remain.
WITH ranked AS (
    SELECT notification_id,
           ROW_NUMBER() OVER (
               PARTITION BY recipient_id, workspace_id, entity_id, coalescing_key
               ORDER BY created_at DESC NULLS LAST, notification_id DESC
           ) AS position
    FROM public.notifications
    WHERE type = 'story_update' AND entity_type = 'story' AND read_at IS NULL
)
DELETE FROM public.notifications AS notification
USING ranked
WHERE notification.notification_id = ranked.notification_id AND ranked.position > 1;

CREATE UNIQUE INDEX idx_notifications_unread_story_update
    ON public.notifications (recipient_id, workspace_id, entity_id, coalescing_key)
    WHERE type = 'story_update' AND entity_type = 'story' AND read_at IS NULL;

COMMENT ON COLUMN public.notifications.event_at IS
    'Producer event time used to prevent delayed or replayed updates from replacing newer inbox content.';
