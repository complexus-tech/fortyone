-- name: ListFeedbackDigestRecipients :many
SELECT DISTINCT account.user_id,
       account.email AS user_email,
       COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), account.email) AS user_name,
       CAST(COALESCE(NULLIF(TRIM(account.timezone), ''), 'UTC') AS text) AS timezone,
       workspace.workspace_id,
       workspace.name AS workspace_name,
       workspace.slug AS workspace_slug
FROM feedback_board_subscriptions subscription
INNER JOIN feedback_boards board ON board.id = subscription.board_id
INNER JOIN teams team ON team.team_id = board.team_id AND team.workspace_id = board.workspace_id
INNER JOIN team_members team_member ON team_member.team_id = board.team_id AND team_member.user_id = subscription.user_id
INNER JOIN workspace_members workspace_member
    ON workspace_member.workspace_id = board.workspace_id
   AND workspace_member.user_id = subscription.user_id
   AND workspace_member.role IN ('admin', 'member')
INNER JOIN users account
    ON account.user_id = subscription.user_id
   AND account.is_active = true
   AND account.is_system = false
INNER JOIN workspaces workspace
    ON workspace.workspace_id = board.workspace_id
   AND workspace.deleted_at IS NULL
WHERE subscription.email_frequency IN ('daily', 'weekly')
  AND NULLIF(TRIM(account.email), '') IS NOT NULL
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR workspace.workspace_id > sqlc.arg(after_workspace_id)
      OR (
          workspace.workspace_id = sqlc.arg(after_workspace_id)
          AND account.user_id > sqlc.arg(after_user_id)
      )
  )
ORDER BY workspace.workspace_id, account.user_id
LIMIT sqlc.arg(row_limit);

-- name: ListFeedbackDigestSubscriptions :many
SELECT subscription.board_id,
       board.team_id,
       CAST(subscription.email_frequency AS text) AS email_frequency,
       subscription.created_at,
       subscription.last_digest_sent_at,
       subscription.last_digest_cursor_at
FROM feedback_board_subscriptions subscription
INNER JOIN feedback_boards board ON board.id = subscription.board_id
INNER JOIN teams team ON team.team_id = board.team_id AND team.workspace_id = board.workspace_id
INNER JOIN team_members team_member ON team_member.team_id = board.team_id AND team_member.user_id = subscription.user_id
INNER JOIN workspace_members workspace_member
    ON workspace_member.workspace_id = board.workspace_id
   AND workspace_member.user_id = subscription.user_id
   AND workspace_member.role IN ('admin', 'member')
INNER JOIN users account
    ON account.user_id = subscription.user_id
   AND account.is_active = true
   AND account.is_system = false
INNER JOIN workspaces workspace
    ON workspace.workspace_id = board.workspace_id
   AND workspace.deleted_at IS NULL
WHERE subscription.user_id = sqlc.arg(recipient_id)
  AND board.workspace_id = sqlc.arg(workspace_id)
  AND subscription.email_frequency IN ('daily', 'weekly')
ORDER BY subscription.board_id;

-- name: ClaimFeedbackDigestDelivery :one
INSERT INTO feedback_digest_deliveries (
    workspace_id,
    recipient_id,
    local_date,
    status,
    window_start,
    window_end
)
SELECT workspace.workspace_id,
       account.user_id,
       sqlc.arg(local_date),
       'processing',
       sqlc.arg(window_start),
       sqlc.arg(window_end)
FROM workspaces workspace
INNER JOIN workspace_members workspace_member
    ON workspace_member.workspace_id = workspace.workspace_id
   AND workspace_member.user_id = sqlc.arg(recipient_id)
   AND workspace_member.role IN ('admin', 'member')
INNER JOIN users account
    ON account.user_id = workspace_member.user_id
   AND account.is_active = true
   AND account.is_system = false
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL
ON CONFLICT (workspace_id, recipient_id, local_date)
DO UPDATE SET status = 'processing',
              window_start = EXCLUDED.window_start,
              window_end = EXCLUDED.window_end,
              item_count = 0,
              sent_at = NULL,
              last_error = NULL,
              updated_at = NOW()
WHERE feedback_digest_deliveries.status = 'failed'
   OR (
       feedback_digest_deliveries.status = 'processing'
       AND feedback_digest_deliveries.updated_at < sqlc.arg(stale_before)
   )
RETURNING id;

