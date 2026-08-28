-- name: EnsureEstimationSettings :execrows
INSERT INTO public.team_estimation_settings (
    team_id,
    workspace_id
)
SELECT
    team.team_id,
    team.workspace_id
FROM public.teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
ON CONFLICT (team_id) DO NOTHING;

-- name: GetEstimationSettings :one
SELECT
    settings.team_id,
    settings.workspace_id,
    settings.scheme,
    settings.created_at,
    settings.updated_at
FROM public.team_estimation_settings AS settings
INNER JOIN public.teams AS team
    ON team.team_id = settings.team_id
   AND team.workspace_id = settings.workspace_id
WHERE settings.team_id = sqlc.arg(team_id)
  AND settings.workspace_id = sqlc.arg(workspace_id);

-- name: UpdateEstimationSettings :one
UPDATE public.team_estimation_settings AS settings
SET
    scheme = CASE
        WHEN CAST(sqlc.arg(set_scheme) AS boolean)
            THEN CAST(sqlc.arg(scheme) AS text)
        ELSE settings.scheme
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
    settings.scheme,
    settings.created_at,
    settings.updated_at;
