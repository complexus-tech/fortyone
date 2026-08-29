ALTER TABLE public.slack_workspaces
    ADD COLUMN credential_payload text,
    ADD COLUMN credential_key_version smallint NOT NULL DEFAULT 0,
    ADD COLUMN installation_generation uuid NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN installation_authorized_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN slack_app_id text,
    ADD COLUMN enterprise_id text,
    ADD COLUMN authed_user_id text,
    ADD COLUMN revoked_at timestamptz;

DROP INDEX IF EXISTS public.slack_workspaces_slack_team_id_key;

CREATE UNIQUE INDEX slack_workspaces_slack_team_id_key
    ON public.slack_workspaces USING btree (slack_team_id)
    WHERE is_active = true;

ALTER TABLE public.stories
    ADD COLUMN external_creation_key text;

CREATE UNIQUE INDEX stories_external_creation_key_key
    ON public.stories USING btree (workspace_id, external_creation_key)
    WHERE external_creation_key IS NOT NULL;

ALTER TABLE public.story_links
    ADD COLUMN external_source_key text;

CREATE UNIQUE INDEX story_links_external_source_key_unique
    ON public.story_links USING btree (external_source_key)
    WHERE external_source_key IS NOT NULL;

UPDATE public.slack_workspaces
SET credential_payload = bot_access_token
WHERE credential_payload IS NULL;

UPDATE public.slack_request_logs
SET trigger_id = NULL,
    request_body = NULL,
    headers = '{}'::jsonb;

CREATE TABLE public.messaging_nonces (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    provider text NOT NULL,
    purpose text NOT NULL,
    nonce_hash bytea NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid,
    external_workspace_id text,
    external_user_id text,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_nonces_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_nonces_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT messaging_nonces_purpose_check CHECK (purpose IN ('oauth_install', 'account_link')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_nonces_provider_hash_key
    ON public.messaging_nonces USING btree (provider, nonce_hash);
CREATE INDEX messaging_nonces_expiry_idx
	ON public.messaging_nonces USING btree (expires_at);

CREATE TABLE public.messaging_inbound_events (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    provider text NOT NULL,
    workspace_id uuid,
    installation_generation uuid,
    external_workspace_id text NOT NULL,
    external_event_id text NOT NULL,
    event_type text NOT NULL,
    payload_encrypted text,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    recovery_generation integer NOT NULL DEFAULT 0,
    recovery_enqueued_at timestamptz,
    last_error text,
    received_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_inbound_events_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE SET NULL,
    CONSTRAINT messaging_inbound_events_status_check CHECK (status IN ('pending', 'processing', 'completed', 'ignored', 'failed', 'cancelled')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_inbound_events_external_key
    ON public.messaging_inbound_events USING btree (provider, external_workspace_id, external_event_id);
CREATE INDEX messaging_inbound_events_status_idx
	ON public.messaging_inbound_events USING btree (status, received_at);
CREATE INDEX messaging_inbound_events_received_idx
	ON public.messaging_inbound_events USING btree (received_at);
CREATE INDEX messaging_inbound_events_recovery_idx
	ON public.messaging_inbound_events USING btree (provider, status, updated_at)
	WHERE payload_encrypted IS NOT NULL AND attempt_count < 20;

CREATE TABLE public.messaging_conversations (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    provider text NOT NULL,
    workspace_id uuid NOT NULL,
    external_workspace_id text NOT NULL,
    external_channel_id text NOT NULL,
    external_thread_id text NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_conversations_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_conversations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_conversations_external_key
    ON public.messaging_conversations USING btree (
        provider,
		workspace_id,
        external_workspace_id,
        external_channel_id,
        external_thread_id,
		user_id
	);
CREATE INDEX messaging_conversations_updated_idx
	ON public.messaging_conversations USING btree (updated_at);

CREATE TABLE public.messaging_messages (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    conversation_id uuid NOT NULL,
    external_message_id text,
    role text NOT NULL,
    content text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_messages_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.messaging_conversations(id) ON DELETE CASCADE,
    CONSTRAINT messaging_messages_role_check CHECK (role IN ('user', 'assistant')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_messages_external_key
    ON public.messaging_messages USING btree (conversation_id, external_message_id, role)
    WHERE external_message_id IS NOT NULL;
CREATE INDEX messaging_messages_conversation_created_idx
	ON public.messaging_messages USING btree (conversation_id, created_at DESC);
CREATE INDEX messaging_messages_created_idx
	ON public.messaging_messages USING btree (created_at);

CREATE TABLE public.messaging_outbound_deliveries (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    provider text NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid,
    installation_generation uuid,
    external_workspace_id text NOT NULL,
    external_recipient_user_id text,
    inbound_event_id uuid,
    idempotency_key text NOT NULL,
    external_channel_id text NOT NULL,
    external_thread_id text,
    external_message_id text,
    content text,
    purpose text NOT NULL DEFAULT 'provider_message',
    expires_at timestamptz,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    last_error text,
    delivered_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messaging_outbound_deliveries_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT messaging_outbound_deliveries_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT messaging_outbound_deliveries_inbound_event_id_fkey FOREIGN KEY (inbound_event_id) REFERENCES public.messaging_inbound_events(id) ON DELETE SET NULL,
    CONSTRAINT messaging_outbound_deliveries_status_check CHECK (status IN ('pending', 'delivering', 'delivered', 'failed', 'cancelled')),
    CONSTRAINT messaging_outbound_deliveries_purpose_check CHECK (purpose IN ('provider_message', 'assistant', 'account_link', 'access', 'creation_confirmation')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX messaging_outbound_deliveries_idempotency_key
    ON public.messaging_outbound_deliveries USING btree (provider, workspace_id, idempotency_key);
CREATE INDEX messaging_outbound_deliveries_status_idx
	ON public.messaging_outbound_deliveries USING btree (status, created_at);
CREATE INDEX messaging_outbound_deliveries_created_idx
	ON public.messaging_outbound_deliveries USING btree (created_at);

CREATE TABLE public.slack_uninstall_outbox (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    slack_workspace_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    installation_generation uuid NOT NULL,
    slack_team_id text NOT NULL,
    uninstall_kind text NOT NULL DEFAULT 'disconnect',
    credential_payload text,
    credential_key_version smallint NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    attempt_count integer NOT NULL DEFAULT 0,
    last_error text,
    next_attempt_at timestamptz,
    processing_started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT slack_uninstall_outbox_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'revocation_required')),
    CONSTRAINT slack_uninstall_outbox_kind_check CHECK (uninstall_kind IN ('disconnect', 'workspace_delete', 'orphaned_oauth')),
    CONSTRAINT slack_uninstall_outbox_attempt_count_check CHECK (attempt_count >= 0),
    CONSTRAINT slack_uninstall_outbox_credential_key_version_check CHECK (credential_key_version > 0),
    CONSTRAINT slack_uninstall_outbox_credential_lifecycle_check CHECK (
        (status = 'completed' AND credential_payload IS NULL)
        OR (status <> 'completed' AND NULLIF(credential_payload, '') IS NOT NULL)
    ),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX slack_uninstall_outbox_installation_key
    ON public.slack_uninstall_outbox USING btree (slack_workspace_id, installation_generation, uninstall_kind);

CREATE INDEX slack_uninstall_outbox_recovery_idx
    ON public.slack_uninstall_outbox USING btree (status, next_attempt_at, updated_at)
    WHERE status IN ('pending', 'processing', 'failed');
CREATE INDEX slack_uninstall_outbox_team_idx
    ON public.slack_uninstall_outbox USING btree (slack_team_id, created_at DESC);
