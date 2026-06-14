ALTER TABLE public.user_automation_preferences
	DROP COLUMN IF EXISTS auto_assign_maya;

ALTER TABLE public.team_members
	DROP COLUMN IF EXISTS ai_role_description,
	DROP COLUMN IF EXISTS ai_role_title;
