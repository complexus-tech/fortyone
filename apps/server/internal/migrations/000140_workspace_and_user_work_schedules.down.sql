ALTER TABLE public.users
    DROP CONSTRAINT IF EXISTS users_working_hours_check,
    DROP CONSTRAINT IF EXISTS users_working_days_check,
    DROP COLUMN IF EXISTS working_end_minute,
    DROP COLUMN IF EXISTS working_start_minute,
    DROP COLUMN IF EXISTS working_days;

ALTER TABLE public.workspace_settings
    DROP CONSTRAINT IF EXISTS workspace_settings_working_hours_check,
    DROP CONSTRAINT IF EXISTS workspace_settings_working_days_check,
    DROP COLUMN IF EXISTS working_end_minute,
    DROP COLUMN IF EXISTS working_start_minute,
    DROP COLUMN IF EXISTS working_days;
