-- name: ListFields :many
SELECT id, team_id, name, field_type, currency, icon, show_on_create, archived_at, created_at, updated_at
FROM custom_fields WHERE workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id)
ORDER BY created_at, id LIMIT 200;

-- name: ListFieldOptions :many
SELECT option.id, option.field_id, option.name, option.archived_at
FROM custom_field_options AS option
JOIN custom_fields AS field ON field.id = option.field_id
WHERE field.workspace_id = sqlc.arg(workspace_id) AND field.team_id = sqlc.arg(team_id)
ORDER BY option.field_id, option.position, option.id;

-- name: LockFields :many
SELECT id, team_id, name, field_type, currency, icon, show_on_create, archived_at, created_at, updated_at
FROM custom_fields WHERE workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id)
ORDER BY created_at, id LIMIT 200 FOR SHARE;

-- name: LockField :one
SELECT id, team_id, name, field_type, currency, icon, show_on_create, archived_at, created_at, updated_at
FROM custom_fields WHERE id = sqlc.arg(field_id) AND workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id)
FOR UPDATE;

-- name: FindField :one
SELECT id, team_id, name, field_type, currency, icon, show_on_create, archived_at, created_at, updated_at
FROM custom_fields WHERE id = sqlc.arg(field_id) AND workspace_id = sqlc.arg(workspace_id);

-- name: CountFields :one
SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE archived_at IS NULL) AS active
FROM custom_fields WHERE workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id);

-- name: InsertField :one
INSERT INTO custom_fields (id, workspace_id, team_id, name, field_type, currency, icon, show_on_create)
VALUES (sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(team_id), sqlc.arg(name), sqlc.arg(field_type), sqlc.narg(currency), sqlc.narg(icon), sqlc.arg(show_on_create))
RETURNING id, team_id, name, field_type, currency, icon, show_on_create, archived_at, created_at, updated_at;

-- name: UpdateField :one
UPDATE custom_fields SET name = sqlc.arg(name), show_on_create = sqlc.arg(show_on_create),
    icon = CASE WHEN CAST(sqlc.arg(icon_set) AS boolean) THEN sqlc.narg(icon) ELSE icon END, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(field_id) AND workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id) AND archived_at IS NULL
RETURNING id, team_id, name, field_type, currency, icon, show_on_create, archived_at, created_at, updated_at;

-- name: ArchiveField :execrows
UPDATE custom_fields SET archived_at = COALESCE(archived_at,CURRENT_TIMESTAMP), updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(field_id) AND workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id);

-- name: ArchiveOptions :exec
UPDATE custom_field_options SET archived_at = COALESCE(archived_at,CURRENT_TIMESTAMP)
WHERE field_id = sqlc.arg(field_id);

-- name: UpsertOption :exec
INSERT INTO custom_field_options (id, field_id, name, position) VALUES (sqlc.arg(id), sqlc.arg(field_id), sqlc.arg(name), sqlc.arg(position))
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, position = EXCLUDED.position, archived_at = NULL
WHERE custom_field_options.field_id = EXCLUDED.field_id;

-- name: ListOptionsForField :many
SELECT id, name, archived_at FROM custom_field_options
WHERE field_id = sqlc.arg(field_id) ORDER BY position, id LIMIT 500;

-- name: AppendDefinitionAudit :exec
INSERT INTO custom_field_definition_audit (id, workspace_id, field_id, actor_id, operation, metadata)
VALUES (sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(field_id), sqlc.arg(actor_id), sqlc.arg(operation), sqlc.arg(metadata));
