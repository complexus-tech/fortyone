DROP INDEX IF EXISTS public.idx_sprints_automation_managed_schedule;

ALTER TABLE public.sprints
    DROP COLUMN IF EXISTS schedule_managed_by_automation;
