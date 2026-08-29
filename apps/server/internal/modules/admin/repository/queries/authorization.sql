-- name: LockActiveInternalAdmin :one
SELECT actor.user_id
FROM users AS actor
WHERE actor.user_id = CAST(sqlc.arg(actor_id) AS uuid)
  AND actor.is_active = TRUE
  AND actor.is_internal = TRUE
FOR SHARE OF actor;

-- name: LockAdminUserMutationParticipants :many
SELECT
    target.user_id,
    target.is_active,
    target.is_internal
FROM users AS target
WHERE target.user_id = CAST(sqlc.arg(actor_id) AS uuid)
   OR target.user_id = CAST(sqlc.arg(target_user_id) AS uuid)
ORDER BY target.user_id
FOR UPDATE OF target;

-- name: LockAdminNoteWorkspaceTarget :one
SELECT workspace.workspace_id
FROM workspaces AS workspace
WHERE workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
FOR KEY SHARE OF workspace;

-- name: LockAdminNoteUserTarget :one
SELECT target.user_id
FROM users AS target
WHERE target.user_id = CAST(sqlc.arg(user_id) AS uuid)
FOR KEY SHARE OF target;

-- name: LockAdminNoteUserWorkspaceTarget :one
SELECT target.user_id
FROM users AS target
JOIN workspace_members AS membership
  ON membership.user_id = target.user_id
 AND membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
JOIN workspaces AS workspace
  ON workspace.workspace_id = membership.workspace_id
WHERE target.user_id = CAST(sqlc.arg(user_id) AS uuid)
FOR KEY SHARE OF target, membership, workspace;

-- name: LockUserTarget :one
SELECT target.user_id
FROM users AS target
WHERE target.user_id = CAST(sqlc.arg(user_id) AS uuid)
FOR KEY SHARE OF target;
