-- 000063_sprint_automation_metadata.down.sql
DROP TABLE IF EXISTS public.audit_events;

ALTER TABLE public.team_sprint_settings
    DROP CONSTRAINT IF EXISTS team_sprint_settings_next_auto_sprint_number_check,
    DROP COLUMN IF EXISTS next_auto_sprint_number,
    DROP COLUMN IF EXISTS auto_create_disabled_at,
    DROP COLUMN IF EXISTS auto_create_disabled_reason;
