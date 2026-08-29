-- name: CreateMayaRun :one
INSERT INTO maya_agent_runs (
    workspace_id,
    story_id,
    triggered_by_user_id,
    trigger_type,
    status,
    context
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(triggered_by_user_id),
    sqlc.arg(trigger_type),
    sqlc.arg(status),
    CAST(sqlc.arg(context_payload) AS jsonb)
)
RETURNING
    run_id,
    workspace_id,
    story_id,
    triggered_by_user_id,
    trigger_type,
    status,
    summary,
    context,
    error_message,
    started_at,
    completed_at,
    created_at,
    updated_at;

-- name: CompleteMayaRun :one
UPDATE maya_agent_runs
SET status = sqlc.arg(status),
    summary = sqlc.arg(summary),
    error_message = sqlc.narg(error_message),
    completed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE run_id = sqlc.arg(run_id)
    AND status = 'running'
RETURNING
    run_id,
    workspace_id,
    story_id,
    triggered_by_user_id,
    trigger_type,
    status,
    summary,
    context,
    error_message,
    started_at,
    completed_at,
    created_at,
    updated_at;

-- name: CreateMayaAction :one
INSERT INTO maya_agent_actions (
    run_id,
    workspace_id,
    story_id,
    action_type,
    status,
    reason,
    payload
)
VALUES (
    sqlc.arg(run_id),
    sqlc.arg(workspace_id),
    sqlc.arg(story_id),
    sqlc.arg(action_type),
    sqlc.arg(status),
    sqlc.arg(reason),
    CAST(sqlc.arg(payload) AS jsonb)
)
RETURNING
    action_id,
    run_id,
    workspace_id,
    story_id,
    action_type,
    status,
    reason,
    payload,
    error_message,
    applied_at,
    created_at,
    updated_at;

-- name: GetMayaWorkPlanRun :one
SELECT
    run_id,
    workspace_id,
    story_id,
    triggered_by_user_id,
    trigger_type,
    status,
    summary,
    context,
    error_message,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM maya_agent_runs
WHERE run_id = sqlc.arg(run_id)
    AND workspace_id = sqlc.arg(workspace_id)
    AND triggered_by_user_id = sqlc.arg(triggered_by_user_id);

-- name: ListMayaWorkPlanActions :many
SELECT
    action_id,
    run_id,
    workspace_id,
    story_id,
    action_type,
    status,
    reason,
    payload,
    error_message,
    applied_at,
    created_at,
    updated_at
FROM maya_agent_actions
WHERE run_id = sqlc.arg(run_id)
    AND workspace_id = sqlc.arg(workspace_id)
    AND story_id = sqlc.arg(story_id)
ORDER BY created_at ASC, action_id ASC;

-- name: MarkMayaActionApplied :execrows
UPDATE maya_agent_actions
SET status = 'applied',
    applied_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP,
    error_message = NULL
WHERE action_id = sqlc.arg(action_id)
    AND status = 'proposed';

-- name: MarkMayaActionFailed :execrows
UPDATE maya_agent_actions
SET status = 'failed',
    error_message = sqlc.arg(error_message),
    updated_at = CURRENT_TIMESTAMP
WHERE action_id = sqlc.arg(action_id)
    AND status = 'proposed';
