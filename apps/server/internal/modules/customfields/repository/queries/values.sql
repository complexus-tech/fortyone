-- name: ListStoryValues :many
SELECT value.field_id, COALESCE(value.text_value, CAST(value.numeric_value AS text), CAST(value.date_value AS text), CAST(value.option_id AS text), CAST(value.person_id AS text)) AS value
FROM story_custom_field_values AS value
WHERE value.story_id = sqlc.arg(story_id) AND value.workspace_id = sqlc.arg(workspace_id) AND value.team_id = sqlc.arg(team_id)
ORDER BY value.field_id;

-- name: GetStoryFieldValue :one
SELECT COALESCE(text_value, CAST(numeric_value AS text), CAST(date_value AS text), CAST(option_id AS text), CAST(person_id AS text)) AS value
FROM story_custom_field_values
WHERE story_id = sqlc.arg(story_id) AND field_id = sqlc.arg(field_id) AND workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id);

-- name: UpsertStoryValue :exec
INSERT INTO story_custom_field_values (story_id, field_id, workspace_id, team_id, field_type, text_value, numeric_value, date_value, option_id, person_id)
VALUES (sqlc.arg(story_id), sqlc.arg(field_id), sqlc.arg(workspace_id), sqlc.arg(team_id), sqlc.arg(field_type),
    CASE WHEN CAST(sqlc.arg(field_type) AS text) = 'text' THEN CAST(sqlc.narg(value) AS text) ELSE NULL END,
    CASE WHEN CAST(sqlc.arg(field_type) AS text) IN ('number','money') THEN CAST(sqlc.narg(value) AS numeric) ELSE NULL END,
    CASE WHEN CAST(sqlc.arg(field_type) AS text) = 'date' THEN CAST(sqlc.narg(value) AS date) ELSE NULL END,
    CASE WHEN CAST(sqlc.arg(field_type) AS text) = 'select' THEN CAST(sqlc.narg(value) AS uuid) ELSE NULL END,
    CASE WHEN CAST(sqlc.arg(field_type) AS text) = 'person' THEN CAST(sqlc.narg(value) AS uuid) ELSE NULL END)
ON CONFLICT (story_id, field_id) DO UPDATE SET text_value = EXCLUDED.text_value, numeric_value = EXCLUDED.numeric_value, date_value = EXCLUDED.date_value, option_id = EXCLUDED.option_id, person_id = EXCLUDED.person_id, updated_at = CURRENT_TIMESTAMP;

-- name: AppendValueAudit :exec
INSERT INTO story_custom_field_audit (story_id, field_id, workspace_id, actor_id, old_value, new_value, version)
VALUES (sqlc.arg(story_id), sqlc.arg(field_id), sqlc.arg(workspace_id), sqlc.arg(actor_id), sqlc.narg(old_value), sqlc.narg(new_value), sqlc.arg(version));

-- name: ListBatchStoryValues :many
SELECT story.id AS story_id, story.custom_fields_version, value.field_id,
    COALESCE(value.text_value, CAST(value.numeric_value AS text), CAST(value.date_value AS text), CAST(value.option_id AS text), CAST(value.person_id AS text)) AS value
FROM stories AS story
JOIN teams AS team ON team.team_id = story.team_id AND team.workspace_id = story.workspace_id
JOIN workspaces AS workspace ON workspace.workspace_id = story.workspace_id AND workspace.deleted_at IS NULL
JOIN workspace_members AS member ON member.workspace_id = story.workspace_id AND member.user_id = sqlc.arg(actor_id)
JOIN users AS actor ON actor.user_id = member.user_id AND actor.is_active = TRUE
LEFT JOIN story_custom_field_values AS value ON value.story_id = story.id AND value.workspace_id = story.workspace_id AND value.team_id = story.team_id
WHERE story.workspace_id = sqlc.arg(workspace_id) AND story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[])) AND story.deleted_at IS NULL
  AND member.role IN ('admin','member','guest')
  AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM team_members AS joined WHERE joined.team_id = story.team_id AND joined.user_id = actor.user_id))
ORDER BY story.id, value.field_id;
