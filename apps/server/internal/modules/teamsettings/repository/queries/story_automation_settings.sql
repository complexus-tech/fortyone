-- name: EnsureStoryAutomationSettings :execrows
INSERT INTO public.team_story_automation_settings (
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

-- name: GetStoryAutomationSettings :one
SELECT
    settings.team_id,
    settings.workspace_id,
    settings.auto_close_inactive_enabled,
    settings.auto_close_inactive_months,
    settings.auto_archive_enabled,
    settings.auto_archive_months,
    settings.created_at,
    settings.updated_at
FROM public.team_story_automation_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id);

-- name: UpdateStoryAutomationSettings :one
UPDATE public.team_story_automation_settings AS settings
SET
    auto_close_inactive_enabled = CASE
        WHEN CAST(sqlc.arg(set_auto_close_inactive_enabled) AS boolean)
            THEN CAST(sqlc.arg(auto_close_inactive_enabled) AS boolean)
        ELSE settings.auto_close_inactive_enabled
    END,
    auto_close_inactive_months = CASE
        WHEN CAST(sqlc.arg(set_auto_close_inactive_months) AS boolean)
            THEN CAST(sqlc.arg(auto_close_inactive_months) AS integer)
        ELSE settings.auto_close_inactive_months
    END,
    auto_archive_enabled = CASE
        WHEN CAST(sqlc.arg(set_auto_archive_enabled) AS boolean)
            THEN CAST(sqlc.arg(auto_archive_enabled) AS boolean)
        ELSE settings.auto_archive_enabled
    END,
    auto_archive_months = CASE
        WHEN CAST(sqlc.arg(set_auto_archive_months) AS boolean)
            THEN CAST(sqlc.arg(auto_archive_months) AS integer)
        ELSE settings.auto_archive_months
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
    settings.auto_close_inactive_enabled,
    settings.auto_close_inactive_months,
    settings.auto_archive_enabled,
    settings.auto_archive_months,
    settings.created_at,
    settings.updated_at;