-- name: ListFeedbackDigestItems :many
WITH board_windows AS (
    SELECT boards.board_id, starts.window_start
    FROM UNNEST(CAST(sqlc.arg(board_ids) AS uuid[])) WITH ORDINALITY boards(board_id, ordinal)
    INNER JOIN UNNEST(CAST(sqlc.arg(window_starts) AS timestamptz[])) WITH ORDINALITY starts(window_start, ordinal)
        ON starts.ordinal = boards.ordinal
), eligible_items AS (
    SELECT item.id,
           board.team_id,
           item.title,
           item.description,
           CAST(COALESCE(
               NULLIF(TRIM(author.full_name), ''),
               NULLIF(TRIM(author.username), ''),
               NULLIF(TRIM(author.email), ''),
               'Customer'
           ) AS text) AS author_name,
           team.name AS team_name,
           item.status,
           item.created_at
    FROM board_windows board_window
    INNER JOIN feedback_items item ON item.board_id = board_window.board_id
    INNER JOIN feedback_boards board ON board.id = item.board_id AND board.workspace_id = item.workspace_id
    INNER JOIN feedback_board_subscriptions subscription ON subscription.board_id = board.id AND subscription.user_id = sqlc.arg(recipient_id)
    INNER JOIN teams team ON team.team_id = board.team_id AND team.workspace_id = board.workspace_id
    INNER JOIN team_members team_member ON team_member.team_id = board.team_id AND team_member.user_id = subscription.user_id
    INNER JOIN workspace_members workspace_member
        ON workspace_member.workspace_id = board.workspace_id
       AND workspace_member.user_id = subscription.user_id
       AND workspace_member.role IN ('admin', 'member')
    INNER JOIN users recipient
        ON recipient.user_id = subscription.user_id
       AND recipient.is_active = true
       AND recipient.is_system = false
    LEFT JOIN users author ON author.user_id = item.author_id
    WHERE item.workspace_id = sqlc.arg(workspace_id)
      AND item.deleted_at IS NULL
      AND item.submission_source IN ('portal', 'widget', 'integration')
      AND item.created_at > board_window.window_start
      AND item.created_at <= sqlc.arg(window_end)
)
SELECT eligible_items.id,
       eligible_items.team_id,
       eligible_items.title,
       eligible_items.description,
       eligible_items.author_name,
       eligible_items.team_name,
       eligible_items.status,
       eligible_items.created_at,
       CAST(COUNT(*) OVER () AS integer) AS total_count,
       CAST(COUNT(*) FILTER (WHERE eligible_items.status IN ('pending', 'reviewing')) OVER () AS integer) AS pending_review_count
FROM eligible_items
ORDER BY eligible_items.created_at DESC, eligible_items.id DESC
LIMIT sqlc.arg(row_limit);

-- name: AdvanceFeedbackDigestSubscriptionCursors :execrows
UPDATE feedback_board_subscriptions subscription
SET last_digest_sent_at = sqlc.arg(delivery_at),
    last_digest_cursor_at = sqlc.arg(window_end),
    updated_at = NOW()
FROM feedback_boards board
INNER JOIN teams team ON team.team_id = board.team_id AND team.workspace_id = board.workspace_id
INNER JOIN team_members team_member ON team_member.team_id = board.team_id AND team_member.user_id = sqlc.arg(recipient_id)
INNER JOIN workspace_members workspace_member
    ON workspace_member.workspace_id = board.workspace_id
   AND workspace_member.user_id = sqlc.arg(recipient_id)
   AND workspace_member.role IN ('admin', 'member')
INNER JOIN users account
    ON account.user_id = workspace_member.user_id
   AND account.is_active = true
   AND account.is_system = false
WHERE subscription.board_id = board.id
  AND subscription.user_id = sqlc.arg(recipient_id)
  AND board.workspace_id = sqlc.arg(workspace_id)
  AND subscription.board_id = ANY(CAST(sqlc.arg(board_ids) AS uuid[]));

-- name: CompleteFeedbackDigestDelivery :execrows
UPDATE feedback_digest_deliveries
SET status = sqlc.arg(status),
    item_count = sqlc.arg(item_count),
    sent_at = CAST(sqlc.narg(sent_at) AS timestamptz),
    last_error = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(delivery_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND recipient_id = sqlc.arg(recipient_id)
  AND status = 'processing';

-- name: FailFeedbackDigestDelivery :execrows
UPDATE feedback_digest_deliveries
SET status = 'failed',
    last_error = LEFT(sqlc.arg(failure), 2000),
    updated_at = NOW()
WHERE id = sqlc.arg(delivery_id)
  AND status = 'processing';
