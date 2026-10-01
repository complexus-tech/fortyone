CREATE TABLE public.workspace_scim_credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    issuer_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    token_digest bytea NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    token_prefix text NOT NULL CHECK (char_length(token_prefix) <= 24),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);
CREATE INDEX workspace_scim_credentials_tenant_idx ON public.workspace_scim_credentials(workspace_id, created_at DESC, id DESC);
CREATE TABLE public.workspace_scim_users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    user_name text NOT NULL CHECK (char_length(user_name) BETWEEN 1 AND 255),
    external_id text,
    active boolean NOT NULL DEFAULT TRUE,
    profile jsonb NOT NULL CHECK (jsonb_typeof(profile) = 'object' AND octet_length(CAST(profile AS text)) <= 16384),
    previous_role text NOT NULL DEFAULT 'member' CHECK (previous_role IN ('member', 'guest', 'admin')),
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
CREATE UNIQUE INDEX workspace_scim_users_name_key ON public.workspace_scim_users(workspace_id, lower(user_name)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX workspace_scim_users_account_key ON public.workspace_scim_users(workspace_id, user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX workspace_scim_users_external_key ON public.workspace_scim_users(workspace_id, external_id) WHERE external_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX workspace_scim_users_list_idx ON public.workspace_scim_users(workspace_id, created_at, id) WHERE deleted_at IS NULL;
CREATE TABLE public.workspace_scim_audit_events (
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    credential_id uuid,
    resource_id uuid NOT NULL,
    operation text NOT NULL CHECK (char_length(operation) <= 64),
    metadata jsonb NOT NULL DEFAULT CAST('{}' AS jsonb) CHECK (jsonb_typeof(metadata) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX workspace_scim_audit_scope_idx ON public.workspace_scim_audit_events(workspace_id, created_at DESC, event_id DESC);
CREATE FUNCTION public.reject_workspace_scim_audit_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'workspace SCIM audit events are immutable' USING ERRCODE = '55000';
END;
$$;
CREATE TRIGGER workspace_scim_audit_immutable BEFORE UPDATE OR DELETE ON public.workspace_scim_audit_events
FOR EACH ROW EXECUTE FUNCTION public.reject_workspace_scim_audit_mutation();
CREATE TABLE public.workspace_scim_seat_reconciliation (
    workspace_id uuid PRIMARY KEY REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    generation bigint NOT NULL DEFAULT 1,
    last_error text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now()
);
