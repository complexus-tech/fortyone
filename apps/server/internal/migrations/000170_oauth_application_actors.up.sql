-- OAuth application actors are installation-owned machine principals. The
-- installer is retained as lifecycle metadata, never as the runtime actor.
ALTER TABLE public.principals
    DROP CONSTRAINT principals_kind_check,
    DROP CONSTRAINT principals_identity_shape_check,
    DROP CONSTRAINT principals_service_account_role_check;

ALTER TABLE public.principals
    ADD CONSTRAINT principals_kind_check CHECK (
        kind IN ('human_user', 'service_account', 'oauth_application')
    ),
    ADD CONSTRAINT principals_identity_shape_check CHECK (
        (
            kind = 'human_user'
            AND subject_user_id IS NOT NULL
            AND workspace_role IS NULL
        )
        OR
        (
            kind IN ('service_account', 'oauth_application')
            AND subject_user_id IS NULL
            AND workspace_role IS NOT NULL
        )
    ),
    ADD CONSTRAINT principals_machine_role_check CHECK (
        kind NOT IN ('service_account', 'oauth_application')
        OR workspace_role IN ('guest', 'member')
    );

CREATE INDEX principals_workspace_oauth_applications_idx
    ON public.principals (workspace_id, created_at DESC, principal_id DESC)
    WHERE kind = 'oauth_application';

ALTER TABLE public.oauth_applications
    ADD COLUMN owner_workspace_id uuid,
    ADD COLUMN owner_user_id uuid,
    ADD CONSTRAINT oauth_applications_owner_workspace_id_fkey
        FOREIGN KEY (owner_workspace_id)
        REFERENCES public.workspaces(workspace_id) ON DELETE RESTRICT,
    ADD CONSTRAINT oauth_applications_owner_user_id_fkey
        FOREIGN KEY (owner_user_id)
        REFERENCES public.users(user_id) ON DELETE RESTRICT,
    ADD CONSTRAINT oauth_applications_owner_shape_check CHECK (
        (
            owner_workspace_id IS NULL
            AND owner_user_id IS NULL
        )
        OR
        (
            registration_kind = 'confidential'
            AND owner_workspace_id IS NOT NULL
            AND owner_user_id IS NOT NULL
        )
    );

CREATE INDEX oauth_applications_owner_created_idx
    ON public.oauth_applications (
        owner_workspace_id,
        created_at DESC,
        application_id DESC
    )
    WHERE owner_workspace_id IS NOT NULL;

DROP TRIGGER principals_identity_immutable ON public.principals;

CREATE TRIGGER principals_identity_immutable
BEFORE UPDATE OF principal_id, workspace_id, kind, subject_user_id ON public.principals
FOR EACH ROW
EXECUTE FUNCTION public.reject_principal_identity_mutation();

CREATE FUNCTION public.reject_oauth_application_identity_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'OAuth application identity and ownership fields are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER oauth_applications_identity_immutable
BEFORE UPDATE OF
    application_id,
    client_id,
    registration_kind,
    owner_workspace_id,
    owner_user_id,
    created_at
ON public.oauth_applications
FOR EACH ROW
EXECUTE FUNCTION public.reject_oauth_application_identity_mutation();

CREATE TABLE public.oauth_client_secrets (
    secret_id uuid NOT NULL,
    application_id uuid NOT NULL,
    lookup_prefix varchar(12) NOT NULL,
    secret_digest bytea NOT NULL,
    digest_key_id varchar(64) NOT NULL,
    expires_at timestamptz NOT NULL,
    last_used_at timestamptz,
    rotated_from_id uuid,
    overlap_expires_at timestamptz,
    revoked_at timestamptz,
    revoked_by_user_id uuid,
    revoked_reason varchar(240),
    created_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT oauth_client_secrets_pkey PRIMARY KEY (secret_id),
    CONSTRAINT oauth_client_secrets_application_id_fkey
        FOREIGN KEY (application_id)
        REFERENCES public.oauth_applications(application_id) ON DELETE CASCADE,
    CONSTRAINT oauth_client_secrets_rotated_from_id_fkey
        FOREIGN KEY (rotated_from_id)
        REFERENCES public.oauth_client_secrets(secret_id) ON DELETE RESTRICT,
    CONSTRAINT oauth_client_secrets_lookup_prefix_key UNIQUE (lookup_prefix),
    CONSTRAINT oauth_client_secrets_single_rotation_key UNIQUE (rotated_from_id),
    CONSTRAINT oauth_client_secrets_lookup_prefix_check
        CHECK (lookup_prefix ~ '^[a-f0-9]{12}$'),
    CONSTRAINT oauth_client_secrets_digest_check
        CHECK (octet_length(secret_digest) = 32),
    CONSTRAINT oauth_client_secrets_digest_key_id_check CHECK (
        char_length(btrim(digest_key_id)) BETWEEN 1 AND 64
        AND digest_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'
    ),
    CONSTRAINT oauth_client_secrets_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT oauth_client_secrets_last_used_check CHECK (
        last_used_at IS NULL OR last_used_at >= created_at
    ),
    CONSTRAINT oauth_client_secrets_rotation_check CHECK (
        rotated_from_id IS NULL OR rotated_from_id <> secret_id
    ),
    CONSTRAINT oauth_client_secrets_overlap_check CHECK (
        overlap_expires_at IS NULL OR overlap_expires_at > created_at
    ),
    CONSTRAINT oauth_client_secrets_revocation_check CHECK (
        (
            revoked_at IS NULL
            AND revoked_by_user_id IS NULL
            AND revoked_reason IS NULL
        )
        OR
        (
            revoked_at IS NOT NULL
            AND revoked_by_user_id IS NOT NULL
            AND char_length(btrim(revoked_reason)) BETWEEN 1 AND 240
        )
    )
);

