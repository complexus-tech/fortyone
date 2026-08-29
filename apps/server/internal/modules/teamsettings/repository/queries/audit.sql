-- name: InsertTeamSettingsAuditEvent :exec
INSERT INTO public.audit_events (
    workspace_id,
    team_id,
    actor_type,
    actor_id,
    entity_type,
    entity_id,
    event_type,
    metadata
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    sqlc.arg(actor_type),
    CAST(sqlc.narg(actor_id) AS uuid),
    sqlc.arg(entity_type),
    sqlc.arg(entity_id),
    sqlc.arg(event_type),
    CAST(sqlc.arg(metadata) AS jsonb)
);
