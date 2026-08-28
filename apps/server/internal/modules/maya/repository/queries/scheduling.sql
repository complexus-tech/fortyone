-- name: LockMayaStorySchedule :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'maya-schedule:'
            || CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text)
            || ':'
            || CAST(CAST(sqlc.arg(story_id) AS uuid) AS text),
        0
    )
);

-- name: ListScheduleStoryRefsForUser :many
WITH scheduled_stories AS (
    SELECT ownership.workspace_id, ownership.story_id
    FROM calendar_maya_schedule_ownerships ownership
    WHERE ownership.user_id = sqlc.arg(user_id)

    UNION

    SELECT block.workspace_id, block.story_id
    FROM calendar_schedule_blocks block
    WHERE block.user_id = sqlc.arg(user_id)
        AND block.source = 'maya'
        AND block.story_id IS NOT NULL
        AND block.completed_at IS NULL

    UNION

    SELECT story.workspace_id, story.id AS story_id
    FROM stories story
    WHERE story.assignee_id = sqlc.arg(user_id)
        AND story.auto_scheduling_enabled = TRUE
)
SELECT scheduled.workspace_id, scheduled.story_id
FROM scheduled_stories scheduled
INNER JOIN stories story
    ON story.id = scheduled.story_id
    AND story.workspace_id = scheduled.workspace_id
ORDER BY
    story.end_date NULLS LAST,
    CASE story.priority
        WHEN 'Urgent' THEN 0
        WHEN 'High' THEN 1
        WHEN 'Medium' THEN 2
        WHEN 'Low' THEN 3
        WHEN 'No Priority' THEN 4
        ELSE 5
    END,
    scheduled.story_id;

