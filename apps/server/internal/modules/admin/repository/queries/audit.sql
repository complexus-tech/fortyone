-- name: ListAdminAuditLogs :many
SELECT
    audit.id,
    audit.actor_user_id,
    actor.email AS actor_email,
    actor.full_name AS actor_name,
    audit.target_type,
    audit.target_id,
    audit.workspace_id,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    audit.action,
    audit.field_name,
    audit.old_value,
    audit.new_value,
    audit.reason,
    audit.metadata,
    audit.created_at,
    CAST(COUNT(*) OVER () AS bigint) AS total_count
FROM admin_audit_logs AS audit
JOIN users AS actor ON actor.user_id = audit.actor_user_id
LEFT JOIN workspaces AS workspace ON workspace.workspace_id = audit.workspace_id
WHERE (
      NOT CAST(sqlc.arg(workspace_id_set) AS boolean)
      OR audit.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  )
  AND (
      CAST(sqlc.arg(target_type_filter) AS text) = ''
      OR audit.target_type = CAST(sqlc.arg(target_type_filter) AS text)
  )
  AND (
      CAST(sqlc.arg(action_filter) AS text) = ''
      OR audit.action = CAST(sqlc.arg(action_filter) AS text)
  )
  AND (
      CAST(sqlc.arg(search_text) AS text) = ''
      OR audit.action ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
      OR audit.field_name ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
      OR audit.reason ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
      OR workspace.name ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
      OR workspace.slug ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
      OR CAST(audit.target_id AS text) ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
  )
  AND (
      CAST(sqlc.arg(actor_search) AS text) = ''
      OR actor.email ILIKE '%' || CAST(sqlc.arg(actor_search) AS text) || '%'
      OR actor.full_name ILIKE '%' || CAST(sqlc.arg(actor_search) AS text) || '%'
      OR actor.username ILIKE '%' || CAST(sqlc.arg(actor_search) AS text) || '%'
  )
  AND (
      NOT CAST(sqlc.arg(from_set) AS boolean)
      OR audit.created_at >= CAST(sqlc.arg(from_at) AS timestamptz)
  )
  AND (
      NOT CAST(sqlc.arg(to_set) AS boolean)
      OR audit.created_at <= CAST(sqlc.arg(to_at) AS timestamptz)
  )
ORDER BY audit.created_at DESC, audit.id DESC
LIMIT CAST(sqlc.arg(row_limit) AS integer)
OFFSET CAST(sqlc.arg(row_offset) AS integer);

-- name: ListAdminNotes :many
SELECT
    note.id,
    note.target_type,
    note.target_id,
    note.workspace_id,
    note.body,
    note.created_by_user_id,
    creator.full_name AS created_by_name,
    creator.email AS created_by_email,
    note.created_at,
    CAST(COUNT(*) OVER () AS bigint) AS total_count
FROM admin_notes AS note
JOIN users AS creator ON creator.user_id = note.created_by_user_id
WHERE (
      CAST(sqlc.arg(target_type_filter) AS text) = ''
      OR note.target_type = CAST(sqlc.arg(target_type_filter) AS text)
  )
  AND (
      NOT CAST(sqlc.arg(target_id_set) AS boolean)
      OR note.target_id = CAST(sqlc.arg(target_id) AS uuid)
  )
  AND (
      NOT CAST(sqlc.arg(workspace_id_set) AS boolean)
      OR note.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  )
ORDER BY note.created_at DESC, note.id DESC
LIMIT CAST(sqlc.arg(row_limit) AS integer)
OFFSET CAST(sqlc.arg(row_offset) AS integer);

-- name: InsertAdminAuditLog :one
INSERT INTO admin_audit_logs (
    actor_user_id,
    target_type,
    target_id,
    workspace_id,
    action,
    field_name,
    old_value,
    new_value,
    reason,
    metadata
) VALUES (
    CAST(sqlc.arg(actor_user_id) AS uuid),
    CAST(sqlc.arg(target_type) AS text),
    CAST(sqlc.narg(target_id) AS uuid),
    CAST(sqlc.narg(workspace_id) AS uuid),
    CAST(sqlc.arg(action) AS text),
    CAST(sqlc.narg(field_name) AS text),
    CAST(sqlc.arg(old_value) AS jsonb),
    CAST(sqlc.arg(new_value) AS jsonb),
    CAST(sqlc.narg(reason) AS text),
    CAST(sqlc.arg(metadata) AS jsonb)
)
RETURNING id, created_at;
