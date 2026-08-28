-- ListPortalFeedbackNotifications keeps the public portal inbox isolated from
-- workspace membership. Access instead requires an active account contributor
-- that has not been blocked from the exact public portal.
-- name: ListPortalFeedbackNotifications :many
WITH portal_scope AS (
    SELECT portal.id, portal.workspace_id
    FROM public.users AS actor
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.user_id = actor.user_id
       AND contributor.kind = 'account'
       AND contributor.blocked_at IS NULL
    INNER JOIN public.feedback_portals AS portal
        ON portal.id = contributor.portal_id
       AND portal.is_public = TRUE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = portal.workspace_id
       AND workspace.slug = CAST(sqlc.arg(portal_slug) AS text)
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
), visible_notifications AS (
    SELECT
        notification.notification_id,
        feedback_item.title AS feedback_title,
        feedback_item.slug AS feedback_slug
    FROM public.notifications AS notification
    INNER JOIN public.feedback_items AS feedback_item
        ON feedback_item.id = notification.entity_id
       AND feedback_item.workspace_id = notification.workspace_id
       AND feedback_item.deleted_at IS NULL
    INNER JOIN portal_scope AS scope
        ON scope.id = feedback_item.portal_id
       AND scope.workspace_id = notification.workspace_id
    WHERE notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND CAST(notification.entity_type AS text) = 'feedback'
      AND CAST(notification.type AS text) IN (
          'feedback_comment',
          'feedback_status_update',
          'feedback_update_published',
          'feedback_item_merged'
      )
      AND notification.in_app_enabled = TRUE
      AND (
          NOT CAST(sqlc.arg(unread_only) AS boolean)
          OR notification.read_at IS NULL
      )
)
SELECT
    notification.notification_id,
    notification.recipient_id,
    notification.workspace_id,
    notification.type,
    notification.entity_type,
    notification.entity_id,
    notification.actor_id,
    notification.title,
    notification.message,
    notification.in_app_enabled,
    notification.created_at,
    notification.read_at,
    COALESCE(
        NULLIF(actor.full_name, ''),
        NULLIF(actor.username, ''),
        actor.email,
        'Someone'
    ) AS actor_name,
    actor.avatar_url AS actor_avatar,
    visible.feedback_title,
    visible.feedback_slug
FROM public.notifications AS notification
INNER JOIN visible_notifications AS visible
    ON visible.notification_id = notification.notification_id
LEFT JOIN public.users AS actor
    ON actor.user_id = notification.actor_id
ORDER BY notification.created_at DESC NULLS LAST, notification.notification_id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer)
OFFSET CAST(sqlc.arg(result_offset) AS integer);

-- name: CountUnreadPortalFeedbackNotifications :one
WITH portal_scope AS (
    SELECT portal.id, portal.workspace_id
    FROM public.users AS actor
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.user_id = actor.user_id
       AND contributor.kind = 'account'
       AND contributor.blocked_at IS NULL
    INNER JOIN public.feedback_portals AS portal
        ON portal.id = contributor.portal_id
       AND portal.is_public = TRUE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = portal.workspace_id
       AND workspace.slug = CAST(sqlc.arg(portal_slug) AS text)
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
)
SELECT COUNT(*)
FROM public.notifications AS notification
INNER JOIN public.feedback_items AS feedback_item
    ON feedback_item.id = notification.entity_id
   AND feedback_item.workspace_id = notification.workspace_id
   AND feedback_item.deleted_at IS NULL
INNER JOIN portal_scope AS scope
    ON scope.id = feedback_item.portal_id
   AND scope.workspace_id = notification.workspace_id
WHERE notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
  AND CAST(notification.entity_type AS text) = 'feedback'
  AND CAST(notification.type AS text) IN (
      'feedback_comment',
      'feedback_status_update',
      'feedback_update_published',
      'feedback_item_merged'
  )
  AND notification.in_app_enabled = TRUE
  AND notification.read_at IS NULL;

