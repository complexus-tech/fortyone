ALTER TABLE public.team_sprint_settings
    DROP CONSTRAINT IF EXISTS team_sprint_settings_working_days_check,
    DROP COLUMN IF EXISTS working_days;
