CREATE TABLE public.figma_connections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    figma_user_id text NOT NULL,
    figma_email text,
    figma_handle text,
    token_payload text NOT NULL,
    scopes text[] NOT NULL DEFAULT '{}'::text[],
    expires_at timestamptz NOT NULL,
    connected_by_user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE RESTRICT,
    is_active boolean NOT NULL DEFAULT true,
    disconnected_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX figma_connections_one_active_per_workspace
    ON public.figma_connections (workspace_id)
    WHERE is_active = true;

CREATE TABLE public.figma_oauth_states (
    state_hash text PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    workspace_slug text NOT NULL,
    code_verifier text NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX figma_oauth_states_expiry ON public.figma_oauth_states (expires_at);

CREATE TABLE public.story_figma_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    story_id uuid NOT NULL REFERENCES public.stories(id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE RESTRICT,
    story_link_id uuid REFERENCES public.story_links(link_id) ON DELETE SET NULL,
    file_key text NOT NULL,
    node_id text,
    original_url text NOT NULL,
    canonical_url text NOT NULL,
    file_name text NOT NULL,
    node_name text,
    node_type text,
    thumbnail_url text,
    version text,
    last_modified timestamptz,
    dev_status text,
    dev_resource_id text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_synced_at timestamptz NOT NULL DEFAULT now(),
    unavailable_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX story_figma_links_unique_artifact
    ON public.story_figma_links (story_id, file_key, COALESCE(node_id, ''));
CREATE INDEX story_figma_links_story ON public.story_figma_links (story_id);
CREATE INDEX story_figma_links_file ON public.story_figma_links (file_key);

CREATE TABLE public.figma_webhooks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id uuid NOT NULL REFERENCES public.figma_connections(id) ON DELETE CASCADE,
    file_key text NOT NULL,
    event_type text NOT NULL,
    figma_webhook_id bigint NOT NULL,
    passcode_hash text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (connection_id, file_key, event_type),
    UNIQUE (figma_webhook_id)
);

CREATE TABLE public.figma_webhook_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    figma_webhook_id bigint NOT NULL,
    event_type text NOT NULL,
    event_key text NOT NULL UNIQUE,
    payload jsonb NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz
);
