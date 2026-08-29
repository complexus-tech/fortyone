-- MutateWorkspaceNotification performs one finite notification mutation. The
-- adapter maps the typed domain intent to delete/read flags; SQL never accepts
-- identifiers, predicates, or ordering fragments from callers.
-- name: MutateWorkspaceNotification :one
WITH actor_scope AS (
    SELECT membership.role
    FROM public.users AS actor
    INNER JOIN public.workspace_members AS membership
        ON membership.user_id = actor.user_id
       AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
      AND membership.role IN ('admin', 'member', 'guest')
), authorized_notification AS (
    SELECT notification.notification_id
    FROM public.notifications AS notification
    CROSS JOIN actor_scope AS scope
    WHERE notification.notification_id = CAST(sqlc.arg(notification_id) AS uuid)
      AND notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND CAST(notification.entity_type AS text) <> 'feedback'
      AND notification.in_app_enabled = TRUE
      AND (
          (
              CAST(notification.entity_type AS text) = 'story'
              AND EXISTS (
                  SELECT 1 FROM public.stories AS story
                  WHERE story.id = notification.entity_id
                    AND story.workspace_id = notification.workspace_id
                    AND story.deleted_at IS NULL
                    AND (
                        scope.role = 'admin'
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
                        scope.role = 'admin'
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
                        scope.role = 'admin'
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
                        scope.role = 'admin'
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
                  scope.role = 'admin'
                  OR notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
              )
          )
      )
), deleted_notification AS (
    DELETE FROM public.notifications AS notification
    USING authorized_notification AS authorized
    WHERE CAST(sqlc.arg(delete_notification) AS boolean)
      AND notification.notification_id = authorized.notification_id
    RETURNING notification.notification_id
), updated_notification AS (
    UPDATE public.notifications AS notification
    SET read_at = CASE
        WHEN CAST(sqlc.arg(mark_read) AS boolean)
            THEN COALESCE(notification.read_at, CAST(sqlc.arg(mutated_at) AS timestamptz))
        ELSE NULL
    END
    FROM authorized_notification AS authorized
    WHERE NOT CAST(sqlc.arg(delete_notification) AS boolean)
      AND notification.notification_id = authorized.notification_id
    RETURNING notification.notification_id
)
SELECT notification_id FROM deleted_notification
UNION ALL
SELECT notification_id FROM updated_notification
LIMIT 1;

-- MutateWorkspaceNotifications performs bulk read/delete operations against
-- the same currently visible set used by inbox reads.
-- name: MutateWorkspaceNotifications :one
WITH actor_scope AS (
    SELECT membership.role
    FROM public.users AS actor
    INNER JOIN public.workspace_members AS membership
        ON membership.user_id = actor.user_id
       AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND actor.is_active = TRUE
      AND membership.role IN ('admin', 'member', 'guest')
), authorized_notifications AS (
    SELECT notification.notification_id
    FROM public.notifications AS notification
    CROSS JOIN actor_scope AS scope
    WHERE notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
      AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND CAST(notification.entity_type AS text) <> 'feedback'
      AND notification.in_app_enabled = TRUE
      AND (
          (
              CAST(notification.entity_type AS text) = 'story'
              AND EXISTS (
                  SELECT 1 FROM public.stories AS story
                  WHERE story.id = notification.entity_id
                    AND story.workspace_id = notification.workspace_id
                    AND story.deleted_at IS NULL
                    AND (
                        scope.role = 'admin'
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
                        scope.role = 'admin'
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
                        scope.role = 'admin'
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
                        scope.role = 'admin'
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
                  scope.role = 'admin'
                  OR notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
              )
          )
      )
), deleted_notifications AS (
    DELETE FROM public.notifications AS notification
    USING authorized_notifications AS authorized
    WHERE CAST(sqlc.arg(delete_notifications) AS boolean)
      AND (
          NOT CAST(sqlc.arg(only_read) AS boolean)
          OR notification.read_at IS NOT NULL
      )
      AND notification.notification_id = authorized.notification_id
    RETURNING notification.notification_id
), updated_notifications AS (
    UPDATE public.notifications AS notification
    SET read_at = COALESCE(notification.read_at, CAST(sqlc.arg(mutated_at) AS timestamptz))
    FROM authorized_notifications AS authorized
    WHERE NOT CAST(sqlc.arg(delete_notifications) AS boolean)
      AND notification.notification_id = authorized.notification_id
      AND notification.read_at IS NULL
    RETURNING notification.notification_id
)
SELECT
    CAST((SELECT COUNT(*) FROM deleted_notifications) AS bigint)
    + CAST((SELECT COUNT(*) FROM updated_notifications) AS bigint) AS affected_count;
