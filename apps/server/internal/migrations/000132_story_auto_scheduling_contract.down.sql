ALTER TABLE public.stories
    DROP CONSTRAINT IF EXISTS stories_auto_scheduling_status_check,
    DROP CONSTRAINT IF EXISTS stories_auto_scheduling_locked_requires_enabled_check,
    DROP COLUMN IF EXISTS auto_scheduling_updated_at,
    DROP COLUMN IF EXISTS auto_scheduling_reason,
    DROP COLUMN IF EXISTS auto_scheduling_status,
    DROP COLUMN IF EXISTS auto_scheduling_locked,
    DROP COLUMN IF EXISTS auto_scheduling_enabled;
