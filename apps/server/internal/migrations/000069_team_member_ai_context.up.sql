ALTER TABLE public.team_members
	ADD COLUMN ai_role_title text NOT NULL DEFAULT '',
	ADD COLUMN ai_role_description text NOT NULL DEFAULT '';

ALTER TABLE public.user_automation_preferences
	ADD COLUMN auto_assign_maya bool NOT NULL DEFAULT false;
