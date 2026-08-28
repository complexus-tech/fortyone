-- GetNotificationEmailDelivery returns one pending email only when the task
-- payload's recipient/workspace scope still owns the notification and retains
-- current resource or public-portal access.
-- name: GetNotificationEmailDelivery :one
WITH eligible_notification AS (
    SELECT
        notification.notification_id,
        notification.recipient_id,
        notification.workspace_id,
        notification.type,
        notification.entity_type,
        notification.entity_id,
        notification.title,
        notification.message,
        recipient.email AS user_email,
        COALESCE(NULLIF(recipient.full_name, ''), recipient.username) AS user_name,
        COALESCE(NULLIF(actor.full_name, ''), actor.username) AS actor_name,
        workspace.name AS workspace_name,
        workspace.slug AS workspace_slug,
        CAST(COALESCE(CAST(membership.role AS text), CAST('' AS text)) AS text) AS workspace_role,
        COALESCE(feedback_item.slug, '') AS feedback_slug,
        CASE
            WHEN jsonb_typeof(
                preference.preferences
                    -> CAST(notification.type AS text)
                    -> 'email'
            ) = 'boolean'
                THEN CAST(
                    preference.preferences
                        -> CAST(notification.type AS text)
                        ->> 'email'
                    AS boolean
                )
            ELSE TRUE
        END AS email_enabled
    FROM public.notifications AS notification
    INNER JOIN public.users AS recipient
        ON recipient.user_id = notification.recipient_id
       AND recipient.is_active = TRUE
       AND recipient.is_system = FALSE
       AND NULLIF(BTRIM(recipient.email), '') IS NOT NULL
    INNER JOIN public.users AS actor
        ON actor.user_id = notification.actor_id
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = notification.workspace_id
       AND workspace.deleted_at IS NULL
    LEFT JOIN public.workspace_members AS membership
        ON membership.workspace_id = notification.workspace_id
       AND membership.user_id = notification.recipient_id
       AND membership.role IN ('admin', 'member', 'guest')
    LEFT JOIN public.notification_preferences AS preference
        ON preference.user_id = notification.recipient_id
       AND preference.workspace_id = notification.workspace_id
    LEFT JOIN public.feedback_items AS feedback_item
        ON CAST(notification.entity_type AS text) = 'feedback'
       AND feedback_item.id = notification.entity_id
       AND feedback_item.workspace_id = notification.workspace_id
       AND feedback_item.deleted_at IS NULL
    WHERE notification.notification_id = CAST(sqlc.arg(notification_id) AS uuid)
      AND notification.recipient_id = CAST(sqlc.arg(recipient_id) AS uuid)
      AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND notification.read_at IS NULL
      AND notification.email_sent_at IS NULL
      AND (
          (
              CAST(notification.entity_type AS text) = 'feedback'
              AND feedback_item.id IS NOT NULL
              AND EXISTS (
                  SELECT 1
                  FROM public.feedback_portals AS portal
                  INNER JOIN public.feedback_contributors AS contributor
                      ON contributor.portal_id = portal.id
                     AND contributor.user_id = notification.recipient_id
                     AND contributor.kind = 'account'
                     AND contributor.blocked_at IS NULL
                  WHERE portal.id = feedback_item.portal_id
                    AND portal.workspace_id = notification.workspace_id
                    AND portal.is_public = TRUE
              )
          )
          OR (
              CAST(notification.entity_type AS text) <> 'feedback'
              AND membership.user_id IS NOT NULL
              AND (
                  (
                      CAST(notification.entity_type AS text) = 'story'
                      AND EXISTS (
                          SELECT 1 FROM public.stories AS story
                          WHERE story.id = notification.entity_id
                            AND story.workspace_id = notification.workspace_id
                            AND story.deleted_at IS NULL
                            AND (
                                membership.role = 'admin'
                                OR EXISTS (
                                    SELECT 1 FROM public.team_members AS team_membership
                                    WHERE team_membership.team_id = story.team_id
                                      AND team_membership.user_id = notification.recipient_id
                                )
                            )
                      )
                  )
                  OR (
                      CAST(notification.entity_type AS text) = 'comment'
                      AND EXISTS (
                          SELECT 1
                          FROM public.story_comments AS comment
                          INNER JOIN public.stories AS story
                              ON story.id = comment.story_id
                             AND story.workspace_id = notification.workspace_id
                             AND story.deleted_at IS NULL
                          WHERE comment.comment_id = notification.entity_id
                            AND (
                                membership.role = 'admin'
                                OR EXISTS (
                                    SELECT 1 FROM public.team_members AS team_membership
                                    WHERE team_membership.team_id = story.team_id
                                      AND team_membership.user_id = notification.recipient_id
                                )
                            )
                      )
                  )
                  OR (
                      CAST(notification.entity_type AS text) = 'objective'
                      AND EXISTS (
                          SELECT 1 FROM public.objectives AS objective
                          WHERE objective.objective_id = notification.entity_id
                            AND objective.workspace_id = notification.workspace_id
                            AND (
                                membership.role = 'admin'
                                OR EXISTS (
                                    SELECT 1 FROM public.team_members AS team_membership
                                    WHERE team_membership.team_id = objective.team_id
                                      AND team_membership.user_id = notification.recipient_id
                                )
                            )
                      )
                  )
                  OR (
                      CAST(notification.entity_type AS text) = 'key_result'
                      AND EXISTS (
                          SELECT 1
                          FROM public.key_results AS key_result
                          INNER JOIN public.objectives AS objective
                              ON objective.objective_id = key_result.objective_id
                             AND objective.workspace_id = notification.workspace_id
                          WHERE key_result.id = notification.entity_id
                            AND (
                                membership.role = 'admin'
                                OR EXISTS (
                                    SELECT 1 FROM public.team_members AS team_membership
                                    WHERE team_membership.team_id = objective.team_id
                                      AND team_membership.user_id = notification.recipient_id
                                )
                            )
                      )
                  )
                  OR (
                      CAST(notification.entity_type AS text) = 'strategy'
                      AND (
                          membership.role = 'admin'
                          OR notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
                      )
                  )
              )
          )
      )
)
SELECT
    eligible.notification_id,
    eligible.recipient_id,
    eligible.workspace_id,
    eligible.type,
    eligible.entity_type,
    eligible.entity_id,
    eligible.title,
    eligible.message,
    eligible.user_email,
    eligible.user_name,
    eligible.actor_name,
    eligible.workspace_name,
    eligible.workspace_slug,
    eligible.workspace_role,
    eligible.email_enabled,
    eligible.feedback_slug
