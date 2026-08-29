-- name: ListMayaAssignmentCandidates :many
SELECT
    story.id,
    story.workspace_id,
    story.team_id,
    story.assignee_id
FROM stories story
INNER JOIN workspaces workspace
    ON workspace.workspace_id = story.workspace_id
WHERE story.auto_scheduling_enabled = TRUE
    AND story.assignee_id IS NOT NULL
    AND (
        story.assignee_id = sqlc.arg(maya_actor_id)
        OR NOT EXISTS (
            SELECT 1
            FROM calendar_maya_schedule_ownerships ownership
            WHERE ownership.workspace_id = story.workspace_id
                AND ownership.story_id = story.id
        )
    )
    AND story.id > sqlc.arg(after_story_id)
    AND story.deleted_at IS NULL
    AND story.archived_at IS NULL
    AND story.is_draft = FALSE
    AND workspace.deleted_at IS NULL
    AND (
        workspace.trial_ends_on > CURRENT_TIMESTAMP
        OR EXISTS (
            SELECT 1
            FROM workspace_subscriptions subscription
            WHERE subscription.workspace_id = workspace.workspace_id
                AND subscription.subscription_tier <> 'free'
                AND subscription.subscription_status IN ('active', 'trialing', 'past_due')
        )
    )
    AND NOT EXISTS (
        SELECT 1
        FROM statuses status
        WHERE status.status_id = story.status_id
            AND status.category IN ('completed', 'cancelled')
    )
ORDER BY story.id
LIMIT CAST(sqlc.arg(row_limit) AS integer);

-- name: ListWorkspaceScheduleCandidates :many
SELECT
    story.id,
    story.workspace_id,
    story.team_id,
    story.assignee_id
FROM stories story
INNER JOIN workspaces workspace
    ON workspace.workspace_id = story.workspace_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
    AND story.id > sqlc.arg(after_story_id)
    AND story.auto_scheduling_enabled = TRUE
    AND story.assignee_id IS NOT NULL
    AND story.deleted_at IS NULL
    AND story.archived_at IS NULL
    AND story.is_draft = FALSE
    AND workspace.deleted_at IS NULL
    AND (
        workspace.trial_ends_on > CURRENT_TIMESTAMP
        OR EXISTS (
            SELECT 1
            FROM workspace_subscriptions subscription
            WHERE subscription.workspace_id = workspace.workspace_id
                AND subscription.subscription_tier <> 'free'
                AND subscription.subscription_status IN ('active', 'trialing', 'past_due')
        )
    )
    AND NOT EXISTS (
        SELECT 1
        FROM statuses status
        WHERE status.status_id = story.status_id
            AND status.category IN ('completed', 'cancelled')
    )
ORDER BY story.id
LIMIT CAST(sqlc.arg(row_limit) AS integer);
