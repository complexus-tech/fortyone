-- These state queries run only after AuthorizeSecondaryStoryTargets has
-- locked and authorized the story in the same transaction. Keeping the state
-- SQL small makes the compare-and-swap contract easy to review.

-- name: MayaScheduleBlocksExist :one
SELECT EXISTS (
    SELECT 1
    FROM public.calendar_schedule_blocks AS schedule_block
    INNER JOIN public.stories AS story
        ON story.id = schedule_block.story_id
       AND story.workspace_id = schedule_block.workspace_id
    WHERE schedule_block.story_id = sqlc.arg(story_id)
      AND schedule_block.workspace_id = sqlc.arg(workspace_id)
      AND schedule_block.source = 'maya'
      AND schedule_block.completed_at IS NULL
      AND story.deleted_at IS NULL
);

-- name: UpdateAutoSchedulingState :execrows
UPDATE public.stories AS story
SET
    auto_scheduling_status = sqlc.arg(auto_scheduling_status),
    auto_scheduling_reason = sqlc.narg(auto_scheduling_reason),
    auto_scheduling_updated_at = sqlc.arg(auto_scheduling_updated_at),
    auto_scheduling_locked = CASE
        WHEN CAST(sqlc.arg(set_auto_scheduling_locked) AS boolean)
            THEN CAST(sqlc.arg(auto_scheduling_locked) AS boolean)
        ELSE story.auto_scheduling_locked
    END
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.updated_at = sqlc.arg(expected_updated_at)
  AND story.deleted_at IS NULL;
