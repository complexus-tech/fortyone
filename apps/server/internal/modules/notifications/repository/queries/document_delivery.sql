-- name: NextDocumentNotificationDelivery :one
SELECT notification.notification_id,notification.recipient_id,notification.workspace_id,
 notification.type,notification.entity_type,notification.entity_id,notification.actor_id,
 notification.title,notification.message,notification.in_app_enabled,notification.created_at,notification.read_at,
 public.can_read_document(notification.workspace_id,notification.recipient_id,notification.entity_id) AS authorized
FROM public.document_notification_deliveries AS delivery
JOIN public.notifications AS notification ON notification.notification_id=delivery.notification_id
WHERE delivery.delivered_at IS NULL
ORDER BY delivery.created_at,delivery.notification_id
LIMIT 1 FOR UPDATE OF delivery SKIP LOCKED;

-- name: CompleteDocumentNotificationDelivery :exec
UPDATE public.document_notification_deliveries SET delivered_at=clock_timestamp() WHERE notification_id=sqlc.arg(notification_id);
