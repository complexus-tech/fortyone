-- name: CreateAgentInvocation :one
INSERT INTO public.agent_invocations (
    workspace_id, requester_id, team_id, surface, source_id, conversation_id, entity_id, parent_run_id,
    source_revision, source_hash, idempotency_key, request_hash, payload
)
SELECT sqlc.arg(workspace_id), sqlc.arg(requester_id), CAST(sqlc.narg(team_id) AS uuid), CAST(sqlc.arg(surface) AS text),
    sqlc.arg(source_id), sqlc.arg(conversation_id), sqlc.arg(entity_id), sqlc.narg(parent_run_id), CAST(sqlc.arg(source_revision) AS text),
    sqlc.arg(source_hash), sqlc.arg(idempotency_key), sqlc.arg(request_hash), sqlc.arg(payload)
WHERE EXISTS (
    SELECT 1 FROM public.workspace_members AS member
    INNER JOIN public.users AS account ON account.user_id = member.user_id AND account.is_active = TRUE AND account.is_system = FALSE
    INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
    WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(requester_id)
      AND member.role <> CAST('guest' AS public.user_role)
      AND ((CAST(sqlc.arg(surface) AS text) = 'document_comment' AND CAST(sqlc.narg(team_id) AS uuid) IS NULL
          AND public.can_read_document(member.workspace_id, member.user_id, sqlc.arg(entity_id)))
        OR (CAST(sqlc.arg(surface) AS text) <> 'document_comment' AND EXISTS (
          SELECT 1 FROM public.teams AS team
          JOIN public.team_members AS team_member ON team_member.team_id = team.team_id AND team_member.user_id = member.user_id
          WHERE team.workspace_id = workspace.workspace_id AND team.team_id = CAST(sqlc.narg(team_id) AS uuid)
        )))
)
AND (CAST(sqlc.arg(surface) AS text) <> 'chat_message' OR EXISTS (
    SELECT 1 FROM public.chat_conversation_messages AS message
    JOIN public.chat_conversations AS conversation ON conversation.id = message.conversation_id AND conversation.workspace_id = message.workspace_id
    WHERE message.id = sqlc.arg(source_id) AND message.workspace_id = sqlc.arg(workspace_id)
      AND conversation.id = sqlc.arg(entity_id) AND conversation.kind = 'team' AND conversation.team_id = CAST(sqlc.narg(team_id) AS uuid)
      AND message.author_id = sqlc.arg(requester_id) AND message.deleted_at IS NULL
      AND COALESCE(message.thread_root_id, message.id) = sqlc.arg(conversation_id)
      AND CAST(message.version AS text) = CAST(sqlc.arg(source_revision) AS text)
      AND public.can_read_chat_conversation(message.workspace_id, message.author_id, conversation.id)
))
ON CONFLICT (workspace_id, surface, idempotency_key) DO NOTHING
RETURNING *;

-- name: LockAgentInvocationContinuation :one
SELECT * FROM public.agent_invocations
WHERE id = sqlc.arg(id) AND workspace_id = sqlc.arg(workspace_id)
    AND team_id IS NOT DISTINCT FROM CAST(sqlc.narg(team_id) AS uuid) AND surface = sqlc.arg(surface) AND entity_id = sqlc.arg(entity_id)
    AND conversation_id = sqlc.arg(conversation_id)
FOR UPDATE;

-- name: GetAgentInvocationSourceParent :one
SELECT parent_run_id FROM public.agent_invocations
WHERE workspace_id = sqlc.arg(workspace_id) AND requester_id = sqlc.arg(requester_id) AND surface = sqlc.arg(surface)
    AND entity_id = sqlc.arg(entity_id) AND conversation_id = sqlc.arg(conversation_id) AND source_id = sqlc.arg(source_id)
ORDER BY created_at, id LIMIT 1;

-- name: FindAgentInvocationClarification :one
SELECT id FROM public.agent_invocations
WHERE workspace_id = sqlc.arg(workspace_id) AND surface = sqlc.arg(surface)
    AND entity_id = sqlc.arg(entity_id) AND conversation_id = sqlc.arg(conversation_id) AND status IN ('needs_input','awaiting_approval')
    AND (source_id = sqlc.arg(reply_to_id) OR reply_id = sqlc.arg(reply_to_id) OR proposal_reply_id = sqlc.arg(reply_to_id))
ORDER BY created_at DESC, id DESC LIMIT 1;

-- name: CompleteAgentInvocationContinuation :execrows
UPDATE public.agent_invocations SET status = 'succeeded', summary = 'Continued in a new request.',
    version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status IN ('needs_input','awaiting_approval');

-- name: SaveAgentInvocationProposalPublication :one
UPDATE public.agent_invocations SET proposal_reply_id = sqlc.arg(proposal_reply_id), version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp() AND proposal_reply_id IS NULL
RETURNING *;

