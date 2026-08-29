-- name: EnsureSprintSettings :execrows
INSERT INTO public.team_sprint_settings (
    team_id,
    workspace_id
)
SELECT
    team.team_id,
    team.workspace_id
FROM public.teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
ON CONFLICT (team_id, workspace_id) DO NOTHING;

-- name: GetSprintSettings :one
SELECT
    settings.team_id,
    settings.workspace_id,
    settings.auto_create_sprints,
    settings.upcoming_sprints_count,
    settings.sprint_duration_weeks,
    settings.sprint_start_day,
    settings.working_days,
    settings.move_incomplete_stories_enabled,
    settings.last_auto_sprint_number,
    settings.next_auto_sprint_number,
    settings.auto_create_disabled_at,
    settings.auto_create_disabled_reason,
    settings.created_at,
    settings.updated_at
FROM public.team_sprint_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id);

-- name: LockSprintSettings :one
SELECT
    settings.team_id,
    settings.workspace_id,
    settings.auto_create_sprints,
    settings.upcoming_sprints_count,
    settings.sprint_duration_weeks,
    settings.sprint_start_day,
    settings.working_days,
    settings.move_incomplete_stories_enabled,
    settings.last_auto_sprint_number,
    settings.next_auto_sprint_number,
    settings.auto_create_disabled_at,
    settings.auto_create_disabled_reason,
    settings.created_at,
    settings.updated_at
FROM public.team_sprint_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE OF settings;

-- name: UpdateSprintSettings :one
UPDATE public.team_sprint_settings AS settings
SET
    auto_create_sprints = CASE
        WHEN CAST(sqlc.arg(set_auto_create_sprints) AS boolean)
            THEN CAST(sqlc.arg(auto_create_sprints) AS boolean)
        ELSE settings.auto_create_sprints
    END,
    upcoming_sprints_count = CASE
        WHEN CAST(sqlc.arg(set_upcoming_sprints_count) AS boolean)
            THEN CAST(sqlc.arg(upcoming_sprints_count) AS integer)
        ELSE settings.upcoming_sprints_count
    END,
    sprint_duration_weeks = CASE
        WHEN CAST(sqlc.arg(set_sprint_duration_weeks) AS boolean)
            THEN CAST(sqlc.arg(sprint_duration_weeks) AS integer)
        ELSE settings.sprint_duration_weeks
    END,
    sprint_start_day = CASE
        WHEN CAST(sqlc.arg(set_sprint_start_day) AS boolean)
            THEN CAST(sqlc.arg(sprint_start_day) AS text)
        ELSE settings.sprint_start_day
    END,
    working_days = CASE
        WHEN CAST(sqlc.arg(set_working_days) AS boolean)
            THEN CAST(sqlc.arg(working_days) AS smallint[])
        ELSE settings.working_days
    END,
    move_incomplete_stories_enabled = CASE
        WHEN CAST(sqlc.arg(set_move_incomplete_stories_enabled) AS boolean)
            THEN CAST(sqlc.arg(move_incomplete_stories_enabled) AS boolean)
        ELSE settings.move_incomplete_stories_enabled
    END,
    next_auto_sprint_number = CASE
        WHEN CAST(sqlc.arg(set_next_auto_sprint_number) AS boolean)
            THEN CAST(sqlc.arg(next_auto_sprint_number) AS integer)
        ELSE settings.next_auto_sprint_number
    END,
    auto_create_disabled_at = CASE
        WHEN CAST(sqlc.arg(set_auto_create_sprints) AS boolean)
         AND CAST(sqlc.arg(auto_create_sprints) AS boolean)
            THEN NULL
        ELSE settings.auto_create_disabled_at
    END,
    auto_create_disabled_reason = CASE
        WHEN CAST(sqlc.arg(set_auto_create_sprints) AS boolean)
         AND CAST(sqlc.arg(auto_create_sprints) AS boolean)
            THEN NULL
        ELSE settings.auto_create_disabled_reason
    END,
    updated_at = CURRENT_TIMESTAMP
FROM public.teams AS team
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id)
  AND team.team_id = settings.team_id
  AND team.workspace_id = settings.workspace_id
RETURNING
    settings.team_id,
    settings.workspace_id,
    settings.auto_create_sprints,
    settings.upcoming_sprints_count,
    settings.sprint_duration_weeks,
    settings.sprint_start_day,
    settings.working_days,
    settings.move_incomplete_stories_enabled,
    settings.last_auto_sprint_number,
    settings.next_auto_sprint_number,
    settings.auto_create_disabled_at,
    settings.auto_create_disabled_reason,
    settings.created_at,
    settings.updated_at;
