DROP TABLE IF EXISTS public.slack_file_imports;

DELETE FROM public.messaging_story_mutation_confirmations
WHERE operation = 'attach_story_file';

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
        ));

ALTER TABLE public.attachments
    DROP COLUMN IF EXISTS slack_file_import_id;
