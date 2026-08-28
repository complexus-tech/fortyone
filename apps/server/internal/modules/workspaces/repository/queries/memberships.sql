-- name: AddWorkspaceMember :execrows
INSERT INTO public.workspace_members (
    workspace_id,
    user_id,
    role
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    CAST(sqlc.arg(role) AS user_role)
);

-- name: RemoveWorkspaceMember :execrows
DELETE FROM public.workspace_members
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id);

-- name: RemoveWorkspaceTeamMemberships :exec
DELETE FROM public.team_members AS team_membership
WHERE team_membership.user_id = sqlc.arg(user_id)
  AND EXISTS (
      SELECT 1
      FROM public.teams AS team
      WHERE team.team_id = team_membership.team_id
        AND team.workspace_id = sqlc.arg(workspace_id)
  );

-- name: UpdateWorkspaceMemberRole :execrows
UPDATE public.workspace_members
SET
    role = CAST(sqlc.arg(role) AS user_role),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id);

-- name: ResolveCurrentWorkspaceMembership :one
SELECT
    workspace.workspace_id,
    workspace.name,
    workspace.slug,
    CAST(membership.role AS text) AS user_role
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = workspace.workspace_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE workspace.slug = CAST(sqlc.arg(slug) AS text)
  AND membership.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE;

-- name: RecordWorkspaceAccess :exec
WITH current_membership AS (
    SELECT membership.workspace_id
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS account
        ON account.user_id = membership.user_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(user_id)
      AND account.is_active = TRUE
), touched_account AS (
    UPDATE public.users AS account
    SET
        last_login_at = CURRENT_TIMESTAMP,
        inactivity_warning_sent_at = NULL
    WHERE account.user_id = sqlc.arg(user_id)
      AND EXISTS (SELECT 1 FROM current_membership)
    RETURNING account.user_id
)
UPDATE public.workspaces AS workspace
SET
    last_accessed_at = CURRENT_TIMESTAMP,
    inactivity_warning_sent_at = NULL
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (SELECT 1 FROM touched_account);

-- name: ListWorkspaceAdminEmails :many
SELECT account.email
FROM public.users AS account
INNER JOIN public.workspace_members AS membership
    ON membership.user_id = account.user_id
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.role = 'admin'
  AND account.user_id <> sqlc.arg(actor_id)
  AND account.is_active = TRUE
ORDER BY account.user_id ASC;
