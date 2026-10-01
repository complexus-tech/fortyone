ALTER TABLE custom_fields ADD COLUMN icon text
    CHECK (icon IS NULL OR icon IN ('text','number','calendar','list','person','team','workspace','goal','star','checklist','link','email','clock','work','attachment','globe'));

CREATE TABLE custom_field_definition_audit (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    field_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    operation text NOT NULL CHECK (operation IN ('custom_field.created','custom_field.updated','custom_field.archived')),
    metadata jsonb NOT NULL CHECK (jsonb_typeof(metadata) = 'object'),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX custom_field_definition_audit_workspace ON custom_field_definition_audit (workspace_id, created_at DESC, id DESC);

CREATE FUNCTION reject_custom_field_definition_audit_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'custom field definition audit records are immutable';
END;
$$;
CREATE TRIGGER custom_field_definition_audit_immutable
    BEFORE UPDATE OR DELETE ON custom_field_definition_audit
    FOR EACH ROW EXECUTE FUNCTION reject_custom_field_definition_audit_mutation();
