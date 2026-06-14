-- 000065_calendar_integrations.up.sql

CREATE TABLE public.calendar_connections (
    connection_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    provider varchar(32) NOT NULL,
    connected_email varchar(255) NOT NULL,
    timezone varchar(128) NOT NULL DEFAULT 'UTC',
    token_payload text NOT NULL,
    scopes text[] NOT NULL DEFAULT '{}',
    sync_status varchar(32) NOT NULL DEFAULT 'connected',
    sync_error text,
    last_synced_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_connections_pkey PRIMARY KEY (connection_id),
    CONSTRAINT calendar_connections_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT calendar_connections_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT calendar_connections_provider_check
        CHECK (provider IN ('google')),
    CONSTRAINT calendar_connections_sync_status_check
        CHECK (sync_status IN ('connected', 'synced', 'failed', 'revoked'))
);

CREATE UNIQUE INDEX calendar_connections_one_active_provider_per_user
    ON public.calendar_connections (workspace_id, user_id, provider)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_calendar_connections_workspace_user
    ON public.calendar_connections (workspace_id, user_id)
    WHERE revoked_at IS NULL;

CREATE TABLE public.calendar_busy_windows (
    window_id uuid NOT NULL DEFAULT gen_random_uuid(),
    connection_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    provider varchar(32) NOT NULL,
    provider_event_id text NOT NULL,
    calendar_id text,
    title text,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    status varchar(32) NOT NULL,
    transparency varchar(32) NOT NULL,
    is_private bool NOT NULL DEFAULT true,
    source_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_busy_windows_pkey PRIMARY KEY (window_id),
    CONSTRAINT calendar_busy_windows_connection_id_fkey
        FOREIGN KEY (connection_id) REFERENCES public.calendar_connections(connection_id) ON DELETE CASCADE,
    CONSTRAINT calendar_busy_windows_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT calendar_busy_windows_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT calendar_busy_windows_provider_check
        CHECK (provider IN ('google')),
    CONSTRAINT calendar_busy_windows_status_check
        CHECK (status IN ('busy')),
    CONSTRAINT calendar_busy_windows_transparency_check
        CHECK (transparency IN ('opaque')),
    CONSTRAINT calendar_busy_windows_valid_range_check
        CHECK (end_at > start_at)
);

CREATE UNIQUE INDEX calendar_busy_windows_provider_event_unique
    ON public.calendar_busy_windows (connection_id, provider_event_id);

CREATE INDEX idx_calendar_busy_windows_workspace_user_range
    ON public.calendar_busy_windows (workspace_id, user_id, start_at, end_at);
