-- name: ListOverdueStoryGuidanceRecipients :many
SELECT DISTINCT
    story.assignee_id,
    assignee.email AS assignee_email,
    COALESCE(NULLIF(assignee.full_name, ''), assignee.username) AS assignee_name,
    workspace.workspace_id,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    CAST(COALESCE(preference.preferences -> 'reminders' ->> 'email', 'true') AS boolean) AS email_enabled
FROM stories AS story
JOIN users AS assignee
    ON assignee.user_id = story.assignee_id
JOIN workspaces AS workspace
    ON workspace.workspace_id = story.workspace_id
JOIN workspace_members AS membership
    ON membership.workspace_id = story.workspace_id
    AND membership.user_id = story.assignee_id
    AND membership.role IN ('admin', 'member', 'guest')
JOIN statuses AS status
    ON status.status_id = story.status_id
LEFT JOIN notification_preferences AS preference
    ON preference.user_id = story.assignee_id
    AND preference.workspace_id = story.workspace_id
WHERE story.end_date IS NOT NULL
    AND status.category NOT IN ('completed', 'cancelled', 'paused')
    AND workspace.deleted_at IS NULL
    AND (
        membership.role = 'admin'
        OR EXISTS (
            SELECT 1
            FROM team_members AS team_membership
            WHERE team_membership.team_id = story.team_id
                AND team_membership.user_id = story.assignee_id
        )
    )
    AND story.deleted_at IS NULL
    AND story.archived_at IS NULL
    AND story.completed_at IS NULL
    AND story.assignee_id IS NOT NULL
    AND story.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '3 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '3 days'
    AND assignee.is_active = true
    AND assignee.is_system = false
    AND NULLIF(TRIM(assignee.email), '') IS NOT NULL
    AND CAST(COALESCE(preference.preferences -> 'reminders' ->> 'email', 'true') AS boolean) = true
    AND (
        NOT CAST(sqlc.arg(has_cursor) AS boolean)
        OR story.assignee_id > sqlc.arg(after_assignee_id)
        OR (
            story.assignee_id = sqlc.arg(after_assignee_id)
            AND story.workspace_id > sqlc.arg(after_workspace_id)
        )
    )
ORDER BY story.assignee_id, workspace.workspace_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: ListOverdueStoryGuidanceItems :many
WITH story_deadlines AS (
    SELECT
        story.id,
        story.sequence_id,
        story.title,
        story.end_date,
        story.assignee_id,
        story.workspace_id,
        story.team_id,
        assignee.email AS assignee_email,
        COALESCE(NULLIF(assignee.full_name, ''), assignee.username) AS assignee_name,
        workspace.name AS workspace_name,
        workspace.slug AS workspace_slug,
        team.name AS team_name,
        team.code AS team_code,
        status.name AS status_name,
        status.category AS status_category,
        CASE
            WHEN story.end_date = CAST(sqlc.arg(as_of) AS date) THEN 'due_today'
            WHEN story.end_date = CAST(sqlc.arg(as_of) AS date) + INTERVAL '1 day' THEN 'due_tomorrow'
            WHEN story.end_date = CAST(sqlc.arg(as_of) AS date) + INTERVAL '3 days' THEN 'due_in_3_days'
            WHEN story.end_date < CAST(sqlc.arg(as_of) AS date) THEN 'overdue'
            ELSE 'future'
        END AS deadline_status,
        CASE
            WHEN story.end_date < CAST(sqlc.arg(as_of) AS date) THEN CAST(CAST(sqlc.arg(as_of) AS date) - story.end_date AS integer)
            ELSE CAST(story.end_date - CAST(sqlc.arg(as_of) AS date) AS integer)
        END AS days_difference
    FROM stories AS story
    JOIN users AS assignee
        ON assignee.user_id = story.assignee_id
    JOIN workspaces AS workspace
        ON workspace.workspace_id = story.workspace_id
    JOIN workspace_members AS membership
        ON membership.workspace_id = story.workspace_id
        AND membership.user_id = story.assignee_id
        AND membership.role IN ('admin', 'member', 'guest')
    JOIN teams AS team
        ON team.team_id = story.team_id
    JOIN statuses AS status
        ON status.status_id = story.status_id
    LEFT JOIN notification_preferences AS preference
        ON preference.user_id = story.assignee_id
        AND preference.workspace_id = story.workspace_id
    WHERE story.assignee_id = sqlc.arg(assignee_id)
        AND story.workspace_id = sqlc.arg(workspace_id)
        AND workspace.deleted_at IS NULL
        AND (
            membership.role = 'admin'
            OR EXISTS (
                SELECT 1
                FROM team_members AS team_membership
                WHERE team_membership.team_id = story.team_id
                    AND team_membership.user_id = story.assignee_id
            )
        )
        AND story.end_date IS NOT NULL
        AND status.category NOT IN ('completed', 'cancelled', 'paused')
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND story.completed_at IS NULL
        AND story.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '3 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '3 days'
        AND assignee.is_active = true
        AND assignee.is_system = false
        AND NULLIF(TRIM(assignee.email), '') IS NOT NULL
        AND CAST(COALESCE(preference.preferences -> 'reminders' ->> 'email', 'true') AS boolean) = true
)
SELECT
    id,
    sequence_id,
    title,
    end_date,
    assignee_id,
    workspace_id,
    team_id,
    assignee_email,
    assignee_name,
    workspace_name,
    workspace_slug,
    team_name,
    team_code,
    status_name,
    status_category,
    deadline_status,
    days_difference
FROM story_deadlines
WHERE deadline_status IN ('due_today', 'due_tomorrow', 'due_in_3_days', 'overdue')
ORDER BY deadline_status, end_date, id;
