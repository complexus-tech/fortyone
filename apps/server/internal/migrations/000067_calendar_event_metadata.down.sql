-- 000067_calendar_event_metadata.down.sql

ALTER TABLE public.calendar_busy_windows
    DROP COLUMN IF EXISTS is_private,
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS calendar_id;