-- name: PortalNotificationActorAuthorized :one
SELECT EXISTS (
    SELECT 1
    FROM public.users AS actor
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.user_id = actor.user_id
       AND contributor.kind = 'account'
       AND contributor.blocked_at IS NULL
    INNER JOIN public.feedback_portals AS portal
        ON portal.id = contributor.portal_id
       AND portal.is_public = TRUE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = portal.workspace_id
       AND workspace.slug = CAST(sqlc.arg(portal_slug) AS text)
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
);

-- name: MarkPortalFeedbackNotificationRead :one
WITH portal_scope AS (
    SELECT portal.id, portal.workspace_id
    FROM public.users AS actor
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.user_id = actor.user_id
       AND contributor.kind = 'account'
       AND contributor.blocked_at IS NULL
    INNER JOIN public.feedback_portals AS portal
        ON portal.id = contributor.portal_id
       AND portal.is_public = TRUE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = portal.workspace_id
       AND workspace.slug = CAST(sqlc.arg(portal_slug) AS text)
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
), authorized_notification AS (
    SELECT notification.notification_id
    FROM public.notifications AS notification
    INNER JOIN public.feedback_items AS feedback_item
        ON feedback_item.id = notification.entity_id
       AND feedback_item.workspace_id = notification.workspace_id
       AND feedback_item.deleted_at IS NULL
    INNER JOIN portal_scope AS scope
        ON scope.id = feedback_item.portal_id
       AND scope.workspace_id = notification.workspace_id
    WHERE notification.notification_id = CAST(sqlc.arg(notification_id) AS uuid)
      AND notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND CAST(notification.entity_type AS text) = 'feedback'
      AND CAST(notification.type AS text) IN (
          'feedback_comment',
          'feedback_status_update',
          'feedback_update_published',
          'feedback_item_merged'
      )
      AND notification.in_app_enabled = TRUE
)
UPDATE public.notifications AS notification
SET read_at = COALESCE(notification.read_at, CAST(sqlc.arg(read_at) AS timestamptz))
FROM authorized_notification AS authorized
WHERE notification.notification_id = authorized.notification_id
RETURNING notification.notification_id;

-- name: MarkAllPortalFeedbackNotificationsRead :one
WITH portal_scope AS (
    SELECT portal.id, portal.workspace_id
    FROM public.users AS actor
    INNER JOIN public.feedback_contributors AS contributor
        ON contributor.user_id = actor.user_id
       AND contributor.kind = 'account'
       AND contributor.blocked_at IS NULL
    INNER JOIN public.feedback_portals AS portal
        ON portal.id = contributor.portal_id
       AND portal.is_public = TRUE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = portal.workspace_id
       AND workspace.slug = CAST(sqlc.arg(portal_slug) AS text)
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
), authorized_notifications AS (
    SELECT notification.notification_id
    FROM public.notifications AS notification
    INNER JOIN public.feedback_items AS feedback_item
        ON feedback_item.id = notification.entity_id
       AND feedback_item.workspace_id = notification.workspace_id
       AND feedback_item.deleted_at IS NULL
    INNER JOIN portal_scope AS scope
        ON scope.id = feedback_item.portal_id
       AND scope.workspace_id = notification.workspace_id
    WHERE notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND CAST(notification.entity_type AS text) = 'feedback'
      AND CAST(notification.type AS text) IN (
          'feedback_comment',
          'feedback_status_update',
          'feedback_update_published',
          'feedback_item_merged'
      )
      AND notification.in_app_enabled = TRUE
      AND notification.read_at IS NULL
), updated_notifications AS (
    UPDATE public.notifications AS notification
    SET read_at = CAST(sqlc.arg(read_at) AS timestamptz)
    FROM authorized_notifications AS authorized
    WHERE notification.notification_id = authorized.notification_id
    RETURNING notification.notification_id
)
SELECT CAST(COUNT(*) AS bigint) AS affected_count
FROM updated_notifications;
