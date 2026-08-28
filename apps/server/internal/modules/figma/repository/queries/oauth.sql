-- name: SaveOAuthState :execrows
INSERT INTO public.figma_oauth_states (
    state_hash,
    workspace_id,
    user_id,
    workspace_slug,
    code_verifier,
    expires_at
)
SELECT
    sqlc.arg(state_hash),
    workspace.workspace_id,
    member.user_id,
    workspace.slug,
    sqlc.arg(code_verifier),
    sqlc.arg(expires_at)
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS member
    ON member.workspace_id = workspace.workspace_id
   AND member.user_id = sqlc.arg(user_id)
INNER JOIN public.users AS account
    ON account.user_id = member.user_id
   AND account.is_active = TRUE
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.slug = sqlc.arg(workspace_slug)
  AND workspace.deleted_at IS NULL;

-- name: ConsumeOAuthState :one
UPDATE public.figma_oauth_states
SET consumed_at = sqlc.arg(consumed_at)
WHERE state_hash = sqlc.arg(state_hash)
  AND consumed_at IS NULL
  AND expires_at > sqlc.arg(consumed_at)
RETURNING state_hash, workspace_id, user_id, workspace_slug, code_verifier, expires_at;
