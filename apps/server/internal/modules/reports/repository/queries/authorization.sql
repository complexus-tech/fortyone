-- name: GetReportsActorAccess :one
SELECT CAST(membership.role AS text) AS role
FROM workspace_members AS membership
INNER JOIN users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
INNER JOIN workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = sqlc.arg(workspace_id)::uuid
  AND membership.user_id = sqlc.arg(actor_id)::uuid
  AND membership.role IN ('admin', 'member');

-- name: ListReportsVisibleTeamIDs :many
SELECT team.team_id
FROM teams AS team
INNER JOIN workspace_members AS membership
    ON membership.workspace_id = team.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)::uuid
   AND membership.role IN ('admin', 'member')
INNER JOIN users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
   AND actor.is_system = FALSE
INNER JOIN workspaces AS workspace
    ON workspace.workspace_id = team.workspace_id
   AND workspace.deleted_at IS NULL
WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (
      membership.role = 'admin'
      OR team.is_private = FALSE
      OR EXISTS (
          SELECT 1
          FROM team_members AS actor_team_membership
          WHERE actor_team_membership.team_id = team.team_id
            AND actor_team_membership.user_id = sqlc.arg(actor_id)::uuid
      )
  )
  AND (
      cardinality(sqlc.arg(requested_team_ids)::uuid[]) = 0
      OR team.team_id = ANY(sqlc.arg(requested_team_ids)::uuid[])
  )
ORDER BY team.team_id;
