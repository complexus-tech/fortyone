-- name: LockAdminWorkspaceMutationTarget :one
SELECT
    workspace.workspace_id,
    workspace.name,
    workspace.slug,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.updated_at
FROM workspaces AS workspace
WHERE workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
FOR UPDATE OF workspace;

-- name: UpdateAdminWorkspaceTrial :one
UPDATE workspaces
SET
    trial_ends_on = CAST(sqlc.arg(new_trial_ends_on) AS timestamptz),
    updated_at = CAST(sqlc.arg(changed_at) AS timestamptz)
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND updated_at = CAST(sqlc.arg(expected_updated_at) AS timestamptz)
RETURNING trial_ends_on, updated_at;

-- name: SetAdminWorkspaceDeleted :one
UPDATE workspaces
SET
    deleted_at = CASE
        WHEN CAST(sqlc.arg(delete_requested) AS boolean)
            THEN COALESCE(deleted_at, CAST(sqlc.arg(changed_at) AS timestamp))
        ELSE NULL
    END,
    deleted_by = CASE
        WHEN CAST(sqlc.arg(delete_requested) AS boolean)
            THEN CAST(sqlc.arg(actor_user_id) AS uuid)
        ELSE NULL
    END,
    updated_at = CAST(sqlc.arg(changed_at) AS timestamptz)
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND updated_at = CAST(sqlc.arg(expected_updated_at) AS timestamptz)
RETURNING deleted_at, updated_at;

-- name: UpdateAdminUserState :one
UPDATE users
SET
    is_active = CASE
        WHEN CAST(sqlc.arg(is_active_set) AS boolean)
            THEN CAST(sqlc.arg(new_is_active) AS boolean)
        ELSE is_active
    END,
    is_internal = CASE
        WHEN CAST(sqlc.arg(is_internal_set) AS boolean)
            THEN CAST(sqlc.arg(new_is_internal) AS boolean)
        ELSE is_internal
    END,
    login_reactivation_policy = CASE
        WHEN CAST(sqlc.arg(is_active_set) AS boolean)
         AND CAST(sqlc.arg(new_is_active) AS boolean) = TRUE
            THEN 'verified_sign_in'
        WHEN CAST(sqlc.arg(is_active_set) AS boolean)
            THEN 'admin_only'
        ELSE login_reactivation_policy
    END,
    inactivity_warning_sent_at = CASE
        WHEN CAST(sqlc.arg(is_active_set) AS boolean)
         AND is_active = FALSE
         AND CAST(sqlc.arg(new_is_active) AS boolean) = TRUE
            THEN NULL
        ELSE inactivity_warning_sent_at
    END,
    auth_session_version = CASE
        WHEN CAST(sqlc.arg(is_active_set) AS boolean)
         AND is_active <> CAST(sqlc.arg(new_is_active) AS boolean)
            THEN auth_session_version + 1
        ELSE auth_session_version
    END,
    updated_at = CAST(sqlc.arg(changed_at) AS timestamptz)
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
  AND updated_at = CAST(sqlc.arg(expected_updated_at) AS timestamptz)
RETURNING is_active, is_internal, updated_at;

-- name: RevokeAdminUserBrowserSessions :one
UPDATE users
SET
    auth_session_version = auth_session_version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
RETURNING auth_session_version;

-- name: CreateAdminNote :one
INSERT INTO admin_notes (
    target_type,
    target_id,
    workspace_id,
    body,
    created_by_user_id
) VALUES (
    CAST(sqlc.arg(target_type) AS text),
    CAST(sqlc.arg(target_id) AS uuid),
    CAST(sqlc.narg(workspace_id) AS uuid),
    CAST(sqlc.arg(body) AS text),
    CAST(sqlc.arg(created_by_user_id) AS uuid)
)
RETURNING id, target_type, target_id, workspace_id, body, created_by_user_id, created_at;
