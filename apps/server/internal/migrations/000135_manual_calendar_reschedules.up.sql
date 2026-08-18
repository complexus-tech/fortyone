ALTER TABLE public.calendar_schedule_blocks
    ADD COLUMN manual_override_at timestamptz,
    ADD COLUMN manual_override_by uuid;

ALTER TABLE public.calendar_schedule_blocks
    ADD CONSTRAINT calendar_schedule_blocks_manual_override_by_fkey
        FOREIGN KEY (manual_override_by) REFERENCES public.users(user_id) ON DELETE SET NULL;

CREATE INDEX idx_calendar_schedule_blocks_manual_override
    ON public.calendar_schedule_blocks (workspace_id, user_id, manual_override_at)
    WHERE manual_override_at IS NOT NULL;

CREATE TABLE public.calendar_schedule_reschedule_events (
    event_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    story_id uuid,
    schedule_block_id uuid,
    action varchar(32) NOT NULL,
    source varchar(32) NOT NULL,
    timezone varchar(128) NOT NULL DEFAULT 'UTC',
    previous_start_at timestamptz NOT NULL,
    previous_end_at timestamptz NOT NULL,
    next_start_at timestamptz NOT NULL,
    next_end_at timestamptz NOT NULL,
    client_mutation_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_schedule_reschedule_events_pkey PRIMARY KEY (event_id),
    CONSTRAINT calendar_schedule_reschedule_events_workspace_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT calendar_schedule_reschedule_events_user_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT calendar_schedule_reschedule_events_story_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE SET NULL,
    CONSTRAINT calendar_schedule_reschedule_events_block_fkey
        FOREIGN KEY (schedule_block_id) REFERENCES public.calendar_schedule_blocks(block_id) ON DELETE SET NULL,
    CONSTRAINT calendar_schedule_reschedule_events_action_check
        CHECK (action IN ('move', 'resize')),
    CONSTRAINT calendar_schedule_reschedule_events_source_check
        CHECK (source IN ('user', 'maya')),
    CONSTRAINT calendar_schedule_reschedule_events_range_check
        CHECK (next_end_at > next_start_at),
    CONSTRAINT calendar_schedule_reschedule_events_client_mutation_key
        UNIQUE (client_mutation_id)
);

CREATE INDEX idx_calendar_schedule_reschedule_events_user_created
    ON public.calendar_schedule_reschedule_events (workspace_id, user_id, created_at DESC);

CREATE INDEX idx_calendar_schedule_reschedule_events_story_created
    ON public.calendar_schedule_reschedule_events (workspace_id, story_id, created_at DESC)
    WHERE story_id IS NOT NULL;
