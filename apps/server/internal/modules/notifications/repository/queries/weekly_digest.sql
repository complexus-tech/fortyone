-- ListWeeklyDigestRecipients pages active admin/member recipients using the
-- same composite key for filtering and ordering. Weekly digest email defaults
-- to enabled when no explicit boolean preference exists.
-- name: ListWeeklyDigestRecipients :many
SELECT
    membership.user_id,
    recipient.email AS user_email,
    COALESCE(NULLIF(recipient.full_name, ''), recipient.username) AS user_name,
    workspace.workspace_id,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug
FROM public.workspace_members AS membership
INNER JOIN public.users AS recipient
    ON recipient.user_id = membership.user_id
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
LEFT JOIN public.notification_preferences AS preference
    ON preference.user_id = membership.user_id
   AND preference.workspace_id = membership.workspace_id
WHERE recipient.is_active = TRUE
  AND recipient.is_system = FALSE
  AND membership.role IN ('admin', 'member')
  AND workspace.deleted_at IS NULL
  AND NULLIF(BTRIM(recipient.email), '') IS NOT NULL
  AND CAST(COALESCE(preference.preferences -> 'weekly_digest' ->> 'email', 'true') AS boolean) = TRUE
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR workspace.workspace_id > CAST(sqlc.arg(after_workspace_id) AS uuid)
      OR (
          workspace.workspace_id = CAST(sqlc.arg(after_workspace_id) AS uuid)
          AND membership.user_id > CAST(sqlc.arg(after_user_id) AS uuid)
      )
  )
