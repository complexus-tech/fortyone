ALTER TABLE public.feedback_votes
    DROP CONSTRAINT IF EXISTS feedback_votes_direction_check;

ALTER TABLE public.feedback_votes
    DROP COLUMN IF EXISTS direction;
