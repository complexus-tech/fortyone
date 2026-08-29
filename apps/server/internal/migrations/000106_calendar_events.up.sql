-- 000106_calendar_events.up.sql

ALTER TABLE public.calendar_connections
    ADD COLUMN provider_account_id text NOT NULL DEFAULT '',
    ADD COLUMN credential_generation uuid NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE public.calendar_connections
    ADD CONSTRAINT calendar_connections_scope_unique
        UNIQUE (connection_id, workspace_id, user_id);

-- Credentials and cached calendar data cannot outlive workspace membership.
DELETE FROM public.calendar_connections cc
WHERE NOT EXISTS (
    SELECT 1
    FROM public.workspace_members wm
    WHERE wm.workspace_id = cc.workspace_id
      AND wm.user_id = cc.user_id
);

ALTER TABLE public.calendar_connections
    ADD CONSTRAINT calendar_connections_workspace_member_fkey
        FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members (workspace_id, user_id)
        ON DELETE CASCADE;

DELETE FROM public.calendar_busy_windows cbw
USING public.calendar_connections cc
WHERE cbw.connection_id = cc.connection_id
  AND cc.revoked_at IS NOT NULL;

UPDATE public.calendar_connections
SET token_payload = '',
    scopes = '{}',
    sync_error = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE revoked_at IS NOT NULL;

-- Work blocks must point to a story in the same workspace. Invalid legacy
-- links are removed because their intended tenant cannot be inferred safely.
DELETE FROM public.calendar_schedule_blocks csb
WHERE NOT EXISTS (
    SELECT 1
    FROM public.workspace_members wm
    WHERE wm.workspace_id = csb.workspace_id
      AND wm.user_id = csb.user_id
);

DELETE FROM public.calendar_schedule_blocks csb
WHERE csb.story_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.stories s
      WHERE s.id = csb.story_id
        AND s.workspace_id = csb.workspace_id
  );

ALTER TABLE public.stories
    ADD CONSTRAINT stories_id_workspace_unique
        UNIQUE (id, workspace_id);

ALTER TABLE public.calendar_schedule_blocks
    DROP CONSTRAINT calendar_schedule_blocks_story_id_fkey,
    ADD CONSTRAINT calendar_schedule_blocks_workspace_member_fkey
        FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members (workspace_id, user_id)
        ON DELETE CASCADE,
    ADD CONSTRAINT calendar_schedule_blocks_story_workspace_fkey
        FOREIGN KEY (story_id, workspace_id)
        REFERENCES public.stories (id, workspace_id)
        ON DELETE CASCADE;

-- Availability is intentionally title-free. Owner-visible details live only in
-- calendar_events and never enter Maya or capacity-planning data paths.
UPDATE public.calendar_busy_windows
SET title = NULL
WHERE title IS NOT NULL;

CREATE TABLE public.calendar_events (
    event_id uuid NOT NULL DEFAULT gen_random_uuid(),
    connection_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    provider varchar(32) NOT NULL,
    calendar_id text NOT NULL,
    provider_event_id text NOT NULL,
    title text,
    description text,
    location text,
    meeting_url text,
    html_link text,
    organizer jsonb,
    attendees jsonb NOT NULL DEFAULT '[]'::jsonb,
    attendees_omitted bool NOT NULL DEFAULT false,
    is_all_day bool NOT NULL DEFAULT false,
    start_date date,
    end_date date,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    visibility varchar(32) NOT NULL DEFAULT 'default',
    is_private bool NOT NULL DEFAULT false,
    source_hash text NOT NULL,
    sync_generation uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT calendar_events_connection_scope_fkey
        FOREIGN KEY (connection_id, workspace_id, user_id)
        REFERENCES public.calendar_connections (connection_id, workspace_id, user_id)
        ON DELETE CASCADE,
    CONSTRAINT calendar_events_provider_check
        CHECK (provider IN ('google')),
    CONSTRAINT calendar_events_visibility_check
        CHECK (visibility IN ('default', 'public', 'private', 'confidential')),
    CONSTRAINT calendar_events_valid_range_check
        CHECK (end_at > start_at),
    CONSTRAINT calendar_events_all_day_dates_check
        CHECK (
            (
                is_all_day
                AND start_date IS NOT NULL
                AND end_date IS NOT NULL
                AND end_date > start_date
            )
            OR (
                NOT is_all_day
                AND start_date IS NULL
                AND end_date IS NULL
            )
        ),
    CONSTRAINT calendar_events_organizer_object_check
        CHECK (organizer IS NULL OR jsonb_typeof(organizer) = 'object'),
    CONSTRAINT calendar_events_attendees_array_check
        CHECK (jsonb_typeof(attendees) = 'array'),
    CONSTRAINT calendar_events_private_visibility_check
        CHECK (
            is_private = (visibility IN ('private', 'confidential'))
        )
);

CREATE UNIQUE INDEX calendar_events_provider_event_unique
    ON public.calendar_events (connection_id, calendar_id, provider_event_id);

CREATE INDEX idx_calendar_events_workspace_user_range
    ON public.calendar_events (workspace_id, user_id, start_at, end_at);

CREATE INDEX idx_calendar_events_connection_generation
    ON public.calendar_events (connection_id, sync_generation);
