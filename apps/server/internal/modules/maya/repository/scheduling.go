package mayarepository

import (
	"context"
	"fmt"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
)

const claimScheduleRecoveryStoryRefsQuery = `
	WITH candidates AS (
		SELECT ownership.workspace_id, ownership.story_id, interrupted.run_id AS interrupted_run_id
		FROM calendar_maya_schedule_ownerships ownership
		INNER JOIN stories story ON
			story.id = ownership.story_id
			AND story.workspace_id = ownership.workspace_id
		LEFT JOIN statuses status ON status.status_id = story.status_id
		LEFT JOIN users owner_user ON owner_user.user_id = ownership.user_id
		LEFT JOIN workspace_members workspace_member ON
			workspace_member.workspace_id = ownership.workspace_id
			AND workspace_member.user_id = ownership.user_id
		LEFT JOIN team_members team_member ON
			team_member.team_id = story.team_id
			AND team_member.user_id = ownership.user_id
		LEFT JOIN team_sprint_settings team_settings ON
			team_settings.team_id = story.team_id
			AND team_settings.workspace_id = story.workspace_id
		LEFT JOIN sprints sprint ON sprint.sprint_id = story.sprint_id
		LEFT JOIN LATERAL (
			SELECT run.run_id
			FROM maya_agent_runs run
			WHERE run.workspace_id = ownership.workspace_id
				AND run.story_id = ownership.story_id
				AND run.status = 'running'
				AND run.updated_at <= $3
			ORDER BY run.updated_at, run.run_id
			LIMIT 1
		) interrupted ON TRUE
		WHERE COALESCE(ownership.recovery_attempted_at, TIMESTAMP 'epoch') <= $2
			AND (
				interrupted.run_id IS NOT NULL
				OR story.updated_at > ownership.updated_at
				OR team_settings.updated_at > ownership.updated_at
				OR sprint.updated_at > ownership.updated_at
				OR (story.assignee_id IS NOT NULL AND story.assignee_id <> ownership.user_id)
				OR story.deleted_at IS NOT NULL
				OR story.archived_at IS NOT NULL
				OR story.completed_at IS NOT NULL
				OR status.status_id IS NULL
				OR status.deleted_at IS NOT NULL
				OR status.category IN ('completed', 'cancelled')
				OR owner_user.user_id IS NULL
				OR owner_user.is_active = FALSE
				OR workspace_member.user_id IS NULL
				OR team_member.user_id IS NULL
				OR EXISTS (
					SELECT 1
					FROM calendar_schedule_blocks block
					WHERE block.workspace_id = ownership.workspace_id
						AND block.story_id = ownership.story_id
						AND block.source = 'maya'
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
		LIMIT $1
		FOR UPDATE OF ownership SKIP LOCKED
	), claimed AS (
		UPDATE calendar_maya_schedule_ownerships ownership
		SET recovery_attempted_at = CURRENT_TIMESTAMP
		FROM candidates
		WHERE ownership.workspace_id = candidates.workspace_id
			AND ownership.story_id = candidates.story_id
		RETURNING ownership.workspace_id, ownership.story_id
	)
	SELECT claimed.workspace_id, claimed.story_id, candidates.interrupted_run_id
	FROM claimed
	INNER JOIN candidates ON
		candidates.workspace_id = claimed.workspace_id
		AND candidates.story_id = claimed.story_id
	ORDER BY claimed.workspace_id, claimed.story_id
`

const listScheduleStoryRefsForUserQuery = `
	WITH scheduled_stories AS (
		SELECT workspace_id, story_id
		FROM calendar_maya_schedule_ownerships
		WHERE user_id = $1
		UNION
		SELECT workspace_id, story_id
		FROM calendar_schedule_blocks
		WHERE user_id = $1 AND source = 'maya' AND story_id IS NOT NULL
	)
	SELECT scheduled.workspace_id, scheduled.story_id
	FROM scheduled_stories scheduled
	INNER JOIN stories story ON story.id = scheduled.story_id AND story.workspace_id = scheduled.workspace_id
	ORDER BY story.end_date NULLS LAST,
		CASE story.priority
			WHEN 'Urgent' THEN 0
			WHEN 'High' THEN 1
			WHEN 'Medium' THEN 2
			WHEN 'Low' THEN 3
			WHEN 'No Priority' THEN 4
			ELSE 5
		END,
	scheduled.story_id
`

const storyScheduleOwnershipRetainableQuery = `
	SELECT EXISTS (
		SELECT 1
		FROM stories story
		INNER JOIN statuses status ON status.status_id = story.status_id AND status.deleted_at IS NULL
		INNER JOIN users owner_user ON owner_user.user_id = $3 AND owner_user.is_active = TRUE
		INNER JOIN workspace_members workspace_member ON
			workspace_member.workspace_id = story.workspace_id
			AND workspace_member.user_id = $3
		INNER JOIN team_members team_member ON
			team_member.team_id = story.team_id
			AND team_member.user_id = $3
		WHERE story.workspace_id = $1
			AND story.id = $2
			AND story.deleted_at IS NULL
			AND story.archived_at IS NULL
			AND story.completed_at IS NULL
			AND status.category NOT IN ('completed', 'cancelled')
			AND (story.assignee_id IS NULL OR story.assignee_id = $3)
	)
`