-- name: ClaimScheduleRecoveryStoryRefs :many
WITH candidates AS (
    SELECT
        ownership.workspace_id,
        ownership.story_id,
        interrupted.run_id AS interrupted_run_id
    FROM calendar_maya_schedule_ownerships ownership
    INNER JOIN stories story
        ON story.id = ownership.story_id
        AND story.workspace_id = ownership.workspace_id
    INNER JOIN workspaces workspace
        ON workspace.workspace_id = story.workspace_id
    LEFT JOIN statuses status
        ON status.status_id = story.status_id
    LEFT JOIN users owner_user
        ON owner_user.user_id = ownership.user_id
    LEFT JOIN workspace_members workspace_member
        ON workspace_member.workspace_id = ownership.workspace_id
        AND workspace_member.user_id = ownership.user_id
    LEFT JOIN team_members team_member
        ON team_member.team_id = story.team_id
        AND team_member.user_id = ownership.user_id
    LEFT JOIN team_sprint_settings team_settings
        ON team_settings.team_id = story.team_id
        AND team_settings.workspace_id = story.workspace_id
    LEFT JOIN sprints sprint
        ON sprint.sprint_id = story.sprint_id
    LEFT JOIN LATERAL (
        SELECT run.run_id
        FROM maya_agent_runs run
        WHERE run.workspace_id = ownership.workspace_id
            AND run.story_id = ownership.story_id
            AND run.status = 'running'
            AND run.updated_at <= sqlc.arg(interrupted_run_before)
        ORDER BY run.updated_at, run.run_id
        LIMIT 1
    ) interrupted ON TRUE
    WHERE COALESCE(ownership.recovery_attempted_at, TIMESTAMP 'epoch')
            <= sqlc.arg(retry_before)
        AND (
            interrupted.run_id IS NOT NULL
            OR story.updated_at > ownership.updated_at
            OR team_settings.updated_at > ownership.updated_at
            OR sprint.updated_at > ownership.updated_at
            OR (story.assignee_id IS NOT NULL AND story.assignee_id <> ownership.user_id)
            OR story.deleted_at IS NOT NULL
            OR story.archived_at IS NOT NULL
            OR story.completed_at IS NOT NULL
            OR story.auto_scheduling_enabled = FALSE
            OR NOT (
                workspace.deleted_at IS NULL
                AND (
                    workspace.trial_ends_on > CURRENT_TIMESTAMP
                    OR EXISTS (
                        SELECT 1
                        FROM workspace_subscriptions subscription
                        WHERE subscription.workspace_id = workspace.workspace_id
                            AND subscription.subscription_tier <> 'free'
                            AND subscription.subscription_status IN ('active', 'trialing', 'past_due')
                    )
                )
            )
            OR status.status_id IS NULL
            OR status.category IN ('completed', 'cancelled')
            OR owner_user.user_id IS NULL
            OR owner_user.is_active = FALSE
            OR workspace_member.user_id IS NULL
            OR team_member.user_id IS NULL
            OR (
                story.auto_scheduling_enabled = TRUE
                AND (
                    story.auto_scheduling_status = 'planning'
                    OR (
                        story.auto_scheduling_status = 'cannot_fit'
                        AND COALESCE(ownership.recovery_attempted_at, TIMESTAMP 'epoch')
                            <= CURRENT_TIMESTAMP - INTERVAL '1 hour'
                        AND NOT EXISTS (
                            SELECT 1
                            FROM calendar_schedule_blocks retry_block
                            WHERE retry_block.workspace_id = ownership.workspace_id
                                AND retry_block.story_id = ownership.story_id
                                AND retry_block.source = 'maya'
                                AND retry_block.completed_at IS NULL
                                AND retry_block.end_at > CURRENT_TIMESTAMP
                        )
                    )
                    OR (
                        story.auto_scheduling_status IN ('scheduled', 'locked')
                        AND EXISTS (
                            SELECT 1
                            FROM calendar_schedule_blocks elapsed_block
                            WHERE elapsed_block.workspace_id = ownership.workspace_id
                                AND elapsed_block.story_id = ownership.story_id
                                AND elapsed_block.source = 'maya'
                                AND elapsed_block.completed_at IS NULL
                        )
                        AND NOT EXISTS (
                            SELECT 1
                            FROM calendar_schedule_blocks future_block
                            WHERE future_block.workspace_id = ownership.workspace_id
                                AND future_block.story_id = ownership.story_id
                                AND future_block.source = 'maya'
                                AND future_block.completed_at IS NULL
                                AND future_block.end_at > CURRENT_TIMESTAMP
                        )
                    )
                )
            )
            OR EXISTS (
                SELECT 1
                FROM calendar_schedule_blocks block
                WHERE block.workspace_id = ownership.workspace_id
                    AND block.story_id = ownership.story_id
                    AND block.source = 'maya'
                    AND block.completed_at IS NULL
                    AND (
                        block.user_id <> ownership.user_id
                        OR (story.assignee_id IS NOT NULL AND block.user_id <> story.assignee_id)
                    )
            )
        )
    ORDER BY
        CASE WHEN interrupted.run_id IS NULL THEN 1 ELSE 0 END,
        ownership.updated_at,
        ownership.workspace_id,
        ownership.story_id
    LIMIT CAST(sqlc.arg(row_limit) AS integer)
    FOR UPDATE OF ownership SKIP LOCKED
), claimed AS (
    UPDATE calendar_maya_schedule_ownerships ownership
    SET recovery_attempted_at = CURRENT_TIMESTAMP
    FROM candidates
    WHERE ownership.workspace_id = candidates.workspace_id
        AND ownership.story_id = candidates.story_id
    RETURNING ownership.workspace_id, ownership.story_id
)
SELECT
    claimed.workspace_id,
    claimed.story_id,
    interrupted_run.run_id AS interrupted_run_id
FROM claimed
INNER JOIN candidates
    ON candidates.workspace_id = claimed.workspace_id
    AND candidates.story_id = claimed.story_id
LEFT JOIN maya_agent_runs interrupted_run
    ON interrupted_run.run_id = candidates.interrupted_run_id
ORDER BY claimed.workspace_id, claimed.story_id;

-- name: FailInterruptedScheduleActions :execrows
UPDATE maya_agent_actions
SET status = 'failed',
    error_message = sqlc.arg(error_message),
    updated_at = CURRENT_TIMESTAMP
