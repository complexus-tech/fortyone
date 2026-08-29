-- name: InsertStoryMutationConfirmation :exec
INSERT INTO messaging_story_mutation_confirmations (
    confirmation_id,
    workspace_id,
    user_id,
    team_id,
    operation,
    token_hash,
    proposal,
    status,
    expires_at
) VALUES (
    sqlc.arg(confirmation_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(team_id),
    sqlc.arg(operation),
    sqlc.arg(token_hash),
    CAST(NULLIF(CAST(sqlc.arg(proposal) AS text), '') AS jsonb),
    'pending',
    sqlc.arg(expires_at)
);

-- name: LoadStoryMutationConfirmation :one
SELECT team_id,
       operation,
       proposal,
       status,
       COALESCE(result, CAST('null' AS jsonb)) AS result,
       last_error
FROM messaging_story_mutation_confirmations
WHERE confirmation_id = sqlc.arg(confirmation_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND token_hash = sqlc.arg(token_hash);

-- name: LockStoryMutationConfirmation :one
SELECT status,
       COALESCE(result, CAST('null' AS jsonb)) AS result,
       last_error,
       CAST(proposal IS NOT NULL AS boolean) AS has_proposal,
       expires_at
FROM messaging_story_mutation_confirmations
WHERE confirmation_id = sqlc.arg(confirmation_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND token_hash = sqlc.arg(token_hash)
FOR UPDATE;

-- name: TransitionStoryMutationConfirmation :execrows
UPDATE messaging_story_mutation_confirmations
SET status = sqlc.arg(status),
    applied_at = CASE
        WHEN sqlc.arg(status) = 'applied' THEN CAST(sqlc.arg(now) AS timestamptz)
        ELSE NULL
    END,
    cancelled_at = CASE
        WHEN sqlc.arg(status) = 'cancelled' THEN CAST(sqlc.arg(now) AS timestamptz)
        ELSE NULL
    END,
    expired_at = CASE
        WHEN sqlc.arg(status) = 'expired' THEN CAST(sqlc.arg(now) AS timestamptz)
        ELSE NULL
    END,
    proposal = CASE
        WHEN sqlc.arg(status) IN ('cancelled', 'expired') THEN NULL
        ELSE proposal
    END,
    updated_at = NOW()
WHERE confirmation_id = sqlc.arg(confirmation_id)
  AND status = sqlc.arg(current_status);

-- name: RecordStoryMutationApplyFailure :exec
UPDATE messaging_story_mutation_confirmations
SET result = COALESCE(
        CAST(NULLIF(CAST(sqlc.arg(result) AS text), '') AS jsonb),
        result
    ),
    last_error = sqlc.arg(last_error),
    updated_at = NOW()
WHERE confirmation_id = sqlc.arg(confirmation_id)
  AND status = 'applied';

-- name: CompleteStoryMutationApply :execrows
UPDATE messaging_story_mutation_confirmations
SET result = CAST(sqlc.arg(result) AS jsonb),
    proposal = NULL,
    last_error = NULL,
    updated_at = NOW()
WHERE confirmation_id = sqlc.arg(confirmation_id)
  AND status = 'applied';