func (r *Repo) WithScheduleStoryLock(ctx context.Context, workspaceID, storyID uuid.UUID, reconcile func() error) error {
	if workspaceID == uuid.Nil || storyID == uuid.Nil || reconcile == nil {
		return maya.ErrInvalidPlanInput
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin Maya story schedule lock: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	const query = `
		SELECT pg_advisory_xact_lock(
			hashtextextended(
				CONCAT('maya-schedule:', CAST($1 AS text), ':', CAST($2 AS text)),
				0
			)
		)
	`
	if _, err := tx.ExecContext(ctx, query, workspaceID, storyID); err != nil {
		return fmt.Errorf("lock Maya story schedule: %w", err)
	}
	if err := reconcile(); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit Maya story schedule lock: %w", err)
	}
	return nil
}

func (r *Repo) ListScheduleStoryRefsForUser(ctx context.Context, userID uuid.UUID) ([]maya.ScheduleStoryRef, error) {
	refs := []maya.ScheduleStoryRef{}
	if err := r.db.SelectContext(ctx, &refs, listScheduleStoryRefsForUserQuery, userID); err != nil {
		return nil, fmt.Errorf("list stories for schedule reconciliation: %w", err)
	}
	return refs, nil
}

// ClaimScheduleRecoveryStoryRefs leases inconsistent durable Maya ownerships
// without advancing their last-successful-reconciliation watermark. Failed
// rows therefore remain discoverable after the bounded retry delay.
func (r *Repo) ClaimScheduleRecoveryStoryRefs(ctx context.Context, limit int, retryBefore, interruptedRunBefore time.Time) ([]maya.ScheduleRecoveryRef, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin Maya schedule recovery claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	refs := []maya.ScheduleRecoveryRef{}
	if err := tx.SelectContext(ctx, &refs, claimScheduleRecoveryStoryRefsQuery, limit, retryBefore.UTC(), interruptedRunBefore.UTC()); err != nil {
		return nil, fmt.Errorf("claim Maya schedule recovery stories: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit Maya schedule recovery claim: %w", err)
	}
	return refs, nil
}

func (r *Repo) CompleteInterruptedScheduleRun(ctx context.Context, runID uuid.UUID, message string) error {
	if runID == uuid.Nil {
		return nil
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin interrupted Maya run completion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		UPDATE maya_agent_actions
		SET status = 'failed',
			error_message = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE run_id = $1 AND status = 'proposed'
	`, runID, message); err != nil {
		return fmt.Errorf("fail interrupted Maya actions: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE maya_agent_runs run
		SET status = CASE
				WHEN EXISTS (
					SELECT 1 FROM maya_agent_actions action
					WHERE action.run_id = run.run_id AND action.status = 'failed'
				) THEN 'failed'
				ELSE 'succeeded'
			END,
			summary = CASE
				WHEN EXISTS (
					SELECT 1 FROM maya_agent_actions action
					WHERE action.run_id = run.run_id AND action.status = 'failed'
				) THEN 'Maya recovered an interrupted scheduling operation against the story''s current state.'
				ELSE 'Maya verified the completed scheduling operation after worker recovery.'
			END,
			error_message = CASE
				WHEN EXISTS (
					SELECT 1 FROM maya_agent_actions action
					WHERE action.run_id = run.run_id AND action.status = 'failed'
				) THEN $2
				ELSE NULL
			END,
			completed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE run.run_id = $1 AND run.status = 'running'
	`, runID, message); err != nil {
		return fmt.Errorf("complete interrupted Maya run: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit interrupted Maya run completion: %w", err)
	}
	return nil
}

func (r *Repo) ListMayaScheduleOwners(ctx context.Context, workspaceID, storyID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT user_id
		FROM calendar_maya_schedule_ownerships
		WHERE workspace_id = $1 AND story_id = $2
		UNION
		SELECT user_id
		FROM calendar_schedule_blocks
		WHERE workspace_id = $1 AND story_id = $2 AND source = 'maya'
		ORDER BY user_id
	`
	owners := []uuid.UUID{}
	if err := r.db.SelectContext(ctx, &owners, query, workspaceID, storyID); err != nil {
		return nil, fmt.Errorf("list Maya schedule owners: %w", err)
	}
	return owners, nil
}

func (r *Repo) StoryIsSchedulableForUser(ctx context.Context, workspaceID, storyID, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM stories story
			INNER JOIN statuses status ON status.status_id = story.status_id AND status.deleted_at IS NULL
			INNER JOIN users assignee ON assignee.user_id = $3 AND assignee.is_active = TRUE
			INNER JOIN workspace_members membership ON
				membership.workspace_id = story.workspace_id AND membership.user_id = $3
			INNER JOIN team_members team_membership ON
				team_membership.team_id = story.team_id AND team_membership.user_id = $3
			WHERE story.workspace_id = $1
				AND story.id = $2
				AND story.assignee_id = $3
				AND story.deleted_at IS NULL
				AND story.archived_at IS NULL
				AND story.completed_at IS NULL
				AND status.category NOT IN ('completed', 'cancelled')
		)
	`
	var schedulable bool
	if err := r.db.GetContext(ctx, &schedulable, query, workspaceID, storyID, userID); err != nil {
		return false, fmt.Errorf("check story schedule eligibility: %w", err)
	}
	return schedulable, nil
}

// StoryScheduleOwnershipIsRetainable distinguishes temporary placement gaps
// from lifecycle states that must permanently retire Maya's ownership. An
// unassigned active story remains enrolled so a later assignment can resume
// scheduling, but terminal stories and removed or inactive owners do not.
func (r *Repo) StoryScheduleOwnershipIsRetainable(ctx context.Context, workspaceID, storyID, userID uuid.UUID) (bool, error) {
	var retainable bool
	if err := r.db.GetContext(ctx, &retainable, storyScheduleOwnershipRetainableQuery, workspaceID, storyID, userID); err != nil {
		return false, fmt.Errorf("check Maya schedule ownership retention: %w", err)
	}
	return retainable, nil
}
