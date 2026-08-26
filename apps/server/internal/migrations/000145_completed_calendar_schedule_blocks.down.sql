DELETE FROM public.calendar_schedule_blocks
WHERE completed_at IS NOT NULL;

DROP INDEX IF EXISTS public.idx_calendar_schedule_blocks_completed_history;
DROP INDEX IF EXISTS public.uq_calendar_maya_story_segment;

CREATE UNIQUE INDEX uq_calendar_maya_story_segment
    ON public.calendar_schedule_blocks (workspace_id, user_id, story_id, segment_index)
    WHERE source = 'maya' AND story_id IS NOT NULL;

CREATE OR REPLACE FUNCTION public.cleanup_terminal_story_calendar_schedule()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL
       OR EXISTS (
           SELECT 1
           FROM public.statuses AS status
           WHERE status.status_id = NEW.status_id
               AND status.category IN ('completed', 'cancelled')
       ) THEN
        DELETE FROM public.calendar_schedule_blocks
        WHERE story_id = NEW.id
            AND workspace_id = NEW.workspace_id
            AND block_type = 'work';
    END IF;

    RETURN NEW;
END;
$$;

ALTER TABLE public.calendar_schedule_blocks
    DROP COLUMN completed_at;

