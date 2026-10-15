-- A receipt and its domain mutation commit together. Receipts intentionally
-- have no TTL: a laptop may replay an operation after a long offline interval.
CREATE TABLE desktop_operation_receipts (
    actor_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    operation_id UUID NOT NULL,
    request_hash BYTEA NOT NULL CHECK (octet_length(request_hash) = 32),
    resource_id UUID NOT NULL,
    operation_version TEXT CHECK (operation_version IS NULL OR length(operation_version) <= 100),
    completed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (actor_id, workspace_id, operation_id)
);
COMMENT ON TABLE desktop_operation_receipts IS 'Permanent deduplication identity for committed desktop operations; removed only with their account/workspace.';