-- name: AwaitAgentInvocationApproval :one
UPDATE public.agent_invocations SET status = 'awaiting_approval', proposal_delivered_at = clock_timestamp(),
    summary = COALESCE(proposal->>'summary', ''),
    lease_token = NULL, lease_expires_at = NULL, version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
    AND proposal_reply_id IS NOT NULL AND outcome IS NULL AND approved_at IS NULL
    AND proposal->>'requiresApproval' = 'true'
RETURNING *;

-- name: GetAgentInvocationByKey :one
SELECT invocation.* FROM public.agent_invocations AS invocation
WHERE invocation.workspace_id = sqlc.arg(workspace_id) AND invocation.surface = sqlc.arg(surface)
    AND invocation.idempotency_key = sqlc.arg(idempotency_key) AND invocation.requester_id = sqlc.arg(requester_id)
    AND EXISTS (
        SELECT 1 FROM public.workspace_members AS member
        INNER JOIN public.users AS account ON account.user_id = member.user_id AND account.is_active = TRUE AND account.is_system = FALSE
        INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
        WHERE member.workspace_id = invocation.workspace_id AND member.user_id = invocation.requester_id
          AND member.role <> CAST('guest' AS public.user_role)
          AND ((invocation.surface = 'document_comment' AND invocation.team_id IS NULL
              AND public.can_read_document(member.workspace_id, member.user_id, invocation.entity_id))
            OR (invocation.surface <> 'document_comment' AND EXISTS (
              SELECT 1 FROM public.teams AS team
              JOIN public.team_members AS team_member ON team_member.team_id = team.team_id AND team_member.user_id = member.user_id
              WHERE team.workspace_id = workspace.workspace_id AND team.team_id = invocation.team_id
            )))
    )
    AND (invocation.surface <> 'chat_message' OR public.can_read_chat_conversation(invocation.workspace_id, invocation.requester_id, invocation.entity_id));

-- name: GetAgentInvocation :one
SELECT * FROM public.agent_invocations WHERE id = sqlc.arg(id) AND workspace_id = sqlc.arg(workspace_id);

-- name: ListAgentInvocations :many
SELECT * FROM public.agent_invocations
WHERE workspace_id = sqlc.arg(workspace_id) AND entity_id = sqlc.arg(entity_id)
  AND COALESCE(proposal->>'ignore', 'false') <> 'true'
ORDER BY created_at DESC, id DESC LIMIT 100;

-- name: ListPendingAgentInvocationIDs :many
SELECT id FROM public.agent_invocations
WHERE ((status = 'queued' AND available_at <= clock_timestamp()) OR (status = 'running' AND lease_expires_at < clock_timestamp()))
  AND (surface <> 'chat_message' OR proposal->>'ignore' = 'true' OR EXISTS (
      SELECT 1 FROM public.workspaces AS workspace
      LEFT JOIN public.workspace_settings AS settings ON settings.workspace_id = workspace.workspace_id
      WHERE workspace.workspace_id = agent_invocations.workspace_id
        AND workspace.deleted_at IS NULL AND COALESCE(settings.chat_enabled, TRUE)
  ))
ORDER BY CASE WHEN status = 'running' THEN lease_expires_at ELSE available_at END, created_at, id LIMIT sqlc.arg(row_limit);

-- name: ClaimAgentInvocation :one
UPDATE public.agent_invocations SET status = 'running', lease_token = sqlc.arg(lease_token),
    lease_expires_at = clock_timestamp() + INTERVAL '2 minutes', lease_generation = lease_generation + 1,
    attempt_count = attempt_count + 1, retry_attempt_count = retry_attempt_count + 1,
    version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND ((status = 'queued' AND available_at <= clock_timestamp()) OR (status = 'running' AND lease_expires_at < clock_timestamp()))
  AND (surface <> 'chat_message' OR proposal->>'ignore' = 'true' OR EXISTS (
      SELECT 1 FROM public.workspaces AS workspace
      LEFT JOIN public.workspace_settings AS settings ON settings.workspace_id = workspace.workspace_id
      WHERE workspace.workspace_id = agent_invocations.workspace_id
        AND workspace.deleted_at IS NULL AND COALESCE(settings.chat_enabled, TRUE)
  ))
RETURNING *;

-- name: SaveAgentInvocationProposal :one
UPDATE public.agent_invocations SET proposal = sqlc.arg(proposal), summary = sqlc.arg(summary),
    status = CAST(sqlc.arg(next_status) AS text), version = version + 1, updated_at = clock_timestamp(),
    lease_token = CASE WHEN CAST(sqlc.arg(next_status) AS text) = 'running' THEN lease_token ELSE NULL END,
    lease_expires_at = CASE WHEN CAST(sqlc.arg(next_status) AS text) = 'running' THEN lease_expires_at ELSE NULL END
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp() AND proposal IS NULL
RETURNING *;