FROM eligible_notification AS eligible;

-- ListNotificationEmailDigestDeliveries applies the same current-access policy
-- to every pending row and uses a deterministic oldest-first order.
-- name: ListNotificationEmailDigestDeliveries :many
SELECT
    notification.notification_id,
    notification.recipient_id,
    notification.workspace_id,
    notification.type,
    notification.entity_type,
    notification.entity_id,
    notification.title,
    notification.message,
    notification.created_at,
    recipient.email AS user_email,
    COALESCE(NULLIF(recipient.full_name, ''), recipient.username) AS user_name,
    COALESCE(NULLIF(actor.full_name, ''), actor.username) AS actor_name,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    CAST(COALESCE(CAST(membership.role AS text), CAST('' AS text)) AS text) AS workspace_role,
    COALESCE(feedback_item.slug, '') AS feedback_slug
FROM public.notifications AS notification
INNER JOIN public.users AS recipient
    ON recipient.user_id = notification.recipient_id
   AND recipient.is_active = TRUE
   AND recipient.is_system = FALSE
   AND NULLIF(BTRIM(recipient.email), '') IS NOT NULL
INNER JOIN public.users AS actor
    ON actor.user_id = notification.actor_id
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = notification.workspace_id
   AND workspace.deleted_at IS NULL
LEFT JOIN public.workspace_members AS membership
    ON membership.workspace_id = notification.workspace_id
   AND membership.user_id = notification.recipient_id
   AND membership.role IN ('admin', 'member', 'guest')
LEFT JOIN public.notification_preferences AS preference
    ON preference.user_id = notification.recipient_id
   AND preference.workspace_id = notification.workspace_id
LEFT JOIN public.feedback_items AS feedback_item
    ON CAST(notification.entity_type AS text) = 'feedback'
   AND feedback_item.id = notification.entity_id
   AND feedback_item.workspace_id = notification.workspace_id
   AND feedback_item.deleted_at IS NULL
