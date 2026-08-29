-- ListWorkspaceNotifications returns only currently visible in-app rows. The
-- authorization predicate deliberately rechecks active account, live workspace
-- membership, and current team/resource access so stale notifications cannot
-- preserve access after revocation.
-- name: ListWorkspaceNotifications :many
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
), visible_notifications AS (
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
                  SELECT 1
                  FROM public.stories AS story
                  WHERE story.id = notification.entity_id
                    AND story.workspace_id = notification.workspace_id
                    AND story.deleted_at IS NULL
                    AND (
                        scope.role = 'admin'
                        OR EXISTS (
                            SELECT 1
                            FROM public.team_members AS team_membership
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
                            SELECT 1
                            FROM public.team_members AS team_membership
                            WHERE team_membership.team_id = story.team_id
                              AND team_membership.user_id = notification.recipient_id
                        )
                    )
              )
          )
          OR (
              CAST(notification.entity_type AS text) = 'objective'
              AND EXISTS (
                  SELECT 1
                  FROM public.objectives AS objective
                  WHERE objective.objective_id = notification.entity_id
                    AND objective.workspace_id = notification.workspace_id
                    AND (
                        scope.role = 'admin'
                        OR EXISTS (
                            SELECT 1
                            FROM public.team_members AS team_membership
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
                            SELECT 1
                            FROM public.team_members AS team_membership
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
    notification.read_at
FROM public.notifications AS notification
INNER JOIN visible_notifications AS visible
    ON visible.notification_id = notification.notification_id
WHERE (
    CAST(sqlc.arg(search) AS text) = ''
    OR notification.title ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
    OR (
        CAST(notification.entity_type AS text) <> 'strategy'
        AND CAST(notification.message AS text) ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
    )
)
ORDER BY notification.created_at DESC NULLS LAST, notification.notification_id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer)
OFFSET CAST(sqlc.arg(result_offset) AS integer);

-- name: CountUnreadWorkspaceNotifications :one
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
)
SELECT COUNT(*)
FROM public.notifications AS notification
CROSS JOIN actor_scope AS scope
WHERE notification.recipient_id = CAST(sqlc.arg(actor_id) AS uuid)
  AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND CAST(notification.entity_type AS text) <> 'feedback'
  AND notification.in_app_enabled = TRUE
  AND notification.read_at IS NULL
  AND (
      (
          CAST(notification.entity_type AS text) = 'story'
          AND EXISTS (
              SELECT 1
              FROM public.stories AS story
              WHERE story.id = notification.entity_id
                AND story.workspace_id = notification.workspace_id
                AND story.deleted_at IS NULL
                AND (
                    scope.role = 'admin'
                    OR EXISTS (
                        SELECT 1
                        FROM public.team_members AS team_membership
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
                        SELECT 1
                        FROM public.team_members AS team_membership
                        WHERE team_membership.team_id = story.team_id
                          AND team_membership.user_id = notification.recipient_id
                    )
                )
          )
      )
      OR (
          CAST(notification.entity_type AS text) = 'objective'
          AND EXISTS (
              SELECT 1
              FROM public.objectives AS objective
              WHERE objective.objective_id = notification.entity_id
                AND objective.workspace_id = notification.workspace_id
                AND (
                    scope.role = 'admin'
                    OR EXISTS (
                        SELECT 1
                        FROM public.team_members AS team_membership
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
                        SELECT 1
                        FROM public.team_members AS team_membership
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
  );

-- name: WorkspaceNotificationActorAuthorized :one
SELECT EXISTS (
    SELECT 1
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
);
