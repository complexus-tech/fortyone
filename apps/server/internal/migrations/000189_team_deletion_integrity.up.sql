-- Provider cleanup must remain accurate during cascades and the status repair
-- below, which can itself retire calendar blocks through existing triggers.
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
        OLD.external_provider,
        OLD.external_calendar_id,
        OLD.external_event_id,
        jsonb_build_object(
            'CalendarID', OLD.external_calendar_id,
            'EventID', OLD.external_event_id,
            'BlockID', OLD.block_id,
            'StoryID', OLD.story_id,
            'WorkspaceID', OLD.workspace_id
        ),
        CONCAT(OLD.external_provider, ':delete:', OLD.block_id, ':')
    )
    ON CONFLICT (dedupe_key) DO UPDATE SET
        workspace_id = EXCLUDED.workspace_id,
        user_id = EXCLUDED.user_id,
        schedule_block_id = EXCLUDED.schedule_block_id,
        provider = EXCLUDED.provider,
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

ALTER TABLE public.objectives
    DROP CONSTRAINT objectives_team_id_fkey,
    ADD CONSTRAINT objectives_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;

ALTER TABLE public.labels
    DROP CONSTRAINT labels_team_id_fkey,
    ADD CONSTRAINT labels_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;

-- A merge can cross team boards. Deleting its target must preserve the other
-- team's source item without returning it to an actionable state. The durable
-- merge outbox retains the original merge event and attribution unchanged.
-- A statement trigger only updates surviving rows, so deleting a source and
-- target together cannot modify a row still being deleted by that statement.
CREATE FUNCTION public.close_feedback_sources_of_deleted_merge_targets()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE public.feedback_items AS source
    SET merged_into_item_id = NULL,
        merged_at = NULL,
        merged_by_user_id = NULL,
        status = 'closed',
        updated_at = CURRENT_TIMESTAMP
    FROM deleted_feedback_items AS target
    WHERE source.workspace_id = target.workspace_id
      AND source.portal_id = target.portal_id
      AND source.merged_into_item_id = target.id;

    RETURN NULL;
END;
$$;

CREATE TRIGGER feedback_items_close_deleted_merge_sources
AFTER DELETE ON public.feedback_items
REFERENCING OLD TABLE AS deleted_feedback_items
FOR EACH STATEMENT
EXECUTE FUNCTION public.close_feedback_sources_of_deleted_merge_targets();

-- Preserve every story and prefer the owning team's equivalent status category
-- before its default. Stable ordering makes historical repair deterministic.
-- A team without statuses leaves status_id NULL, which stories already allow.
UPDATE public.stories AS story
SET status_id = (
        SELECT replacement.status_id
        FROM public.statuses AS replacement
        WHERE replacement.team_id = story.team_id
          AND replacement.workspace_id = story.workspace_id
        ORDER BY
            (replacement.category IS NOT DISTINCT FROM previous_status.category) DESC,
            replacement.is_default DESC,
            replacement.order_index NULLS LAST,
            replacement.created_at,
            replacement.status_id
        LIMIT 1
    ),
    updated_at = CURRENT_TIMESTAMP
FROM public.statuses AS previous_status
WHERE story.status_id = previous_status.status_id
  AND story.team_id IS DISTINCT FROM previous_status.team_id;

-- Enforce this invariant at every writer, including duplication, imports and
-- team transfers. MATCH SIMPLE deliberately permits nullable story statuses.
-- NO ACTION lets same-statement team cascades delete stories and statuses in
-- either order without cascading through status references to another team.
ALTER TABLE public.statuses
    ADD CONSTRAINT statuses_team_status_key UNIQUE (team_id, status_id);

ALTER TABLE public.stories
    DROP CONSTRAINT stories_status_id_fkey,
    ADD CONSTRAINT stories_team_status_fkey
        FOREIGN KEY (team_id, status_id)
        REFERENCES public.statuses(team_id, status_id)
        ON DELETE NO ACTION;
