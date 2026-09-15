-- Historical status repairs and closed surviving feedback remain intact:
-- rollback must not recreate invalid references or reopen merged feedback.
ALTER TABLE public.stories
    DROP CONSTRAINT stories_team_status_fkey,
    ADD CONSTRAINT stories_status_id_fkey
        FOREIGN KEY (status_id) REFERENCES public.statuses(status_id);

ALTER TABLE public.statuses DROP CONSTRAINT statuses_team_status_key;

DROP TRIGGER feedback_items_close_deleted_merge_sources ON public.feedback_items;
DROP FUNCTION public.close_feedback_sources_of_deleted_merge_targets();

ALTER TABLE public.labels
    DROP CONSTRAINT labels_team_id_fkey,
    ADD CONSTRAINT labels_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id);

ALTER TABLE public.objectives
    DROP CONSTRAINT objectives_team_id_fkey,
    ADD CONSTRAINT objectives_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id);

CREATE OR REPLACE FUNCTION public.enqueue_deleted_maya_schedule_event()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.source <> 'maya' OR OLD.external_event_id IS NULL THEN
        RETURN OLD;
    END IF;

    UPDATE public.calendar_schedule_event_outbox
    SET processed_at = CURRENT_TIMESTAMP,
        last_error = 'Superseded by schedule block deletion.',
        updated_at = CURRENT_TIMESTAMP
    WHERE schedule_block_id = OLD.block_id
      AND processed_at IS NULL;

    INSERT INTO public.calendar_schedule_event_outbox (
        workspace_id, user_id, schedule_block_id, operation, provider,
        calendar_id, provider_event_id, payload, dedupe_key
    ) VALUES (
        OLD.workspace_id,
        OLD.user_id,
        OLD.block_id,
        'delete',
        'google',
        COALESCE(OLD.external_calendar_id, 'primary'),
        OLD.external_event_id,
        jsonb_build_object(
            'CalendarID', COALESCE(OLD.external_calendar_id, 'primary'),
            'EventID', OLD.external_event_id,
            'BlockID', OLD.block_id,
            'StoryID', OLD.story_id,
            'WorkspaceID', OLD.workspace_id
        ),
        CONCAT('delete:', OLD.external_event_id, ':')
    )
    ON CONFLICT (dedupe_key) DO UPDATE SET
        workspace_id = EXCLUDED.workspace_id,
        user_id = EXCLUDED.user_id,
        schedule_block_id = EXCLUDED.schedule_block_id,
        calendar_id = EXCLUDED.calendar_id,
        provider_event_id = EXCLUDED.provider_event_id,
        payload = EXCLUDED.payload,
        processed_at = NULL,
        dead_lettered_at = NULL,
        attempt_count = 0,
        last_error = NULL,
        available_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP;

    RETURN OLD;
END;
$$;
