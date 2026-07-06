-- 000072_estimation_default_hours.up.sql

ALTER TABLE public.team_estimation_settings
    ALTER COLUMN scheme SET DEFAULT 'hours';
