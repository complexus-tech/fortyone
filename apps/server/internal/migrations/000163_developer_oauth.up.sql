CREATE TABLE public.oauth_applications (
    application_id uuid NOT NULL,
    client_id varchar(96) NOT NULL,
    name varchar(120) NOT NULL,
    registration_kind varchar(24) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT oauth_applications_pkey PRIMARY KEY (application_id),
    CONSTRAINT oauth_applications_client_id_key UNIQUE (client_id),
    CONSTRAINT oauth_applications_registration_kind_check
        CHECK (registration_kind IN ('dynamic_public', 'managed_public', 'confidential')),
    CONSTRAINT oauth_applications_status_check CHECK (status IN ('active', 'revoked')),
    CONSTRAINT oauth_applications_client_id_check
        CHECK (char_length(btrim(client_id)) BETWEEN 24 AND 96),
    CONSTRAINT oauth_applications_name_check
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    CONSTRAINT oauth_applications_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT oauth_applications_timestamps_check CHECK (updated_at >= created_at),
    CONSTRAINT oauth_applications_revocation_check CHECK (
        (status = 'active' AND revoked_at IS NULL)
        OR (status = 'revoked' AND revoked_at IS NOT NULL)
    )
);

CREATE INDEX oauth_applications_active_expiry_idx
    ON public.oauth_applications (expires_at, application_id)
    WHERE status = 'active';

CREATE TABLE public.oauth_application_redirect_uris (
    application_id uuid NOT NULL,
    redirect_uri varchar(2048) NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT oauth_application_redirect_uris_pkey
        PRIMARY KEY (application_id, redirect_uri),
    CONSTRAINT oauth_application_redirect_uris_application_id_fkey
        FOREIGN KEY (application_id)
        REFERENCES public.oauth_applications(application_id) ON DELETE CASCADE,
    CONSTRAINT oauth_application_redirect_uris_value_check
        CHECK (char_length(btrim(redirect_uri)) BETWEEN 1 AND 2048)
);

CREATE TABLE public.oauth_grants (
    grant_id uuid NOT NULL,
    application_id uuid NOT NULL,
    user_id uuid NOT NULL,
    actor_kind varchar(32) NOT NULL,
    resource varchar(512) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    last_used_at timestamptz,
    revoked_at timestamptz,
    revoked_reason varchar(240),
    CONSTRAINT oauth_grants_pkey PRIMARY KEY (grant_id),
    CONSTRAINT oauth_grants_application_id_fkey
        FOREIGN KEY (application_id)
        REFERENCES public.oauth_applications(application_id) ON DELETE CASCADE,
    CONSTRAINT oauth_grants_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT oauth_grants_identity_key UNIQUE (application_id, user_id, resource),
    CONSTRAINT oauth_grants_actor_kind_check
        CHECK (actor_kind = 'oauth_user'),
    CONSTRAINT oauth_grants_resource_check
        CHECK (char_length(btrim(resource)) BETWEEN 1 AND 512),
    CONSTRAINT oauth_grants_status_check CHECK (status IN ('active', 'revoked')),
    CONSTRAINT oauth_grants_timestamps_check CHECK (
        updated_at >= created_at
        AND (last_used_at IS NULL OR last_used_at >= created_at)
    ),
    CONSTRAINT oauth_grants_revocation_check CHECK (
        (status = 'active' AND revoked_at IS NULL AND revoked_reason IS NULL)
        OR (
            status = 'revoked'
            AND revoked_at IS NOT NULL
            AND char_length(btrim(revoked_reason)) BETWEEN 1 AND 240
        )
    )
);

CREATE INDEX oauth_grants_user_created_idx
    ON public.oauth_grants (user_id, created_at DESC, grant_id DESC);

CREATE INDEX oauth_grants_application_active_idx
    ON public.oauth_grants (application_id, user_id, grant_id)
    WHERE status = 'active';

