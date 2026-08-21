DROP INDEX IF EXISTS public.calendar_connections_one_primary_per_account;

ALTER TABLE public.calendar_connections
    DROP COLUMN IF EXISTS is_primary;