WHERE run_id = sqlc.arg(run_id)
    AND status = 'proposed';

-- name: CompleteInterruptedScheduleRun :execrows
UPDATE maya_agent_runs run
SET status = CASE
        WHEN EXISTS (
            SELECT 1
            FROM maya_agent_actions action
            WHERE action.run_id = run.run_id
                AND action.status = 'failed'
        ) THEN 'failed'
        ELSE 'succeeded'
    END,
    summary = CASE
        WHEN EXISTS (
            SELECT 1
            FROM maya_agent_actions action
            WHERE action.run_id = run.run_id
                AND action.status = 'failed'
        ) THEN 'Maya recovered an interrupted scheduling operation against the story''s current state.'
        ELSE 'Maya verified the completed scheduling operation after worker recovery.'
    END,
    error_message = CASE
        WHEN EXISTS (
            SELECT 1
            FROM maya_agent_actions action
            WHERE action.run_id = run.run_id
                AND action.status = 'failed'
        ) THEN sqlc.arg(error_message)
        ELSE NULL
    END,
    completed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE run.run_id = sqlc.arg(run_id)
    AND run.status = 'running';

-- name: ListMayaScheduleOwners :many
SELECT ownership.user_id
FROM calendar_maya_schedule_ownerships ownership
WHERE ownership.workspace_id = sqlc.arg(workspace_id)
    AND ownership.story_id = sqlc.arg(story_id)

UNION

SELECT block.user_id
FROM calendar_schedule_blocks block
WHERE block.workspace_id = sqlc.arg(workspace_id)
    AND block.story_id = sqlc.arg(story_id)
    AND block.source = 'maya'
    AND block.completed_at IS NULL
ORDER BY user_id;

-- name: StoryIsSchedulableForUser :one
SELECT EXISTS (
    SELECT 1
    FROM stories story
    INNER JOIN statuses status
        ON status.status_id = story.status_id
    INNER JOIN users assignee
        ON assignee.user_id = sqlc.arg(user_id)
        AND assignee.is_active = TRUE
    INNER JOIN workspace_members membership
        ON membership.workspace_id = story.workspace_id
        AND membership.user_id = sqlc.arg(user_id)
    INNER JOIN team_members team_membership
        ON team_membership.team_id = story.team_id
        AND team_membership.user_id = sqlc.arg(user_id)
    WHERE story.workspace_id = sqlc.arg(workspace_id)
        AND story.id = sqlc.arg(story_id)
        AND story.assignee_id = sqlc.arg(user_id)
        AND story.auto_scheduling_enabled = TRUE
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND story.completed_at IS NULL
        AND status.category NOT IN ('completed', 'cancelled')
);

-- name: StoryIsActiveForAutoScheduling :one
SELECT EXISTS (
    SELECT 1
    FROM stories story
    INNER JOIN statuses status
        ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)
        AND story.id = sqlc.arg(story_id)
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND story.completed_at IS NULL
        AND status.category NOT IN ('completed', 'cancelled')
);

-- name: StoryScheduleOwnershipIsRetainable :one
SELECT EXISTS (
    SELECT 1
    FROM stories story
    INNER JOIN statuses status
        ON status.status_id = story.status_id
    INNER JOIN users owner_user
        ON owner_user.user_id = sqlc.arg(user_id)
        AND owner_user.is_active = TRUE
    INNER JOIN workspace_members workspace_member
        ON workspace_member.workspace_id = story.workspace_id
        AND workspace_member.user_id = sqlc.arg(user_id)
    INNER JOIN team_members team_member
        ON team_member.team_id = story.team_id
        AND team_member.user_id = sqlc.arg(user_id)
    WHERE story.workspace_id = sqlc.arg(workspace_id)
        AND story.id = sqlc.arg(story_id)
        AND story.deleted_at IS NULL
        AND story.archived_at IS NULL
        AND story.completed_at IS NULL
        AND status.category NOT IN ('completed', 'cancelled')
        AND (story.assignee_id IS NULL OR story.assignee_id = sqlc.arg(user_id))
);
