-- name: IsActiveTeamMember :one
SELECT EXISTS (
    SELECT 1
    FROM public.teams AS team
    INNER JOIN public.team_members AS team_member
        ON team_member.team_id = team.team_id
    INNER JOIN public.workspace_members AS workspace_member
        ON workspace_member.workspace_id = team.workspace_id
       AND workspace_member.user_id = team_member.user_id
    INNER JOIN public.users AS actor
        ON actor.user_id = team_member.user_id
       AND actor.is_active = TRUE
    WHERE team.team_id = sqlc.arg(team_id)
      AND team.workspace_id = sqlc.arg(workspace_id)
      AND team_member.user_id = sqlc.arg(actor_id)
) AS is_member;
