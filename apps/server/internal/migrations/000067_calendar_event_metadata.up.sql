-- 000067_calendar_event_metadata.up.sql

ALTER TABLE public.calendar_busy_windows
    ADD COLUMN IF NOT EXISTS calendar_id text,
    ADD COLUMN IF NOT EXISTS title text,
    ADD COLUMN IF NOT EXISTS is_private bool NOT NULL DEFAULT true;
