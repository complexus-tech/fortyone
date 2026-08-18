ALTER TABLE public.user_automation_preferences
	RENAME COLUMN auto_assign_maya TO auto_scheduling;

ALTER TABLE public.user_automation_preferences
	ALTER COLUMN auto_scheduling SET DEFAULT TRUE;

UPDATE public.user_automation_preferences
	SET auto_scheduling = TRUE;
