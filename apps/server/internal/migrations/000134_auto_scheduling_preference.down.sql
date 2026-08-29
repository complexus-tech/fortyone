ALTER TABLE public.user_automation_preferences
	RENAME COLUMN auto_scheduling TO auto_assign_maya;

ALTER TABLE public.user_automation_preferences
	ALTER COLUMN auto_assign_maya SET DEFAULT FALSE;
