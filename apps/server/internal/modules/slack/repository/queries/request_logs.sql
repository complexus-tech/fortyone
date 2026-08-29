-- name: InsertRequestLog :exec
INSERT INTO public.slack_request_logs (
    request_type,
    endpoint,
    workspace_id,
    slack_team_id,
    slack_user_id,
    slack_channel_id,
    command,
    trigger_id,
    request_body,
    headers,
    response_code,
    outcome,
    error_message
) VALUES (
    CAST(sqlc.arg(request_type) AS text),
    CAST(sqlc.arg(endpoint) AS text),
    sqlc.narg(workspace_id),
    sqlc.narg(slack_team_id),
    sqlc.narg(slack_user_id),
    sqlc.narg(slack_channel_id),
    sqlc.narg(command),
    sqlc.narg(trigger_id),
    sqlc.narg(request_body),
    CAST(sqlc.arg(headers) AS jsonb),
    CAST(sqlc.arg(response_code) AS integer),
    CAST(sqlc.arg(outcome) AS text),
    sqlc.narg(error_message)
);

-- name: HasSlackUserOnboardingReceipt :one
SELECT EXISTS (
    SELECT 1
    FROM public.slack_user_onboarding_receipts
    WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND external_identity_digest = CAST(sqlc.arg(identity_digest) AS bytea)
);

