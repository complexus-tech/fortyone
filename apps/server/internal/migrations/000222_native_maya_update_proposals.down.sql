DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM public.messaging_story_mutation_confirmations
        WHERE operation = 'update_story' AND proposal IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Native Maya update proposals remain; cancel or finish them before rollback';
    END IF;
END $$;

ALTER TABLE public.messaging_story_mutation_confirmations
    DROP CONSTRAINT messaging_story_mutation_confirmations_proposal_check;

ALTER TABLE public.messaging_story_mutation_confirmations
    ADD CONSTRAINT messaging_story_mutation_confirmations_proposal_check
        CHECK (
            (operation <> 'create_stories' AND proposal IS NULL)
            OR (
                operation = 'create_stories'
                AND (
                    (status IN ('pending', 'applied') AND proposal IS NOT NULL AND jsonb_typeof(proposal) = 'object')
                    OR (status = 'applied' AND proposal IS NULL AND result IS NOT NULL AND last_error IS NULL)
                    OR (status IN ('cancelled', 'expired') AND proposal IS NULL)
                )
            )
        );
