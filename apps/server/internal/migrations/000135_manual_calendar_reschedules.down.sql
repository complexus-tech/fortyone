DROP TABLE IF EXISTS public.calendar_schedule_reschedule_events;

DROP INDEX IF EXISTS public.idx_calendar_schedule_blocks_manual_override;

ALTER TABLE public.calendar_schedule_blocks
    DROP CONSTRAINT IF EXISTS calendar_schedule_blocks_manual_override_by_fkey,
    DROP COLUMN IF EXISTS manual_override_at,
    DROP COLUMN IF EXISTS manual_override_by;
