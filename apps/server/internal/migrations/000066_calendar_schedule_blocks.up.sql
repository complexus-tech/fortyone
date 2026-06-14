-- 000066_calendar_schedule_blocks.up.sql

CREATE TABLE public.calendar_schedule_blocks (
    block_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    story_id uuid,
    block_type varchar(32) NOT NULL,
    title varchar(255) NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    is_locked bool NOT NULL DEFAULT true,
    source varchar(32) NOT NULL DEFAULT 'user',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_schedule_blocks_pkey PRIMARY KEY (block_id),
    CONSTRAINT calendar_schedule_blocks_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT calendar_schedule_blocks_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT calendar_schedule_blocks_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE SET NULL,
    CONSTRAINT calendar_schedule_blocks_type_check
        CHECK (block_type IN ('work', 'focus')),
    CONSTRAINT calendar_schedule_blocks_source_check
        CHECK (source IN ('user', 'maya')),
    CONSTRAINT calendar_schedule_blocks_valid_range_check
        CHECK (end_at > start_at),
    CONSTRAINT calendar_schedule_blocks_work_story_check
        CHECK ((block_type = 'work' AND story_id IS NOT NULL) OR (block_type = 'focus'))
);

CREATE INDEX idx_calendar_schedule_blocks_workspace_user_range
    ON public.calendar_schedule_blocks (workspace_id, user_id, start_at, end_at);

CREATE INDEX idx_calendar_schedule_blocks_story
    ON public.calendar_schedule_blocks (workspace_id, story_id)
    WHERE story_id IS NOT NULL;
