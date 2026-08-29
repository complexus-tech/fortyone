-- Shared API idempotency receipts retain only digests of caller-supplied keys
-- and request bytes. Route adoption is deliberately separate: a handler must
-- first coordinate its domain write with this receipt lifecycle.
CREATE TABLE public.api_idempotency_receipts (
    receipt_id uuid NOT NULL,
    principal_kind varchar(32) NOT NULL,
    principal_id uuid NOT NULL,
    workspace_id uuid,
    http_method varchar(16) NOT NULL,
    route_operation varchar(128) NOT NULL,
    key_digest bytea NOT NULL,
    request_hash bytea NOT NULL,
    state varchar(16) NOT NULL,
    lease_generation bigint NOT NULL,
    lease_expires_at timestamptz,
    response_status integer,
    response_body bytea,
    response_content_type varchar(128),
    completed_at timestamptz,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT api_idempotency_receipts_pkey PRIMARY KEY (receipt_id),
    CONSTRAINT api_idempotency_receipts_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT api_idempotency_receipts_principal_kind_check CHECK (
        principal_kind IN (
            'human_user',
            'personal_token',
            'service_account',
            'oauth_user',
            'oauth_application',
            'system',
            'external_contributor'
        )
    ),
    CONSTRAINT api_idempotency_receipts_principal_id_check CHECK (
        principal_id <> '00000000-0000-0000-0000-000000000000'
    ),
    CONSTRAINT api_idempotency_receipts_workspace_id_check CHECK (
        workspace_id IS NULL OR workspace_id <> '00000000-0000-0000-0000-000000000000'
    ),
    CONSTRAINT api_idempotency_receipts_http_method_check CHECK (
        http_method IN ('POST', 'PUT', 'PATCH', 'DELETE')
    ),
    CONSTRAINT api_idempotency_receipts_route_operation_check CHECK (
        route_operation ~ '^[a-z][a-z0-9._:-]{0,127}$'
    ),
    CONSTRAINT api_idempotency_receipts_key_digest_check CHECK (
        octet_length(key_digest) = 32
    ),
    CONSTRAINT api_idempotency_receipts_request_hash_check CHECK (
        octet_length(request_hash) = 32
    ),
    CONSTRAINT api_idempotency_receipts_lease_generation_check CHECK (
        lease_generation > 0
    ),
    CONSTRAINT api_idempotency_receipts_state_check CHECK (
        state IN ('in_progress', 'completed')
    ),
    CONSTRAINT api_idempotency_receipts_response_status_check CHECK (
        response_status IS NULL OR response_status BETWEEN 200 AND 599
    ),
    CONSTRAINT api_idempotency_receipts_response_body_check CHECK (
        response_body IS NULL OR octet_length(response_body) <= 65536
    ),
    CONSTRAINT api_idempotency_receipts_response_content_type_check CHECK (
        response_content_type IS NULL
        OR (
            response_content_type <> ''
            AND response_content_type !~ '[\r\n]'
        )
    ),
    CONSTRAINT api_idempotency_receipts_lifecycle_check CHECK (
        (
            state = 'in_progress'
            AND lease_expires_at IS NOT NULL
            AND response_status IS NULL
            AND response_body IS NULL
            AND response_content_type IS NULL
            AND completed_at IS NULL
        )
        OR
        (
            state = 'completed'
            AND lease_expires_at IS NULL
            AND response_status IS NOT NULL
            AND response_body IS NOT NULL
            AND response_content_type IS NOT NULL
            AND completed_at IS NOT NULL
        )
    ),
    CONSTRAINT api_idempotency_receipts_time_order_check CHECK (
        updated_at >= created_at
        AND expires_at > updated_at
        AND (lease_expires_at IS NULL OR lease_expires_at <= expires_at)
        AND (completed_at IS NULL OR completed_at >= created_at)
    ),
    CONSTRAINT api_idempotency_receipts_scope_key
        UNIQUE NULLS NOT DISTINCT (
            principal_kind,
            principal_id,
            workspace_id,
            http_method,
            route_operation,
            key_digest
        )
);

CREATE INDEX idx_api_idempotency_receipts_expiry
    ON public.api_idempotency_receipts (expires_at, receipt_id);

CREATE INDEX idx_api_idempotency_receipts_stale_lease
    ON public.api_idempotency_receipts (lease_expires_at, receipt_id)
    WHERE state = 'in_progress';

COMMENT ON TABLE public.api_idempotency_receipts IS
    'Bounded, principal-scoped API idempotency state. Raw keys and request payloads are never stored.';
COMMENT ON COLUMN public.api_idempotency_receipts.key_digest IS
    'SHA-256 digest of the exact caller-supplied idempotency key bytes.';
COMMENT ON COLUMN public.api_idempotency_receipts.request_hash IS
    'SHA-256 digest of the exact request bytes used for conflict detection.';
COMMENT ON COLUMN public.api_idempotency_receipts.route_operation IS
    'Stable route operation identifier; it must not contain path parameters or raw URLs.';
COMMENT ON COLUMN public.api_idempotency_receipts.response_body IS
    'Replay-safe response body only. Response headers other than content type are never retained.';