ORDER BY workspace.workspace_id, membership.user_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- GetWeeklyDigestStats revalidates delivery eligibility and computes every
-- signal against one caller-supplied UTC as-of time. Current entity access is
-- applied before an unread notification can contribute to the aggregate.
-- name: GetWeeklyDigestStats :one
WITH digest_period AS (
    SELECT
        CAST(sqlc.arg(as_of) AS timestamptz) AS as_of,
        CAST(sqlc.arg(as_of) AS timestamptz) - INTERVAL '7 days' AS window_start,
        CAST(CAST(sqlc.arg(as_of) AS timestamptz) AT TIME ZONE 'UTC' AS date) AS as_of_date
),
recipient_access AS (
    SELECT membership.role
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS recipient
        ON recipient.user_id = membership.user_id
       AND recipient.is_active = TRUE
       AND recipient.is_system = FALSE
       AND NULLIF(BTRIM(recipient.email), '') IS NOT NULL
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    LEFT JOIN public.notification_preferences AS preference
        ON preference.user_id = membership.user_id
       AND preference.workspace_id = membership.workspace_id
    WHERE membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
      AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND membership.role IN ('admin', 'member')
      AND CAST(COALESCE(preference.preferences -> 'weekly_digest' ->> 'email', 'true') AS boolean) = TRUE
),
visible_teams AS (
    SELECT team_membership.team_id
    FROM public.team_members AS team_membership
    INNER JOIN public.teams AS team
        ON team.team_id = team_membership.team_id
       AND team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
    WHERE team_membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
),
accessible_notifications AS (
    SELECT notification.notification_id, notification.type, notification.created_at
    FROM public.notifications AS notification
    WHERE notification.recipient_id = CAST(sqlc.arg(user_id) AS uuid)
      AND notification.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND notification.read_at IS NULL
      AND EXISTS (SELECT 1 FROM recipient_access)
      AND (
          (
              CAST(notification.entity_type AS text) = 'feedback'
              AND EXISTS (
                  SELECT 1
                  FROM public.feedback_items AS feedback
                  WHERE feedback.id = notification.entity_id
                    AND feedback.workspace_id = notification.workspace_id
                    AND feedback.deleted_at IS NULL
              )
          )
          OR (
              CAST(notification.entity_type AS text) = 'story'
              AND EXISTS (
                  SELECT 1
                  FROM public.stories AS story
                  WHERE story.id = notification.entity_id
                    AND story.workspace_id = notification.workspace_id
                    AND story.deleted_at IS NULL
                    AND (
                        EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
                        OR story.team_id IN (SELECT team_id FROM visible_teams)
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
                  WHERE comment.comment_id = notification.entity_id
                    AND story.workspace_id = notification.workspace_id
                    AND story.deleted_at IS NULL
                    AND (
                        EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
                        OR story.team_id IN (SELECT team_id FROM visible_teams)
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
                        EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
                        OR objective.team_id IN (SELECT team_id FROM visible_teams)
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
                  WHERE key_result.id = notification.entity_id
                    AND objective.workspace_id = notification.workspace_id
                    AND (
                        EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
                        OR objective.team_id IN (SELECT team_id FROM visible_teams)
                    )
              )
          )
          OR (
              CAST(notification.entity_type AS text) = 'strategy'
              AND (
                  EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
                  OR notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
              )
          )
      )
)
SELECT
    CAST((
        SELECT COUNT(*)
        FROM accessible_notifications
    ) AS integer) AS unread_notifications,
    CAST((
        SELECT COUNT(*)
        FROM accessible_notifications AS notification
        CROSS JOIN digest_period AS period
        WHERE notification.created_at >= period.window_start
          AND notification.created_at <= period.as_of
          AND notification.type IN ('mention', 'comment_reply')
    ) AS integer) AS unread_priority_notifications,
    CAST((
        SELECT COUNT(*)
        FROM public.stories AS story
        INNER JOIN public.statuses AS status
            ON status.status_id = story.status_id
        CROSS JOIN digest_period AS period
        WHERE story.assignee_id = CAST(sqlc.arg(user_id) AS uuid)
          AND story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
          AND EXISTS (SELECT 1 FROM recipient_access)
          AND story.end_date < period.as_of_date
          AND status.category NOT IN ('completed', 'cancelled', 'paused')
          AND story.deleted_at IS NULL
          AND story.archived_at IS NULL
          AND story.completed_at IS NULL
          AND (
              EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
              OR story.team_id IN (SELECT team_id FROM visible_teams)
          )
    ) AS integer) AS overdue_stories,
    CAST((
        SELECT COUNT(*)
        FROM public.stories AS story
        INNER JOIN public.statuses AS status
            ON status.status_id = story.status_id
        CROSS JOIN digest_period AS period
        WHERE story.assignee_id = CAST(sqlc.arg(user_id) AS uuid)
          AND story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
          AND EXISTS (SELECT 1 FROM recipient_access)
          AND story.end_date BETWEEN period.as_of_date AND period.as_of_date + 7
          AND status.category NOT IN ('completed', 'cancelled', 'paused')
          AND story.deleted_at IS NULL
          AND story.archived_at IS NULL
          AND story.completed_at IS NULL
          AND (
              EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
              OR story.team_id IN (SELECT team_id FROM visible_teams)
          )
    ) AS integer) AS due_this_week_stories,
    CAST((
        SELECT COUNT(*)
        FROM public.objectives AS objective
        INNER JOIN public.objective_statuses AS status
            ON status.status_id = objective.status_id
        INNER JOIN public.workspace_settings AS settings
            ON settings.workspace_id = objective.workspace_id
        CROSS JOIN digest_period AS period
        WHERE objective.lead_user_id = CAST(sqlc.arg(user_id) AS uuid)
          AND objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
          AND EXISTS (SELECT 1 FROM recipient_access)
          AND objective.end_date BETWEEN period.as_of_date - 7 AND period.as_of_date + 7
          AND status.category NOT IN ('completed', 'cancelled', 'paused')
          AND settings.objective_enabled = TRUE
          AND (
              EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
              OR objective.team_id IN (SELECT team_id FROM visible_teams)
          )
    ) AS integer) AS objective_risks,
    CAST((
        SELECT COUNT(*)
        FROM public.story_comments AS comment
        INNER JOIN public.stories AS story
            ON story.id = comment.story_id
        CROSS JOIN digest_period AS period
        WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
          AND EXISTS (SELECT 1 FROM recipient_access)
          AND comment.commenter_id <> CAST(sqlc.arg(user_id) AS uuid)
          AND comment.created_at >= period.window_start
          AND comment.created_at <= period.as_of
          AND story.deleted_at IS NULL
          AND (
              EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
              OR story.team_id IN (SELECT team_id FROM visible_teams)
          )
    ) AS integer) AS team_comments;
