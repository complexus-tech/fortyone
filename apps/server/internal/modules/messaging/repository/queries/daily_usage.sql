-- name: ClaimAssistantUsageEvent :one
INSERT INTO messaging_assistant_usage_events (
    inbound_event_id,
    workspace_id,
    provider,
    external_workspace_id,
    external_event_id,
    attempt_count,
    usage_date,
    input_tokens,
    output_tokens,
    total_tokens
) VALUES (
    sqlc.arg(inbound_event_id),
    sqlc.arg(workspace_id),
    sqlc.arg(provider),
    sqlc.arg(external_workspace_id),
    sqlc.arg(external_event_id),
    sqlc.arg(attempt_count),
    CAST(NOW() AT TIME ZONE 'UTC' AS date),
    sqlc.arg(input_tokens),
    sqlc.arg(output_tokens),
    sqlc.arg(total_tokens)
)
ON CONFLICT (inbound_event_id, attempt_count) DO NOTHING
RETURNING 1;

-- name: AddAssistantDailyUsage :one
INSERT INTO messaging_assistant_daily_usage (
    workspace_id,
    usage_date,
    input_tokens,
    output_tokens,
    total_tokens,
    request_count
) VALUES (
    sqlc.arg(workspace_id),
    CAST(NOW() AT TIME ZONE 'UTC' AS date),
    sqlc.arg(input_tokens),
    sqlc.arg(output_tokens),
    sqlc.arg(total_tokens),
    1
)
ON CONFLICT (workspace_id, usage_date) DO UPDATE
SET input_tokens = messaging_assistant_daily_usage.input_tokens + EXCLUDED.input_tokens,
    output_tokens = messaging_assistant_daily_usage.output_tokens + EXCLUDED.output_tokens,
    total_tokens = messaging_assistant_daily_usage.total_tokens + EXCLUDED.total_tokens,
    request_count = messaging_assistant_daily_usage.request_count + 1,
    updated_at = NOW()
RETURNING input_tokens, output_tokens, total_tokens, request_count;

-- name: GetAssistantDailyUsage :one
SELECT CAST(COALESCE(SUM(input_tokens), 0) AS bigint) AS input_tokens,
       CAST(COALESCE(SUM(output_tokens), 0) AS bigint) AS output_tokens,
       CAST(COALESCE(SUM(total_tokens), 0) AS bigint) AS total_tokens,
       CAST(COALESCE(SUM(request_count), 0) AS bigint) AS request_count
FROM messaging_assistant_daily_usage
WHERE workspace_id = sqlc.arg(workspace_id)
  AND usage_date = CAST(NOW() AT TIME ZONE 'UTC' AS date);
