-- name: AggregateCustomField :many
WITH selected AS (
    SELECT story.id, story.status_id, story.assignee_id,
        COALESCE(status.name,'No status') AS status_name,
        COALESCE(assignee.full_name,'Unassigned') AS assignee_name,
        value.numeric_value,
        (value.text_value IS NOT NULL OR value.numeric_value IS NOT NULL OR value.date_value IS NOT NULL OR value.option_id IS NOT NULL OR value.person_id IS NOT NULL) AS has_value,
        CASE CAST(sqlc.arg(date_basis) AS text)
            WHEN 'created' THEN CAST(story.created_at AT TIME ZONE 'UTC' AS date)
            WHEN 'completed' THEN CAST(story.completed_at AT TIME ZONE 'UTC' AS date)
            ELSE date_value.date_value
        END AS report_date
    FROM stories AS story
    LEFT JOIN story_custom_field_values AS value ON value.story_id = story.id AND value.field_id = sqlc.arg(field_id) AND value.workspace_id = story.workspace_id AND value.team_id = story.team_id
    LEFT JOIN story_custom_field_values AS date_value ON date_value.story_id = story.id AND date_value.field_id = sqlc.narg(date_field_id) AND date_value.workspace_id = story.workspace_id AND date_value.team_id = story.team_id
    LEFT JOIN statuses AS status ON status.status_id = story.status_id AND status.workspace_id = story.workspace_id
    LEFT JOIN users AS assignee ON assignee.user_id = story.assignee_id
    WHERE story.workspace_id = sqlc.arg(workspace_id) AND story.team_id = sqlc.arg(team_id)
      AND story.deleted_at IS NULL AND story.archived_at IS NULL AND story.is_draft = FALSE
      AND (cardinality(CAST(sqlc.arg(status_ids) AS uuid[])) = 0 OR story.status_id = ANY(CAST(sqlc.arg(status_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
), scoped AS (
    SELECT id, status_id, assignee_id, status_name, assignee_name, numeric_value, has_value, report_date
    FROM selected
    WHERE (CAST(sqlc.narg(start_date) AS date) IS NULL OR report_date >= CAST(sqlc.narg(start_date) AS date))
      AND (CAST(sqlc.narg(end_date) AS date) IS NULL OR report_date <= CAST(sqlc.narg(end_date) AS date))
), grouped AS (
    SELECT CASE CAST(sqlc.arg(group_by) AS text)
        WHEN 'status' THEN COALESCE(CAST(status_id AS text),'none')
        WHEN 'assignee' THEN COALESCE(CAST(assignee_id AS text),'none')
        WHEN 'month' THEN COALESCE(to_char(report_date,'YYYY-MM'),'undated')
        ELSE 'all' END AS group_key,
        CASE CAST(sqlc.arg(group_by) AS text)
        WHEN 'status' THEN status_name
        WHEN 'assignee' THEN assignee_name
        WHEN 'month' THEN COALESCE(to_char(report_date,'YYYY-MM'),'Undated')
        ELSE 'All work' END AS group_label,
        numeric_value, has_value
    FROM scoped
)
SELECT group_key, CAST(MAX(group_label) AS text) AS label, COUNT(*) AS total_count, COUNT(*) FILTER (WHERE has_value) AS valued_count,
    CAST(COALESCE(CAST(CASE CAST(sqlc.arg(aggregation) AS text)
        WHEN 'sum' THEN COALESCE(SUM(numeric_value),0)
        WHEN 'average' THEN AVG(numeric_value)
        WHEN 'min' THEN MIN(numeric_value)
        WHEN 'max' THEN MAX(numeric_value)
        ELSE CAST(COUNT(*) FILTER (WHERE has_value) AS numeric) END AS text),'') AS text) AS result
FROM grouped GROUP BY group_key ORDER BY group_key LIMIT 501;
