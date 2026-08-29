-- name: GetDatabaseDate :one
SELECT CAST(CURRENT_DATE AS date) AS current_date;

-- name: ListManagedFutureSprintsForUpdate :many
SELECT
    sprint.sprint_id,
    sprint.name,
    sprint.start_date,
    sprint.end_date
FROM public.sprints AS sprint
INNER JOIN public.teams AS team
    ON team.team_id = sprint.team_id
   AND team.workspace_id = sprint.workspace_id
WHERE sprint.team_id = sqlc.arg(team_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND sprint.schedule_managed_by_automation = TRUE
  AND sprint.start_date > sqlc.arg(schedule_date)
ORDER BY sprint.start_date, sprint.sprint_id
LIMIT CAST(sqlc.arg(row_limit) AS integer)
FOR UPDATE OF sprint;

-- name: GetActiveSprintEnd :one
SELECT sprint.end_date
FROM public.sprints AS sprint
INNER JOIN public.teams AS team
    ON team.team_id = sprint.team_id
   AND team.workspace_id = sprint.workspace_id
WHERE sprint.team_id = sqlc.arg(team_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND sprint.start_date <= sqlc.arg(schedule_date)
  AND sprint.end_date >= sqlc.arg(schedule_date)
ORDER BY sprint.end_date DESC, sprint.sprint_id
LIMIT 1
FOR SHARE OF sprint;

-- name: FindCustomSprintScheduleConflict :one
SELECT
    sprint.sprint_id,
    sprint.name,
    sprint.start_date,
    sprint.end_date
FROM public.sprints AS sprint
INNER JOIN public.teams AS team
    ON team.team_id = sprint.team_id
   AND team.workspace_id = sprint.workspace_id
WHERE sprint.team_id = sqlc.arg(team_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND sprint.schedule_managed_by_automation = FALSE
  AND sprint.start_date > sqlc.arg(schedule_date)
  AND sprint.start_date <= sqlc.arg(proposed_end_date)
  AND sprint.end_date >= sqlc.arg(proposed_start_date)
ORDER BY sprint.start_date, sprint.sprint_id
LIMIT 1
FOR SHARE OF sprint;

-- name: UpdateManagedSprintSchedule :execrows
UPDATE public.sprints AS sprint
SET
    start_date = sqlc.arg(start_date),
    end_date = sqlc.arg(end_date),
    updated_at = sqlc.arg(updated_at)
FROM public.teams AS team
WHERE sprint.sprint_id = sqlc.arg(sprint_id)
  AND sprint.team_id = sqlc.arg(team_id)
  AND sprint.workspace_id = sqlc.arg(workspace_id)
  AND sprint.schedule_managed_by_automation = TRUE
  AND team.team_id = sprint.team_id
  AND team.workspace_id = sprint.workspace_id;
