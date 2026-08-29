-- name: CreateCalendarScheduleBlock :one
INSERT INTO calendar_schedule_blocks (
    workspace_id,
    user_id,
    story_id,
    block_type,
    title,
    start_at,
    end_at,
    is_locked,
    source
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.narg(story_id),
    sqlc.arg(block_type),
    sqlc.arg(title),
    sqlc.arg(start_at),
    sqlc.arg(end_at),
    sqlc.arg(is_locked),
    sqlc.arg(source)
)
RETURNING block_id;

-- name: UpdateCalendarScheduleBlock :execrows
UPDATE calendar_schedule_blocks
SET story_id = sqlc.narg(story_id),
    block_type = sqlc.arg(block_type),
    title = sqlc.arg(title),
    start_at = sqlc.arg(start_at),
    end_at = sqlc.arg(end_at),
    is_locked = sqlc.arg(is_locked),
    source = sqlc.arg(source),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND block_id = sqlc.arg(block_id)
  AND source <> 'maya';

-- name: GetManualScheduleRescheduleBlockID :one
SELECT event.schedule_block_id
FROM calendar_schedule_reschedule_events event
WHERE event.client_mutation_id = sqlc.arg(client_mutation_id);

-- name: LockManualScheduleBlock :one
SELECT block.block_id,
       block.story_id,
       block.block_type,
       block.title,
       block.start_at,
       block.end_at,
       block.is_locked,
       block.source,
       block.external_provider,
       block.external_calendar_id,
       block.external_event_id,
       block.updated_at
FROM calendar_schedule_blocks block
WHERE block.workspace_id = sqlc.arg(workspace_id)
  AND block.user_id = sqlc.arg(user_id)
  AND block.block_id = sqlc.arg(block_id)
  AND block.completed_at IS NULL
FOR UPDATE;

-- name: LockScheduleStoryTime :one
SELECT story.estimated_duration_minutes,
       story.minimum_focus_block_minutes,
       story.auto_scheduling_enabled
FROM stories story
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.id = sqlc.arg(story_id)
  AND story.deleted_at IS NULL
FOR UPDATE;

-- name: UpdateScheduleStoryEstimate :exec
UPDATE stories
SET estimated_duration_minutes = sqlc.arg(estimated_duration_minutes),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(story_id)
  AND deleted_at IS NULL;

-- name: ManuallyRescheduleCalendarBlock :execrows
UPDATE calendar_schedule_blocks
SET start_at = sqlc.arg(start_at),
    end_at = sqlc.arg(end_at),
    is_locked = TRUE,
    manual_override_at = CURRENT_TIMESTAMP,
    manual_override_by = sqlc.arg(actor_id),
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND block_id = sqlc.arg(block_id);

-- name: RecordManualCalendarReschedule :exec
INSERT INTO calendar_schedule_reschedule_events (
    workspace_id,
    user_id,
    story_id,
    schedule_block_id,
    action,
    source,
    timezone,
    previous_start_at,
    previous_end_at,
    next_start_at,
    next_end_at,
    client_mutation_id
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.narg(story_id),
    sqlc.arg(schedule_block_id),
    sqlc.arg(action),
    'user',
    sqlc.arg(timezone),
    sqlc.arg(previous_start_at),
    sqlc.arg(previous_end_at),
    sqlc.arg(next_start_at),
    sqlc.arg(next_end_at),
    sqlc.arg(client_mutation_id)
);

-- name: DeleteCalendarScheduleBlock :execrows
DELETE FROM calendar_schedule_blocks
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND block_id = sqlc.arg(block_id)
  AND source <> 'maya';
