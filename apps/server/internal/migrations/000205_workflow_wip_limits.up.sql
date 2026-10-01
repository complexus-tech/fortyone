ALTER TABLE public.statuses
    ADD COLUMN wip_limit integer,
    ADD CONSTRAINT statuses_wip_limit_range CHECK (wip_limit BETWEEN 1 AND 10000);

CREATE INDEX stories_active_status_count_idx ON public.stories (workspace_id, status_id)
    WHERE deleted_at IS NULL AND archived_at IS NULL AND is_draft = FALSE;
