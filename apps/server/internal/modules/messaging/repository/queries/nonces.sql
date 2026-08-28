-- name: CreateNonce :exec
INSERT INTO messaging_nonces (
    provider,
    purpose,
    nonce_hash,
    workspace_id,
    user_id,
    external_workspace_id,
    external_user_id,
    payload,
    expires_at
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(purpose),
    sqlc.arg(nonce_hash),
    sqlc.arg(workspace_id),
    sqlc.narg(user_id),
    NULLIF(CAST(sqlc.arg(external_workspace_id) AS text), ''),
    NULLIF(CAST(sqlc.arg(external_user_id) AS text), ''),
    sqlc.arg(payload),
    sqlc.arg(expires_at)
);

-- name: ConsumeNonce :one
UPDATE messaging_nonces
SET consumed_at = sqlc.arg(consumed_at),
    user_id = COALESCE(user_id, sqlc.narg(user_id))
WHERE provider = sqlc.arg(provider)
  AND purpose = sqlc.arg(purpose)
  AND nonce_hash = sqlc.arg(nonce_hash)
  AND consumed_at IS NULL
  AND expires_at > sqlc.arg(consumed_at)
  AND (CAST(sqlc.narg(workspace_id) AS uuid) IS NULL OR workspace_id = sqlc.narg(workspace_id))
  AND (CAST(sqlc.narg(user_id) AS uuid) IS NULL OR user_id IS NULL OR user_id = sqlc.narg(user_id))
RETURNING id,
          provider,
          purpose,
          workspace_id,
          user_id,
          external_workspace_id,
          external_user_id,
          payload,
          expires_at,
          consumed_at;
