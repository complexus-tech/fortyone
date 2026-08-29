-- name: LockWorkspaceForRealtimeVoice :one
SELECT workspace_id
FROM workspaces
WHERE workspace_id = sqlc.arg(workspace_id)
    AND deleted_at IS NULL
FOR UPDATE;

-- name: RealtimeVoiceMonthlyUsageSeconds :one
SELECT CAST(
    COALESCE(
        SUM(
            EXTRACT(
                EPOCH FROM (
                    CASE
                        WHEN ended_at IS NULL THEN
                            started_at
                            + CAST(sqlc.arg(max_session_seconds) AS double precision) * INTERVAL '1 second'
                        ELSE LEAST(
                            ended_at,
                            started_at
                            + CAST(sqlc.arg(max_session_seconds) AS double precision) * INTERVAL '1 second'
                        )
                    END
                    - started_at
                )
            )
        ),
        0
    ) AS bigint
) AS used_seconds
FROM maya_realtime_voice_sessions
WHERE workspace_id = sqlc.arg(workspace_id)
    AND started_at >= date_trunc('month', CURRENT_TIMESTAMP)
    AND started_at < date_trunc('month', CURRENT_TIMESTAMP) + INTERVAL '1 month';

-- name: CreateRealtimeVoiceSession :one
INSERT INTO maya_realtime_voice_sessions (
    workspace_id,
    user_id
)
SELECT
    membership.workspace_id,
    membership.user_id
FROM workspace_members membership
INNER JOIN users actor
    ON actor.user_id = membership.user_id
    AND actor.is_active = TRUE
WHERE membership.workspace_id = sqlc.arg(workspace_id)
    AND membership.user_id = sqlc.arg(user_id)
RETURNING session_id;

-- name: EndRealtimeVoiceSession :execrows
UPDATE maya_realtime_voice_sessions
SET ended_at = COALESCE(
        ended_at,
        LEAST(
            CURRENT_TIMESTAMP,
            started_at + CAST(sqlc.arg(max_session_seconds) AS double precision) * INTERVAL '1 second'
        )
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
    AND user_id = sqlc.arg(user_id)
    AND session_id = sqlc.arg(session_id);

-- name: RealtimeVoiceSessionIsActive :one
SELECT EXISTS (
    SELECT 1
    FROM maya_realtime_voice_sessions
    WHERE session_id = sqlc.arg(session_id)
        AND workspace_id = sqlc.arg(workspace_id)
        AND user_id = sqlc.arg(user_id)
        AND ended_at IS NULL
        AND started_at
            + CAST(sqlc.arg(max_session_seconds) AS double precision) * INTERVAL '1 second'
            > CURRENT_TIMESTAMP
);

-- name: ClaimRealtimeToolCall :execrows
INSERT INTO maya_realtime_voice_tool_calls (
    session_id,
    call_id,
    tool_name,
    request_hash
)
VALUES (
    sqlc.arg(session_id),
    sqlc.arg(call_id),
    sqlc.arg(tool_name),
    sqlc.arg(request_hash)
)
ON CONFLICT (session_id, call_id) DO NOTHING;

-- name: GetRealtimeToolCall :one
SELECT
    request_hash,
    response
FROM maya_realtime_voice_tool_calls
WHERE session_id = sqlc.arg(session_id)
    AND call_id = sqlc.arg(call_id);

-- name: CompleteRealtimeToolCall :execrows
UPDATE maya_realtime_voice_tool_calls
SET response = CAST(sqlc.arg(response) AS jsonb),
    completed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE session_id = sqlc.arg(session_id)
    AND call_id = sqlc.arg(call_id)
    AND response IS NULL;
