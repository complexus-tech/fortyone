-- name: InitializeChatMessages :execrows
INSERT INTO public.chat_messages (
    session_id,
    messages
)
VALUES (
    sqlc.arg(session_id),
    CAST('[]' AS jsonb)
)
ON CONFLICT (session_id) DO NOTHING;

-- name: LockMessageWrite :one
SELECT messages.*
FROM public.chat_messages AS messages
INNER JOIN public.chat_sessions AS session
    ON session.id = messages.session_id
WHERE messages.session_id = sqlc.arg(session_id)
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL
FOR UPDATE OF messages;

-- name: ReserveMessageWrite :execrows
UPDATE public.chat_messages
SET
    messages = sqlc.arg(messages),
    write_generation = sqlc.arg(write_generation),
    write_token = sqlc.arg(write_token),
    write_operation = sqlc.arg(write_operation),
    write_finalized_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE session_id = sqlc.arg(session_id);

-- name: FinalizeMessageWriteCAS :execrows
UPDATE public.chat_messages
SET
    messages = sqlc.arg(messages),
    write_finalized_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE session_id = sqlc.arg(session_id)
  AND write_generation = sqlc.arg(write_generation)
  AND write_token = sqlc.arg(write_token);

-- name: RecoverMessageWrite :execrows
UPDATE public.chat_messages
SET
    messages = sqlc.arg(messages),
    updated_at = CURRENT_TIMESTAMP
WHERE session_id = sqlc.arg(session_id);

-- name: GetCompletedMutationApprovalOutputByFingerprint :one
SELECT execution.output
FROM public.chat_mutation_approval_executions AS execution
INNER JOIN public.chat_sessions AS session
    ON session.id = execution.session_id
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.fingerprint = sqlc.arg(fingerprint)
  AND execution.status = 'completed'
  AND session.user_id = sqlc.arg(user_id)
  AND session.workspace_id = sqlc.arg(workspace_id)
  AND session.deleted_at IS NULL;

-- name: HasPendingMutationApproval :one
SELECT EXISTS (
    SELECT 1
    FROM public.chat_mutation_approval_executions AS execution
    WHERE execution.session_id = sqlc.arg(session_id)
      AND execution.user_id = sqlc.arg(user_id)
      AND execution.workspace_id = sqlc.arg(workspace_id)
      AND (
          execution.status = 'retry_ready'
          OR (
              execution.status IN ('ready', 'executing')
              AND execution.lease_expires_at > CURRENT_TIMESTAMP
          )
      )
);

-- name: PrepareMutationApprovalRetry :one
WITH prepared AS (
    UPDATE public.chat_mutation_approval_executions AS execution
    SET
        status = 'retry_ready',
        lease_token = NULL,
        lease_expires_at = NULL,
        started_at = NULL,
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
      AND execution.reconciliation_count = 0
      AND session.id = execution.session_id
      AND session.user_id = sqlc.arg(user_id)
      AND session.workspace_id = sqlc.arg(workspace_id)
      AND session.deleted_at IS NULL
    RETURNING TRUE AS prepared
), existing AS (
    SELECT TRUE AS prepared
    FROM public.chat_mutation_approval_executions AS execution
    INNER JOIN public.chat_sessions AS session
        ON session.id = execution.session_id
    WHERE execution.session_id = sqlc.arg(session_id)
      AND execution.user_id = sqlc.arg(user_id)
      AND execution.workspace_id = sqlc.arg(workspace_id)
      AND execution.tool_call_id = sqlc.arg(tool_call_id)
      AND execution.fingerprint = sqlc.arg(fingerprint)
      AND execution.status = 'retry_ready'
      AND execution.reconciliation_count = 1
      AND execution.last_reconciliation_resolution = sqlc.arg(resolution)
      AND session.user_id = sqlc.arg(user_id)
      AND session.workspace_id = sqlc.arg(workspace_id)
      AND session.deleted_at IS NULL
)
SELECT prepared
FROM prepared
UNION ALL
SELECT prepared
FROM existing
LIMIT 1;

-- name: ExpireStaleMutationApprovals :execrows
UPDATE public.chat_mutation_approval_executions
SET
    status = 'failed_uncertain',
    failure_code = sqlc.arg(failure_code),
    lease_token = NULL,
    lease_expires_at = NULL,
    failed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE session_id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND status = 'executing'
  AND lease_expires_at <= CURRENT_TIMESTAMP;

-- name: GetDurableApprovalReceipt :one
SELECT
    execution.status,
    execution.output,
    CAST(COALESCE(execution.lease_expires_at <= CURRENT_TIMESTAMP, FALSE) AS boolean) AS lease_expired
FROM public.chat_mutation_approval_executions AS execution
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.status IN ('completed', 'failed_uncertain', 'ready', 'retry_ready');

-- name: GetCompletedApprovalOutput :one
SELECT execution.output
FROM public.chat_mutation_approval_executions AS execution
WHERE execution.session_id = sqlc.arg(session_id)
  AND execution.user_id = sqlc.arg(user_id)
  AND execution.workspace_id = sqlc.arg(workspace_id)
  AND execution.tool_call_id = sqlc.arg(tool_call_id)
  AND execution.status = 'completed';
