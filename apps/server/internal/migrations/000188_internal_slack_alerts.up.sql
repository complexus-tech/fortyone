-- Internal operations alerts are independent of customer notification settings.
-- Only new application signups and collected invoice payments populate this
-- outbox. Existing accounts and invoices are deliberately not backfilled.
CREATE TABLE public.internal_slack_alerts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dedupe_key text NOT NULL UNIQUE,
    kind text NOT NULL CHECK (kind IN ('account_created', 'payment_received')),
    user_id uuid REFERENCES public.users(user_id) ON DELETE CASCADE,
    workspace_id uuid REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    available_at timestamptz NOT NULL DEFAULT now(),
    slack_team_id text,
    slack_channel_id text,
    lease_token uuid,
    lease_until timestamptz,
    attempts integer NOT NULL DEFAULT 0,
    delivered_at timestamptz,
    slack_message_ts text,
    CHECK ((kind = 'account_created' AND user_id IS NOT NULL AND workspace_id IS NULL)
        OR (kind = 'payment_received' AND workspace_id IS NOT NULL AND user_id IS NULL)),
    CHECK ((slack_team_id IS NULL) = (slack_channel_id IS NULL))
);

CREATE INDEX internal_slack_alerts_pending_idx
    ON public.internal_slack_alerts (available_at, created_at, id)
    WHERE delivered_at IS NULL;
