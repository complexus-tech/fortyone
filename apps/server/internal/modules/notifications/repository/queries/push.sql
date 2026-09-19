-- name: RegisterNotificationPushDevice :one
WITH authorized_user AS (
    SELECT users.user_id
    FROM public.users
    WHERE users.user_id = CAST(sqlc.arg(user_id) AS uuid)
      AND users.is_active = TRUE
      AND users.is_system = FALSE
)
INSERT INTO public.notification_push_devices (
    user_id,
    expo_push_token,
    platform
)
SELECT
    authorized_user.user_id,
    CAST(sqlc.arg(expo_push_token) AS text),
    CAST(sqlc.arg(platform) AS text)
FROM authorized_user
ON CONFLICT (expo_push_token) DO UPDATE
SET
    user_id = EXCLUDED.user_id,
    platform = EXCLUDED.platform,
    updated_at = now(),
    disabled_at = NULL
RETURNING device_id, user_id, expo_push_token, platform, created_at, updated_at, disabled_at;

-- name: UnregisterNotificationPushDevice :execrows
DELETE FROM public.notification_push_devices
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
  AND expo_push_token = CAST(sqlc.arg(expo_push_token) AS text);

-- GetNotificationPushDelivery repeats the inbox visibility boundary so a
-- queued delivery cannot reveal content after workspace or team access changes.
-- name: GetNotificationPushDelivery :many
WITH notification_scope AS (
    SELECT
        notification.notification_id,
        notification.recipient_id,
        notification.workspace_id,
        notification.entity_type,
        notification.entity_id,
        notification.title,
        notification.message,
        workspace.slug AS workspace_slug,
        membership.role
    FROM public.notifications AS notification
    INNER JOIN public.users AS recipient
        ON recipient.user_id = notification.recipient_id
       AND recipient.is_active = TRUE
       AND recipient.is_system = FALSE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = notification.workspace_id
       AND workspace.deleted_at IS NULL
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = notification.workspace_id
       AND membership.user_id = notification.recipient_id
       AND membership.role IN ('admin', 'member', 'guest')
    WHERE notification.notification_id = CAST(sqlc.arg(notification_id) AS uuid)
      AND notification.in_app_enabled = TRUE
      AND notification.push_sent_at IS NULL
      AND CAST(notification.entity_type AS text) <> 'feedback'
), visible_notification AS (
    SELECT scope.*
    FROM notification_scope AS scope
    WHERE
        (
            CAST(scope.entity_type AS text) = 'story'
            AND EXISTS (
                SELECT 1
                FROM public.stories AS story
                WHERE story.id = scope.entity_id
                  AND story.workspace_id = scope.workspace_id
                  AND story.deleted_at IS NULL
                  AND (
                      scope.role = 'admin'
                      OR EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_membership
                          WHERE team_membership.team_id = story.team_id
                            AND team_membership.user_id = scope.recipient_id
                      )
                  )
            )
        )
        OR (
            CAST(scope.entity_type AS text) = 'comment'
            AND EXISTS (
                SELECT 1
                FROM public.story_comments AS comment
                INNER JOIN public.stories AS story
                    ON story.id = comment.story_id
                   AND story.workspace_id = scope.workspace_id
                   AND story.deleted_at IS NULL
                WHERE comment.comment_id = scope.entity_id
                  AND (
                      scope.role = 'admin'
                      OR EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_membership
                          WHERE team_membership.team_id = story.team_id
                            AND team_membership.user_id = scope.recipient_id
                      )
                  )
            )
        )
        OR (
            CAST(scope.entity_type AS text) = 'objective'
            AND EXISTS (
                SELECT 1
                FROM public.objectives AS objective
                WHERE objective.objective_id = scope.entity_id
                  AND objective.workspace_id = scope.workspace_id
                  AND (
                      scope.role = 'admin'
                      OR EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_membership
                          WHERE team_membership.team_id = objective.team_id
                            AND team_membership.user_id = scope.recipient_id
                      )
                  )
            )
        )
        OR (
            CAST(scope.entity_type AS text) = 'key_result'
            AND EXISTS (
                SELECT 1
                FROM public.key_results AS key_result
                INNER JOIN public.objectives AS objective
                    ON objective.objective_id = key_result.objective_id
                   AND objective.workspace_id = scope.workspace_id
                WHERE key_result.id = scope.entity_id
                  AND (
                      scope.role = 'admin'
                      OR EXISTS (
                          SELECT 1
                          FROM public.team_members AS team_membership
                          WHERE team_membership.team_id = objective.team_id
                            AND team_membership.user_id = scope.recipient_id
                      )
                  )
            )
        )
        OR (
            CAST(scope.entity_type AS text) = 'strategy'
            AND (
                scope.role = 'admin'
                OR scope.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
            )
        )
)
SELECT
    visible.notification_id,
    visible.recipient_id,
    visible.workspace_id,
    visible.entity_type,
    visible.entity_id,
    visible.title,
    visible.message,
    visible.workspace_slug,
    COALESCE(device.expo_push_token, '') AS expo_push_token
FROM visible_notification AS visible
LEFT JOIN public.notification_push_devices AS device
    ON device.user_id = visible.recipient_id
   AND device.disabled_at IS NULL
ORDER BY device.device_id NULLS LAST;

-- name: MarkNotificationPushSent :execrows
UPDATE public.notifications
SET push_sent_at = CAST(sqlc.arg(sent_at) AS timestamptz)
WHERE notification_id = CAST(sqlc.arg(notification_id) AS uuid)
  AND push_sent_at IS NULL;

-- name: DisableNotificationPushDevices :execrows
UPDATE public.notification_push_devices
SET disabled_at = CAST(sqlc.arg(disabled_at) AS timestamptz),
    updated_at = CAST(sqlc.arg(disabled_at) AS timestamptz)
WHERE expo_push_token = ANY(CAST(sqlc.arg(expo_push_tokens) AS text[]))
  AND disabled_at IS NULL;
