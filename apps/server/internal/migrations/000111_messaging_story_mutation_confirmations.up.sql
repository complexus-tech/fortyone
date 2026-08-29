CREATE TABLE public.messaging_story_mutation_confirmations (
    confirmation_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    team_id uuid NOT NULL,
    operation text NOT NULL,
    token_hash bytea NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    result jsonb,
    last_error text,
    expires_at timestamptz NOT NULL,
    applied_at timestamptz,
    cancelled_at timestamptz,
    expired_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_story_mutation_confirmations_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_story_mutation_confirmations_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT messaging_story_mutation_confirmations_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT messaging_story_mutation_confirmations_operation_check
        CHECK (operation IN ('create_story', 'update_story')),
    CONSTRAINT messaging_story_mutation_confirmations_status_check
        CHECK (status IN ('pending', 'applied', 'cancelled', 'expired')),
    CONSTRAINT messaging_story_mutation_confirmations_token_hash_check
        CHECK (octet_length(token_hash) = 32),
    CONSTRAINT messaging_story_mutation_confirmations_lifecycle_check
        CHECK (
            (status = 'pending' AND applied_at IS NULL AND cancelled_at IS NULL AND expired_at IS NULL)
            OR (status = 'applied' AND applied_at IS NOT NULL AND cancelled_at IS NULL AND expired_at IS NULL)
            OR (status = 'cancelled' AND applied_at IS NULL AND cancelled_at IS NOT NULL AND expired_at IS NULL)
            OR (status = 'expired' AND applied_at IS NULL AND cancelled_at IS NULL AND expired_at IS NOT NULL)
        ),
    PRIMARY KEY (confirmation_id)
);

CREATE INDEX messaging_story_mutation_confirmations_pending_expiry_idx
    ON public.messaging_story_mutation_confirmations USING btree (expires_at)
    WHERE status = 'pending';

CREATE INDEX messaging_story_mutation_confirmations_actor_idx
    ON public.messaging_story_mutation_confirmations USING btree (workspace_id, user_id, created_at DESC);