CREATE TABLE public.oauth_grant_scopes (
    grant_id uuid NOT NULL,
    scope varchar(64) NOT NULL,
    CONSTRAINT oauth_grant_scopes_pkey PRIMARY KEY (grant_id, scope),
    CONSTRAINT oauth_grant_scopes_grant_id_fkey
        FOREIGN KEY (grant_id) REFERENCES public.oauth_grants(grant_id) ON DELETE CASCADE,
    CONSTRAINT oauth_grant_scopes_scope_check CHECK (
        scope IN (
            'mcp:access',
            'offline_access',
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
            'integrations:manage'
        )
    )
);

CREATE TABLE public.oauth_authorization_codes (
    authorization_code_id uuid NOT NULL,
    application_id uuid NOT NULL,
    grant_id uuid NOT NULL,
    lookup_prefix varchar(12) NOT NULL,
    secret_digest bytea NOT NULL,
    digest_key_id varchar(64) NOT NULL,
    redirect_uri varchar(2048) NOT NULL,
    resource varchar(512) NOT NULL,
    code_challenge varchar(128) NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    CONSTRAINT oauth_authorization_codes_pkey PRIMARY KEY (authorization_code_id),
    CONSTRAINT oauth_authorization_codes_application_id_fkey
        FOREIGN KEY (application_id)
        REFERENCES public.oauth_applications(application_id) ON DELETE CASCADE,
    CONSTRAINT oauth_authorization_codes_grant_id_fkey
        FOREIGN KEY (grant_id) REFERENCES public.oauth_grants(grant_id) ON DELETE CASCADE,
    CONSTRAINT oauth_authorization_codes_lookup_prefix_key UNIQUE (lookup_prefix),
    CONSTRAINT oauth_authorization_codes_lookup_prefix_check
        CHECK (lookup_prefix ~ '^[a-f0-9]{12}$'),
    CONSTRAINT oauth_authorization_codes_secret_digest_check
        CHECK (octet_length(secret_digest) = 32),
    CONSTRAINT oauth_authorization_codes_digest_key_id_check
        CHECK (
            char_length(btrim(digest_key_id)) BETWEEN 1 AND 64
            AND digest_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'
        ),
    CONSTRAINT oauth_authorization_codes_redirect_uri_check
        CHECK (char_length(btrim(redirect_uri)) BETWEEN 1 AND 2048),
    CONSTRAINT oauth_authorization_codes_resource_check
        CHECK (char_length(btrim(resource)) BETWEEN 1 AND 512),
    CONSTRAINT oauth_authorization_codes_code_challenge_check
        CHECK (char_length(code_challenge) BETWEEN 43 AND 128),
    CONSTRAINT oauth_authorization_codes_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT oauth_authorization_codes_consumed_check
        CHECK (consumed_at IS NULL OR consumed_at >= created_at)
);

CREATE INDEX oauth_authorization_codes_expiry_idx
    ON public.oauth_authorization_codes (expires_at, authorization_code_id)
    WHERE consumed_at IS NULL;

CREATE TABLE public.oauth_refresh_token_families (
    family_id uuid NOT NULL,
    grant_id uuid NOT NULL,
    resource varchar(512) NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoked_reason varchar(240),
    CONSTRAINT oauth_refresh_token_families_pkey PRIMARY KEY (family_id),
    CONSTRAINT oauth_refresh_token_families_grant_id_fkey
        FOREIGN KEY (grant_id) REFERENCES public.oauth_grants(grant_id) ON DELETE CASCADE,
    CONSTRAINT oauth_refresh_token_families_resource_check
        CHECK (char_length(btrim(resource)) BETWEEN 1 AND 512),
    CONSTRAINT oauth_refresh_token_families_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT oauth_refresh_token_families_revocation_check CHECK (
        (revoked_at IS NULL AND revoked_reason IS NULL)
        OR (
            revoked_at IS NOT NULL
            AND char_length(btrim(revoked_reason)) BETWEEN 1 AND 240
        )
    )
);

