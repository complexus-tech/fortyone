ALTER TABLE public.stories
    ADD COLUMN auto_scheduling_enabled boolean NOT NULL DEFAULT false,
    ADD COLUMN auto_scheduling_locked boolean NOT NULL DEFAULT false,
    ADD COLUMN auto_scheduling_status text NOT NULL DEFAULT 'off',
    ADD COLUMN auto_scheduling_reason text,
    ADD COLUMN auto_scheduling_updated_at timestamptz,
    ADD CONSTRAINT stories_auto_scheduling_locked_requires_enabled_check
        CHECK (NOT auto_scheduling_locked OR auto_scheduling_enabled),
    ADD CONSTRAINT stories_auto_scheduling_status_check
        CHECK (
            auto_scheduling_status IN (
                'off',
                'needs_owner',
                'needs_time',
                'planning',
                'scheduled',
                'at_risk',
                'cannot_fit',
                'locked'
            )
        );

-- Preserve the scheduling enrollment that predates the story-level contract.
-- A Maya block is definitive evidence of a schedule. Ownership without a
-- block preserves the best state that can be inferred without inventing a
-- historical planning result. Active stories still assigned to Maya remain
-- enrolled so the assignment batch can select their human owner.
WITH maya_actor AS (
    SELECT actor.user_id
    FROM public.users AS actor
    WHERE actor.email = 'maya@fortyone.app'
        AND actor.is_system = TRUE
), managed_stories AS (
    SELECT
        story.id,
        story.assignee_id,
        story.estimated_duration_minutes,
        EXISTS (
            SELECT 1
            FROM public.calendar_schedule_blocks AS block
            WHERE block.workspace_id = story.workspace_id
                AND block.story_id = story.id
                AND block.source = 'maya'
        ) AS has_maya_blocks,
        EXISTS (
            SELECT 1
            FROM maya_actor
            WHERE maya_actor.user_id = story.assignee_id
        ) AS assigned_to_maya
    FROM public.stories AS story
    WHERE EXISTS (
            SELECT 1
            FROM public.calendar_maya_schedule_ownerships AS ownership
            WHERE ownership.workspace_id = story.workspace_id
                AND ownership.story_id = story.id
        )
        OR EXISTS (
            SELECT 1
            FROM public.calendar_schedule_blocks AS block
            WHERE block.workspace_id = story.workspace_id
                AND block.story_id = story.id
                AND block.source = 'maya'
        )
        OR (
            story.deleted_at IS NULL
            AND story.archived_at IS NULL
            AND story.completed_at IS NULL
            AND story.is_draft = FALSE
            AND EXISTS (
                SELECT 1
                FROM maya_actor
                WHERE maya_actor.user_id = story.assignee_id
            )
            AND NOT EXISTS (
                SELECT 1
                FROM public.statuses AS status
                WHERE status.status_id = story.status_id
                    AND status.category IN ('completed', 'cancelled')
            )
        )
), classified_stories AS (
    SELECT
        managed.id,
        CASE
            WHEN managed.has_maya_blocks THEN 'scheduled'
            WHEN managed.assignee_id IS NULL OR managed.assigned_to_maya THEN 'needs_owner'
            WHEN managed.estimated_duration_minutes IS NULL THEN 'needs_time'
            ELSE 'planning'
        END AS scheduling_status,
        CASE
            WHEN managed.has_maya_blocks THEN NULL
            WHEN managed.assignee_id IS NULL THEN 'Choose an owner before Maya can schedule this story.'
            WHEN managed.assigned_to_maya THEN 'Maya is selecting an eligible owner for this story.'
            WHEN managed.estimated_duration_minutes IS NULL THEN 'Add a time estimate before Maya can schedule this story.'
            ELSE 'Maya is checking availability and scheduling this story.'
        END AS scheduling_reason
    FROM managed_stories AS managed
)
UPDATE public.stories AS story
SET auto_scheduling_enabled = TRUE,
    auto_scheduling_status = classified.scheduling_status,
    auto_scheduling_reason = classified.scheduling_reason,
    auto_scheduling_updated_at = CURRENT_TIMESTAMP
FROM classified_stories AS classified
WHERE story.id = classified.id;
