-- Google Calendar credentials belong to the account, not to whichever
-- workspace happened to be active when OAuth completed. The historical
-- workspace columns remain as the tenant anchor for cached rows, but removing
-- that membership must not revoke an otherwise usable personal connection.
ALTER TABLE public.calendar_connections
    DROP CONSTRAINT IF EXISTS calendar_connections_workspace_member_fkey,
    DROP CONSTRAINT IF EXISTS calendar_connections_workspace_id_fkey,
    ADD COLUMN cleanup_pending_at timestamptz;

ALTER TABLE public.calendar_events
    DROP CONSTRAINT calendar_events_connection_scope_fkey,
    ADD CONSTRAINT calendar_events_connection_scope_fkey
        FOREIGN KEY (connection_id, workspace_id, user_id)
        REFERENCES public.calendar_connections (connection_id, workspace_id, user_id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

-- Define the durable provider outbox before any trigger function references
-- it. PostgreSQL relation resolution must not depend on deferred PL/pgSQL
-- compilation during a live migration.
CREATE TABLE public.calendar_schedule_event_outbox (
    outbox_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    schedule_block_id uuid,
    operation varchar(16) NOT NULL,
    provider varchar(32) NOT NULL DEFAULT 'google',
    calendar_id varchar(255) NOT NULL DEFAULT 'primary',
    provider_event_id varchar(255) NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    dedupe_key varchar(512) NOT NULL,
    attempt_count integer NOT NULL DEFAULT 0,
    available_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at timestamptz,
    dead_lettered_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_schedule_event_outbox_pkey PRIMARY KEY (outbox_id),
    CONSTRAINT calendar_schedule_event_outbox_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT calendar_schedule_event_outbox_operation_check
        CHECK (operation IN ('upsert', 'delete')),
    CONSTRAINT calendar_schedule_event_outbox_provider_check
        CHECK (provider = 'google'),
    CONSTRAINT calendar_schedule_event_outbox_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT calendar_schedule_event_outbox_dedupe_key_key UNIQUE (dedupe_key)
);

CREATE INDEX idx_calendar_schedule_event_outbox_ready
    ON public.calendar_schedule_event_outbox (available_at, user_id, created_at)
    WHERE processed_at IS NULL AND dead_lettered_at IS NULL;

CREATE FUNCTION public.reanchor_account_calendar_connection()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    replacement_workspace_id uuid;
BEGIN
    -- Coordinate membership changes with snapshots, push sync, and schedule
    -- mirroring, all of which use this account-scoped advisory lock.
    PERFORM pg_advisory_xact_lock(
        hashtextextended(CONCAT('calendar:', CAST(OLD.user_id AS text)), 0)
    );

    SELECT member.workspace_id
    INTO replacement_workspace_id
    FROM public.workspace_members member
    WHERE member.user_id = OLD.user_id
      AND member.workspace_id <> OLD.workspace_id
    ORDER BY member.created_at, member.workspace_id
    LIMIT 1;

    IF replacement_workspace_id IS NULL THEN
        DELETE FROM public.calendar_events cached_event
        USING public.calendar_connections calendar_connection
        WHERE cached_event.connection_id = calendar_connection.connection_id
          AND calendar_connection.workspace_id = OLD.workspace_id
          AND calendar_connection.user_id = OLD.user_id;

        DELETE FROM public.calendar_busy_windows busy_window
        USING public.calendar_connections calendar_connection
        WHERE busy_window.connection_id = calendar_connection.connection_id
          AND calendar_connection.workspace_id = OLD.workspace_id
          AND calendar_connection.user_id = OLD.user_id;

        -- A prior permanent write failure must not cause teardown to discard
        -- the only credential capable of deleting an already-created event.
        UPDATE public.calendar_schedule_event_outbox outbox
        SET dead_lettered_at = NULL,
            attempt_count = 0,
            available_at = CURRENT_TIMESTAMP,
            last_error = 'Retrying provider cleanup after final workspace removal.',
            updated_at = CURRENT_TIMESTAMP
        WHERE outbox.user_id = OLD.user_id
          AND outbox.processed_at IS NULL;

        IF EXISTS (
            SELECT 1
            FROM public.calendar_schedule_blocks block
            WHERE block.user_id = OLD.user_id
              AND block.source = 'maya'
              AND block.external_provider = 'google'
              AND block.external_event_id IS NOT NULL
        ) OR EXISTS (
            SELECT 1
            FROM public.calendar_schedule_event_outbox outbox
            WHERE outbox.user_id = OLD.user_id
              AND outbox.processed_at IS NULL
              AND outbox.dead_lettered_at IS NULL
        ) THEN
            -- Keep the encrypted token only long enough for the block cascade
            -- to enqueue and dispatch provider deletes. Application reads,
            -- snapshots, watches, and upserts exclude cleanup-pending rows.
            UPDATE public.calendar_connections connection
            SET cleanup_pending_at = CURRENT_TIMESTAMP,
                sync_token = NULL,
                notification_channel_id = NULL,
                notification_resource_id = NULL,
                notification_expires_at = NULL,
                updated_at = CURRENT_TIMESTAMP
            WHERE connection.workspace_id = OLD.workspace_id
              AND connection.user_id = OLD.user_id;
        ELSE
            DELETE FROM public.calendar_connections connection
            WHERE connection.workspace_id = OLD.workspace_id
              AND connection.user_id = OLD.user_id;
        END IF;
        RETURN OLD;
    END IF;

    -- Preserve the complete provider snapshot. Busy windows carry their own
    -- workspace scope, while calendar events follow the connection through
    -- the ON UPDATE CASCADE composite foreign key above.
    UPDATE public.calendar_busy_windows busy_window
    SET workspace_id = replacement_workspace_id,
        updated_at = CURRENT_TIMESTAMP
    FROM public.calendar_connections calendar_connection
    WHERE busy_window.connection_id = calendar_connection.connection_id
      AND calendar_connection.workspace_id = OLD.workspace_id
      AND calendar_connection.user_id = OLD.user_id;

    UPDATE public.calendar_connections connection
    SET workspace_id = replacement_workspace_id,
        updated_at = CURRENT_TIMESTAMP
    WHERE connection.workspace_id = OLD.workspace_id
      AND connection.user_id = OLD.user_id;

    RETURN OLD;
END;
$$;

CREATE TRIGGER workspace_members_reanchor_account_calendar_connection
BEFORE DELETE ON public.workspace_members
FOR EACH ROW
EXECUTE FUNCTION public.reanchor_account_calendar_connection();

ALTER TABLE public.calendar_schedule_blocks
    ADD COLUMN segment_index integer NOT NULL DEFAULT 0,
    ADD COLUMN external_provider varchar(32),
    ADD COLUMN external_calendar_id varchar(255),
    ADD COLUMN external_event_id varchar(255),
    ADD COLUMN external_sync_hash varchar(64),
    ADD COLUMN external_synced_at timestamptz,
    ADD CONSTRAINT calendar_schedule_blocks_segment_index_check
        CHECK (segment_index >= 0),
    ADD CONSTRAINT calendar_schedule_blocks_external_mapping_check
        CHECK (
            (external_provider IS NULL AND external_calendar_id IS NULL AND external_event_id IS NULL)
            OR
            (external_provider IS NOT NULL AND external_calendar_id IS NOT NULL AND external_event_id IS NOT NULL)
        );

WITH ranked AS (
    SELECT
        block_id,
        CAST(
            ROW_NUMBER() OVER (
                PARTITION BY workspace_id, user_id, story_id
                ORDER BY start_at, block_id
            ) - 1 AS integer
        ) AS segment_index
    FROM public.calendar_schedule_blocks
    WHERE source = 'maya'
      AND story_id IS NOT NULL
)
UPDATE public.calendar_schedule_blocks AS block
SET segment_index = ranked.segment_index
FROM ranked
WHERE block.block_id = ranked.block_id;

CREATE UNIQUE INDEX uq_calendar_maya_story_segment
    ON public.calendar_schedule_blocks (workspace_id, user_id, story_id, segment_index)
    WHERE source = 'maya' AND story_id IS NOT NULL;

CREATE TABLE public.calendar_maya_schedule_ownerships (
    workspace_id uuid NOT NULL,
    story_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recovery_attempted_at timestamptz,
    CONSTRAINT calendar_maya_schedule_ownerships_pkey PRIMARY KEY (workspace_id, story_id),
    CONSTRAINT calendar_maya_schedule_ownerships_story_workspace_fkey
        FOREIGN KEY (story_id, workspace_id)
        REFERENCES public.stories(id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT calendar_maya_schedule_ownerships_workspace_member_fkey
        FOREIGN KEY (workspace_id, user_id)
        REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE
);

INSERT INTO public.calendar_maya_schedule_ownerships (workspace_id, story_id, user_id, created_at, updated_at)
SELECT DISTINCT ON (workspace_id, story_id)
    workspace_id,
    story_id,
    user_id,
    created_at,
    updated_at
FROM public.calendar_schedule_blocks
WHERE source = 'maya'
  AND story_id IS NOT NULL
ORDER BY workspace_id, story_id, updated_at DESC, block_id;

CREATE INDEX idx_calendar_maya_schedule_ownerships_user
    ON public.calendar_maya_schedule_ownerships (user_id, updated_at, story_id);

CREATE INDEX idx_calendar_maya_schedule_ownerships_recovery
    ON public.calendar_maya_schedule_ownerships (updated_at, workspace_id, story_id);

CREATE INDEX idx_maya_agent_runs_schedule_recovery
    ON public.maya_agent_runs (updated_at, workspace_id, story_id, run_id)
    WHERE status = 'running';

CREATE FUNCTION public.enqueue_deleted_maya_schedule_event()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.source <> 'maya' OR OLD.external_event_id IS NULL THEN
        RETURN OLD;
    END IF;

    UPDATE public.calendar_schedule_event_outbox
    SET processed_at = CURRENT_TIMESTAMP,
        last_error = 'Superseded by schedule block deletion.',
        updated_at = CURRENT_TIMESTAMP
    WHERE schedule_block_id = OLD.block_id
      AND processed_at IS NULL;

    INSERT INTO public.calendar_schedule_event_outbox (
        workspace_id,
        user_id,
        schedule_block_id,
        operation,
        provider,
        calendar_id,
        provider_event_id,
        payload,
        dedupe_key
    ) VALUES (
        OLD.workspace_id,
        OLD.user_id,
        OLD.block_id,
        'delete',
        'google',
        COALESCE(OLD.external_calendar_id, 'primary'),
        OLD.external_event_id,
        jsonb_build_object(
            'CalendarID', COALESCE(OLD.external_calendar_id, 'primary'),
            'EventID', OLD.external_event_id,
            'BlockID', OLD.block_id,
            'StoryID', OLD.story_id,
            'WorkspaceID', OLD.workspace_id
        ),
        CONCAT('delete:', OLD.external_event_id, ':')
    )
    ON CONFLICT (dedupe_key) DO UPDATE SET
        workspace_id = EXCLUDED.workspace_id,
        user_id = EXCLUDED.user_id,
        schedule_block_id = EXCLUDED.schedule_block_id,
        calendar_id = EXCLUDED.calendar_id,
        provider_event_id = EXCLUDED.provider_event_id,
        payload = EXCLUDED.payload,
        processed_at = NULL,
        dead_lettered_at = NULL,
        attempt_count = 0,
        last_error = NULL,
        available_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP;

    RETURN OLD;
END;
$$;

CREATE TRIGGER calendar_schedule_blocks_enqueue_provider_delete
BEFORE DELETE ON public.calendar_schedule_blocks
FOR EACH ROW
EXECUTE FUNCTION public.enqueue_deleted_maya_schedule_event();

CREATE FUNCTION public.cleanup_retired_story_maya_schedule()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
       OR (OLD.archived_at IS NULL AND NEW.archived_at IS NOT NULL) THEN
        DELETE FROM public.calendar_schedule_blocks
        WHERE story_id = NEW.id
          AND workspace_id = NEW.workspace_id
          AND source = 'maya';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER stories_cleanup_retired_maya_schedule
AFTER UPDATE OF deleted_at, archived_at ON public.stories
FOR EACH ROW
EXECUTE FUNCTION public.cleanup_retired_story_maya_schedule();

CREATE FUNCTION public.cleanup_removed_team_member_maya_schedule()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM public.calendar_schedule_blocks block
    USING public.stories story
    WHERE block.story_id = story.id
      AND block.workspace_id = story.workspace_id
      AND block.user_id = OLD.user_id
      AND block.source = 'maya'
      AND story.team_id = OLD.team_id;
    RETURN OLD;
END;
$$;

CREATE TRIGGER team_members_cleanup_maya_schedule
AFTER DELETE ON public.team_members
FOR EACH ROW
EXECUTE FUNCTION public.cleanup_removed_team_member_maya_schedule();
