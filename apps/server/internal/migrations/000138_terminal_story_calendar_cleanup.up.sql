-- Calendar time is an active-work projection. Once a story is terminal, its
-- task blocks must stop reserving the assignee's account-wide availability.
-- The existing schedule-block delete trigger records provider cleanup before
-- Google-backed Maya blocks are removed.
DELETE FROM public.calendar_schedule_blocks AS block
USING public.stories AS story
INNER JOIN public.statuses AS status ON status.status_id = story.status_id
WHERE block.story_id = story.id
    AND block.workspace_id = story.workspace_id
    AND block.block_type = 'work'
    AND (
        story.completed_at IS NOT NULL
        OR status.category IN ('completed', 'cancelled')
    );

CREATE FUNCTION public.cleanup_terminal_story_calendar_schedule()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL
       OR EXISTS (
           SELECT 1
           FROM public.statuses AS status
           WHERE status.status_id = NEW.status_id
               AND status.category IN ('completed', 'cancelled')
       ) THEN
        DELETE FROM public.calendar_schedule_blocks
        WHERE story_id = NEW.id
            AND workspace_id = NEW.workspace_id
            AND block_type = 'work';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER stories_cleanup_terminal_calendar_schedule
AFTER UPDATE OF status_id, completed_at ON public.stories
FOR EACH ROW
EXECUTE FUNCTION public.cleanup_terminal_story_calendar_schedule();
