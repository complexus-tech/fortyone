-- Expand workspace invitation storage so new bearer tokens are persisted only
-- as versioned HMAC-SHA-256 digests. Existing plaintext rows remain readable
-- during the bounded compatibility window; the application never creates a
-- new plaintext row after this migration is deployed.
ALTER TABLE public.workspace_invitations
    ALTER COLUMN token DROP NOT NULL,
    ADD COLUMN token_digest bytea,
    ADD COLUMN token_nonce bytea,
    ADD COLUMN token_key_id varchar(64),
    ADD COLUMN token_version smallint;

ALTER TABLE public.workspace_invitations
    ADD CONSTRAINT workspace_invitations_token_storage_shape_check CHECK (
        (
            token IS NOT NULL
            AND token_digest IS NULL
            AND token_nonce IS NULL
            AND token_key_id IS NULL
            AND token_version IS NULL
        )
        OR
        (
            token IS NULL
            AND octet_length(token_digest) = 32
            AND octet_length(token_nonce) = 32
            AND token_key_id IS NOT NULL
            AND token_key_id <> ''
            AND token_version > 0
        )
    );

CREATE UNIQUE INDEX workspace_invitations_token_digest_key
    ON public.workspace_invitations (token_key_id, token_version, token_digest)
    WHERE token_digest IS NOT NULL;

COMMENT ON COLUMN public.workspace_invitations.token IS
    'Legacy plaintext bearer token. New application writes leave this NULL; remove after the compatibility window.';
COMMENT ON COLUMN public.workspace_invitations.token_digest IS
    'HMAC-SHA-256 digest of the one-time bearer token. Raw token material is never persisted for new invitations.';
COMMENT ON COLUMN public.workspace_invitations.token_nonce IS
    'Non-secret random input used by the email dispatcher to reconstruct a signed one-time token at the delivery boundary.';
COMMENT ON COLUMN public.workspace_invitations.token_key_id IS
    'Non-secret identifier of the HMAC key generation used for token_digest.';
COMMENT ON COLUMN public.workspace_invitations.token_version IS
    'Invitation token format and digest protocol version.';

-- Invitation creation and acceptance write their delivery record in the same
-- transaction as domain state. Payloads deliberately exclude raw bearer
-- tokens; the worker reconstructs an email token from safe invitation metadata
-- only at the delivery boundary.
CREATE TABLE public.workspace_invitation_outbox (
    outbox_id uuid NOT NULL DEFAULT gen_random_uuid(),
    invitation_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    event_type varchar(64) NOT NULL,
    event_payload jsonb NOT NULL,
    idempotency_key varchar(160) NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    claim_token uuid,
    claimed_at timestamptz,
    completed_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT workspace_invitation_outbox_pkey PRIMARY KEY (outbox_id),
    CONSTRAINT workspace_invitation_outbox_invitation_id_fkey
        FOREIGN KEY (invitation_id) REFERENCES public.workspace_invitations(invitation_id) ON DELETE CASCADE,
    CONSTRAINT workspace_invitation_outbox_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT workspace_invitation_outbox_actor_id_fkey
        FOREIGN KEY (actor_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT workspace_invitation_outbox_event_type_check
        CHECK (event_type IN ('invitation.email', 'invitation.accepted')),
    CONSTRAINT workspace_invitation_outbox_payload_check
        CHECK (jsonb_typeof(event_payload) = 'object' AND NOT (event_payload ? 'token')),
    CONSTRAINT workspace_invitation_outbox_idempotency_key_key UNIQUE (idempotency_key),
    CONSTRAINT workspace_invitation_outbox_status_check
        CHECK (status IN ('pending', 'processing', 'retrying', 'completed', 'failed', 'cancelled')),
    CONSTRAINT workspace_invitation_outbox_attempt_count_check CHECK (attempt_count >= 0),
    CONSTRAINT workspace_invitation_outbox_lifecycle_check CHECK (
        (
            status IN ('pending', 'retrying')
            AND next_attempt_at IS NOT NULL
            AND claim_token IS NULL
            AND claimed_at IS NULL
            AND completed_at IS NULL
        )
        OR
        (
            status = 'processing'
            AND next_attempt_at IS NULL
            AND claim_token IS NOT NULL
            AND claimed_at IS NOT NULL
            AND completed_at IS NULL
        )
        OR
        (
            status IN ('completed', 'cancelled')
            AND next_attempt_at IS NULL
            AND claim_token IS NULL
            AND claimed_at IS NULL
            AND completed_at IS NOT NULL
        )
        OR
        (
            status = 'failed'
            AND next_attempt_at IS NULL
            AND claim_token IS NULL
            AND claimed_at IS NULL
            AND completed_at IS NULL
        )
    )
);

CREATE INDEX idx_workspace_invitation_outbox_ready
    ON public.workspace_invitation_outbox (next_attempt_at, created_at, outbox_id)
    WHERE status IN ('pending', 'retrying');

CREATE INDEX idx_workspace_invitation_outbox_stale_claim
    ON public.workspace_invitation_outbox (claimed_at, outbox_id)
    WHERE status = 'processing';

CREATE INDEX idx_workspace_invitation_outbox_retention
    ON public.workspace_invitation_outbox (completed_at, outbox_id)
    WHERE status IN ('completed', 'cancelled');
