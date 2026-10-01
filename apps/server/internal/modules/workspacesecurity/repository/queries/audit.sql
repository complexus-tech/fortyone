-- name: ListAudit :many
WITH scoped_events AS (
    SELECT event_id AS id, CAST('security' AS text) AS source, actor_id,
        CAST('human_user' AS text) AS actor_type, CAST(resource_type AS text) AS resource_type,
        resource_id, CAST(operation AS text) AS operation, metadata, created_at
    FROM public.workspace_security_audit_events WHERE workspace_security_audit_events.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT event_id, CAST('workspace' AS text), actor_id, CAST(actor_type AS text), CAST(entity_type AS text), entity_id, CAST(event_type AS text),
        jsonb_strip_nulls(jsonb_build_object('reason', metadata -> 'reason', 'count', metadata -> 'count', 'teamId', metadata -> 'team_id')), created_at
    FROM public.audit_events WHERE audit_events.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT event_id, CAST('credentials' AS text), actor_id, CAST(actor_kind AS text), CAST(subject_type AS text), subject_id, CAST(operation AS text),
        jsonb_strip_nulls(jsonb_build_object('result', result, 'reasonCode', reason_code)), created_at
    FROM public.developer_credential_audit_events WHERE developer_credential_audit_events.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT audit_event_id, CAST('webhooks' AS text), actor_id, CAST(actor_kind AS text), CAST('webhook' AS text), endpoint_id, CAST(operation AS text),
        jsonb_strip_nulls(jsonb_build_object('result', result, 'reasonCode', reason_code, 'deliveryId', delivery_id)), created_at
    FROM public.outbound_webhook_audit_events WHERE outbound_webhook_audit_events.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT id, CAST('custom_fields' AS text), actor_id, CAST('human_user' AS text), CAST('story' AS text), story_id, CAST('custom_field.value_changed' AS text),
        jsonb_build_object('fieldId', field_id, 'version', version), created_at
    FROM public.story_custom_field_audit WHERE story_custom_field_audit.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT id, CAST('custom_fields' AS text), actor_id, CAST('human_user' AS text), CAST('custom_field' AS text), field_id, operation,
        jsonb_build_object('teamId', metadata -> 'teamId', 'oldIcon', metadata -> 'oldIcon', 'newIcon', metadata -> 'newIcon'), created_at
    FROM public.custom_field_definition_audit WHERE custom_field_definition_audit.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT id, CAST('sso' AS text), actor_id, CAST('human_user' AS text), CAST('sso_connection' AS text), connection_id, CAST(operation AS text), metadata, created_at
    FROM public.workspace_sso_audit_events WHERE workspace_sso_audit_events.workspace_id = sqlc.arg(workspace_id)
    UNION ALL
    SELECT event_id, CAST('scim' AS text), actor_id, CAST('scim_credential' AS text), CAST('scim_resource' AS text), resource_id, CAST(operation AS text),
        jsonb_strip_nulls(jsonb_build_object('credentialId', credential_id)), created_at
    FROM public.workspace_scim_audit_events WHERE workspace_scim_audit_events.workspace_id = sqlc.arg(workspace_id)
)
SELECT id, source,
    COALESCE(actor_id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS actor_id,
    actor_type, resource_type,
    COALESCE(resource_id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS resource_id,
    operation, metadata, created_at
FROM scoped_events
WHERE (CAST(sqlc.narg(actor_id) AS uuid) IS NULL OR actor_id = sqlc.narg(actor_id))
  AND (CAST(sqlc.arg(resource_type) AS text) = '' OR resource_type = sqlc.arg(resource_type))
  AND (CAST(sqlc.narg(resource_id) AS uuid) IS NULL OR resource_id = sqlc.narg(resource_id))
  AND (CAST(sqlc.narg(from_time) AS timestamptz) IS NULL OR created_at >= sqlc.narg(from_time))
  AND (CAST(sqlc.narg(to_time) AS timestamptz) IS NULL OR created_at < sqlc.narg(to_time))
  AND (CAST(sqlc.narg(before_time) AS timestamptz) IS NULL OR (created_at, id, source) < (sqlc.narg(before_time), CAST(sqlc.arg(before_id) AS uuid), CAST(sqlc.arg(before_source) AS text)))
ORDER BY created_at DESC, id DESC, source DESC LIMIT sqlc.arg(max_rows);
