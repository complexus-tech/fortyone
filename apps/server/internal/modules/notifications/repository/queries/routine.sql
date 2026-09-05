-- name: ListRoutineEmailRecipients :many
SELECT
    account.user_id,
    workspace.workspace_id,
    account.email,
    COALESCE(NULLIF(account.full_name, ''), account.username) AS name,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    CAST(COALESCE(NULLIF(account.timezone, ''), 'UTC') AS text) AS timezone,
    CAST(COALESCE(preference.preferences -> 'weekly_digest' ->> 'email', 'true') AS boolean) AS weekly_enabled
FROM users AS account
JOIN workspace_members AS member
    ON member.user_id = account.user_id AND member.role IN ('admin', 'member', 'guest')
JOIN workspaces AS workspace
    ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
LEFT JOIN notification_preferences AS preference
    ON preference.user_id = account.user_id AND preference.workspace_id = workspace.workspace_id
WHERE account.is_active
    AND NOT account.is_system
    AND NULLIF(TRIM(account.email), '') IS NOT NULL
    AND (
        NOT CAST(sqlc.arg(has_cursor) AS boolean)
        OR (workspace.workspace_id, account.user_id) > (sqlc.arg(after_workspace_id), sqlc.arg(after_user_id))
    )
ORDER BY workspace.workspace_id, account.user_id
LIMIT sqlc.arg(row_limit);

-- name: LockRoutineEmailRecipient :exec
SELECT pg_advisory_xact_lock(hashtextextended(CAST(sqlc.arg(recipient_id) AS text), 184));

-- name: HasActiveRoutineEmailClaim :one
SELECT EXISTS (
    SELECT 1 FROM routine_email_deliveries
    WHERE recipient_id = sqlc.arg(recipient_id)
        AND status = 'processing'
        AND claimed_at > sqlc.arg(stale_before)
);

-- name: ClaimRoutineEmail :one
INSERT INTO routine_email_deliveries (
    recipient_id, workspace_id, delivery_key, kind, local_date, status, claimed_at
)
VALUES (
    sqlc.arg(recipient_id), sqlc.arg(workspace_id), sqlc.arg(delivery_key),
    sqlc.arg(kind), sqlc.arg(local_date), 'processing', sqlc.arg(now)
)
ON CONFLICT (recipient_id, workspace_id, delivery_key) DO UPDATE
SET id = gen_random_uuid(), status = 'processing',
    claimed_at = EXCLUDED.claimed_at, local_date = EXCLUDED.local_date
WHERE routine_email_deliveries.status = 'failed'
    OR (routine_email_deliveries.status = 'processing'
        AND routine_email_deliveries.claimed_at <= sqlc.arg(stale_before))
RETURNING id;

-- name: CompleteRoutineEmail :execrows
UPDATE routine_email_deliveries
SET status = sqlc.arg(status), completed_at = sqlc.arg(now)
WHERE id = sqlc.arg(id)
    AND recipient_id = sqlc.arg(recipient_id)
    AND workspace_id = sqlc.arg(workspace_id)
    AND status = 'processing';

-- name: FailRoutineEmail :execrows
UPDATE routine_email_deliveries SET status = 'failed'
WHERE id = sqlc.arg(id) AND status = 'processing';

-- name: GetRoutineEmailRecipient :one
SELECT
    account.user_id,
    workspace.workspace_id,
    account.email,
    COALESCE(NULLIF(account.full_name, ''), account.username) AS name,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    CAST(COALESCE(NULLIF(account.timezone, ''), 'UTC') AS text) AS timezone,
    CAST(COALESCE(preference.preferences -> 'weekly_digest' ->> 'email', 'true') AS boolean) AS weekly_enabled
FROM users AS account
JOIN workspace_members AS member
    ON member.user_id = account.user_id AND member.role IN ('admin', 'member', 'guest')
JOIN workspaces AS workspace
    ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
LEFT JOIN notification_preferences AS preference
    ON preference.user_id = account.user_id AND preference.workspace_id = workspace.workspace_id
WHERE account.is_active
    AND NOT account.is_system
    AND NULLIF(TRIM(account.email), '') IS NOT NULL
    AND account.user_id = sqlc.arg(recipient_id)
    AND workspace.workspace_id = sqlc.arg(workspace_id);

-- name: HasRoutineEmailGuidance :one
SELECT EXISTS (
    SELECT 1 FROM routine_email_deliveries
    WHERE recipient_id = sqlc.arg(recipient_id)
        AND workspace_id = sqlc.arg(workspace_id)
        AND delivery_key = sqlc.arg(delivery_key)
        AND status = 'sent'
);

-- name: RecordRoutineEmailGuidance :exec
INSERT INTO routine_email_deliveries (
    recipient_id, workspace_id, delivery_key, kind, local_date, status, claimed_at, completed_at
) VALUES (
    sqlc.arg(recipient_id), sqlc.arg(workspace_id), sqlc.arg(delivery_key),
    'briefing', sqlc.arg(local_date), 'sent', sqlc.arg(now), sqlc.arg(now)
)
ON CONFLICT (recipient_id, workspace_id, delivery_key) DO UPDATE
SET status = 'sent', completed_at = EXCLUDED.completed_at;
