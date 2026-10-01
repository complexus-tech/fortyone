CREATE TABLE public.workspace_sso_connections (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    issuer text NOT NULL,
    client_id varchar(512) NOT NULL,
    client_secret_envelope text NOT NULL,
    enabled boolean NOT NULL DEFAULT TRUE,
    require_sso boolean NOT NULL DEFAULT FALSE,
    generation bigint NOT NULL DEFAULT 1,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at timestamptz,
    UNIQUE (id, workspace_id),
    CHECK (char_length(issuer) BETWEEN 8 AND 2048 AND issuer LIKE 'https://%'),
    CHECK (char_length(btrim(client_id)) BETWEEN 1 AND 512),
    CHECK (generation > 0 AND version > 0),
    CHECK (NOT require_sso OR (enabled AND archived_at IS NULL))
);
CREATE UNIQUE INDEX workspace_sso_current_connection ON public.workspace_sso_connections(workspace_id) WHERE archived_at IS NULL;

-- Email is never an SSO identity key. Initial binding requires proof of an
-- authenticated existing account; subsequent sign-in uses the IdP subject.
CREATE TABLE public.workspace_sso_identities (
    connection_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    subject varchar(512) NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_authenticated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (connection_id, subject),
    UNIQUE (connection_id, user_id),
    FOREIGN KEY (connection_id, workspace_id) REFERENCES public.workspace_sso_connections(id, workspace_id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id, user_id) REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE,
    CHECK (char_length(btrim(subject)) BETWEEN 1 AND 512)
);
CREATE TABLE public.workspace_sso_audit_events (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    connection_id uuid NOT NULL,
    operation varchar(128) NOT NULL,
    metadata jsonb NOT NULL DEFAULT CAST('{}' AS jsonb),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (jsonb_typeof(metadata) = 'object')
);
CREATE INDEX workspace_sso_audit_scope_idx ON public.workspace_sso_audit_events(workspace_id, created_at DESC, id DESC);
CREATE FUNCTION public.reject_workspace_sso_audit_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'workspace SSO audit events are immutable' USING ERRCODE = '55000';
END;
$$;
CREATE TRIGGER workspace_sso_audit_immutable BEFORE UPDATE OR DELETE ON public.workspace_sso_audit_events
FOR EACH ROW EXECUTE FUNCTION public.reject_workspace_sso_audit_mutation();
