-- name: IntegrationRequestStatusAvailable :one
SELECT EXISTS (
    SELECT 1
    FROM public.statuses
    WHERE status_id = sqlc.arg(status_id)
      AND workspace_id = sqlc.arg(workspace_id)
      AND team_id = sqlc.arg(team_id)
);

-- name: IntegrationRequestAssigneeAvailable :one
SELECT EXISTS (
    SELECT 1
    FROM public.team_members AS team_member
    INNER JOIN public.users AS actor_user
        ON actor_user.user_id = team_member.user_id
       AND actor_user.is_active = TRUE
    INNER JOIN public.workspace_members AS workspace_member
        ON workspace_member.user_id = team_member.user_id
       AND workspace_member.workspace_id = sqlc.arg(workspace_id)
    WHERE team_member.team_id = sqlc.arg(team_id)
      AND team_member.user_id = sqlc.arg(assignee_id)
);

-- name: IntegrationRequestObjectiveAvailable :one
SELECT EXISTS (
    SELECT 1
    FROM public.objectives
    WHERE objective_id = sqlc.arg(objective_id)
      AND workspace_id = sqlc.arg(workspace_id)
      AND team_id = sqlc.arg(team_id)
);

-- name: IntegrationRequestKeyResultAvailable :one
SELECT EXISTS (
    SELECT 1
    FROM public.key_results AS key_result
    INNER JOIN public.objectives AS objective
        ON objective.objective_id = key_result.objective_id
    WHERE key_result.id = sqlc.arg(key_result_id)
      AND key_result.objective_id = sqlc.arg(objective_id)
      AND objective.workspace_id = sqlc.arg(workspace_id)
      AND objective.team_id = sqlc.arg(team_id)
);

-- name: IntegrationRequestSprintAvailable :one
SELECT EXISTS (
    SELECT 1
    FROM public.sprints
    WHERE sprint_id = sqlc.arg(sprint_id)
      AND workspace_id = sqlc.arg(workspace_id)
      AND team_id = sqlc.arg(team_id)
);

-- name: CountAvailableIntegrationRequestLabels :one
SELECT COUNT(*)
FROM public.labels
WHERE workspace_id = sqlc.arg(workspace_id)
  AND (team_id = sqlc.arg(team_id) OR team_id IS NULL)
  AND label_id = ANY(CAST(sqlc.arg(label_ids) AS uuid[]));
