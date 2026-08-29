ALTER TABLE public.integration_requests
    ADD COLUMN label_ids uuid[] NOT NULL DEFAULT '{}';

UPDATE public.integration_requests
SET label_ids = COALESCE(
    ARRAY(
        SELECT DISTINCT value::uuid
        FROM jsonb_array_elements_text(
            CASE
                WHEN jsonb_typeof(metadata -> 'label_ids') = 'array' THEN metadata -> 'label_ids'
                ELSE '[]'::jsonb
            END
        ) AS labels(value)
        WHERE value ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
    ),
    '{}'
);

CREATE TABLE public.integration_request_threads (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    integration_request_id uuid NOT NULL,
    provider text NOT NULL,
    external_workspace_id text NOT NULL,
    installation_generation uuid,
    external_channel_id text NOT NULL,
    external_thread_id text NOT NULL,
    external_source_message_id text,
    source_url text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integration_request_threads_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT integration_request_threads_request_id_fkey FOREIGN KEY (integration_request_id) REFERENCES public.integration_requests(id) ON DELETE CASCADE,
    CONSTRAINT integration_request_threads_provider_check CHECK (provider IN ('slack')),
    CONSTRAINT integration_request_threads_external_workspace_check CHECK (btrim(external_workspace_id) <> ''),
    CONSTRAINT integration_request_threads_external_channel_check CHECK (btrim(external_channel_id) <> ''),
    CONSTRAINT integration_request_threads_external_thread_check CHECK (btrim(external_thread_id) <> ''),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX integration_request_threads_request_provider_key
    ON public.integration_request_threads USING btree (integration_request_id, provider);

CREATE UNIQUE INDEX integration_request_threads_external_key
    ON public.integration_request_threads USING btree (provider, external_workspace_id, external_channel_id, external_thread_id);

CREATE INDEX integration_request_threads_story_idx
    ON public.integration_request_threads USING btree (workspace_id, integration_request_id);

CREATE TABLE public.integration_request_comments (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    thread_id uuid NOT NULL,
    direction text NOT NULL,
    author_user_id uuid,
    external_author_id text,
    external_message_id text,
    outbound_idempotency_key text,
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integration_request_comments_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT integration_request_comments_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.integration_request_threads(id) ON DELETE CASCADE,
    CONSTRAINT integration_request_comments_author_user_id_fkey FOREIGN KEY (author_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT integration_request_comments_direction_check CHECK (direction IN ('inbound', 'outbound')),
    CONSTRAINT integration_request_comments_body_check CHECK (btrim(body) <> ''),
    CONSTRAINT integration_request_comments_outbound_binding_check CHECK (
        (direction = 'outbound' AND outbound_idempotency_key IS NOT NULL)
        OR (direction = 'inbound' AND external_message_id IS NOT NULL)
    ),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX integration_request_comments_external_message_key
    ON public.integration_request_comments USING btree (thread_id, external_message_id)
    WHERE external_message_id IS NOT NULL;

CREATE UNIQUE INDEX integration_request_comments_outbound_key
    ON public.integration_request_comments USING btree (thread_id, outbound_idempotency_key)
    WHERE outbound_idempotency_key IS NOT NULL;

CREATE INDEX integration_request_comments_thread_created_idx
    ON public.integration_request_comments USING btree (thread_id, created_at, id);
