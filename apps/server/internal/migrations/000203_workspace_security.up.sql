CREATE TABLE public.workspace_security_policies (
    workspace_id uuid PRIMARY KEY REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    allowed_domains text[] NOT NULL DEFAULT '{}',
    allow_guests boolean NOT NULL DEFAULT TRUE,
    max_session_age_hours integer NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (cardinality(allowed_domains) <= 50),
    CHECK (max_session_age_hours BETWEEN 0 AND 720),
    CHECK (version >= 0)
);

-- These fences are tenant specific. They must never increment the account's
-- global session version or invalidate a user's sessions in another workspace.
CREATE TABLE public.workspace_member_session_epochs (
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    revoked_before timestamptz NOT NULL,
    PRIMARY KEY (workspace_id, user_id),
    FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE
);

-- Session IDs are random, non-secret metadata. Raw cookies and Redis lookup
-- tokens are deliberately absent from this table and from its audit ledger.
CREATE TABLE public.workspace_browser_sessions (
    workspace_id uuid NOT NULL,
    session_id uuid NOT NULL,
    user_id uuid NOT NULL,
    authenticated_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    PRIMARY KEY (workspace_id, session_id),
    FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE,
    CHECK (expires_at > authenticated_at),
    CHECK (last_seen_at >= authenticated_at)
);
CREATE INDEX workspace_browser_sessions_member_idx
    ON public.workspace_browser_sessions (workspace_id, user_id, last_seen_at DESC, session_id DESC);

-- UUID facts survive member removal; immutable audits never retain credentials.
CREATE TABLE public.workspace_security_audit_events (
    event_id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    resource_type varchar(64) NOT NULL,
    resource_id uuid NOT NULL,
    operation varchar(128) NOT NULL,
    metadata jsonb NOT NULL DEFAULT CAST('{}' AS jsonb),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (jsonb_typeof(metadata) = 'object')
);
CREATE INDEX workspace_security_audit_events_scope_idx
    ON public.workspace_security_audit_events (workspace_id, created_at DESC, event_id DESC);
CREATE FUNCTION public.reject_workspace_security_audit_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'workspace security audit events are immutable' USING ERRCODE = '55000';
END;
$$;
CREATE TRIGGER workspace_security_audit_immutable
BEFORE UPDATE OR DELETE ON public.workspace_security_audit_events
FOR EACH ROW EXECUTE FUNCTION public.reject_workspace_security_audit_mutation();
