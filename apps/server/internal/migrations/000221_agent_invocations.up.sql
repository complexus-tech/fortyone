CREATE TABLE public.agent_invocations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    requester_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    team_id uuid NOT NULL REFERENCES public.teams(team_id) ON DELETE CASCADE,
    surface varchar(64) NOT NULL,
    source_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    entity_id uuid NOT NULL,
    parent_run_id uuid REFERENCES public.agent_invocations(id) ON DELETE SET NULL,
    source_revision varchar(256) NOT NULL,
    source_hash varchar(128) NOT NULL,
    idempotency_key varchar(256) NOT NULL,
    request_hash bytea NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}',
    status varchar(32) NOT NULL DEFAULT 'queued',
    summary text NOT NULL DEFAULT '',
    version bigint NOT NULL DEFAULT 1,
    proposal jsonb,
    outcome jsonb,
    reply_id uuid,
    proposal_reply_id uuid,
    proposal_delivered_at timestamptz,
    failure_reply_id uuid,
    failure_pending boolean NOT NULL DEFAULT FALSE,
    lease_token uuid,
    lease_generation bigint NOT NULL DEFAULT 0,
    lease_expires_at timestamptz,
    attempt_count integer NOT NULL DEFAULT 0,
    retry_attempt_count integer NOT NULL DEFAULT 0,
    approved_at timestamptz,
    available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT agent_invocations_key UNIQUE (workspace_id, surface, idempotency_key),
    CONSTRAINT agent_invocations_status CHECK (status IN ('queued','running','needs_input','awaiting_approval','succeeded','failed','cancelled')),
    CONSTRAINT agent_invocations_lease CHECK (
        (status = 'running' AND lease_token IS NOT NULL AND lease_expires_at IS NOT NULL)
        OR (status <> 'running' AND lease_token IS NULL AND lease_expires_at IS NULL)
    ),
    CONSTRAINT agent_invocations_failure_stage CHECK (
        NOT failure_pending OR (status IN ('queued','running') AND outcome IS NULL AND reply_id IS NULL)
    ),
    CONSTRAINT agent_invocations_proposal_delivery CHECK (proposal_delivered_at IS NULL OR proposal_reply_id IS NOT NULL),
    CONSTRAINT agent_invocations_positive_version CHECK (version > 0 AND lease_generation >= 0 AND attempt_count >= 0 AND retry_attempt_count >= 0),
    CONSTRAINT agent_invocations_request_hash CHECK (octet_length(request_hash) = 32)
);

-- A queued row is the transactional outbox. Recovery also discovers dead leases.
CREATE INDEX agent_invocations_pending ON public.agent_invocations (created_at, id)
    WHERE status IN ('queued', 'running');
CREATE INDEX agent_invocations_entity ON public.agent_invocations (workspace_id, entity_id, created_at DESC, id DESC);
CREATE UNIQUE INDEX agent_invocations_one_continuation ON public.agent_invocations (parent_run_id)
    WHERE parent_run_id IS NOT NULL;
CREATE INDEX agent_invocations_source ON public.agent_invocations (workspace_id, surface, source_id);
CREATE INDEX agent_invocations_clarification ON public.agent_invocations (workspace_id, surface, entity_id, conversation_id)
    WHERE status IN ('needs_input','awaiting_approval');
