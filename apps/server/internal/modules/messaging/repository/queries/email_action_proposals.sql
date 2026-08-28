-- name: SupersedePendingEmailProposals :exec
UPDATE messaging_email_action_proposals
SET status = CASE WHEN expires_at <= sqlc.arg(now) THEN 'expired' ELSE 'superseded' END,
    expired_at = CASE WHEN expires_at <= sqlc.arg(now) THEN sqlc.arg(now) ELSE NULL END,
    superseded_at = CASE WHEN expires_at > sqlc.arg(now) THEN sqlc.arg(now) ELSE NULL END,
    updated_at = NOW()
WHERE thread_id = sqlc.arg(thread_id)
  AND status = 'pending';

-- name: InsertEmailActionProposal :one
INSERT INTO messaging_email_action_proposals (
    thread_id,
    workspace_id,
    user_id,
    source_message_id,
    idempotency_key,
    action_kind,
    entity_type,
    entity_id,
    expected_entity_version,
    proposed_diff,
    status,
    expires_at
) VALUES (
    sqlc.arg(thread_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(source_message_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(action_kind),
    sqlc.arg(entity_type),
    sqlc.arg(entity_id),
    sqlc.arg(expected_entity_version),
    sqlc.arg(proposed_diff),
    'pending',
    sqlc.arg(expires_at)
)
RETURNING id,
          thread_id,
          workspace_id,
          user_id,
          source_message_id,
          idempotency_key,
          action_kind,
          entity_type,
          entity_id,
          expected_entity_version,
          proposed_diff,
          status,
          apply_attempt,
          COALESCE(result, CAST('{}' AS jsonb)) AS result,
          last_error,
          expires_at,
          confirmed_at,
          applying_at,
          applied_at,
          failed_at,
          cancelled_at,
          expired_at,
          superseded_at,
          created_at,
          updated_at;

-- name: GetEmailActionProposalByIdempotencyKey :one
SELECT id,
       thread_id,
       workspace_id,
       user_id,
       source_message_id,
       idempotency_key,
       action_kind,
       entity_type,
       entity_id,
       expected_entity_version,
       proposed_diff,
       status,
       apply_attempt,
       COALESCE(result, CAST('{}' AS jsonb)) AS result,
       last_error,
       expires_at,
       confirmed_at,
       applying_at,
       applied_at,
       failed_at,
       cancelled_at,
       expired_at,
       superseded_at,
       created_at,
       updated_at
FROM messaging_email_action_proposals
WHERE thread_id = sqlc.arg(thread_id)
  AND idempotency_key = sqlc.arg(idempotency_key);

-- name: ExpirePendingEmailProposals :exec
UPDATE messaging_email_action_proposals
SET status = 'expired',
    expired_at = sqlc.arg(now),
    updated_at = NOW()
WHERE thread_id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND status = 'pending'
  AND expires_at <= sqlc.arg(now);

-- name: ListPendingEmailActionProposals :many
SELECT id,
       thread_id,
       workspace_id,
       user_id,
       source_message_id,
       idempotency_key,
       action_kind,
       entity_type,
       entity_id,
       expected_entity_version,
       proposed_diff,
       status,
       apply_attempt,
       COALESCE(result, CAST('{}' AS jsonb)) AS result,
       last_error,
       expires_at,
       confirmed_at,
       applying_at,
       applied_at,
       failed_at,
       cancelled_at,
       expired_at,
       superseded_at,
       created_at,
       updated_at
FROM messaging_email_action_proposals
WHERE thread_id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND status = 'pending'
ORDER BY created_at DESC;

-- name: FindLatestEmailActionProposalForConfirm :one
SELECT id,
       thread_id,
       workspace_id,
       user_id,
       source_message_id,
       idempotency_key,
       action_kind,
       entity_type,
       entity_id,
       expected_entity_version,
       proposed_diff,
       status,
       apply_attempt,
       COALESCE(result, CAST('{}' AS jsonb)) AS result,
       last_error,
       expires_at,
       confirmed_at,
       applying_at,
       applied_at,
       failed_at,
       cancelled_at,
       expired_at,
       superseded_at,
       created_at,
       updated_at
FROM messaging_email_action_proposals
WHERE thread_id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND status IN ('pending', 'confirmed', 'applying', 'applied', 'failed')
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: FindLatestEmailActionProposalForCancel :one
SELECT id,
       thread_id,
       workspace_id,
       user_id,
       source_message_id,
       idempotency_key,
       action_kind,
       entity_type,
       entity_id,
       expected_entity_version,
       proposed_diff,
       status,
       apply_attempt,
       COALESCE(result, CAST('{}' AS jsonb)) AS result,
       last_error,
       expires_at,
       confirmed_at,
       applying_at,
       applied_at,
       failed_at,
       cancelled_at,
       expired_at,
       superseded_at,
       created_at,
       updated_at
FROM messaging_email_action_proposals
WHERE thread_id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND status IN ('pending', 'cancelled')
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: GetEmailActionProposal :one
SELECT id,
       thread_id,
       workspace_id,
       user_id,
       source_message_id,
       idempotency_key,
       action_kind,
       entity_type,
       entity_id,
       expected_entity_version,
       proposed_diff,
       status,
       apply_attempt,
       COALESCE(result, CAST('{}' AS jsonb)) AS result,
       last_error,
       expires_at,
       confirmed_at,
       applying_at,
       applied_at,
       failed_at,
       cancelled_at,
       expired_at,
       superseded_at,
       created_at,
       updated_at
FROM messaging_email_action_proposals
WHERE id = sqlc.arg(proposal_id)
  AND thread_id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id);

-- name: LockEmailActionProposal :one
SELECT id,
       thread_id,
       workspace_id,
       user_id,
       source_message_id,
       idempotency_key,
       action_kind,
       entity_type,
       entity_id,
       expected_entity_version,
       proposed_diff,
       status,
       apply_attempt,
       COALESCE(result, CAST('{}' AS jsonb)) AS result,
       last_error,
       expires_at,
       confirmed_at,
       applying_at,
       applied_at,
       failed_at,
       cancelled_at,
       expired_at,
       superseded_at,
       created_at,
       updated_at
FROM messaging_email_action_proposals
WHERE id = sqlc.arg(proposal_id)
  AND thread_id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: HasActiveEmailReplyToken :one
SELECT EXISTS (
    SELECT 1
    FROM messaging_email_reply_tokens AS token
    INNER JOIN messaging_email_threads AS thread
        ON thread.id = token.thread_id
        AND thread.workspace_id = token.workspace_id
        AND thread.user_id = token.user_id
    INNER JOIN workspaces AS workspace ON workspace.workspace_id = thread.workspace_id
    INNER JOIN users AS actor ON actor.user_id = thread.user_id
    INNER JOIN workspace_members AS member
        ON member.workspace_id = thread.workspace_id
        AND member.user_id = thread.user_id
    WHERE token.thread_id = sqlc.arg(thread_id)
      AND token.workspace_id = sqlc.arg(workspace_id)
      AND token.user_id = sqlc.arg(user_id)
      AND token.token_hash = sqlc.arg(token_hash)
      AND token.revoked_at IS NULL
      AND token.expires_at > sqlc.arg(now)
      AND workspace.deleted_at IS NULL
      AND actor.is_active = true
      AND actor.is_system = false
      AND LOWER(actor.email) = LOWER(thread.recipient_email)
      AND member.role IN ('admin', 'member', 'guest')
);

-- name: TransitionEmailActionProposalDecision :execrows
UPDATE messaging_email_action_proposals
SET status = sqlc.arg(status),
    confirmed_at = CASE WHEN sqlc.arg(status) = 'confirmed' THEN sqlc.arg(now) ELSE confirmed_at END,
    cancelled_at = CASE WHEN sqlc.arg(status) = 'cancelled' THEN sqlc.arg(now) ELSE cancelled_at END,
    expired_at = CASE WHEN sqlc.arg(status) = 'expired' THEN sqlc.arg(now) ELSE expired_at END,
    superseded_at = CASE WHEN sqlc.arg(status) = 'superseded' THEN sqlc.arg(now) ELSE superseded_at END,
    updated_at = NOW()
WHERE id = sqlc.arg(proposal_id)
  AND status = 'pending';

-- name: ClaimEmailActionProposalApply :one
UPDATE messaging_email_action_proposals
SET status = 'applying',
    apply_attempt = apply_attempt + 1,
    applying_at = sqlc.arg(now),
    failed_at = NULL,
    result = NULL,
    last_error = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(proposal_id)
RETURNING id,
          thread_id,
          workspace_id,
          user_id,
          source_message_id,
          idempotency_key,
          action_kind,
          entity_type,
          entity_id,
          expected_entity_version,
          proposed_diff,
          status,
          apply_attempt,
          COALESCE(result, CAST('{}' AS jsonb)) AS result,
          last_error,
          expires_at,
          confirmed_at,
          applying_at,
          applied_at,
          failed_at,
          cancelled_at,
          expired_at,
          superseded_at,
          created_at,
          updated_at;

-- name: MarkEmailActionProposalApplied :one
UPDATE messaging_email_action_proposals
SET status = 'applied',
    result = sqlc.arg(result),
    last_error = NULL,
    applied_at = sqlc.arg(now),
    updated_at = NOW()
WHERE id = sqlc.arg(proposal_id)
  AND apply_attempt = sqlc.arg(apply_attempt)
RETURNING id,
          thread_id,
          workspace_id,
          user_id,
          source_message_id,
          idempotency_key,
          action_kind,
          entity_type,
          entity_id,
          expected_entity_version,
          proposed_diff,
          status,
          apply_attempt,
          COALESCE(result, CAST('{}' AS jsonb)) AS result,
          last_error,
          expires_at,
          confirmed_at,
          applying_at,
          applied_at,
          failed_at,
          cancelled_at,
          expired_at,
          superseded_at,
          created_at,
          updated_at;

-- name: MarkEmailActionProposalFailed :one
UPDATE messaging_email_action_proposals
SET status = 'failed',
    result = sqlc.arg(result),
    last_error = NULLIF(CAST(sqlc.arg(last_error) AS text), ''),
    failed_at = sqlc.arg(now),
    updated_at = NOW()
WHERE id = sqlc.arg(proposal_id)
  AND apply_attempt = sqlc.arg(apply_attempt)
RETURNING id,
          thread_id,
          workspace_id,
          user_id,
          source_message_id,
          idempotency_key,
          action_kind,
          entity_type,
          entity_id,
          expected_entity_version,
          proposed_diff,
          status,
          apply_attempt,
          COALESCE(result, CAST('{}' AS jsonb)) AS result,
          last_error,
          expires_at,
          confirmed_at,
          applying_at,
          applied_at,
          failed_at,
          cancelled_at,
          expired_at,
          superseded_at,
          created_at,
          updated_at;
