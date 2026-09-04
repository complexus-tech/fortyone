CREATE TABLE public.google_drive_document_import_operations (
    operation_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    source_reference_id uuid NOT NULL,
    document_id uuid NOT NULL UNIQUE,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 1 AND 200),
    request_hash text NOT NULL CHECK (char_length(request_hash) = 64),
    visibility text NOT NULL CHECK (visibility IN ('workspace', 'private')),
    attempt_generation uuid NOT NULL,
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'completed', 'failed')),
    error_code text CHECK (error_code IS NULL OR char_length(error_code) <= 100),
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at timestamptz,
    UNIQUE (workspace_id, user_id, idempotency_key),
    CHECK (
        (status = 'completed' AND completed_at IS NOT NULL AND error_code IS NULL)
        OR (status = 'failed' AND completed_at IS NULL AND error_code IS NOT NULL)
        OR (status = 'pending' AND completed_at IS NULL AND error_code IS NULL)
    )
);

-- The source reference and preallocated document identifiers intentionally do
-- not have foreign keys. An operation must outlive source or document deletion
-- so a retried request cannot reuse the same key to create a duplicate snapshot.
COMMENT ON COLUMN public.google_drive_document_import_operations.source_reference_id IS
    'Immutable source identity retained for idempotent replay after reference deletion.';
COMMENT ON COLUMN public.google_drive_document_import_operations.document_id IS
    'Preallocated result identity; the document and provenance are inserted atomically.';