-- name: LockAgentInvocationLease :one
SELECT * FROM public.agent_invocations
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
FOR UPDATE;

-- name: SaveAgentInvocationOutcome :one
UPDATE public.agent_invocations SET outcome = sqlc.arg(outcome), version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp() AND outcome IS NULL
RETURNING *;

-- name: CompleteAgentInvocation :one
UPDATE public.agent_invocations SET status = sqlc.arg(next_status), reply_id = sqlc.arg(reply_id),
    summary = COALESCE(proposal->>'summary', ''),
    lease_token = NULL, lease_expires_at = NULL, version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
RETURNING *;

-- name: FailAgentInvocation :execrows
UPDATE public.agent_invocations SET status = CASE
        WHEN CAST(sqlc.arg(retry) AS boolean) AND (retry_attempt_count < CAST(sqlc.arg(maximum_attempts) AS integer)
            OR outcome IS NOT NULL OR proposal_reply_id IS NOT NULL OR action_started_at IS NOT NULL
            OR proposal->>'ignore' = 'true') THEN 'queued'
        ELSE 'failed' END,
    summary = sqlc.arg(summary),
    failure_pending = FALSE,
    available_at = clock_timestamp() + CAST(sqlc.arg(retry_seconds) AS double precision) * INTERVAL '1 second',
    lease_token = NULL, lease_expires_at = NULL, version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp();

-- name: ControlAgentInvocation :one
UPDATE public.agent_invocations SET
    status = CASE WHEN CAST(sqlc.arg(operation) AS text) = 'retry' AND approved_at IS NULL AND proposal_delivered_at IS NOT NULL
        AND proposal->>'requiresApproval' = 'true' THEN 'awaiting_approval' ELSE sqlc.arg(next_status) END,
    version = version + 1, failure_pending = FALSE,
    lease_token = NULL, lease_expires_at = NULL, lease_generation = lease_generation + 1,
    approved_at = CASE WHEN CAST(sqlc.arg(operation) AS text) = 'approve' THEN clock_timestamp() ELSE approved_at END,
    available_at = clock_timestamp(),
    retry_attempt_count = CASE WHEN CAST(sqlc.arg(operation) AS text) = 'retry' THEN 0 ELSE retry_attempt_count END,
    updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND workspace_id = sqlc.arg(workspace_id) AND requester_id = sqlc.arg(requester_id)
    AND version = sqlc.arg(expected_version)
    AND ((CAST(sqlc.arg(operation) AS text) = 'cancel' AND status IN ('queued','running','awaiting_approval','needs_input') AND outcome IS NULL AND action_started_at IS NULL
        AND COALESCE(proposal->>'ignore', 'false') <> 'true')
      OR (CAST(sqlc.arg(operation) AS text) = 'retry' AND status IN ('failed','cancelled'))
      OR (CAST(sqlc.arg(operation) AS text) = 'approve' AND status = 'awaiting_approval'))
RETURNING *;

-- name: BeginAgentInvocationFailure :one
UPDATE public.agent_invocations SET failure_pending = TRUE, summary = sqlc.arg(summary),
    version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
    AND outcome IS NULL AND reply_id IS NULL
RETURNING *;

-- name: SaveAgentInvocationFailurePublication :one
UPDATE public.agent_invocations SET failure_reply_id = sqlc.arg(failure_reply_id), version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
    AND failure_pending = TRUE AND failure_reply_id IS NULL
RETURNING *;

-- name: DeferAgentInvocationFailure :execrows
UPDATE public.agent_invocations SET status = 'queued', available_at = clock_timestamp() + INTERVAL '30 seconds',
    lease_token = NULL, lease_expires_at = NULL, version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp() AND failure_pending = TRUE;

-- name: CompleteAgentInvocationFailure :one
UPDATE public.agent_invocations SET status = 'failed', failure_pending = FALSE,
    lease_token = NULL, lease_expires_at = NULL, version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
    AND failure_pending = TRUE AND failure_reply_id IS NOT NULL
RETURNING *;

-- name: SaveAgentInvocationPublication :one
UPDATE public.agent_invocations SET reply_id = sqlc.arg(reply_id), version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp() AND reply_id IS NULL
RETURNING *;

-- name: StartAgentInvocationAction :one
UPDATE public.agent_invocations SET action_started_at = clock_timestamp(), version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
    AND action_started_at IS NULL AND outcome IS NULL
RETURNING *;

-- name: DeferAgentInvocationAction :execrows
UPDATE public.agent_invocations SET status = 'queued', available_at = clock_timestamp() + INTERVAL '5 seconds',
    retry_attempt_count = 0, lease_token = NULL, lease_expires_at = NULL, version = version + 1, updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND status = 'running' AND lease_token = sqlc.arg(lease_token)
    AND lease_generation = sqlc.arg(lease_generation) AND lease_expires_at > clock_timestamp()
    AND action_started_at IS NOT NULL AND outcome IS NULL;
