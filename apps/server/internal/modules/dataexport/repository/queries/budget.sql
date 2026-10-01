-- name: GetExportBudget :one
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
SELECT
    CAST(EXISTS (SELECT 1 FROM public.workspace_members AS member INNER JOIN public.users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
        INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = member.workspace_id AND workspace.deleted_at IS NULL
        WHERE member.workspace_id = sqlc.arg(workspace_id) AND member.user_id = sqlc.arg(actor_id) AND member.role = 'admin') AS boolean) AS authorized,
    CAST(CAST(sqlc.narg(team_id) AS uuid) IS NULL OR EXISTS (SELECT 1 FROM visible_teams) AS boolean) AS team_visible,
    CAST(CURRENT_TIMESTAMP AS timestamptz) AS generated_at,
    CAST((SELECT COUNT(*) FROM visible_stories) AS bigint) AS task_count,
    CAST((SELECT COUNT(*) FROM visible_teams) AS integer) AS team_count,
    CAST((SELECT COUNT(*) FROM visible_people) AS integer) AS person_count,
    CAST((SELECT COUNT(*) FROM public.labels AS label WHERE label.workspace_id = sqlc.arg(workspace_id) AND (label.team_id IS NULL OR label.team_id IN (SELECT team_id FROM visible_teams))) AS integer) AS label_count,
    CAST((SELECT COUNT(*) FROM visible_objectives) AS integer) AS objective_count,
    CAST((SELECT COUNT(*) FROM public.key_results AS result WHERE result.objective_id IN (SELECT objective_id FROM visible_objectives)) AS integer) AS key_result_count,
    CAST((SELECT COUNT(*) FROM public.sprints AS sprint WHERE sprint.workspace_id = sqlc.arg(workspace_id) AND sprint.team_id IN (SELECT team_id FROM visible_teams)) AS integer) AS sprint_count,
    CAST((SELECT COUNT(*) FROM public.custom_fields AS field WHERE field.workspace_id = sqlc.arg(workspace_id) AND field.team_id IN (SELECT team_id FROM visible_teams)) AS integer) AS field_count,
    CAST(COALESCE((SELECT SUM(2 * OCTET_LENGTH(COALESCE(description, '')) + OCTET_LENGTH(COALESCE(description_html, '')) + 2048) FROM visible_stories), 0)
        + COALESCE((SELECT SUM(OCTET_LENGTH(comment.content) + 512) FROM public.story_comments AS comment WHERE comment.story_id IN (SELECT id FROM visible_stories)), 0)
        + COALESCE((SELECT SUM(OCTET_LENGTH(COALESCE(value.text_value, CAST(value.numeric_value AS text), '')) + 256) FROM public.story_custom_field_values AS value WHERE value.story_id IN (SELECT id FROM visible_stories)), 0)
        AS bigint) AS estimated_bytes;
