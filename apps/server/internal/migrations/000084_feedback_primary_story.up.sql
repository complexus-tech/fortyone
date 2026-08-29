ALTER TABLE public.feedback_story_links
    ADD COLUMN is_primary boolean NOT NULL DEFAULT false;

WITH ranked_links AS (
    SELECT
        link.id,
        ROW_NUMBER() OVER (
            PARTITION BY link.item_id
            ORDER BY
                CASE WHEN story.deleted_at IS NULL THEN 0 ELSE 1 END,
                CASE link.relationship
                    WHEN 'created_from' THEN 0
                    WHEN 'solves' THEN 1
                    ELSE 2
                END,
                link.created_at ASC,
                link.id ASC
        ) AS position
    FROM public.feedback_story_links AS link
    INNER JOIN public.stories AS story ON story.id = link.story_id
    WHERE link.relationship IN ('created_from', 'solves')
)
UPDATE public.feedback_story_links AS link
SET is_primary = true
FROM ranked_links
WHERE link.id = ranked_links.id
  AND ranked_links.position = 1;

ALTER TABLE public.feedback_story_links
    ADD CONSTRAINT feedback_story_links_primary_relationship_check
    CHECK (NOT is_primary OR relationship IN ('created_from', 'solves'));

CREATE UNIQUE INDEX feedback_story_links_one_primary_per_item
    ON public.feedback_story_links (item_id)
    WHERE is_primary = true;

CREATE INDEX feedback_story_links_primary_story
    ON public.feedback_story_links (story_id)
    WHERE is_primary = true;
