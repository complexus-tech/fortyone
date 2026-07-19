ALTER TABLE public.feedback_votes
    ADD COLUMN direction smallint NOT NULL DEFAULT 1;

ALTER TABLE public.feedback_votes
    ADD CONSTRAINT feedback_votes_direction_check
    CHECK (direction IN (-1, 1));
