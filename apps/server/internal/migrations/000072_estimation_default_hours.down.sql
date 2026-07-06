-- 000072_estimation_default_hours.down.sql

ALTER TABLE public.team_estimation_settings
    ALTER COLUMN scheme SET DEFAULT 'points';
