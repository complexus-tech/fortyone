-- name: SystemCanReadTeamWorkload :one
SELECT EXISTS (
    SELECT 1 FROM public.users AS actor
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = sqlc.arg(workspace_id) AND workspace.deleted_at IS NULL
    INNER JOIN public.teams AS team
        ON team.workspace_id = workspace.workspace_id AND team.team_id = sqlc.arg(team_id)
    WHERE actor.user_id = sqlc.arg(actor_id)
      AND actor.is_active = TRUE AND actor.is_system = TRUE
) AS allowed;
