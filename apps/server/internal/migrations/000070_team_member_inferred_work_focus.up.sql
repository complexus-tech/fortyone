ALTER TABLE public.team_members
	ADD COLUMN inferred_ai_role_title text NOT NULL DEFAULT '',
	ADD COLUMN inferred_ai_role_description text NOT NULL DEFAULT '',
	ADD COLUMN inferred_ai_role_story_count integer NOT NULL DEFAULT 0,
	ADD COLUMN inferred_ai_role_confidence real NOT NULL DEFAULT 0,
	ADD COLUMN inferred_ai_role_generated_at timestamptz;

CREATE INDEX idx_team_members_inferred_work_focus_missing
	ON public.team_members (team_id, user_id)
	WHERE ai_role_title = ''
		AND ai_role_description = '';
