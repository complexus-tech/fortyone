-- 000106_calendar_events.down.sql

DROP TABLE IF EXISTS public.calendar_events;

ALTER TABLE public.calendar_schedule_blocks
    DROP CONSTRAINT IF EXISTS calendar_schedule_blocks_workspace_member_fkey,
    DROP CONSTRAINT IF EXISTS calendar_schedule_blocks_story_workspace_fkey,
    ADD CONSTRAINT calendar_schedule_blocks_story_id_fkey
        FOREIGN KEY (story_id)
        REFERENCES public.stories (id)
        ON DELETE SET NULL;

ALTER TABLE public.stories
    DROP CONSTRAINT IF EXISTS stories_id_workspace_unique;

ALTER TABLE public.calendar_connections
    DROP CONSTRAINT IF EXISTS calendar_connections_workspace_member_fkey;

ALTER TABLE public.calendar_connections
    DROP CONSTRAINT IF EXISTS calendar_connections_scope_unique;

ALTER TABLE public.calendar_connections
    DROP COLUMN IF EXISTS credential_generation,
    DROP COLUMN IF EXISTS provider_account_id;
