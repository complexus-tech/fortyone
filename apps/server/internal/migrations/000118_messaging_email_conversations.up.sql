-- Email replies use the existing durable delivery outbox. Preserve every
-- purpose introduced by earlier migrations when extending its constraint.
ALTER TABLE public.messaging_outbound_deliveries
    DROP CONSTRAINT messaging_outbound_deliveries_purpose_check,
    ADD CONSTRAINT messaging_outbound_deliveries_purpose_check
    CHECK (
        purpose IN (
            'provider_message',
            'assistant',
            'account_link',
            'access',
            'creation_confirmation',
            'onboarding',
            'email_reply'
        )
    );

CREATE TABLE public.messaging_email_threads (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    recipient_email text NOT NULL,
    external_thread_id text NOT NULL,
    root_internet_message_id text NOT NULL DEFAULT '',
    latest_internet_message_id text NOT NULL DEFAULT '',
    context jsonb NOT NULL DEFAULT '{}'::jsonb,
    summary text NOT NULL DEFAULT '',
    summary_through_sequence bigint NOT NULL DEFAULT 0,
    next_message_sequence bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_email_threads_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_email_threads_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT messaging_email_threads_identity_unique
        UNIQUE (id, workspace_id, user_id),
    CONSTRAINT messaging_email_threads_provider_check
        CHECK (btrim(provider) <> '' AND char_length(provider) <= 64),
    CONSTRAINT messaging_email_threads_recipient_email_check
        CHECK (btrim(recipient_email) <> '' AND char_length(recipient_email) <= 320),
    CONSTRAINT messaging_email_threads_external_thread_check
        CHECK (btrim(external_thread_id) <> '' AND char_length(external_thread_id) <= 512),
    CONSTRAINT messaging_email_threads_message_ids_check
        CHECK (
            char_length(root_internet_message_id) <= 998
            AND char_length(latest_internet_message_id) <= 998
        ),
    CONSTRAINT messaging_email_threads_context_check
        CHECK (jsonb_typeof(context) = 'object'),
    CONSTRAINT messaging_email_threads_summary_cursor_check
        CHECK (
            summary_through_sequence >= 0
            AND summary_through_sequence < next_message_sequence
            AND next_message_sequence > 0
        ),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_email_threads_external_key
    ON public.messaging_email_threads USING btree (
        provider, workspace_id, external_thread_id
    );

CREATE INDEX messaging_email_threads_actor_updated_idx
    ON public.messaging_email_threads USING btree (
        workspace_id, user_id, updated_at DESC
    );

CREATE TABLE public.messaging_email_reply_tokens (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    thread_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_email_reply_tokens_thread_fkey
        FOREIGN KEY (thread_id, workspace_id, user_id)
        REFERENCES public.messaging_email_threads(id, workspace_id, user_id)
        ON DELETE CASCADE,
    CONSTRAINT messaging_email_reply_tokens_hash_check
        CHECK (octet_length(token_hash) = 32),
    CONSTRAINT messaging_email_reply_tokens_lifecycle_check
        CHECK (
            expires_at > created_at
            AND (revoked_at IS NULL OR revoked_at >= created_at)
        ),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_email_reply_tokens_token_hash_key
    ON public.messaging_email_reply_tokens USING btree (token_hash);

CREATE INDEX messaging_email_reply_tokens_active_expiry_idx
    ON public.messaging_email_reply_tokens USING btree (expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE public.messaging_email_messages (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    thread_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    sequence bigint NOT NULL,
    inbound_event_id uuid,
    idempotency_key text NOT NULL,
    direction text NOT NULL,
    role text NOT NULL,
    kind text NOT NULL,
    provider_message_id text,
    internet_message_id text,
    in_reply_to_message_id text,
    subject text NOT NULL DEFAULT '',
    content text NOT NULL,
    context jsonb NOT NULL DEFAULT '{}'::jsonb,
    provider_metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_email_messages_thread_fkey
        FOREIGN KEY (thread_id, workspace_id, user_id)
        REFERENCES public.messaging_email_threads(id, workspace_id, user_id)
        ON DELETE CASCADE,
    CONSTRAINT messaging_email_messages_inbound_event_id_fkey
        FOREIGN KEY (inbound_event_id) REFERENCES public.messaging_inbound_events(id) ON DELETE SET NULL,
    CONSTRAINT messaging_email_messages_thread_sequence_unique
        UNIQUE (thread_id, sequence),
    CONSTRAINT messaging_email_messages_thread_identity_unique
        UNIQUE (thread_id, id),
    CONSTRAINT messaging_email_messages_sequence_check
        CHECK (sequence > 0),
    CONSTRAINT messaging_email_messages_idempotency_check
        CHECK (btrim(idempotency_key) <> '' AND char_length(idempotency_key) <= 512),
    CONSTRAINT messaging_email_messages_direction_check
        CHECK (direction IN ('inbound', 'outbound')),
    CONSTRAINT messaging_email_messages_role_check
        CHECK (role IN ('user', 'assistant', 'system')),
    CONSTRAINT messaging_email_messages_kind_check
        CHECK (kind IN ('guidance', 'reply', 'answer', 'proposal', 'confirmation', 'receipt', 'error')),
    CONSTRAINT messaging_email_messages_provider_message_id_check
        CHECK (provider_message_id IS NULL OR char_length(provider_message_id) <= 998),
    CONSTRAINT messaging_email_messages_rfc_message_ids_check
        CHECK (
            (internet_message_id IS NULL OR char_length(internet_message_id) <= 998)
            AND (in_reply_to_message_id IS NULL OR char_length(in_reply_to_message_id) <= 998)
        ),
    CONSTRAINT messaging_email_messages_subject_check
        CHECK (char_length(subject) <= 998),
    CONSTRAINT messaging_email_messages_content_check
        CHECK (btrim(content) <> ''),
    CONSTRAINT messaging_email_messages_context_check
        CHECK (jsonb_typeof(context) = 'object'),
    CONSTRAINT messaging_email_messages_provider_metadata_check
        CHECK (jsonb_typeof(provider_metadata) = 'object'),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_email_messages_idempotency_key
    ON public.messaging_email_messages USING btree (thread_id, idempotency_key);

CREATE UNIQUE INDEX messaging_email_messages_provider_message_key
    ON public.messaging_email_messages USING btree (thread_id, direction, provider_message_id)
    WHERE provider_message_id IS NOT NULL;

CREATE UNIQUE INDEX messaging_email_messages_internet_message_key
    ON public.messaging_email_messages USING btree (thread_id, direction, internet_message_id)
    WHERE internet_message_id IS NOT NULL;

CREATE INDEX messaging_email_messages_history_idx
    ON public.messaging_email_messages USING btree (thread_id, sequence);

CREATE INDEX messaging_email_messages_created_idx
    ON public.messaging_email_messages USING btree (created_at);

CREATE TABLE public.messaging_email_action_proposals (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    thread_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    source_message_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    action_kind text NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    expected_entity_version text NOT NULL,
    proposed_diff jsonb NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    apply_attempt integer NOT NULL DEFAULT 0,
    result jsonb,
    last_error text,
    expires_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    applying_at timestamptz,
    applied_at timestamptz,
    failed_at timestamptz,
    cancelled_at timestamptz,
    expired_at timestamptz,
    superseded_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_email_action_proposals_thread_fkey
        FOREIGN KEY (thread_id, workspace_id, user_id)
        REFERENCES public.messaging_email_threads(id, workspace_id, user_id)
        ON DELETE CASCADE,
    CONSTRAINT messaging_email_action_proposals_source_message_fkey
        FOREIGN KEY (thread_id, source_message_id)
        REFERENCES public.messaging_email_messages(thread_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT messaging_email_action_proposals_idempotency_check
        CHECK (btrim(idempotency_key) <> '' AND char_length(idempotency_key) <= 512),
    CONSTRAINT messaging_email_action_proposals_action_check
        CHECK (btrim(action_kind) <> '' AND char_length(action_kind) <= 128),
    CONSTRAINT messaging_email_action_proposals_entity_type_check
        CHECK (btrim(entity_type) <> '' AND char_length(entity_type) <= 128),
    CONSTRAINT messaging_email_action_proposals_version_check
        CHECK (btrim(expected_entity_version) <> '' AND char_length(expected_entity_version) <= 512),
    CONSTRAINT messaging_email_action_proposals_diff_check
        CHECK (jsonb_typeof(proposed_diff) = 'object'),
    CONSTRAINT messaging_email_action_proposals_status_check
        CHECK (status IN ('pending', 'confirmed', 'applying', 'applied', 'failed', 'cancelled', 'expired', 'superseded')),
    CONSTRAINT messaging_email_action_proposals_apply_attempt_check
        CHECK (
            (status IN ('pending', 'confirmed', 'cancelled', 'expired', 'superseded') AND apply_attempt = 0)
            OR (status IN ('applying', 'applied', 'failed') AND apply_attempt > 0)
        ),
    CONSTRAINT messaging_email_action_proposals_result_check
        CHECK (result IS NULL OR jsonb_typeof(result) = 'object'),
    CONSTRAINT messaging_email_action_proposals_lifecycle_check
        CHECK (
            expires_at > created_at
            AND (
                (status = 'pending'
                    AND confirmed_at IS NULL AND applying_at IS NULL AND applied_at IS NULL
                    AND failed_at IS NULL AND cancelled_at IS NULL AND expired_at IS NULL AND superseded_at IS NULL)
                OR (status = 'confirmed'
                    AND confirmed_at IS NOT NULL AND applying_at IS NULL AND applied_at IS NULL
                    AND failed_at IS NULL AND cancelled_at IS NULL AND expired_at IS NULL AND superseded_at IS NULL)
                OR (status = 'applying'
                    AND confirmed_at IS NOT NULL AND applying_at IS NOT NULL AND applied_at IS NULL
                    AND failed_at IS NULL AND cancelled_at IS NULL AND expired_at IS NULL AND superseded_at IS NULL)
                OR (status = 'applied'
                    AND confirmed_at IS NOT NULL AND applying_at IS NOT NULL AND applied_at IS NOT NULL
                    AND failed_at IS NULL AND cancelled_at IS NULL AND expired_at IS NULL AND superseded_at IS NULL)
                OR (status = 'failed'
                    AND confirmed_at IS NOT NULL AND applying_at IS NOT NULL AND applied_at IS NULL
                    AND failed_at IS NOT NULL AND cancelled_at IS NULL AND expired_at IS NULL AND superseded_at IS NULL)
                OR (status = 'cancelled'
                    AND confirmed_at IS NULL AND applying_at IS NULL AND applied_at IS NULL
                    AND failed_at IS NULL AND cancelled_at IS NOT NULL AND expired_at IS NULL AND superseded_at IS NULL)
                OR (status = 'expired'
                    AND confirmed_at IS NULL AND applying_at IS NULL AND applied_at IS NULL
                    AND failed_at IS NULL AND cancelled_at IS NULL AND expired_at IS NOT NULL AND superseded_at IS NULL)
                OR (status = 'superseded'
                    AND confirmed_at IS NULL AND applying_at IS NULL AND applied_at IS NULL
                    AND failed_at IS NULL AND cancelled_at IS NULL AND expired_at IS NULL AND superseded_at IS NOT NULL)
            )
        ),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_email_action_proposals_idempotency_key
    ON public.messaging_email_action_proposals USING btree (thread_id, idempotency_key);

CREATE UNIQUE INDEX messaging_email_action_proposals_one_pending_idx
    ON public.messaging_email_action_proposals USING btree (thread_id)
    WHERE status = 'pending';

CREATE INDEX messaging_email_action_proposals_actor_idx
    ON public.messaging_email_action_proposals USING btree (
        workspace_id, user_id, created_at DESC
    );

CREATE INDEX messaging_email_action_proposals_apply_recovery_idx
    ON public.messaging_email_action_proposals USING btree (status, updated_at)
    WHERE status IN ('confirmed', 'applying', 'failed');