WHERE notification.recipient_id = CAST(sqlc.arg(recipient_id) AS uuid)
  AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND notification.read_at IS NULL
  AND notification.email_sent_at IS NULL
  AND CASE
      WHEN jsonb_typeof(
          preference.preferences
              -> CAST(notification.type AS text)
              -> 'email'
      ) = 'boolean'
          THEN CAST(
              preference.preferences
                  -> CAST(notification.type AS text)
                  ->> 'email'
              AS boolean
          )
      ELSE TRUE
  END
  AND (
      (
          CAST(notification.entity_type AS text) = 'feedback'
          AND feedback_item.id IS NOT NULL
          AND EXISTS (
              SELECT 1
              FROM public.feedback_portals AS portal
              INNER JOIN public.feedback_contributors AS contributor
                  ON contributor.portal_id = portal.id
                 AND contributor.user_id = notification.recipient_id
                 AND contributor.kind = 'account'
                 AND contributor.blocked_at IS NULL
              WHERE portal.id = feedback_item.portal_id
                AND portal.workspace_id = notification.workspace_id
                AND portal.is_public = TRUE
          )
      )
      OR (
          CAST(notification.entity_type AS text) <> 'feedback'
          AND membership.user_id IS NOT NULL
          AND (
              (
                  CAST(notification.entity_type AS text) = 'story'
                  AND EXISTS (
                      SELECT 1 FROM public.stories AS story
                      WHERE story.id = notification.entity_id
                        AND story.workspace_id = notification.workspace_id
                        AND story.deleted_at IS NULL
                        AND (
                            membership.role = 'admin'
                            OR EXISTS (
                                SELECT 1 FROM public.team_members AS team_membership
                                WHERE team_membership.team_id = story.team_id
                                  AND team_membership.user_id = notification.recipient_id
                            )
                        )
                  )
              )
              OR (
                  CAST(notification.entity_type AS text) = 'comment'
                  AND EXISTS (
                      SELECT 1
                      FROM public.story_comments AS comment
                      INNER JOIN public.stories AS story
                          ON story.id = comment.story_id
                         AND story.workspace_id = notification.workspace_id
                         AND story.deleted_at IS NULL
                      WHERE comment.comment_id = notification.entity_id
                        AND (
                            membership.role = 'admin'
                            OR EXISTS (
                                SELECT 1 FROM public.team_members AS team_membership
                                WHERE team_membership.team_id = story.team_id
                                  AND team_membership.user_id = notification.recipient_id
                            )
                        )
                  )
              )
              OR (
                  CAST(notification.entity_type AS text) = 'objective'
                  AND EXISTS (
                      SELECT 1 FROM public.objectives AS objective
                      WHERE objective.objective_id = notification.entity_id
                        AND objective.workspace_id = notification.workspace_id
                        AND (
                            membership.role = 'admin'
                            OR EXISTS (
                                SELECT 1 FROM public.team_members AS team_membership
                                WHERE team_membership.team_id = objective.team_id
                                  AND team_membership.user_id = notification.recipient_id
                            )
                        )
                  )
              )
              OR (
                  CAST(notification.entity_type AS text) = 'key_result'
                  AND EXISTS (
                      SELECT 1
                      FROM public.key_results AS key_result
                      INNER JOIN public.objectives AS objective
                          ON objective.objective_id = key_result.objective_id
                         AND objective.workspace_id = notification.workspace_id
                      WHERE key_result.id = notification.entity_id
                        AND (
                            membership.role = 'admin'
                            OR EXISTS (
                                SELECT 1 FROM public.team_members AS team_membership
                                WHERE team_membership.team_id = objective.team_id
                                  AND team_membership.user_id = notification.recipient_id
                            )
                        )
                  )
              )
              OR (
                  CAST(notification.entity_type AS text) = 'strategy'
                  AND (
                      membership.role = 'admin'
                      OR notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
                  )
              )
          )
      )
  )
ORDER BY notification.created_at ASC NULLS LAST, notification.notification_id ASC;

-- name: ListNotificationDeliveryTeamIDs :many
SELECT team_membership.team_id
FROM public.users AS recipient
INNER JOIN public.workspace_members AS membership
    ON membership.user_id = recipient.user_id
   AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
   AND membership.role IN ('member', 'guest')
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.team_members AS team_membership
    ON team_membership.user_id = recipient.user_id
INNER JOIN public.teams AS team
    ON team.team_id = team_membership.team_id
   AND team.workspace_id = workspace.workspace_id
WHERE recipient.user_id = CAST(sqlc.arg(recipient_id) AS uuid)
  AND recipient.is_active = TRUE
ORDER BY team_membership.team_id;

-- name: MarkNotificationEmailsSent :execrows
UPDATE public.notifications AS notification
SET email_sent_at = COALESCE(notification.email_sent_at, CAST(sqlc.arg(sent_at) AS timestamptz))
WHERE notification.recipient_id = CAST(sqlc.arg(recipient_id) AS uuid)
  AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND notification.notification_id = ANY(CAST(sqlc.arg(notification_ids) AS uuid[]));