CREATE INDEX oauth_refresh_token_families_grant_active_idx
    ON public.oauth_refresh_token_families (grant_id, expires_at, family_id)
    WHERE revoked_at IS NULL;

CREATE TABLE public.oauth_refresh_tokens (
    refresh_token_id uuid NOT NULL,
    family_id uuid NOT NULL,
    parent_token_id uuid,
    lookup_prefix varchar(12) NOT NULL,
    secret_digest bytea NOT NULL,
    digest_key_id varchar(64) NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    CONSTRAINT oauth_refresh_tokens_pkey PRIMARY KEY (refresh_token_id),
    CONSTRAINT oauth_refresh_tokens_family_id_fkey
        FOREIGN KEY (family_id)
        REFERENCES public.oauth_refresh_token_families(family_id) ON DELETE CASCADE,
    CONSTRAINT oauth_refresh_tokens_parent_token_id_fkey
        FOREIGN KEY (parent_token_id)
        REFERENCES public.oauth_refresh_tokens(refresh_token_id) ON DELETE RESTRICT,
    CONSTRAINT oauth_refresh_tokens_lookup_prefix_key UNIQUE (lookup_prefix),
    CONSTRAINT oauth_refresh_tokens_parent_token_id_key UNIQUE (parent_token_id),
    CONSTRAINT oauth_refresh_tokens_lookup_prefix_check
        CHECK (lookup_prefix ~ '^[a-f0-9]{12}$'),
    CONSTRAINT oauth_refresh_tokens_secret_digest_check
        CHECK (octet_length(secret_digest) = 32),
    CONSTRAINT oauth_refresh_tokens_digest_key_id_check
        CHECK (
            char_length(btrim(digest_key_id)) BETWEEN 1 AND 64
            AND digest_key_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'
        ),
    CONSTRAINT oauth_refresh_tokens_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT oauth_refresh_tokens_parent_check
        CHECK (parent_token_id IS NULL OR parent_token_id <> refresh_token_id),
    CONSTRAINT oauth_refresh_tokens_consumed_check
        CHECK (consumed_at IS NULL OR consumed_at >= created_at)
);

CREATE INDEX oauth_refresh_tokens_family_created_idx
    ON public.oauth_refresh_tokens (family_id, created_at DESC, refresh_token_id DESC);

CREATE INDEX oauth_refresh_tokens_expiry_idx
    ON public.oauth_refresh_tokens (expires_at, refresh_token_id)
    WHERE consumed_at IS NULL;

-- This append-only ledger deliberately retains identity facts after an app,
-- grant, user, code, or refresh-token row is removed.
CREATE TABLE public.oauth_audit_events (
    event_id uuid NOT NULL,
    application_id uuid,
    grant_id uuid,
    user_id uuid,
    operation varchar(80) NOT NULL,
    result varchar(16) NOT NULL,
    reason_code varchar(64),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL,
    CONSTRAINT oauth_audit_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT oauth_audit_events_operation_check
        CHECK (char_length(btrim(operation)) BETWEEN 1 AND 80),
    CONSTRAINT oauth_audit_events_result_check
        CHECK (result IN ('succeeded', 'denied', 'failed')),
    CONSTRAINT oauth_audit_events_metadata_check
        CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX oauth_audit_events_application_created_idx
    ON public.oauth_audit_events (application_id, created_at DESC, event_id DESC);

CREATE INDEX oauth_audit_events_grant_created_idx
    ON public.oauth_audit_events (grant_id, created_at DESC, event_id DESC);

CREATE FUNCTION public.reject_oauth_audit_event_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'OAuth audit events are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER oauth_audit_events_immutable
BEFORE UPDATE OR DELETE ON public.oauth_audit_events
FOR EACH ROW
EXECUTE FUNCTION public.reject_oauth_audit_event_mutation();
