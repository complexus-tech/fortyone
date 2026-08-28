-- name: InsertDeveloperCredentialAuditEvent :exec
INSERT INTO public.developer_credential_audit_events (
    event_id,
    workspace_id,
    actor_kind,
    actor_id,
    actor_credential_id,
    operation,
    subject_type,
    subject_id,
    result,
    reason_code,
    request_id,
    metadata,
    created_at
) VALUES (
    sqlc.arg(event_id),
    sqlc.arg(workspace_id),
    sqlc.arg(actor_kind),
    sqlc.arg(actor_id),
    sqlc.narg(actor_credential_id),
    sqlc.arg(operation),
    sqlc.arg(subject_type),
    sqlc.arg(subject_id),
    sqlc.arg(result),
    sqlc.narg(reason_code),
    sqlc.narg(request_id),
    CAST(sqlc.arg(metadata) AS jsonb),
    sqlc.arg(created_at)
);
