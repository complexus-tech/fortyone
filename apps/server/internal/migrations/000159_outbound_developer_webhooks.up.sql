CREATE TABLE public.outbound_webhook_endpoints (
    endpoint_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    owner_principal_id uuid NOT NULL,
    name varchar(120) NOT NULL,
    endpoint_url text NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    signing_secret_envelope text NOT NULL,
    secret_generation integer NOT NULL DEFAULT 1,
    previous_secret_envelope text,
    previous_secret_generation integer,
    previous_secret_expires_at timestamptz,
    subscription_generation integer NOT NULL DEFAULT 1,
    consecutive_failures integer NOT NULL DEFAULT 0,
    last_delivery_claimed_at timestamptz,
    last_success_at timestamptz,
    disabled_at timestamptz,
    disabled_reason varchar(240),
    created_by_user_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT outbound_webhook_endpoints_pkey PRIMARY KEY (endpoint_id),
    CONSTRAINT outbound_webhook_endpoints_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_endpoints_owner_principal_fkey
        FOREIGN KEY (owner_principal_id, workspace_id)
        REFERENCES public.principals(principal_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_endpoints_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT outbound_webhook_endpoints_workspace_identity_key
        UNIQUE (endpoint_id, workspace_id),
    CONSTRAINT outbound_webhook_endpoints_name_check
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    CONSTRAINT outbound_webhook_endpoints_url_check CHECK (
        char_length(endpoint_url) BETWEEN 12 AND 2048
        AND endpoint_url = btrim(endpoint_url)
        AND endpoint_url LIKE 'https://%'
    ),
    CONSTRAINT outbound_webhook_endpoints_status_check
        CHECK (status IN ('active', 'disabled')),
    CONSTRAINT outbound_webhook_endpoints_secret_check CHECK (
        char_length(signing_secret_envelope) BETWEEN 32 AND 262144
        AND secret_generation > 0
    ),
    CONSTRAINT outbound_webhook_endpoints_previous_secret_check CHECK (
        (
            previous_secret_envelope IS NULL
            AND previous_secret_generation IS NULL
            AND previous_secret_expires_at IS NULL
        )
        OR
        (
            previous_secret_envelope IS NOT NULL
            AND previous_secret_generation IS NOT NULL
            AND previous_secret_generation > 0
            AND previous_secret_generation < secret_generation
            AND previous_secret_expires_at IS NOT NULL
        )
    ),
    CONSTRAINT outbound_webhook_endpoints_generation_check
        CHECK (subscription_generation > 0),
    CONSTRAINT outbound_webhook_endpoints_failure_count_check
        CHECK (consecutive_failures >= 0),
    CONSTRAINT outbound_webhook_endpoints_disabled_state_check CHECK (
        (
            status = 'active'
            AND disabled_at IS NULL
            AND disabled_reason IS NULL
        )
        OR
        (
            status = 'disabled'
            AND disabled_at IS NOT NULL
            AND disabled_reason IS NOT NULL
        )
    ),
    CONSTRAINT outbound_webhook_endpoints_timestamps_check
        CHECK (updated_at >= created_at)
);

CREATE INDEX outbound_webhook_endpoints_workspace_created_idx
    ON public.outbound_webhook_endpoints (workspace_id, created_at DESC, endpoint_id DESC);

CREATE INDEX outbound_webhook_endpoints_dispatch_idx
    ON public.outbound_webhook_endpoints (
        COALESCE(last_delivery_claimed_at, 'epoch'::timestamptz),
        endpoint_id
    )
    WHERE status = 'active';

CREATE TABLE public.outbound_webhook_subscriptions (
    endpoint_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    event_type varchar(80) NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT outbound_webhook_subscriptions_pkey PRIMARY KEY (endpoint_id, event_type),
    CONSTRAINT outbound_webhook_subscriptions_endpoint_workspace_fkey
        FOREIGN KEY (endpoint_id, workspace_id)
        REFERENCES public.outbound_webhook_endpoints(endpoint_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_subscriptions_event_type_check CHECK (
        event_type IN (
            'story.created',
            'story.updated',
            'story.deleted',
            'comment.created',
            'comment.updated',
            'comment.deleted'
        )
    )
);

CREATE INDEX outbound_webhook_subscriptions_workspace_event_idx
    ON public.outbound_webhook_subscriptions (workspace_id, event_type, endpoint_id);

CREATE TABLE public.outbound_webhook_events (
    event_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    event_type varchar(80) NOT NULL,
    payload_version integer NOT NULL DEFAULT 1,
    subject_type varchar(48) NOT NULL,
    subject_id uuid NOT NULL,
    actor_kind varchar(32) NOT NULL,
    actor_id uuid NOT NULL,
    actor_credential_id uuid,
    payload jsonb NOT NULL,
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT outbound_webhook_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT outbound_webhook_events_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_events_workspace_identity_key UNIQUE (event_id, workspace_id),
    CONSTRAINT outbound_webhook_events_event_type_check CHECK (
        event_type IN (
            'story.created',
            'story.updated',
            'story.deleted',
            'comment.created',
            'comment.updated',
            'comment.deleted'
        )
    ),
    CONSTRAINT outbound_webhook_events_payload_version_check
        CHECK (payload_version = 1),
    CONSTRAINT outbound_webhook_events_subject_type_check
        CHECK (subject_type IN ('story', 'comment')),
    CONSTRAINT outbound_webhook_events_actor_kind_check CHECK (
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
    CONSTRAINT outbound_webhook_events_payload_check
        CHECK (jsonb_typeof(payload) = 'object'),
    CONSTRAINT outbound_webhook_events_timestamps_check
        CHECK (created_at >= occurred_at)
);

CREATE INDEX outbound_webhook_events_workspace_created_idx
    ON public.outbound_webhook_events (workspace_id, created_at DESC, event_id DESC);

CREATE TABLE public.outbound_webhook_deliveries (
    delivery_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    event_id uuid NOT NULL,
    endpoint_id uuid NOT NULL,
    subscription_generation integer NOT NULL,
    payload_body bytea NOT NULL,
    status varchar(24) NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL,
    lease_token uuid,
    lease_expires_at timestamptz,
    last_http_status integer,
    last_error_code varchar(64),
    completed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT outbound_webhook_deliveries_pkey PRIMARY KEY (delivery_id),
    CONSTRAINT outbound_webhook_deliveries_event_workspace_fkey
        FOREIGN KEY (event_id, workspace_id)
        REFERENCES public.outbound_webhook_events(event_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_deliveries_endpoint_workspace_fkey
        FOREIGN KEY (endpoint_id, workspace_id)
        REFERENCES public.outbound_webhook_endpoints(endpoint_id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_deliveries_event_endpoint_key UNIQUE (event_id, endpoint_id),
    CONSTRAINT outbound_webhook_deliveries_generation_check
        CHECK (subscription_generation > 0),
    CONSTRAINT outbound_webhook_deliveries_payload_check
        CHECK (octet_length(payload_body) BETWEEN 2 AND 262144),
    CONSTRAINT outbound_webhook_deliveries_status_check CHECK (
        status IN (
            'pending',
            'delivering',
            'retry_scheduled',
            'succeeded',
            'failed',
            'cancelled'
        )
    ),
    CONSTRAINT outbound_webhook_deliveries_attempt_count_check
        CHECK (attempt_count BETWEEN 0 AND 32),
    CONSTRAINT outbound_webhook_deliveries_lease_check CHECK (
        (
            status = 'delivering'
            AND lease_token IS NOT NULL
            AND lease_expires_at IS NOT NULL
            AND completed_at IS NULL
        )
        OR
        (
            status <> 'delivering'
            AND lease_token IS NULL
            AND lease_expires_at IS NULL
        )
    ),
    CONSTRAINT outbound_webhook_deliveries_terminal_check CHECK (
        (status IN ('succeeded', 'failed', 'cancelled') AND completed_at IS NOT NULL)
        OR
        (status NOT IN ('succeeded', 'failed', 'cancelled') AND completed_at IS NULL)
    ),
    CONSTRAINT outbound_webhook_deliveries_http_status_check
        CHECK (last_http_status IS NULL OR last_http_status BETWEEN 100 AND 599),
    CONSTRAINT outbound_webhook_deliveries_timestamps_check
        CHECK (updated_at >= created_at AND available_at >= created_at)
);

CREATE INDEX outbound_webhook_deliveries_dispatch_idx
    ON public.outbound_webhook_deliveries (available_at, created_at, delivery_id)
    WHERE status IN ('pending', 'retry_scheduled');

CREATE INDEX outbound_webhook_deliveries_endpoint_created_idx
    ON public.outbound_webhook_deliveries (endpoint_id, created_at DESC, delivery_id DESC);

CREATE TABLE public.outbound_webhook_delivery_attempts (
    attempt_id uuid NOT NULL,
    delivery_id uuid NOT NULL,
    attempt_number integer NOT NULL,
    outcome varchar(24) NOT NULL,
    resolved_ip inet,
    http_status integer,
    response_bytes integer,
    response_digest bytea,
    error_code varchar(64),
    duration_ms integer NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    CONSTRAINT outbound_webhook_delivery_attempts_pkey PRIMARY KEY (attempt_id),
    CONSTRAINT outbound_webhook_delivery_attempts_delivery_id_fkey
        FOREIGN KEY (delivery_id) REFERENCES public.outbound_webhook_deliveries(delivery_id) ON DELETE CASCADE,
    CONSTRAINT outbound_webhook_delivery_attempts_delivery_number_key
        UNIQUE (delivery_id, attempt_number),
    CONSTRAINT outbound_webhook_delivery_attempts_number_check
        CHECK (attempt_number BETWEEN 1 AND 32),
    CONSTRAINT outbound_webhook_delivery_attempts_outcome_check
        CHECK (outcome IN ('succeeded', 'retry_scheduled', 'failed', 'cancelled')),
    CONSTRAINT outbound_webhook_delivery_attempts_http_status_check
        CHECK (http_status IS NULL OR http_status BETWEEN 100 AND 599),
    CONSTRAINT outbound_webhook_delivery_attempts_response_check CHECK (
        response_bytes IS NULL OR response_bytes BETWEEN 0 AND 65536
    ),
    CONSTRAINT outbound_webhook_delivery_attempts_digest_check CHECK (
        response_digest IS NULL OR octet_length(response_digest) = 32
    ),
    CONSTRAINT outbound_webhook_delivery_attempts_duration_check
        CHECK (duration_ms BETWEEN 0 AND 30000),
    CONSTRAINT outbound_webhook_delivery_attempts_timestamps_check
        CHECK (finished_at >= started_at)
);

CREATE INDEX outbound_webhook_delivery_attempts_delivery_created_idx
    ON public.outbound_webhook_delivery_attempts (delivery_id, attempt_number DESC);

CREATE TABLE public.outbound_webhook_audit_events (
    audit_event_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    actor_kind varchar(32) NOT NULL,
    actor_id uuid NOT NULL,
    actor_credential_id uuid,
    operation varchar(80) NOT NULL,
    endpoint_id uuid,
    delivery_id uuid,
    result varchar(16) NOT NULL,
    reason_code varchar(64),
    request_id varchar(128),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL,
    CONSTRAINT outbound_webhook_audit_events_pkey PRIMARY KEY (audit_event_id),
    CONSTRAINT outbound_webhook_audit_events_actor_kind_check CHECK (
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
    CONSTRAINT outbound_webhook_audit_events_operation_check
        CHECK (char_length(btrim(operation)) BETWEEN 1 AND 80),
    CONSTRAINT outbound_webhook_audit_events_result_check
        CHECK (result IN ('succeeded', 'denied', 'failed')),
    CONSTRAINT outbound_webhook_audit_events_metadata_check
        CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX outbound_webhook_audit_events_workspace_created_idx
    ON public.outbound_webhook_audit_events (workspace_id, created_at DESC, audit_event_id DESC);

CREATE FUNCTION public.reject_outbound_webhook_immutable_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'outbound webhook delivery attempts and audit events are immutable'
        USING ERRCODE = '55000';
END;
$$;

CREATE TRIGGER outbound_webhook_delivery_attempts_immutable
BEFORE UPDATE OR DELETE ON public.outbound_webhook_delivery_attempts
FOR EACH ROW
EXECUTE FUNCTION public.reject_outbound_webhook_immutable_mutation();

CREATE TRIGGER outbound_webhook_audit_events_immutable
BEFORE UPDATE OR DELETE ON public.outbound_webhook_audit_events
FOR EACH ROW
EXECUTE FUNCTION public.reject_outbound_webhook_immutable_mutation();
