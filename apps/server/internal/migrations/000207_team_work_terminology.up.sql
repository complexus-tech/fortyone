ALTER TABLE teams
    ADD COLUMN story_term text,
    ADD CONSTRAINT teams_story_term_check
        CHECK (story_term IS NULL OR story_term IN ('story', 'task', 'issue', 'ticket', 'work item', 'deal'));
