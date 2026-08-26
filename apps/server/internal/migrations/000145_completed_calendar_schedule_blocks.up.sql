ALTER TABLE public.calendar_schedule_blocks
    ADD COLUMN completed_at timestamptz;

DROP INDEX IF EXISTS public.uq_calendar_maya_story_segment;

CREATE UNIQUE INDEX uq_calendar_maya_story_segment
    ON public.calendar_schedule_blocks (workspace_id, user_id, story_id, segment_index)
    WHERE source = 'maya'
      AND story_id IS NOT NULL
      AND completed_at IS NULL;

CREATE INDEX idx_calendar_schedule_blocks_completed_history
    ON public.calendar_schedule_blocks (workspace_id, user_id, start_at)
    WHERE completed_at IS NOT NULL;

-- Completed work remains available as optional calendar history, but it must
-- immediately stop reserving time and disappear from the connected provider.
-- Cancelled work is not completed history and continues to be removed.
CREATE OR REPLACE FUNCTION public.cleanup_terminal_story_calendar_schedule()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    terminal_category text;
    effective_completed_at timestamptz;
BEGIN
    SELECT status.category
    INTO terminal_category
    FROM public.statuses AS status
    WHERE status.status_id = NEW.status_id;

    IF NEW.completed_at IS NOT NULL OR terminal_category = 'completed' THEN
        effective_completed_at := COALESCE(NEW.completed_at, CURRENT_TIMESTAMP);

        -- A retained block will no longer be deleted, so enqueue its provider
        -- cleanup before detaching the external mapping. Future blocks still
        -- flow through the existing BEFORE DELETE cleanup trigger below.
        UPDATE public.calendar_schedule_event_outbox AS outbox
        SET processed_at = CURRENT_TIMESTAMP,
            last_error = 'Superseded by story completion.',
            updated_at = CURRENT_TIMESTAMP
        FROM public.calendar_schedule_blocks AS block
        WHERE block.story_id = NEW.id
          AND block.workspace_id = NEW.workspace_id
          AND block.block_type = 'work'
          AND block.start_at < effective_completed_at
          AND block.source = 'maya'
          AND block.external_provider IS NOT NULL
          AND block.external_event_id IS NOT NULL
          AND outbox.schedule_block_id = block.block_id
          AND outbox.processed_at IS NULL;

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
        )
        SELECT
            block.workspace_id,
            block.user_id,
            block.block_id,
            'delete',
            block.external_provider,
            COALESCE(block.external_calendar_id, 'primary'),
            block.external_event_id,
            jsonb_build_object(
                'CalendarID', COALESCE(block.external_calendar_id, 'primary'),
                'EventID', block.external_event_id,
                'BlockID', block.block_id,
                'StoryID', block.story_id,
                'WorkspaceID', block.workspace_id
            ),
            CONCAT(block.external_provider, ':delete:', block.block_id, ':')
        FROM public.calendar_schedule_blocks AS block
        WHERE block.story_id = NEW.id
          AND block.workspace_id = NEW.workspace_id
          AND block.block_type = 'work'
          AND block.start_at < effective_completed_at
          AND block.source = 'maya'
          AND block.external_provider IS NOT NULL
          AND block.external_event_id IS NOT NULL
        ON CONFLICT (dedupe_key) DO UPDATE SET
            workspace_id = EXCLUDED.workspace_id,
            user_id = EXCLUDED.user_id,
            schedule_block_id = EXCLUDED.schedule_block_id,
            provider = EXCLUDED.provider,
            calendar_id = EXCLUDED.calendar_id,
            provider_event_id = EXCLUDED.provider_event_id,
            payload = EXCLUDED.payload,
            processed_at = NULL,
            dead_lettered_at = NULL,
            attempt_count = 0,
            last_error = NULL,
            available_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP;

        DELETE FROM public.calendar_schedule_blocks
        WHERE story_id = NEW.id
          AND workspace_id = NEW.workspace_id
          AND block_type = 'work'
          AND start_at >= effective_completed_at;

        UPDATE public.calendar_schedule_blocks
        SET end_at = LEAST(end_at, effective_completed_at),
            completed_at = effective_completed_at,
            is_locked = FALSE,
            external_provider = NULL,
            external_calendar_id = NULL,
            external_event_id = NULL,
            external_sync_hash = NULL,
            external_synced_at = NULL,
            updated_at = CURRENT_TIMESTAMP
        WHERE story_id = NEW.id
          AND workspace_id = NEW.workspace_id
          AND block_type = 'work'
          AND start_at < effective_completed_at;
    ELSIF terminal_category = 'cancelled' THEN
        DELETE FROM public.calendar_schedule_blocks
        WHERE story_id = NEW.id
          AND workspace_id = NEW.workspace_id
          AND block_type = 'work';
    END IF;

    RETURN NEW;
END;
$$;

