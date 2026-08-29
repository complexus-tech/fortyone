-- name: WorkspaceCanUseMaya :one
SELECT EXISTS (
    SELECT 1
    FROM workspaces workspace
    WHERE workspace.workspace_id = sqlc.arg(workspace_id)
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
);
