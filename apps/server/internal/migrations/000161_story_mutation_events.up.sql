-- Story writes persist their integration event intent in the same transaction
-- as the business mutation. A separate dispatcher may safely retry publication
-- without replaying the story write or manufacturing a new event identity.
CREATE TABLE public.story_mutation_events (
    event_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    story_id uuid NOT NULL,
    event_type text NOT NULL,
    actor_kind text NOT NULL,
    actor_id uuid NOT NULL,
    actor_credential_id uuid,
    payload jsonb NOT NULL,
    occurred_at timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL,
    claim_token uuid,
    claimed_at timestamptz,
    completed_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT story_mutation_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT story_mutation_events_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT story_mutation_events_type_check
        CHECK (event_type IN ('story.created', 'story.updated', 'story.deleted')),
    CONSTRAINT story_mutation_events_actor_kind_check
        CHECK (actor_kind IN (
            'human_user',
            'personal_token',
            'service_account',
            'oauth_user',
            'oauth_application',
            'system'
        )),
    CONSTRAINT story_mutation_events_payload_check
        CHECK (jsonb_typeof(payload) = 'object'),
    CONSTRAINT story_mutation_events_status_check
        CHECK (status IN ('pending', 'processing', 'completed')),
    CONSTRAINT story_mutation_events_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT story_mutation_events_claim_check
        CHECK (
            (status = 'processing' AND claim_token IS NOT NULL AND claimed_at IS NOT NULL)
            OR (status <> 'processing' AND claim_token IS NULL AND claimed_at IS NULL)
        ),
    CONSTRAINT story_mutation_events_completion_check
        CHECK (
            (status = 'completed' AND completed_at IS NOT NULL)
            OR (status <> 'completed' AND completed_at IS NULL)
        ),
    CONSTRAINT story_mutation_events_timestamps_check
        CHECK (
            occurred_at <= created_at
            AND updated_at >= created_at
            AND (claimed_at IS NULL OR claimed_at >= created_at)
            AND (completed_at IS NULL OR completed_at >= created_at)
        ),
    CONSTRAINT story_mutation_events_last_error_check
        CHECK (last_error IS NULL OR char_length(last_error) <= 1000)
);

CREATE UNIQUE INDEX story_mutation_events_claim_token_key
    ON public.story_mutation_events (claim_token)
    WHERE claim_token IS NOT NULL;

CREATE INDEX idx_story_mutation_events_ready
    ON public.story_mutation_events (next_attempt_at, created_at, event_id)
    WHERE status = 'pending';

CREATE INDEX idx_story_mutation_events_stale_claim
    ON public.story_mutation_events (claimed_at, event_id)
    WHERE status = 'processing';

CREATE INDEX idx_story_mutation_events_story
    ON public.story_mutation_events (workspace_id, story_id, occurred_at DESC, event_id);

CREATE INDEX idx_story_mutation_events_retention
    ON public.story_mutation_events (completed_at, event_id)
    WHERE status = 'completed';

CREATE FUNCTION public.reject_story_mutation_event_identity_change()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF ROW(
        NEW.event_id,
        NEW.workspace_id,
        NEW.story_id,
        NEW.event_type,
        NEW.actor_kind,
        NEW.actor_id,
        NEW.actor_credential_id,
        NEW.payload,
        NEW.occurred_at,
        NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.event_id,
        OLD.workspace_id,
        OLD.story_id,
        OLD.event_type,
        OLD.actor_kind,
        OLD.actor_id,
        OLD.actor_credential_id,
        OLD.payload,
        OLD.occurred_at,
        OLD.created_at
    ) THEN
        RAISE EXCEPTION 'story mutation event identity is immutable'
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER story_mutation_events_identity_immutable
BEFORE UPDATE ON public.story_mutation_events
FOR EACH ROW
EXECUTE FUNCTION public.reject_story_mutation_event_identity_change();
