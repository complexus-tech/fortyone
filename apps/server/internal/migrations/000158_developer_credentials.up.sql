CREATE TABLE public.principals (
    principal_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    kind varchar(32) NOT NULL,
    name varchar(120) NOT NULL,
    subject_user_id uuid,
    workspace_role public.user_role,
    status varchar(16) NOT NULL DEFAULT 'active',
    created_by_user_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    disabled_at timestamptz,
    disabled_by_user_id uuid,
    disabled_reason varchar(240),
    CONSTRAINT principals_pkey PRIMARY KEY (principal_id),
    CONSTRAINT principals_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT principals_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT principals_disabled_by_user_id_fkey
        FOREIGN KEY (disabled_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT principals_workspace_subject_fkey
        FOREIGN KEY (workspace_id, subject_user_id)
        REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE,
    CONSTRAINT principals_workspace_identity_key UNIQUE (principal_id, workspace_id),
    CONSTRAINT principals_kind_check
        CHECK (kind IN ('human_user', 'service_account')),
    CONSTRAINT principals_status_check
        CHECK (status IN ('active', 'disabled')),
    CONSTRAINT principals_name_check
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    CONSTRAINT principals_identity_shape_check CHECK (
        (
            kind = 'human_user'
            AND subject_user_id IS NOT NULL
            AND workspace_role IS NULL
        )
        OR
        (
            kind = 'service_account'
            AND subject_user_id IS NULL
            AND workspace_role IS NOT NULL
        )
    ),
    CONSTRAINT principals_service_account_role_check CHECK (
        kind <> 'service_account'
        OR workspace_role IN ('guest', 'member')
    ),
    CONSTRAINT principals_disabled_state_check CHECK (
        (
            status = 'active'
            AND disabled_at IS NULL
            AND disabled_by_user_id IS NULL
            AND disabled_reason IS NULL
        )
        OR
        (
            status = 'disabled'
            AND disabled_at IS NOT NULL
            AND disabled_reason IS NOT NULL
        )
    ),
    CONSTRAINT principals_timestamps_check CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX principals_workspace_human_subject_key
    ON public.principals (workspace_id, subject_user_id)
    WHERE kind = 'human_user';

CREATE INDEX principals_workspace_service_accounts_idx
    ON public.principals (workspace_id, created_at DESC, principal_id DESC)
    WHERE kind = 'service_account';

CREATE FUNCTION public.reject_principal_identity_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'principal identity fields are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER principals_identity_immutable
BEFORE UPDATE OF principal_id, workspace_id, kind, subject_user_id ON public.principals
FOR EACH ROW
EXECUTE FUNCTION public.reject_principal_identity_mutation();

CREATE TABLE public.api_credentials (
    credential_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    principal_id uuid NOT NULL,
    kind varchar(32) NOT NULL,
    name varchar(120) NOT NULL,
    lookup_prefix varchar(12) NOT NULL,
    secret_digest bytea NOT NULL,
    token_version smallint NOT NULL,
    digest_key_id varchar(64) NOT NULL,
    digest_key_version integer NOT NULL,
    expires_at timestamptz NOT NULL,
    last_used_at timestamptz,
    rotated_from_id uuid,
    rotated_at timestamptz,
    revoked_at timestamptz,
    revoked_by_user_id uuid,
    revoked_reason varchar(240),
    created_by_user_id uuid,
    created_at timestamptz NOT NULL,
    CONSTRAINT api_credentials_pkey PRIMARY KEY (credential_id),
    CONSTRAINT api_credentials_principal_workspace_fkey
        FOREIGN KEY (principal_id, workspace_id)
        REFERENCES public.principals(principal_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT api_credentials_rotated_from_id_fkey
        FOREIGN KEY (rotated_from_id) REFERENCES public.api_credentials(credential_id) ON DELETE SET NULL,
    CONSTRAINT api_credentials_revoked_by_user_id_fkey
        FOREIGN KEY (revoked_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT api_credentials_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT api_credentials_workspace_identity_key UNIQUE (credential_id, workspace_id),
    CONSTRAINT api_credentials_lookup_prefix_key UNIQUE (lookup_prefix),
    CONSTRAINT api_credentials_single_rotation_key UNIQUE (rotated_from_id),
    CONSTRAINT api_credentials_kind_check
        CHECK (kind IN ('personal_access_token', 'service_account_key')),
    CONSTRAINT api_credentials_name_check
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    CONSTRAINT api_credentials_lookup_prefix_check
        CHECK (lookup_prefix ~ '^[a-f0-9]{12}$'),
    CONSTRAINT api_credentials_secret_digest_check
        CHECK (octet_length(secret_digest) = 32),
    CONSTRAINT api_credentials_token_version_check
        CHECK (token_version = 1),
    CONSTRAINT api_credentials_digest_key_check CHECK (
        char_length(btrim(digest_key_id)) BETWEEN 1 AND 64
        AND digest_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'
        AND digest_key_version > 0
    ),
    CONSTRAINT api_credentials_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT api_credentials_last_used_check CHECK (
        last_used_at IS NULL OR last_used_at >= created_at
    ),
    CONSTRAINT api_credentials_rotation_check CHECK (
        (rotated_from_id IS NULL) OR (rotated_from_id <> credential_id)
    ),
    CONSTRAINT api_credentials_revocation_check CHECK (
        (
            revoked_at IS NULL
            AND revoked_by_user_id IS NULL
            AND revoked_reason IS NULL
        )
        OR
        (
            revoked_at IS NOT NULL
            AND revoked_reason IS NOT NULL
        )
    )
);

CREATE FUNCTION public.enforce_api_credential_principal_kind()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    stored_principal_kind varchar(32);
BEGIN
    SELECT principal.kind
    INTO stored_principal_kind
    FROM public.principals AS principal
    WHERE principal.principal_id = NEW.principal_id
      AND principal.workspace_id = NEW.workspace_id;

    IF stored_principal_kind IS NULL OR NOT (
        (NEW.kind = 'personal_access_token' AND stored_principal_kind = 'human_user')
        OR
        (NEW.kind = 'service_account_key' AND stored_principal_kind = 'service_account')
    ) THEN
        RAISE EXCEPTION 'credential kind does not match principal kind'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'api_credentials_principal_kind_check';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER api_credentials_principal_kind_enforced
BEFORE INSERT OR UPDATE OF principal_id, workspace_id, kind ON public.api_credentials
FOR EACH ROW
EXECUTE FUNCTION public.enforce_api_credential_principal_kind();

CREATE FUNCTION public.reject_api_credential_identity_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'API credential identity and secret fields are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER api_credentials_identity_immutable
BEFORE UPDATE OF
    credential_id,
    workspace_id,
    principal_id,
    kind,
    lookup_prefix,
    secret_digest,
    token_version,
    digest_key_id,
    digest_key_version,
    rotated_from_id,
    created_by_user_id,
    created_at
ON public.api_credentials
FOR EACH ROW
EXECUTE FUNCTION public.reject_api_credential_identity_mutation();

CREATE INDEX api_credentials_principal_created_idx
    ON public.api_credentials (principal_id, created_at DESC, credential_id DESC);

CREATE INDEX api_credentials_workspace_active_idx
    ON public.api_credentials (workspace_id, expires_at, credential_id)
    WHERE revoked_at IS NULL;

CREATE TABLE public.api_credential_scopes (
    credential_id uuid NOT NULL,
    scope varchar(64) NOT NULL,
    CONSTRAINT api_credential_scopes_pkey PRIMARY KEY (credential_id, scope),
    CONSTRAINT api_credential_scopes_credential_id_fkey
        FOREIGN KEY (credential_id) REFERENCES public.api_credentials(credential_id) ON DELETE CASCADE,
    CONSTRAINT api_credential_scopes_scope_check CHECK (
        scope IN (
            'workspaces:read',
            'teams:read',
            'stories:read',
            'stories:write',
            'comments:read',
            'comments:write',
            'labels:read',
            'labels:write',
            'sprints:read',
            'objectives:read',
            'objectives:write',
            'webhooks:manage',
            'integrations:manage',
            'service_accounts:manage'
        )
    )
);

CREATE FUNCTION public.enforce_service_account_scope_boundary()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.scope = 'service_accounts:manage' AND EXISTS (
        SELECT 1
        FROM public.api_credentials AS credential
        WHERE credential.credential_id = NEW.credential_id
          AND credential.kind = 'service_account_key'
    ) THEN
        RAISE EXCEPTION 'service-account keys cannot manage service accounts'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'api_credential_scopes_service_account_management_check';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER api_credential_scopes_service_account_boundary
BEFORE INSERT OR UPDATE OF credential_id, scope ON public.api_credential_scopes
FOR EACH ROW
EXECUTE FUNCTION public.enforce_service_account_scope_boundary();

ALTER TABLE public.teams
    ADD CONSTRAINT teams_team_id_workspace_id_key UNIQUE (team_id, workspace_id);

CREATE TABLE public.api_credential_team_restrictions (
    credential_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    team_id uuid NOT NULL,
    CONSTRAINT api_credential_team_restrictions_pkey PRIMARY KEY (credential_id, team_id),
    CONSTRAINT api_credential_team_restrictions_credential_workspace_fkey
        FOREIGN KEY (credential_id, workspace_id)
        REFERENCES public.api_credentials(credential_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT api_credential_team_restrictions_team_workspace_fkey
        FOREIGN KEY (team_id, workspace_id)
        REFERENCES public.teams(team_id, workspace_id) ON DELETE CASCADE
);

CREATE INDEX api_credential_team_restrictions_team_idx
    ON public.api_credential_team_restrictions (workspace_id, team_id, credential_id);

-- This ledger intentionally retains UUID facts instead of foreign keys. Actor,
-- principal, credential, user, or workspace deletion must not rewrite history.
CREATE TABLE public.developer_credential_audit_events (
    event_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    actor_kind varchar(32) NOT NULL,
    actor_id uuid NOT NULL,
    actor_credential_id uuid,
    operation varchar(80) NOT NULL,
    subject_type varchar(32) NOT NULL,
    subject_id uuid NOT NULL,
    result varchar(16) NOT NULL,
    reason_code varchar(64),
    request_id varchar(128),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL,
    CONSTRAINT developer_credential_audit_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT developer_credential_audit_events_actor_kind_check CHECK (
        actor_kind IN (
            'human_user',
            'personal_token',
            'service_account',
            'oauth_user',
            'oauth_application',
            'system',
            'external_contributor'
        )
    ),
    CONSTRAINT developer_credential_audit_events_operation_check
        CHECK (char_length(btrim(operation)) BETWEEN 1 AND 80),
    CONSTRAINT developer_credential_audit_events_subject_type_check
        CHECK (subject_type IN ('principal', 'api_credential')),
    CONSTRAINT developer_credential_audit_events_result_check
        CHECK (result IN ('succeeded', 'denied', 'failed')),
    CONSTRAINT developer_credential_audit_events_attribution_check CHECK (
        (actor_kind <> 'human_user' OR actor_credential_id IS NULL)
        AND
        (actor_kind NOT IN ('personal_token', 'service_account') OR actor_credential_id IS NOT NULL)
    ),
    CONSTRAINT developer_credential_audit_events_metadata_check
        CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX developer_credential_audit_workspace_created_idx
    ON public.developer_credential_audit_events (workspace_id, created_at DESC, event_id DESC);

CREATE INDEX developer_credential_audit_subject_created_idx
    ON public.developer_credential_audit_events (subject_type, subject_id, created_at DESC);

CREATE FUNCTION public.reject_developer_credential_audit_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'developer credential audit events are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER developer_credential_audit_events_immutable
BEFORE UPDATE OR DELETE ON public.developer_credential_audit_events
FOR EACH ROW
EXECUTE FUNCTION public.reject_developer_credential_audit_mutation();
