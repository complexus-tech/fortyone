-- name: AuthorizeSprintCreate :one
SELECT team.team_id
FROM teams AS team
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = team.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS team_member
  ON team_member.team_id = team.team_id
 AND team_member.user_id = workspace_member.user_id
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND workspace_member.role IN ('member', 'admin')
FOR KEY SHARE OF team, workspace_member, actor, team_member;

-- name: CreateSprint :one
INSERT INTO sprints (
    name,
    goal,
    objective_id,
    team_id,
    workspace_id,
    start_date,
    end_date
) VALUES (
    sqlc.arg(name),
    sqlc.narg(goal),
    sqlc.narg(objective_id),
    sqlc.arg(team_id),
    sqlc.arg(workspace_id),
    sqlc.arg(start_date),
    sqlc.arg(end_date)
)
RETURNING
    sprint_id,
    name,
    goal,
    objective_id,
    team_id,
    workspace_id,
    start_date,
    end_date,
    created_at,
    updated_at,
    schedule_managed_by_automation;

-- name: LockSprintForMutation :one
SELECT
    sprint.sprint_id,
    sprint.name,
    sprint.goal,
    sprint.objective_id,
    sprint.team_id,
    sprint.workspace_id,
    sprint.start_date,
    sprint.end_date,
    sprint.created_at,
    sprint.updated_at,
    sprint.schedule_managed_by_automation
FROM sprints AS sprint
JOIN workspace_members AS workspace_member
  ON workspace_member.workspace_id = sprint.workspace_id
 AND workspace_member.user_id = sqlc.arg(actor_id)
JOIN users AS actor
  ON actor.user_id = workspace_member.user_id
 AND actor.is_active = TRUE
JOIN team_members AS team_member
  ON team_member.team_id = sprint.team_id
 AND team_member.user_id = workspace_member.user_id
WHERE sprint.sprint_id = sqlc.arg(sprint_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND workspace_member.role IN ('member', 'admin')
FOR UPDATE OF sprint
FOR KEY SHARE OF workspace_member, actor, team_member;

-- name: SprintObjectiveBelongsToTeam :one
SELECT EXISTS (
    SELECT 1
    FROM objectives AS objective
    WHERE objective.objective_id = sqlc.arg(objective_id)
      AND objective.workspace_id = sqlc.arg(workspace_id)
      AND objective.team_id = sqlc.arg(team_id)
);

-- name: UpdateSprint :one
UPDATE sprints
SET
    name = CASE WHEN sqlc.arg(name_set)::boolean THEN sqlc.arg(name)::text ELSE name END,
    goal = CASE WHEN sqlc.arg(goal_set)::boolean THEN sqlc.narg(goal)::text ELSE goal END,
    objective_id = CASE WHEN sqlc.arg(objective_set)::boolean THEN sqlc.narg(objective_id)::uuid ELSE objective_id END,
    start_date = CASE WHEN sqlc.arg(start_date_set)::boolean THEN sqlc.arg(start_date)::date ELSE start_date END,
    end_date = CASE WHEN sqlc.arg(end_date_set)::boolean THEN sqlc.arg(end_date)::date ELSE end_date END,
    schedule_managed_by_automation = CASE
        WHEN sqlc.arg(start_date_set)::boolean OR sqlc.arg(end_date_set)::boolean THEN FALSE
        ELSE schedule_managed_by_automation
    END,
    updated_at = NOW()
WHERE sprint_id = sqlc.arg(sprint_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND (
      NOT sqlc.arg(expected_updated_at_set)::boolean
      OR updated_at = sqlc.arg(expected_updated_at)::timestamptz
  )
RETURNING
    sprint_id,
    name,
    goal,
    objective_id,
    team_id,
    workspace_id,
    start_date,
    end_date,
    created_at,
    updated_at,
    schedule_managed_by_automation;

-- name: DeleteSprint :one
DELETE FROM sprints
WHERE sprint_id = sqlc.arg(sprint_id)
  AND workspace_id = sqlc.arg(workspace_id)
RETURNING team_id, name, start_date, end_date;

-- name: InsertSprintAuditEvent :exec
INSERT INTO audit_events (
    workspace_id,
    team_id,
    actor_type,
    actor_id,
    entity_type,
    entity_id,
    event_type,
    metadata
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    'user',
    sqlc.arg(actor_id),
    'sprint',
    sqlc.arg(sprint_id),
    sqlc.arg(event_type),
    sqlc.arg(metadata)::jsonb
);
