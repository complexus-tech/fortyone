-- name: ExportTaskData :one
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
    'sourceId', story.id, 'title', story.title, 'description', COALESCE(story.description, ''),
    'descriptionHTML', COALESCE(story.description_html, ''), 'estimateValue', story.estimate_unit,
    'estimatedDurationMinutes', story.estimated_duration_minutes, 'minimumFocusBlockMinutes', story.minimum_focus_block_minutes,
    'createdAt', story.created_at, 'updatedAt', story.updated_at, 'completedAt', story.completed_at, 'archivedAt', story.archived_at,
    'comments', COALESCE((SELECT jsonb_agg(jsonb_build_object('sourceId', comment.comment_id, 'content', comment.content,
        'authorName', COALESCE(author.full_name, 'Former member'), 'createdAt', comment.created_at, 'parentSourceId', comment.parent_id, 'format', 'html') ORDER BY comment.created_at, comment.comment_id)
        FROM public.story_comments AS comment LEFT JOIN public.users AS author ON author.user_id = comment.commenter_id WHERE comment.story_id = story.id), '[]'),
    'customFieldValues', COALESCE((SELECT jsonb_agg(jsonb_build_object('sourceFieldId', value.field_id,
        'value', COALESCE(value.text_value, CAST(value.numeric_value AS text), CAST(value.date_value AS text), CAST(value.option_id AS text), CAST(value.person_id AS text))) ORDER BY value.field_id)
        FROM public.story_custom_field_values AS value WHERE value.story_id = story.id AND value.workspace_id = story.workspace_id AND value.team_id = story.team_id), '[]')
) ORDER BY story.created_at, story.id), '[]') AS jsonb)
FROM visible_stories AS story;