CREATE INDEX oauth_client_secrets_application_created_idx
    ON public.oauth_client_secrets (
        application_id,
        created_at DESC,
        secret_id DESC
    );

CREATE INDEX oauth_client_secrets_active_expiry_idx
    ON public.oauth_client_secrets (expires_at, secret_id)
    WHERE revoked_at IS NULL;

CREATE FUNCTION public.enforce_oauth_client_secret_lifecycle()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.overlap_expires_at IS NOT NULL
       AND NEW.overlap_expires_at IS DISTINCT FROM OLD.overlap_expires_at THEN
        RAISE EXCEPTION 'OAuth client-secret overlap cutoff is immutable once set'
            USING ERRCODE = '55000';
    END IF;
    IF OLD.revoked_at IS NOT NULL AND (
        NEW.revoked_at IS DISTINCT FROM OLD.revoked_at
        OR NEW.revoked_by_user_id IS DISTINCT FROM OLD.revoked_by_user_id
        OR NEW.revoked_reason IS DISTINCT FROM OLD.revoked_reason
    ) THEN
        RAISE EXCEPTION 'OAuth client-secret revocation is immutable once set'
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER oauth_client_secrets_lifecycle_fenced
BEFORE UPDATE ON public.oauth_client_secrets
FOR EACH ROW
EXECUTE FUNCTION public.enforce_oauth_client_secret_lifecycle();

CREATE FUNCTION public.reject_oauth_client_secret_identity_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'OAuth client-secret identity and digest fields are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER oauth_client_secrets_identity_immutable
BEFORE UPDATE OF
    secret_id,
    application_id,
    lookup_prefix,
    secret_digest,
    digest_key_id,
    expires_at,
    rotated_from_id,
    created_by_user_id,
    created_at
ON public.oauth_client_secrets
FOR EACH ROW
EXECUTE FUNCTION public.reject_oauth_client_secret_identity_mutation();

CREATE TABLE public.oauth_application_installations (
    installation_id uuid NOT NULL,
    application_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    principal_id uuid NOT NULL,
    resource varchar(512) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    installed_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    last_used_at timestamptz,
    revoked_at timestamptz,
    revoked_by_user_id uuid,
    revoked_reason varchar(240),
    CONSTRAINT oauth_application_installations_pkey PRIMARY KEY (installation_id),
    CONSTRAINT oauth_application_installations_application_id_fkey
        FOREIGN KEY (application_id)
        REFERENCES public.oauth_applications(application_id) ON DELETE RESTRICT,
    CONSTRAINT oauth_application_installations_workspace_id_fkey
        FOREIGN KEY (workspace_id)
        REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT oauth_application_installations_principal_workspace_fkey
        FOREIGN KEY (principal_id, workspace_id)
        REFERENCES public.principals(principal_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT oauth_application_installations_principal_id_key UNIQUE (principal_id),
    CONSTRAINT oauth_application_installations_workspace_identity_key
        UNIQUE (installation_id, workspace_id),
    CONSTRAINT oauth_application_installations_resource_check
        CHECK (char_length(btrim(resource)) BETWEEN 1 AND 512),
    CONSTRAINT oauth_application_installations_status_check
        CHECK (status IN ('active', 'revoked')),
    CONSTRAINT oauth_application_installations_timestamps_check CHECK (
        updated_at >= created_at
        AND (last_used_at IS NULL OR last_used_at >= created_at)
    ),
    CONSTRAINT oauth_application_installations_revocation_check CHECK (
        (
            status = 'active'
            AND revoked_at IS NULL
            AND revoked_by_user_id IS NULL
            AND revoked_reason IS NULL
        )
        OR
        (
            status = 'revoked'
            AND revoked_at IS NOT NULL
            AND revoked_by_user_id IS NOT NULL
            AND char_length(btrim(revoked_reason)) BETWEEN 1 AND 240
        )
    )
);

CREATE UNIQUE INDEX oauth_application_installations_active_identity_key
    ON public.oauth_application_installations (
        application_id,
        workspace_id,
        resource
    )
    WHERE status = 'active';

CREATE INDEX oauth_application_installations_workspace_created_idx
    ON public.oauth_application_installations (
        workspace_id,
        created_at DESC,
        installation_id DESC
    );

CREATE FUNCTION public.enforce_oauth_application_installation_principal()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM public.principals AS principal
        WHERE principal.principal_id = NEW.principal_id
          AND principal.workspace_id = NEW.workspace_id
          AND principal.kind = 'oauth_application'
    ) THEN
        RAISE EXCEPTION 'OAuth installation requires an OAuth application principal'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'oauth_application_installations_principal_kind_check';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER oauth_application_installations_principal_enforced
