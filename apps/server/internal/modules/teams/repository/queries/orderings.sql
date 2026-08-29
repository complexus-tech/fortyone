-- name: ActorCanOrderWorkspaceTeams :one
SELECT EXISTS (
    SELECT 1
    FROM workspace_members AS actor_membership
    INNER JOIN users AS actor ON actor.user_id = actor_membership.user_id
    WHERE actor_membership.workspace_id = sqlc.arg(workspace_id)
      AND actor_membership.user_id = sqlc.arg(actor_id)
      AND actor.is_active = TRUE
) AS allowed;

-- name: DeleteActorTeamOrdering :exec
DELETE FROM user_team_orders
WHERE user_id = sqlc.arg(actor_id)
  AND workspace_id = sqlc.arg(workspace_id);

-- name: InsertActorTeamOrder :execrows
INSERT INTO user_team_orders (
    user_id,
    team_id,
    workspace_id,
    order_index
)
SELECT
    sqlc.arg(actor_id),
    team.team_id,
    team.workspace_id,
    CAST(sqlc.arg(order_index) AS integer)
FROM teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND (
      EXISTS (
          SELECT 1
          FROM team_members AS actor_team_membership
          WHERE actor_team_membership.team_id = team.team_id
            AND actor_team_membership.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM workspace_members AS admin_membership
          WHERE admin_membership.workspace_id = team.workspace_id
            AND admin_membership.user_id = sqlc.arg(actor_id)
            AND admin_membership.role = 'admin'
      )
  );
