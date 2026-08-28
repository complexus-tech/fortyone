-- name: GetState :one
SELECT
    status.status_id,
    status.name,
    status.category,
    status.order_index,
    status.team_id,
    status.workspace_id,
    status.is_default,
    status.color,
    status.created_at,
    status.updated_at
FROM public.statuses AS status
WHERE status.status_id = sqlc.arg(status_id)
  AND status.workspace_id = sqlc.arg(workspace_id);

-- name: ListStatesForMember :many
SELECT
    status.status_id,
    status.name,
    status.category,
    status.order_index,
    status.team_id,
    status.workspace_id,
    status.is_default,
    status.color,
    status.created_at,
    status.updated_at
FROM public.statuses AS status
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = status.team_id
   AND team_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.workspace_id = status.workspace_id
   AND workspace_membership.user_id = team_membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = workspace_membership.user_id
   AND actor.is_active = TRUE
WHERE status.workspace_id = sqlc.arg(workspace_id)
ORDER BY status.order_index ASC, status.status_id ASC;

-- name: ListTeamStates :many
SELECT
    status.status_id,
    status.name,
    status.category,
    status.order_index,
    status.team_id,
    status.workspace_id,
    status.is_default,
    status.color,
    status.created_at,
    status.updated_at
FROM public.statuses AS status
WHERE status.workspace_id = sqlc.arg(workspace_id)
  AND status.team_id = sqlc.arg(team_id)
ORDER BY status.order_index ASC, status.status_id ASC;

-- name: ListTeamStatesForMember :many
SELECT
    status.status_id,
    status.name,
    status.category,
    status.order_index,
    status.team_id,
    status.workspace_id,
    status.is_default,
    status.color,
    status.created_at,
    status.updated_at
FROM public.statuses AS status
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = status.team_id
   AND team_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.workspace_id = status.workspace_id
   AND workspace_membership.user_id = team_membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = workspace_membership.user_id
   AND actor.is_active = TRUE
WHERE status.workspace_id = sqlc.arg(workspace_id)
  AND status.team_id = sqlc.arg(team_id)
ORDER BY status.order_index ASC, status.status_id ASC;