BEFORE INSERT OR UPDATE OF principal_id, workspace_id
ON public.oauth_application_installations
FOR EACH ROW
EXECUTE FUNCTION public.enforce_oauth_application_installation_principal();

CREATE FUNCTION public.reject_oauth_application_installation_identity_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'OAuth installation identity and installer fields are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER oauth_application_installations_identity_immutable
BEFORE UPDATE OF
    installation_id,
    application_id,
    workspace_id,
    principal_id,
    installed_by_user_id,
    created_at
ON public.oauth_application_installations
FOR EACH ROW
EXECUTE FUNCTION public.reject_oauth_application_installation_identity_mutation();

CREATE TABLE public.oauth_application_installation_scopes (
    installation_id uuid NOT NULL,
    scope varchar(64) NOT NULL,
    CONSTRAINT oauth_application_installation_scopes_pkey
        PRIMARY KEY (installation_id, scope),
    CONSTRAINT oauth_application_installation_scopes_installation_id_fkey
        FOREIGN KEY (installation_id)
        REFERENCES public.oauth_application_installations(installation_id) ON DELETE CASCADE,
    CONSTRAINT oauth_application_installation_scopes_scope_check
        CHECK (scope = 'stories:write')
);

-- New columns deliberately have no foreign keys. Audit identity must survive
-- deletion of the operational application, installation, secret, or user.
ALTER TABLE public.oauth_audit_events
    ADD COLUMN workspace_id uuid,
    ADD COLUMN installation_id uuid,
    ADD COLUMN principal_id uuid,
    ADD COLUMN secret_id uuid,
    ADD COLUMN actor_kind varchar(32),
    ADD COLUMN actor_id uuid,
    ADD COLUMN actor_credential_id uuid,
    ADD COLUMN request_id varchar(128),
    ADD COLUMN subject_type varchar(32),
    ADD COLUMN subject_id uuid,
    ADD CONSTRAINT oauth_audit_events_actor_kind_check CHECK (
        actor_kind IS NULL OR actor_kind IN (
            'human_user',
            'personal_token',
            'service_account',
            'oauth_user',
            'oauth_application',
            'system',
            'external_contributor'
        )
    ),
    ADD CONSTRAINT oauth_audit_events_actor_shape_check CHECK (
        (
            actor_kind IS NULL
            AND actor_id IS NULL
            AND actor_credential_id IS NULL
        )
        OR
        (
            actor_kind IS NOT NULL
            AND actor_id IS NOT NULL
            AND (actor_kind <> 'human_user' OR actor_credential_id IS NULL)
            AND (actor_kind <> 'oauth_application' OR actor_credential_id IS NOT NULL)
        )
    ),
    ADD CONSTRAINT oauth_audit_events_request_id_check CHECK (
        request_id IS NULL OR (
            char_length(btrim(request_id)) BETWEEN 1 AND 128
            AND request_id !~ '[\r\n]'
        )
    ),
    ADD CONSTRAINT oauth_audit_events_subject_check CHECK (
        (subject_type IS NULL AND subject_id IS NULL)
        OR (
            subject_type IN (
                'application',
                'client_secret',
                'installation',
                'access_token'
            )
            AND subject_id IS NOT NULL
        )
    );

CREATE INDEX oauth_audit_events_workspace_created_idx
    ON public.oauth_audit_events (workspace_id, created_at DESC, event_id DESC)
    WHERE workspace_id IS NOT NULL;

CREATE INDEX oauth_audit_events_installation_created_idx
    ON public.oauth_audit_events (installation_id, created_at DESC, event_id DESC)
    WHERE installation_id IS NOT NULL;

CREATE INDEX oauth_audit_events_subject_created_idx
    ON public.oauth_audit_events (subject_type, subject_id, created_at DESC, event_id DESC)
    WHERE subject_type IS NOT NULL;
