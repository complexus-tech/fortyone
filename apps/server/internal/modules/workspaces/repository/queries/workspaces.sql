-- name: ListWorkspacesForUser :many
SELECT
    workspace.workspace_id,
    workspace.slug,
    workspace.name,
    workspace.color,
    workspace.team_size,
    workspace.avatar_url,
    CASE
        WHEN account.last_used_workspace_id = workspace.workspace_id THEN TRUE
        WHEN account.last_used_workspace_id IS NULL
            AND workspace.workspace_id = (
                SELECT first_workspace.workspace_id
                FROM public.workspaces AS first_workspace
                INNER JOIN public.workspace_members AS first_membership
                    ON first_membership.workspace_id = first_workspace.workspace_id
                WHERE first_membership.user_id = sqlc.arg(user_id)
                ORDER BY first_workspace.created_at ASC, first_workspace.workspace_id ASC
                LIMIT 1
            ) THEN TRUE
        ELSE FALSE
    END AS is_active,
    CAST(membership.role AS text) AS user_role,
    workspace.created_by,
    workspace.created_at,
    workspace.updated_at,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.deleted_by
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = workspace.workspace_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE membership.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE
ORDER BY workspace.created_at ASC, workspace.workspace_id ASC;

-- name: GetWorkspaceForMember :one
SELECT
    workspace.workspace_id,
    workspace.slug,
    workspace.name,
    workspace.color,
    workspace.team_size,
    workspace.avatar_url,
    CASE
        WHEN account.last_used_workspace_id = workspace.workspace_id THEN TRUE
        WHEN account.last_used_workspace_id IS NULL
            AND workspace.workspace_id = (
                SELECT first_workspace.workspace_id
                FROM public.workspaces AS first_workspace
                INNER JOIN public.workspace_members AS first_membership
                    ON first_membership.workspace_id = first_workspace.workspace_id
                WHERE first_membership.user_id = sqlc.arg(user_id)
                ORDER BY first_workspace.created_at ASC, first_workspace.workspace_id ASC
                LIMIT 1
            ) THEN TRUE
        ELSE FALSE
    END AS is_active,
    CAST(membership.role AS text) AS user_role,
    workspace.created_by,
    workspace.created_at,
    workspace.updated_at,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.deleted_by
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = workspace.workspace_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE;

-- name: GetWorkspaceForMemberBySlug :one
SELECT
    workspace.workspace_id,
    workspace.slug,
    workspace.name,
    workspace.color,
    workspace.team_size,
    workspace.avatar_url,
    CASE
        WHEN account.last_used_workspace_id = workspace.workspace_id THEN TRUE
        WHEN account.last_used_workspace_id IS NULL
            AND workspace.workspace_id = (
                SELECT first_workspace.workspace_id
                FROM public.workspaces AS first_workspace
                INNER JOIN public.workspace_members AS first_membership
                    ON first_membership.workspace_id = first_workspace.workspace_id
                WHERE first_membership.user_id = sqlc.arg(user_id)
                ORDER BY first_workspace.created_at ASC, first_workspace.workspace_id ASC
                LIMIT 1
            ) THEN TRUE
        ELSE FALSE
    END AS is_active,
    CAST(membership.role AS text) AS user_role,
    workspace.created_by,
    workspace.created_at,
    workspace.updated_at,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.deleted_by
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = workspace.workspace_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE workspace.slug = CAST(sqlc.arg(slug) AS text)
  AND membership.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE;

-- name: GetWorkspaceByID :one
SELECT
    workspace.workspace_id,
    workspace.slug,
    workspace.name,
    workspace.color,
    workspace.team_size,
    workspace.avatar_url,
    workspace.created_by,
    workspace.created_at,
    workspace.updated_at,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.deleted_by
FROM public.workspaces AS workspace
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL;

-- name: GetPublicWorkspaceBySlug :one
SELECT
    workspace.workspace_id,
    workspace.slug,
    workspace.name,
    workspace.color,
    workspace.team_size,
    workspace.avatar_url,
    workspace.created_by,
    workspace.created_at,
    workspace.updated_at,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.deleted_by
FROM public.workspaces AS workspace
WHERE workspace.slug = CAST(sqlc.arg(slug) AS text)
  AND workspace.deleted_at IS NULL;

-- name: WorkspaceSlugExists :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspaces AS workspace
    WHERE workspace.slug = CAST(sqlc.arg(slug) AS text)
) AS exists;

-- name: CreateWorkspace :one
INSERT INTO public.workspaces (
    name,
    slug,
    color,
    team_size,
    trial_ends_on,
    created_by
)
VALUES (
    CAST(sqlc.arg(name) AS text),
    CAST(sqlc.arg(slug) AS text),
    CAST(sqlc.arg(color) AS text),
    CAST(sqlc.arg(team_size) AS text),
    CURRENT_TIMESTAMP + INTERVAL '14 days',
    sqlc.arg(created_by)
)
RETURNING
    workspace_id,
    slug,
    name,
    color,
    team_size,
    avatar_url,
    created_by,
    created_at,
    updated_at,
    trial_ends_on,
    deleted_at,
    deleted_by;

-- name: UpdateWorkspace :one
UPDATE public.workspaces
SET
    name = CASE
        WHEN CAST(sqlc.arg(update_name) AS boolean) THEN CAST(sqlc.arg(name) AS text)
        ELSE name
    END,
    avatar_url = CASE
        WHEN CAST(sqlc.arg(update_avatar_url) AS boolean) THEN CAST(sqlc.arg(avatar_url) AS text)
        ELSE avatar_url
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND (CAST(sqlc.arg(update_name) AS boolean) OR CAST(sqlc.arg(update_avatar_url) AS boolean))
RETURNING
    workspace_id,
    slug,
    name,
    color,
    team_size,
    avatar_url,
    created_by,
    created_at,
    updated_at,
    trial_ends_on,
    deleted_at,
    deleted_by;

-- name: SoftDeleteWorkspace :execrows
UPDATE public.workspaces
SET
    deleted_at = CURRENT_TIMESTAMP,
    deleted_by = sqlc.arg(deleted_by),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND deleted_at IS NULL;

-- name: RestoreWorkspace :execrows
UPDATE public.workspaces
SET
    deleted_at = NULL,
    deleted_by = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND deleted_at IS NOT NULL;

-- name: CreateDefaultObjectiveStatus :execrows
INSERT INTO public.objective_statuses (
    name,
    category,
    order_index,
    color,
    workspace_id
)
VALUES (
    CAST(sqlc.arg(name) AS text),
    CAST(sqlc.arg(category) AS text),
    sqlc.arg(order_index),
    CAST(sqlc.arg(color) AS text),
    sqlc.arg(workspace_id)
);
