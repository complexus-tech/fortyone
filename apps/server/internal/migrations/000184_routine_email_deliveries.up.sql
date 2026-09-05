CREATE TABLE routine_email_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id uuid NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    delivery_key text NOT NULL,
    kind text NOT NULL CHECK (kind IN ('briefing', 'activity')),
    local_date date NOT NULL,
    status text NOT NULL CHECK (status IN ('processing', 'sent', 'skipped', 'failed')),
    claimed_at timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (recipient_id, workspace_id, delivery_key)
);
CREATE INDEX routine_email_recipient_claims
    ON routine_email_deliveries (recipient_id, status, claimed_at);
