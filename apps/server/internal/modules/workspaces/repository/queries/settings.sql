-- name: GetWorkspaceSettings :one
SELECT
    settings.workspace_id,
    settings.story_term,
    settings.sprint_term,
    settings.objective_term,
    settings.key_result_term,
    settings.objective_enabled,
    settings.key_result_enabled,
    settings.working_days,
    settings.working_start_minute,
    settings.working_end_minute,
    settings.created_at,
    settings.updated_at
FROM public.workspace_settings AS settings
WHERE settings.workspace_id = sqlc.arg(workspace_id);

-- name: UpdateWorkspaceSettings :one
UPDATE public.workspace_settings
SET
    story_term = CASE
        WHEN CAST(sqlc.arg(story_term) AS text) = '' THEN story_term
        ELSE CAST(sqlc.arg(story_term) AS text)
    END,
    sprint_term = CASE
        WHEN CAST(sqlc.arg(sprint_term) AS text) = '' THEN sprint_term
        ELSE CAST(sqlc.arg(sprint_term) AS text)
    END,
    objective_term = CASE
        WHEN CAST(sqlc.arg(objective_term) AS text) = '' THEN objective_term
        ELSE CAST(sqlc.arg(objective_term) AS text)
    END,
    key_result_term = CASE
        WHEN CAST(sqlc.arg(key_result_term) AS text) = '' THEN key_result_term
        ELSE CAST(sqlc.arg(key_result_term) AS text)
    END,
    objective_enabled = sqlc.arg(objective_enabled),
    key_result_enabled = sqlc.arg(key_result_enabled),
    working_days = CAST(sqlc.arg(working_days) AS smallint[]),
    working_start_minute = sqlc.arg(working_start_minute),
    working_end_minute = sqlc.arg(working_end_minute),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
RETURNING
    workspace_id,
    story_term,
    sprint_term,
    objective_term,
    key_result_term,
    objective_enabled,
    key_result_enabled,
    working_days,
    working_start_minute,
    working_end_minute,
    created_at,
    updated_at;

-- name: InitializeWorkspaceSettings :execrows
INSERT INTO public.workspace_settings (workspace_id)
VALUES (sqlc.arg(workspace_id))
ON CONFLICT (workspace_id) DO NOTHING;

-- name: GetOrCreateWorkspaceSettings :one
INSERT INTO public.workspace_settings (workspace_id)
VALUES (sqlc.arg(workspace_id))
ON CONFLICT (workspace_id) DO UPDATE
SET workspace_id = EXCLUDED.workspace_id
RETURNING
    workspace_id,
    story_term,
    sprint_term,
    objective_term,
    key_result_term,
    objective_enabled,
    key_result_enabled,
    working_days,
    working_start_minute,
    working_end_minute,
    created_at,
    updated_at;
