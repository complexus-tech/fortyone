DROP INDEX IF EXISTS public.idx_team_members_inferred_work_focus_missing;

ALTER TABLE public.team_members
	DROP COLUMN IF EXISTS inferred_ai_role_generated_at,
	DROP COLUMN IF EXISTS inferred_ai_role_confidence,
	DROP COLUMN IF EXISTS inferred_ai_role_story_count,
	DROP COLUMN IF EXISTS inferred_ai_role_description,
	DROP COLUMN IF EXISTS inferred_ai_role_title;
