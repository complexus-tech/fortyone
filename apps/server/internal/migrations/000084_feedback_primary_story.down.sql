DROP INDEX IF EXISTS public.feedback_story_links_primary_story;
DROP INDEX IF EXISTS public.feedback_story_links_one_primary_per_item;

ALTER TABLE public.feedback_story_links
    DROP CONSTRAINT IF EXISTS feedback_story_links_primary_relationship_check,
    DROP COLUMN IF EXISTS is_primary;
