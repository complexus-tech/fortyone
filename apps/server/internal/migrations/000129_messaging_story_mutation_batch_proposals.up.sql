ALTER TABLE public.messaging_story_mutation_confirmations
    ADD COLUMN proposal jsonb;

ALTER TABLE public.messaging_story_mutation_confirmations
    DROP CONSTRAINT messaging_story_mutation_confirmations_operation_check;

ALTER TABLE public.messaging_story_mutation_confirmations
    ADD CONSTRAINT messaging_story_mutation_confirmations_operation_check
        CHECK (operation IN (
            'create_story',
            'create_stories',
            'update_story',
            'add_story_comment',
            'add_story_relationship'
        )),
    ADD CONSTRAINT messaging_story_mutation_confirmations_proposal_check
        CHECK (
            (operation <> 'create_stories' AND proposal IS NULL)
            OR (
                operation = 'create_stories'
                AND (
                    (
                        status IN ('pending', 'applied')
                        AND proposal IS NOT NULL
                        AND jsonb_typeof(proposal) = 'object'
                    )
                    OR (
                        status = 'applied'
                        AND proposal IS NULL
                        AND result IS NOT NULL
                        AND last_error IS NULL
                    )
                    OR (
                        status IN ('cancelled', 'expired')
                        AND proposal IS NULL
                    )
                )
            )
        );
