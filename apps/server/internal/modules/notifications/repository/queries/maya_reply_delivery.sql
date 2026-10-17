-- name: EnqueueMayaReplyNotification :one
WITH eligible AS (
    SELECT comment.comment_id, comment.created_at, story.workspace_id
    FROM public.story_comments AS comment
    JOIN public.stories AS story ON story.id = comment.story_id AND story.deleted_at IS NULL
    JOIN public.workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
    JOIN public.users AS author ON author.user_id = comment.commenter_id AND author.is_active = TRUE AND author.is_system = TRUE AND author.username = 'maya'
    WHERE comment.comment_id = sqlc.arg(comment_id)
      AND story.id = sqlc.arg(story_id) AND story.workspace_id = sqlc.arg(workspace_id)
      AND author.user_id = sqlc.arg(actor_id)
), inserted AS (
    INSERT INTO public.maya_reply_notification_deliveries
        (dedupe_key, comment_id, recipient_id, workspace_id, payload, occurred_at, available_at)
    SELECT sqlc.arg(dedupe_key), eligible.comment_id, sqlc.arg(recipient_id), eligible.workspace_id,
        sqlc.arg(payload), sqlc.arg(occurred_at), eligible.created_at + INTERVAL '5 seconds'
    FROM eligible
    ON CONFLICT (comment_id, recipient_id) DO UPDATE SET comment_id = EXCLUDED.comment_id
    RETURNING *
)
SELECT dedupe_key = CAST(sqlc.arg(dedupe_key) AS text)
    AND payload = CAST(sqlc.arg(payload) AS jsonb)
    AND occurred_at = CAST(sqlc.arg(occurred_at) AS timestamptz) AS matches_payload
FROM inserted;

-- name: NextMayaReplyNotification :one
SELECT delivery.*,
    EXISTS (SELECT 1 FROM public.story_comment_seen AS seen
            WHERE seen.comment_id = delivery.comment_id AND seen.user_id = delivery.recipient_id) AS seen,
    EXISTS (
        SELECT 1 FROM public.story_comments AS comment
        JOIN public.stories AS story ON story.id = comment.story_id AND story.workspace_id = delivery.workspace_id AND story.deleted_at IS NULL
        JOIN public.workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
        JOIN public.users AS recipient ON recipient.user_id = delivery.recipient_id AND recipient.is_active = TRUE AND recipient.is_system = FALSE
        JOIN public.workspace_members AS member ON member.workspace_id = workspace.workspace_id AND member.user_id = recipient.user_id AND member.role IN ('admin', 'member', 'guest')
        WHERE comment.comment_id = delivery.comment_id
          AND (member.role = 'admin' OR EXISTS (
              SELECT 1 FROM public.team_members AS team_member
              WHERE team_member.team_id = story.team_id AND team_member.user_id = recipient.user_id
          ))
    ) AS authorized,
    CAST(COALESCE((
        SELECT CASE WHEN jsonb_typeof(preference.preferences -> 'comment_reply' -> 'in_app') = 'boolean'
            THEN CAST(preference.preferences -> 'comment_reply' ->> 'in_app' AS boolean) ELSE TRUE END
        FROM public.notification_preferences AS preference
        WHERE preference.user_id = delivery.recipient_id AND preference.workspace_id = delivery.workspace_id
    ), TRUE) AS boolean) AS in_app_enabled
FROM public.maya_reply_notification_deliveries AS delivery
WHERE delivery.completed_at IS NULL AND delivery.available_at <= clock_timestamp()
ORDER BY delivery.available_at, delivery.dedupe_key
LIMIT 1 FOR UPDATE OF delivery SKIP LOCKED;

-- name: CompleteMayaReplyNotification :exec
UPDATE public.maya_reply_notification_deliveries SET completed_at = clock_timestamp()
WHERE dedupe_key = sqlc.arg(dedupe_key);

-- name: RetryMayaReplyNotification :exec
UPDATE public.maya_reply_notification_deliveries
SET attempts = LEAST(attempts + 1, 1000),
    available_at = clock_timestamp() + LEAST(300, 5 * power(2, LEAST(attempts, 6))) * INTERVAL '1 second'
WHERE dedupe_key = sqlc.arg(dedupe_key);
