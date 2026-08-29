-- CreateNotification persists a notification only while its recipient can
-- still access the owning workspace resource. The notification row is the
-- durable email-delivery intent; dedupe replays return the original row without
-- changing read/email timestamps or publishing a second realtime mutation.
-- name: CreateNotification :one
WITH eligible_notification AS (
    SELECT
        CAST(sqlc.arg(dedupe_key) AS text) AS dedupe_key,
        recipient.user_id AS recipient_id,
        workspace.workspace_id,
        CAST(sqlc.arg(notification_type) AS notification_type) AS notification_type,
        CAST(sqlc.arg(entity_type) AS entity_type) AS entity_type,
        CAST(sqlc.arg(entity_id) AS uuid) AS entity_id,
        actor.user_id AS actor_id,
        CAST(sqlc.arg(title) AS text) AS title,
        CAST(sqlc.arg(message) AS jsonb) AS message,
        CASE
            WHEN CAST(sqlc.arg(notification_type) AS text) IN (
                'strategy_update'
            ) THEN FALSE
            WHEN CAST(sqlc.narg(in_app_enabled) AS boolean) IS NOT NULL
                THEN CAST(sqlc.narg(in_app_enabled) AS boolean)
            WHEN jsonb_typeof(
                preference.preferences
                    -> CAST(sqlc.arg(notification_type) AS text)
                    -> 'in_app'
            ) = 'boolean'
                THEN CAST(
                    preference.preferences
                        -> CAST(sqlc.arg(notification_type) AS text)
                        ->> 'in_app'
                    AS boolean
                )
            ELSE TRUE
        END AS in_app_enabled
    FROM public.users AS recipient
    INNER JOIN public.users AS actor
        ON actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
       AND actor.is_active = TRUE
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
       AND workspace.deleted_at IS NULL
    LEFT JOIN public.notification_preferences AS preference
        ON preference.user_id = recipient.user_id
       AND preference.workspace_id = workspace.workspace_id
    WHERE recipient.user_id = CAST(sqlc.arg(recipient_id) AS uuid)
      AND recipient.is_active = TRUE
      AND (
          actor.is_system = TRUE
          OR (
              CAST(sqlc.arg(entity_type) AS text) = 'feedback'
              AND (
                  EXISTS (
                      SELECT 1
                      FROM public.feedback_items AS actor_feedback_item
                      INNER JOIN public.feedback_portals AS actor_feedback_portal
                          ON actor_feedback_portal.id = actor_feedback_item.portal_id
                         AND actor_feedback_portal.workspace_id = actor_feedback_item.workspace_id
                         AND actor_feedback_portal.is_public = TRUE
                      INNER JOIN public.feedback_contributors AS actor_contributor
                          ON actor_contributor.portal_id = actor_feedback_portal.id
                         AND actor_contributor.user_id = actor.user_id
                         AND actor_contributor.kind = 'account'
                         AND actor_contributor.blocked_at IS NULL
                      WHERE actor_feedback_item.id = CAST(sqlc.arg(entity_id) AS uuid)
                        AND actor_feedback_item.workspace_id = workspace.workspace_id
                        AND actor_feedback_item.deleted_at IS NULL
                  )
                  OR EXISTS (
                      SELECT 1
                      FROM public.workspace_members AS actor_membership
                      INNER JOIN public.feedback_items AS actor_feedback_item
                          ON actor_feedback_item.id = CAST(sqlc.arg(entity_id) AS uuid)
                         AND actor_feedback_item.workspace_id = actor_membership.workspace_id
                         AND actor_feedback_item.deleted_at IS NULL
                      INNER JOIN public.feedback_boards AS actor_feedback_board
                          ON actor_feedback_board.id = actor_feedback_item.board_id
                         AND actor_feedback_board.workspace_id = actor_feedback_item.workspace_id
                         AND actor_feedback_board.portal_id = actor_feedback_item.portal_id
                      WHERE actor_membership.workspace_id = workspace.workspace_id
                        AND actor_membership.user_id = actor.user_id
                        AND actor_membership.role IN ('admin', 'member', 'guest')
                        AND (
                            actor_membership.role = 'admin'
                            OR EXISTS (
                                SELECT 1
                                FROM public.team_members AS actor_team_membership
                                WHERE actor_team_membership.team_id = actor_feedback_board.team_id
                                  AND actor_team_membership.user_id = actor.user_id
                            )
                        )
                  )
              )
          )
          OR (
              CAST(sqlc.arg(entity_type) AS text) <> 'feedback'
              AND EXISTS (
                  SELECT 1
                  FROM public.workspace_members AS actor_membership
                  WHERE actor_membership.workspace_id = workspace.workspace_id
                    AND actor_membership.user_id = actor.user_id
                    AND actor_membership.role IN ('admin', 'member', 'guest')
                    AND (
                        (
                            CAST(sqlc.arg(entity_type) AS text) = 'story'
                            AND EXISTS (
                                SELECT 1
                                FROM public.stories AS actor_story
                                WHERE actor_story.id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND actor_story.workspace_id = workspace.workspace_id
                                  AND actor_story.deleted_at IS NULL
                                  AND (
                                      actor_membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS actor_team_membership
                                          WHERE actor_team_membership.team_id = actor_story.team_id
                                            AND actor_team_membership.user_id = actor.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'comment'
                            AND EXISTS (
                                SELECT 1
                                FROM public.story_comments AS actor_comment
                                INNER JOIN public.stories AS actor_story
                                    ON actor_story.id = actor_comment.story_id
                                   AND actor_story.workspace_id = workspace.workspace_id
                                   AND actor_story.deleted_at IS NULL
                                WHERE actor_comment.comment_id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND (
                                      actor_membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS actor_team_membership
                                          WHERE actor_team_membership.team_id = actor_story.team_id
                                            AND actor_team_membership.user_id = actor.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'objective'
                            AND EXISTS (
                                SELECT 1
                                FROM public.objectives AS actor_objective
                                WHERE actor_objective.objective_id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND actor_objective.workspace_id = workspace.workspace_id
                                  AND (
                                      actor_membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS actor_team_membership
                                          WHERE actor_team_membership.team_id = actor_objective.team_id
                                            AND actor_team_membership.user_id = actor.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'key_result'
                            AND EXISTS (
                                SELECT 1
                                FROM public.key_results AS actor_key_result
                                INNER JOIN public.objectives AS actor_objective
                                    ON actor_objective.objective_id = actor_key_result.objective_id
                                   AND actor_objective.workspace_id = workspace.workspace_id
                                WHERE actor_key_result.id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND (
                                      actor_membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS actor_team_membership
                                          WHERE actor_team_membership.team_id = actor_objective.team_id
                                            AND actor_team_membership.user_id = actor.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'strategy'
                            AND CAST(sqlc.arg(entity_id) AS uuid) = workspace.workspace_id
                        )
                    )
              )
          )
      )
      AND (
          (
              CAST(sqlc.arg(entity_type) AS text) = 'feedback'
              AND EXISTS (
                  SELECT 1
                  FROM public.feedback_items AS feedback_item
                  INNER JOIN public.feedback_portals AS feedback_portal
                      ON feedback_portal.id = feedback_item.portal_id
                     AND feedback_portal.workspace_id = feedback_item.workspace_id
                     AND feedback_portal.is_public = TRUE
                  INNER JOIN public.feedback_contributors AS contributor
                      ON contributor.portal_id = feedback_portal.id
                     AND contributor.user_id = recipient.user_id
                     AND contributor.kind = 'account'
                     AND contributor.blocked_at IS NULL
                  WHERE feedback_item.id = CAST(sqlc.arg(entity_id) AS uuid)
                    AND feedback_item.workspace_id = workspace.workspace_id
                    AND feedback_item.deleted_at IS NULL
              )
          )
          OR (
              CAST(sqlc.arg(entity_type) AS text) <> 'feedback'
              AND EXISTS (
                  SELECT 1
                  FROM public.workspace_members AS membership
                  WHERE membership.workspace_id = workspace.workspace_id
                    AND membership.user_id = recipient.user_id
                    AND membership.role IN ('admin', 'member', 'guest')
                    AND (
                        (
                            CAST(sqlc.arg(entity_type) AS text) = 'story'
                            AND EXISTS (
                                SELECT 1
                                FROM public.stories AS story
                                WHERE story.id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND story.workspace_id = workspace.workspace_id
                                  AND story.deleted_at IS NULL
                                  AND (
                                      membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS team_membership
                                          WHERE team_membership.team_id = story.team_id
                                            AND team_membership.user_id = recipient.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'comment'
                            AND EXISTS (
                                SELECT 1
                                FROM public.story_comments AS comment
                                INNER JOIN public.stories AS story
                                    ON story.id = comment.story_id
                                   AND story.workspace_id = workspace.workspace_id
                                   AND story.deleted_at IS NULL
                                WHERE comment.comment_id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND (
                                      membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS team_membership
                                          WHERE team_membership.team_id = story.team_id
                                            AND team_membership.user_id = recipient.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'objective'
                            AND EXISTS (
                                SELECT 1
                                FROM public.objectives AS objective
                                WHERE objective.objective_id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND objective.workspace_id = workspace.workspace_id
                                  AND (
                                      membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS team_membership
                                          WHERE team_membership.team_id = objective.team_id
                                            AND team_membership.user_id = recipient.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'key_result'
                            AND EXISTS (
                                SELECT 1
                                FROM public.key_results AS key_result
                                INNER JOIN public.objectives AS objective
                                    ON objective.objective_id = key_result.objective_id
                                   AND objective.workspace_id = workspace.workspace_id
                                WHERE key_result.id = CAST(sqlc.arg(entity_id) AS uuid)
                                  AND (
                                      membership.role = 'admin'
                                      OR EXISTS (
                                          SELECT 1
                                          FROM public.team_members AS team_membership
                                          WHERE team_membership.team_id = objective.team_id
                                            AND team_membership.user_id = recipient.user_id
                                      )
                                  )
                            )
                        )
                        OR (
                            CAST(sqlc.arg(entity_type) AS text) = 'strategy'
                            AND CAST(sqlc.arg(entity_id) AS uuid) = workspace.workspace_id
                            AND (
                                membership.role = 'admin'
                                OR CAST(sqlc.arg(message) AS jsonb)
                                    -> 'strategy' ->> 'kind' = 'weekly_check_in'
                            )
                        )
                    )
              )
          )
      )
), inserted_notification AS (
    INSERT INTO public.notifications (
        dedupe_key,
        recipient_id,
        workspace_id,
        type,
        entity_type,
        entity_id,
        actor_id,
        title,
        message,
        in_app_enabled
    )
    SELECT
        eligible.dedupe_key,
        eligible.recipient_id,
        eligible.workspace_id,
        eligible.notification_type,
        eligible.entity_type,
        eligible.entity_id,
        eligible.actor_id,
        eligible.title,
        eligible.message,
        eligible.in_app_enabled
    FROM eligible_notification AS eligible
    ON CONFLICT (dedupe_key) DO NOTHING
    RETURNING
        notification_id,
        recipient_id,
        workspace_id,
        type,
        entity_type,
        entity_id,
        actor_id,
        title,
        message,
        in_app_enabled,
        created_at,
        read_at
)
SELECT
    inserted.notification_id,
    inserted.recipient_id,
    inserted.workspace_id,
    inserted.type,
    inserted.entity_type,
    inserted.entity_id,
    inserted.actor_id,
    inserted.title,
    inserted.message,
    inserted.in_app_enabled,
    inserted.created_at,
    inserted.read_at,
    TRUE AS inserted
FROM inserted_notification AS inserted
UNION ALL
SELECT
    existing.notification_id,
    existing.recipient_id,
    existing.workspace_id,
    existing.type,
    existing.entity_type,
    existing.entity_id,
    existing.actor_id,
    existing.title,
    existing.message,
    existing.in_app_enabled,
    existing.created_at,
    existing.read_at,
    FALSE AS inserted
FROM public.notifications AS existing
WHERE existing.dedupe_key = CAST(sqlc.arg(dedupe_key) AS text)
  AND existing.recipient_id = CAST(sqlc.arg(recipient_id) AS uuid)
  AND existing.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND existing.type = CAST(sqlc.arg(notification_type) AS notification_type)
  AND existing.entity_type = CAST(sqlc.arg(entity_type) AS entity_type)
  AND existing.entity_id = CAST(sqlc.arg(entity_id) AS uuid)
  AND existing.actor_id = CAST(sqlc.arg(actor_id) AS uuid)
  AND existing.title = CAST(sqlc.arg(title) AS text)
  AND existing.message = CAST(sqlc.arg(message) AS jsonb)
  AND NOT EXISTS (SELECT 1 FROM inserted_notification)
LIMIT 1;

-- name: NotificationDedupeKeyExists :one
SELECT EXISTS (
    SELECT 1
    FROM public.notifications AS notification
    WHERE notification.dedupe_key = CAST(sqlc.arg(dedupe_key) AS text)
      AND notification.recipient_id = CAST(sqlc.arg(recipient_id) AS uuid)
      AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND notification.actor_id = CAST(sqlc.arg(actor_id) AS uuid)
);
