DELETE FROM public.messaging_story_mutation_confirmations
WHERE operation NOT IN ('create_story', 'update_story');

ALTER TABLE public.messaging_story_mutation_confirmations
    DROP CONSTRAINT messaging_story_mutation_confirmations_proposal_check,
    DROP CONSTRAINT messaging_story_mutation_confirmations_operation_check,
    DROP COLUMN proposal;

ALTER TABLE public.messaging_story_mutation_confirmations
    ADD CONSTRAINT messaging_story_mutation_confirmations_operation_check
        CHECK (operation IN ('create_story', 'update_story'));
