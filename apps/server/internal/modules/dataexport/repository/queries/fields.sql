-- name: ExportFieldDefinitions :one
WITH visible_teams AS (
    SELECT team.* FROM public.teams AS team
    INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = team.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.workspace_members AS member ON member.workspace_id = team.workspace_id AND member.user_id = sqlc.arg(actor_id) AND member.role = 'admin'
    INNER JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
    INNER JOIN public.team_members AS team_member ON team_member.team_id = team.team_id AND team_member.user_id = actor.user_id
    WHERE team.workspace_id = sqlc.arg(workspace_id)
      AND (CAST(sqlc.narg(team_id) AS uuid) IS NULL OR team.team_id = sqlc.narg(team_id))
), visible_stories AS (
    SELECT story.* FROM public.stories AS story
    INNER JOIN visible_teams AS team ON team.team_id = story.team_id
    WHERE story.workspace_id = sqlc.arg(workspace_id) AND story.deleted_at IS NULL AND story.is_draft = FALSE
), visible_objectives AS (
    SELECT objective.* FROM public.objectives AS objective
    INNER JOIN visible_teams AS team ON team.team_id = objective.team_id
    WHERE objective.workspace_id = sqlc.arg(workspace_id)
), visible_people AS (
    SELECT DISTINCT account.user_id, account.full_name, account.email
    FROM public.users AS account
    INNER JOIN public.workspace_members AS member ON member.user_id = account.user_id AND member.workspace_id = sqlc.arg(workspace_id)
    INNER JOIN public.team_members AS team_member ON team_member.user_id = account.user_id
    INNER JOIN visible_teams AS team ON team.team_id = team_member.team_id
)
SELECT CAST(COALESCE(jsonb_agg(jsonb_build_object(
    'sourceId', field.id, 'name', field.name, 'type', field.field_type, 'currency', field.currency, 'icon', field.icon,
    'teamSourceId', field.team_id, 'archivedAt', field.archived_at,
    'options', COALESCE((SELECT jsonb_agg(jsonb_build_object('sourceId', option.id, 'name', option.name, 'archivedAt', option.archived_at) ORDER BY option.position, option.id)
        FROM public.custom_field_options AS option WHERE option.field_id = field.id), '[]')
) ORDER BY field.team_id, field.created_at, field.id), '[]') AS jsonb)
FROM public.custom_fields AS field
WHERE field.workspace_id = sqlc.arg(workspace_id) AND field.team_id IN (SELECT team_id FROM visible_teams);
