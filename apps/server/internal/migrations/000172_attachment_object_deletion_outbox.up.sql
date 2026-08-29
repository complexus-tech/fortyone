-- Durable deletion metadata bridges the PostgreSQL transaction that removes
-- the final attachment reference and the non-transactional object store. The
-- row deliberately survives deletion of the attachment and workspace, so it
-- has no foreign keys and stores only routing metadata, never credentials.
CREATE TABLE public.attachment_object_deletion_outbox (
    outbox_id uuid NOT NULL DEFAULT gen_random_uuid(),
    attachment_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    storage_provider varchar(32) NOT NULL,
    container_name varchar(255) NOT NULL,
    blob_name varchar(1024) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processing_started_at timestamptz,
    claim_token uuid,
    completed_at timestamptz,
    last_error varchar(255),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT attachment_object_deletion_outbox_pkey PRIMARY KEY (outbox_id),
    CONSTRAINT attachment_object_deletion_outbox_attachment_id_key UNIQUE (attachment_id),
    CONSTRAINT attachment_object_deletion_outbox_provider_check
        CHECK (NULLIF(BTRIM(storage_provider), '') IS NOT NULL),
    CONSTRAINT attachment_object_deletion_outbox_container_check
        CHECK (NULLIF(BTRIM(container_name), '') IS NOT NULL),
    CONSTRAINT attachment_object_deletion_outbox_blob_check
        CHECK (NULLIF(BTRIM(blob_name), '') IS NOT NULL),
    CONSTRAINT attachment_object_deletion_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'retrying', 'completed')),
    CONSTRAINT attachment_object_deletion_outbox_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT attachment_object_deletion_outbox_lifecycle_check CHECK (
        (
            status = 'pending'
            AND attempt_count = 0
            AND next_attempt_at IS NOT NULL
            AND processing_started_at IS NULL
            AND claim_token IS NULL
            AND completed_at IS NULL
            AND last_error IS NULL
        )
        OR
        (
            status = 'processing'
            AND attempt_count > 0
            AND next_attempt_at IS NULL
            AND processing_started_at IS NOT NULL
            AND claim_token IS NOT NULL
            AND completed_at IS NULL
            AND last_error IS NULL
        )
        OR
        (
            status = 'retrying'
            AND attempt_count > 0
            AND next_attempt_at IS NOT NULL
            AND processing_started_at IS NULL
            AND claim_token IS NULL
            AND completed_at IS NULL
            AND NULLIF(BTRIM(last_error), '') IS NOT NULL
        )
        OR
        (
            status = 'completed'
            AND attempt_count > 0
            AND next_attempt_at IS NULL
            AND processing_started_at IS NULL
            AND claim_token IS NULL
            AND completed_at IS NOT NULL
            AND last_error IS NULL
        )
    )
);

CREATE INDEX idx_attachment_object_deletion_outbox_due
    ON public.attachment_object_deletion_outbox (next_attempt_at, created_at, outbox_id)
    WHERE status IN ('pending', 'retrying');

CREATE INDEX idx_attachment_object_deletion_outbox_lease
    ON public.attachment_object_deletion_outbox (processing_started_at, outbox_id)
    WHERE status = 'processing';

CREATE INDEX idx_attachment_object_deletion_outbox_completed
    ON public.attachment_object_deletion_outbox (completed_at, outbox_id)
    WHERE status = 'completed';

COMMENT ON TABLE public.attachment_object_deletion_outbox IS
    'Credential-free durable work for deleting attachment objects after their final database reference is removed.';
COMMENT ON COLUMN public.attachment_object_deletion_outbox.storage_provider IS
    'Non-secret storage adapter identifier captured when the attachment row is retired.';
COMMENT ON COLUMN public.attachment_object_deletion_outbox.container_name IS
    'Non-secret bucket or container name; credentials remain exclusively in runtime configuration.';
COMMENT ON COLUMN public.attachment_object_deletion_outbox.blob_name IS
    'Opaque object key. Application logs, traces, metrics, and errors must not emit this value.';
