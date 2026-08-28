-- name: ClaimNewMutationApproval :one
INSERT INTO public.chat_mutation_approval_executions AS execution (
    session_id,
    user_id,
    workspace_id,
    tool_call_id,
    fingerprint,
    status,
    lease_token,
    lease_expires_at,
    attempt_count
)
SELECT
    session.id,
    session.user_id,
    session.workspace_id,
    sqlc.arg(tool_call_id),
    sqlc.arg(fingerprint),
    'ready',
    sqlc.arg(lease_token),
    CURRENT_TIMESTAMP + INTERVAL '30 seconds',
    1
FROM public.chat_sessions AS session
WHERE session.id = sqlc.arg(session_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
ON CONFLICT DO NOTHING
RETURNING execution.*;

-- name: LockMutationApproval :one
SELECT execution.*
FROM public.chat_mutation_approval_executions AS execution
INNER JOIN public.chat_sessions AS session
    ON session.id = execution.session_id
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
FOR UPDATE OF execution;

-- name: ReclaimExpiredReadyMutationApproval :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    lease_token = sqlc.arg(lease_token),
    lease_expires_at = CURRENT_TIMESTAMP + INTERVAL '30 seconds',
    attempt_count = attempt_count + 1,
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status = 'ready'
  AND execution.lease_expires_at <= CURRENT_TIMESTAMP
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
RETURNING execution.*;

-- name: ClaimRetryReadyMutationApproval :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    lease_token = sqlc.arg(lease_token),
    lease_expires_at = CURRENT_TIMESTAMP + INTERVAL '30 seconds',
    attempt_count = attempt_count + 1,
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status = 'retry_ready'
  AND (
      execution.lease_token IS NULL
      OR execution.lease_expires_at <= CURRENT_TIMESTAMP
  )
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
RETURNING execution.*;

-- name: FailExpiredExecutingMutationApproval :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    status = 'failed_uncertain',
    lease_token = NULL,
    lease_expires_at = NULL,
    failed_at = CURRENT_TIMESTAMP,
    failure_code = sqlc.arg(failure_code),
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status = 'executing'
  AND execution.lease_expires_at <= CURRENT_TIMESTAMP
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
RETURNING execution.*;

-- name: StartMutationApproval :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    status = 'executing',
    started_at = CURRENT_TIMESTAMP,
    failed_at = NULL,
    failure_code = NULL,
    lease_expires_at = CURRENT_TIMESTAMP + INTERVAL '3 minutes',
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.lease_token = sqlc.arg(lease_token)
  AND execution.status IN ('ready', 'retry_ready')
  AND execution.lease_expires_at > CURRENT_TIMESTAMP
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
RETURNING execution.*;

-- name: CompleteMutationApproval :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    status = 'completed',
    output = CAST(sqlc.arg(output) AS jsonb),
    completed_at = CURRENT_TIMESTAMP,
    lease_token = NULL,
    lease_expires_at = NULL,
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.lease_token = sqlc.arg(lease_token)
  AND execution.status = 'executing'
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
RETURNING execution.*;

-- name: FailMutationApproval :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    status = 'failed_uncertain',
    lease_token = NULL,
    lease_expires_at = NULL,
    failed_at = CURRENT_TIMESTAMP,
    failure_code = sqlc.arg(failure_code),
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.lease_token = sqlc.arg(lease_token)
  AND (
      execution.status = 'executing'
      OR (
          execution.status IN ('ready', 'retry_ready')
          AND sqlc.arg(failure_code) = 'start_transition_uncertain'
      )
  )
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
RETURNING execution.*;

-- name: ReconcileMutationApprovalVerifiedCompleted :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    status = 'completed',
    output = CAST(sqlc.arg(output) AS jsonb),
    completed_at = CURRENT_TIMESTAMP,
    lease_token = NULL,
    lease_expires_at = NULL,
    failed_at = NULL,
    failure_code = NULL,
    last_reconciliation_resolution = sqlc.arg(resolution),
    last_reconciliation_evidence = CAST(sqlc.arg(evidence) AS jsonb),
    last_reconciled_at = CURRENT_TIMESTAMP,
    reconciliation_count = reconciliation_count + 1,
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status = 'failed_uncertain'
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
RETURNING execution.*;

-- name: ReconcileMutationApprovalVerifiedNotApplied :one
UPDATE public.chat_mutation_approval_executions AS execution
SET
    status = 'ready',
    output = NULL,
    completed_at = NULL,
    lease_token = sqlc.arg(lease_token),
    lease_expires_at = CURRENT_TIMESTAMP,
    started_at = NULL,
    failed_at = NULL,
    failure_code = NULL,
    attempt_count = attempt_count + 1,
    last_reconciliation_resolution = sqlc.arg(resolution),
    last_reconciliation_evidence = CAST(sqlc.arg(evidence) AS jsonb),
    last_reconciled_at = CURRENT_TIMESTAMP,
    reconciliation_count = reconciliation_count + 1,
    updated_at = CURRENT_TIMESTAMP
FROM public.chat_sessions AS session
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status = 'failed_uncertain'
  AND session.id = execution.session_id
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
RETURNING execution.*;

-- name: LockUnresolvedMutationApprovalFingerprint :one
SELECT execution.*
FROM public.chat_mutation_approval_executions AS execution
INNER JOIN public.chat_sessions AS session
    ON session.id = execution.session_id
WHERE execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND (
      execution.session_id <> sqlc.arg(session_id)
      OR execution.tool_call_id <> sqlc.arg(tool_call_id)
  )
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status IN ('ready', 'retry_ready', 'executing', 'failed_uncertain')
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE OF execution;

-- name: TerminalizeExpiredReadyAndClaimReplacement :one
WITH terminalized AS (
    UPDATE public.chat_mutation_approval_executions AS execution
    SET
        status = 'completed',
        output = CAST(sqlc.arg(expired_output) AS jsonb),
        completed_at = CURRENT_TIMESTAMP,
        lease_token = NULL,
        lease_expires_at = NULL,
        failed_at = NULL,
        failure_code = NULL,
        updated_at = CURRENT_TIMESTAMP
    WHERE execution.session_id = sqlc.arg(previous_session_id)
      AND execution.user_id = sqlc.arg(user_id)
      AND execution.workspace_id = sqlc.arg(workspace_id)
      AND execution.tool_call_id = sqlc.arg(previous_tool_call_id)
      AND execution.fingerprint = sqlc.arg(fingerprint)
      AND execution.status = 'ready'
      AND execution.lease_expires_at <= CURRENT_TIMESTAMP
    RETURNING 1
)
INSERT INTO public.chat_mutation_approval_executions AS replacement (
    session_id,
    user_id,
    workspace_id,
    tool_call_id,
    fingerprint,
    status,
    lease_token,
    lease_expires_at,
    attempt_count
)
SELECT
    destination.id,
    destination.user_id,
    destination.workspace_id,
    sqlc.arg(tool_call_id),
    sqlc.arg(fingerprint),
    'ready',
    sqlc.arg(lease_token),
    CURRENT_TIMESTAMP + INTERVAL '30 seconds',
    1
FROM public.chat_sessions AS destination
CROSS JOIN terminalized
WHERE destination.id = sqlc.arg(session_id)
  AND destination.user_id = sqlc.arg(user_id)
  AND destination.workspace_id = sqlc.arg(workspace_id)
  AND destination.deleted_at IS NULL
RETURNING replacement.*;
