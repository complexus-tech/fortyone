-- Retain email coverage independently of deletable inbox notifications.
CREATE TABLE public.notification_email_receipts (
    recipient_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    content_hash bytea NOT NULL CHECK (octet_length(content_hash) = 32),
    sent_at timestamptz NOT NULL,
    PRIMARY KEY (recipient_id, workspace_id, content_hash)
);

-- Protect existing deliveries as well as mail sent after this migration.
INSERT INTO public.notification_email_receipts (recipient_id, workspace_id, content_hash, sent_at)
SELECT notification.recipient_id, notification.workspace_id,
       sha256(convert_to(CAST(jsonb_build_array(notification.type, notification.entity_type, notification.entity_id, notification.actor_id, notification.title, notification.message) AS text), 'UTF8')), MIN(notification.email_sent_at)
FROM public.notifications AS notification
WHERE notification.email_sent_at IS NOT NULL
GROUP BY notification.recipient_id, notification.workspace_id, sha256(convert_to(CAST(jsonb_build_array(notification.type, notification.entity_type, notification.entity_id, notification.actor_id, notification.title, notification.message) AS text), 'UTF8'))
ON CONFLICT (recipient_id, workspace_id, content_hash) DO NOTHING;
