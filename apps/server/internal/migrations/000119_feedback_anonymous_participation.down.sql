DROP INDEX IF EXISTS public.idx_feedback_items_portal_contributor_created;

ALTER TABLE public.feedback_items
    DROP CONSTRAINT IF EXISTS feedback_items_contributor_fkey,
    DROP COLUMN IF EXISTS contributor_id;

DROP TABLE IF EXISTS public.feedback_contributors;

ALTER TABLE public.feedback_portals
    DROP CONSTRAINT IF EXISTS feedback_portals_participation_mode_check,
    DROP COLUMN IF EXISTS participation_mode;
