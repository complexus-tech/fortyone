-- name: RecordOutboundWebhookAuditEvent :exec
INSERT INTO public.outbound_webhook_audit_events (
    audit_event_id,
    workspace_id,
    actor_kind,
    actor_id,
    actor_credential_id,
    operation,
    endpoint_id,
    delivery_id,
    result,
    reason_code,
    request_id,
    metadata,
    created_at
) VALUES (
    sqlc.arg(audit_event_id),
    sqlc.arg(workspace_id),
    sqlc.arg(actor_kind),
    sqlc.arg(actor_id),
    sqlc.narg(actor_credential_id),
    sqlc.arg(operation),
    sqlc.narg(endpoint_id),
    sqlc.narg(delivery_id),
    sqlc.arg(result),
    sqlc.narg(reason_code),
    sqlc.narg(request_id),
    sqlc.arg(metadata),
    sqlc.arg(created_at)
);
