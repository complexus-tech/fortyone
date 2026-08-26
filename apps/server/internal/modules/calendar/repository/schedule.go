package calendarrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

const scheduleBlockSelect = `
	SELECT
		csb.block_id,
		csb.workspace_id,
		csb.user_id,
		CASE WHEN s.id IS NOT NULL THEN csb.story_id ELSE NULL END AS story_id,
		s.title AS story_title,
		CASE
			WHEN s.id IS NOT NULL AND t.code IS NOT NULL THEN CONCAT(t.code, '-', CAST(s.sequence_id AS text))
			ELSE NULL
		END AS story_code,
		status.color AS story_status_color,
		COALESCE(s.priority, '') AS story_priority,
		s.end_date AS story_end_date,
		t.team_id,
		t.name AS team_name,
		t.code AS team_code,
		csb.block_type,
		CASE
			WHEN csb.block_type = 'work' AND s.id IS NULL THEN 'Work'
			ELSE csb.title
		END AS title,
		csb.start_at,
		csb.end_at,
		csb.completed_at,
		csb.completed_at IS NULL AND EXISTS (
			SELECT 1
			FROM calendar_busy_windows conflict_window
			INNER JOIN calendar_connections conflict_connection ON
				conflict_connection.connection_id = conflict_window.connection_id
				AND conflict_connection.user_id = conflict_window.user_id
				AND conflict_connection.revoked_at IS NULL
				AND conflict_connection.cleanup_pending_at IS NULL
			WHERE conflict_window.user_id = csb.user_id
				AND conflict_window.start_at < csb.end_at
				AND conflict_window.end_at > csb.start_at
		) AS has_conflict,
		csb.is_locked,
		CASE WHEN csb.source = 'maya' THEN s.auto_scheduling_status ELSE NULL END AS auto_scheduling_status,
		CASE WHEN csb.source = 'maya' THEN s.auto_scheduling_reason ELSE NULL END AS auto_scheduling_reason,
		csb.source,
		csb.segment_index,
		csb.external_provider,
		csb.external_calendar_id,
		csb.external_event_id,
		csb.external_sync_hash,
		csb.external_synced_at,
		csb.created_at,
		csb.updated_at,
		csb.manual_override_at,
		csb.manual_override_by
	FROM calendar_schedule_blocks csb
	LEFT JOIN stories s ON
		s.id = csb.story_id
		AND s.workspace_id = csb.workspace_id
		AND s.deleted_at IS NULL
		AND EXISTS (
			SELECT 1
			FROM team_members viewer_membership
			WHERE viewer_membership.team_id = s.team_id
				AND viewer_membership.user_id = csb.user_id
		)
	LEFT JOIN teams t ON t.team_id = s.team_id
	LEFT JOIN statuses status ON
		status.status_id = s.status_id
		AND status.team_id = s.team_id
`

func (r *Repo) ListBusyWindows(ctx context.Context, _ uuid.UUID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreBusyWindow, error) {
	const query = `
		SELECT cbw.window_id, cbw.connection_id, cbw.workspace_id, cbw.user_id,
		       cbw.provider, cbw.provider_event_id, cbw.calendar_id, cbw.title,
		       cbw.start_at, cbw.end_at, cbw.status, cbw.transparency,
		       cbw.is_private, cbw.source_hash, cbw.created_at, cbw.updated_at
		FROM calendar_busy_windows cbw
		INNER JOIN calendar_connections cc ON
			cc.connection_id = cbw.connection_id
			AND cc.user_id = cbw.user_id
			AND cc.revoked_at IS NULL
			AND cc.cleanup_pending_at IS NULL
		WHERE cbw.user_id = $1
			AND cbw.start_at < $3
			AND cbw.end_at > $2
		ORDER BY cbw.start_at ASC
	`
	rows := []dbBusyWindow{}
	if err := r.db.SelectContext(ctx, &rows, query, userID, startAt, endAt); err != nil {
		return nil, fmt.Errorf("list calendar busy windows: %w", err)
	}
	return toCoreBusyWindows(rows), nil
}

func (r *Repo) ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreScheduleBlock, error) {
	query := scheduleBlockSelect + `
		WHERE csb.workspace_id = $1
			AND csb.user_id = $2
			AND csb.start_at < $4
			AND csb.end_at > $3
		ORDER BY csb.start_at ASC
	`
	rows := []dbScheduleBlock{}
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID, userID, startAt, endAt); err != nil {
		return nil, fmt.Errorf("list calendar schedule blocks: %w", err)
	}
	return toCoreScheduleBlocks(rows), nil
}

func (r *Repo) ListScheduleIssues(ctx context.Context, workspaceID, userID uuid.UUID) ([]calendar.CoreScheduleIssue, error) {
	const query = `
		SELECT
			story.id AS story_id,
			story.title AS story_title,
			CONCAT(team.code, '-', CAST(story.sequence_id AS text)) AS story_code,
			team.team_id,
			team.name AS team_name,
			team.code AS team_code,
			story.estimated_duration_minutes,
			COALESCE(scheduled.scheduled_duration_minutes, 0)::int AS scheduled_duration_minutes,
			GREATEST(
				COALESCE(story.estimated_duration_minutes, 0) - COALESCE(scheduled.scheduled_duration_minutes, 0),
				0
			)::int AS remaining_duration_minutes,
			story.auto_scheduling_status,
			story.auto_scheduling_reason,
			story.updated_at
		FROM stories story
		INNER JOIN teams team ON team.team_id = story.team_id
		INNER JOIN statuses status ON status.status_id = story.status_id
		INNER JOIN team_members membership ON
			membership.team_id = story.team_id
			AND membership.user_id = $2
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (block.end_at - block.start_at)) / 60), 0)::int AS scheduled_duration_minutes
			FROM calendar_schedule_blocks block
			WHERE block.workspace_id = story.workspace_id
				AND block.user_id = story.assignee_id
				AND block.story_id = story.id
				AND block.source = 'maya'
				AND block.completed_at IS NULL
		) scheduled ON TRUE
		WHERE story.workspace_id = $1
			AND story.assignee_id = $2
			AND story.auto_scheduling_enabled = TRUE
			AND story.auto_scheduling_status = 'cannot_fit'
			AND story.deleted_at IS NULL
			AND story.archived_at IS NULL
			AND story.completed_at IS NULL
			AND status.category NOT IN ('completed', 'cancelled')
		ORDER BY story.auto_scheduling_updated_at DESC NULLS LAST, story.updated_at DESC, story.id
	`
	issues := []calendar.CoreScheduleIssue{}
	if err := r.db.SelectContext(ctx, &issues, query, workspaceID, userID); err != nil {
		return nil, fmt.Errorf("list Maya schedule issues: %w", err)
	}
	return issues, nil
}

func (r *Repo) ListSchedulingBlocksForUser(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreScheduleBlock, error) {
	query := scheduleBlockSelect + `
		WHERE csb.user_id = $1
			AND csb.start_at < $3
			AND csb.end_at > $2
			AND csb.completed_at IS NULL
			AND EXISTS (
				SELECT 1 FROM workspace_members owner_membership
				WHERE owner_membership.workspace_id = csb.workspace_id
					AND owner_membership.user_id = csb.user_id
			)
		ORDER BY csb.start_at ASC
	`
	rows := []dbScheduleBlock{}
	if err := r.db.SelectContext(ctx, &rows, query, userID, startAt, endAt); err != nil {
		return nil, fmt.Errorf("list account-wide scheduling blocks: %w", err)
	}
	blocks := toCoreScheduleBlocks(rows)
	redactCrossWorkspaceScheduleBlocks(blocks, workspaceID)
	return blocks, nil
}

func redactCrossWorkspaceScheduleBlocks(blocks []calendar.CoreScheduleBlock, workspaceID uuid.UUID) {
	for index := range blocks {
		if blocks[index].WorkspaceID == workspaceID {
			continue
		}
		blocks[index].StoryID = nil
		blocks[index].StoryTitle = nil
		blocks[index].StoryCode = nil
		blocks[index].StoryStatusColor = nil
		blocks[index].StoryPriority = ""
		blocks[index].StoryEndDate = nil
		blocks[index].TeamID = nil
		blocks[index].TeamName = nil
		blocks[index].TeamCode = nil
		blocks[index].AutoSchedulingStatus = nil
		blocks[index].AutoSchedulingReason = nil
		blocks[index].ManualOverrideBy = nil
		blocks[index].ManualOverrideAt = nil
		blocks[index].Title = "Scheduled elsewhere"
		blocks[index].IsCrossWorkspace = true
	}
}

func (r *Repo) ListManualScheduleRescheduleEvents(ctx context.Context, workspaceID, userID uuid.UUID, since time.Time) ([]calendar.CoreScheduleRescheduleEvent, error) {
	const query = `
		SELECT next_start_at, timezone, created_at
		FROM calendar_schedule_reschedule_events
		WHERE workspace_id = $1
			AND user_id = $2
			AND source = 'user'
			AND created_at >= $3
		ORDER BY created_at DESC
		LIMIT 100
	`
	type row struct {
		NextStartAt time.Time `db:"next_start_at"`
		Timezone    string    `db:"timezone"`
		CreatedAt   time.Time `db:"created_at"`
	}
	rows := []row{}
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID, userID, since.UTC()); err != nil {
		return nil, fmt.Errorf("list manual calendar reschedule events: %w", err)
	}
	events := make([]calendar.CoreScheduleRescheduleEvent, 0, len(rows))
	for _, item := range rows {
		events = append(events, calendar.CoreScheduleRescheduleEvent{
			NextStartAt: item.NextStartAt,
			Timezone:    item.Timezone,
			CreatedAt:   item.CreatedAt,
		})
	}
	return events, nil
}

func (r *Repo) ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]calendar.CoreScheduleBlock, error) {
	query := scheduleBlockSelect + `
		WHERE csb.workspace_id = $1
			AND csb.user_id = $2
			AND csb.story_id = $3
			AND csb.source = 'maya'
			AND csb.completed_at IS NULL
		ORDER BY csb.segment_index, csb.start_at
	`
	rows := []dbScheduleBlock{}
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID, userID, storyID); err != nil {
		return nil, fmt.Errorf("list Maya schedule blocks for story: %w", err)
	}
	return toCoreScheduleBlocks(rows), nil
}

func (r *Repo) MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM calendar_maya_schedule_ownerships
			WHERE workspace_id = $1 AND user_id = $2 AND story_id = $3
		)
	`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, workspaceID, userID, storyID); err != nil {
		return false, fmt.Errorf("check Maya schedule ownership: %w", err)
	}
	return exists, nil
}

func (r *Repo) ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM stories s
			INNER JOIN team_members tm ON
				tm.team_id = s.team_id
				AND tm.user_id = $2
			WHERE s.workspace_id = $1
				AND s.id = $3
				AND s.deleted_at IS NULL
		)
	`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, workspaceID, userID, storyID); err != nil {
		return false, fmt.Errorf("check calendar schedule story: %w", err)
	}
	return exists, nil
}

func (r *Repo) CreateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("begin create calendar schedule block: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, input.WorkspaceID, input.UserID); err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	conflicts, err := scheduleBlockConflicts(ctx, tx, input, uuid.Nil)
	if err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	if conflicts {
		return calendar.CoreScheduleBlock{}, calendar.ErrCalendarScheduleConflict
	}

	const query = `
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
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING block_id
	`
	var blockID uuid.UUID
	if err := tx.GetContext(
		ctx,
		&blockID,
		query,
		input.WorkspaceID,
		input.UserID,
		input.StoryID,
		string(input.BlockType),
		input.Title,
		input.StartAt,
		input.EndAt,
		input.IsLocked,
		string(input.Source),
	); err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("create calendar schedule block: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("commit create calendar schedule block: %w", err)
	}
	return r.getScheduleBlock(ctx, input.WorkspaceID, input.UserID, blockID)
}

func (r *Repo) UpdateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("begin update calendar schedule block: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, input.WorkspaceID, input.UserID); err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	conflicts, err := scheduleBlockConflicts(ctx, tx, input, input.ID)
	if err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	if conflicts {
		return calendar.CoreScheduleBlock{}, calendar.ErrCalendarScheduleConflict
	}

	const query = `
		UPDATE calendar_schedule_blocks
		SET story_id = $4,
			block_type = $5,
			title = $6,
			start_at = $7,
			end_at = $8,
			is_locked = $9,
			source = $10,
			updated_at = CURRENT_TIMESTAMP
		WHERE workspace_id = $1
			AND user_id = $2
			AND block_id = $3
			AND source <> 'maya'
	`
	result, err := tx.ExecContext(
		ctx,
		query,
		input.WorkspaceID,
		input.UserID,
		input.ID,
		input.StoryID,
		string(input.BlockType),
		input.Title,
		input.StartAt,
		input.EndAt,
		input.IsLocked,
		string(input.Source),
	)
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("update calendar schedule block: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("read updated calendar schedule block count: %w", err)
	}
	if rows == 0 {
		return calendar.CoreScheduleBlock{}, calendar.ErrCalendarScheduleBlockNotFound
	}
	if err := tx.Commit(); err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("commit update calendar schedule block: %w", err)
	}
	return r.getScheduleBlock(ctx, input.WorkspaceID, input.UserID, input.ID)
}

func (r *Repo) ManuallyRescheduleScheduleBlock(ctx context.Context, input calendar.ManualScheduleBlockInput) (calendar.ManualScheduleBlockResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("begin manual calendar reschedule: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, input.WorkspaceID, input.UserID); err != nil {
		return calendar.ManualScheduleBlockResult{}, err
	}

	var existingEventBlockID uuid.UUID
	if err := tx.GetContext(ctx, &existingEventBlockID, `
		SELECT schedule_block_id
		FROM calendar_schedule_reschedule_events
		WHERE client_mutation_id = $1
	`, input.ClientMutationID); err == nil {
		if err := tx.Commit(); err != nil {
			return calendar.ManualScheduleBlockResult{}, fmt.Errorf("commit idempotent manual calendar reschedule: %w", err)
		}
		block, err := r.getScheduleBlock(ctx, input.WorkspaceID, input.UserID, existingEventBlockID)
		if err != nil {
			return calendar.ManualScheduleBlockResult{}, err
		}
		return calendar.ManualScheduleBlockResult{Block: block}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("check manual calendar reschedule idempotency: %w", err)
	}

	var current struct {
		ID                 uuid.UUID  `db:"block_id"`
		StoryID            *uuid.UUID `db:"story_id"`
		BlockType          string     `db:"block_type"`
		Title              string     `db:"title"`
		StartAt            time.Time  `db:"start_at"`
		EndAt              time.Time  `db:"end_at"`
		IsLocked           bool       `db:"is_locked"`
		Source             string     `db:"source"`
		ExternalProvider   *string    `db:"external_provider"`
		ExternalCalendarID *string    `db:"external_calendar_id"`
		ExternalEventID    *string    `db:"external_event_id"`
		UpdatedAt          time.Time  `db:"updated_at"`
	}
	if err := tx.GetContext(ctx, &current, `
		SELECT block_id, story_id, block_type, title, start_at, end_at, is_locked, source,
		       external_provider, external_calendar_id, external_event_id, updated_at
		FROM calendar_schedule_blocks
		WHERE workspace_id = $1 AND user_id = $2 AND block_id = $3
			AND completed_at IS NULL
		FOR UPDATE
	`, input.WorkspaceID, input.UserID, input.BlockID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.ManualScheduleBlockResult{}, calendar.ErrCalendarScheduleBlockNotFound
		}
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("load manual calendar reschedule block: %w", err)
	}
	if input.ExpectedUpdatedAt != nil && !current.UpdatedAt.Equal(input.ExpectedUpdatedAt.UTC()) {
		return calendar.ManualScheduleBlockResult{}, calendar.ErrCalendarScheduleStalePlan
	}

	var storyScheduleReconcileID *uuid.UUID
	if input.Change == calendar.ManualScheduleBlockChangeResize && current.StoryID != nil {
		var storyTime struct {
			EstimatedDurationMinutes *int `db:"estimated_duration_minutes"`
			MinimumFocusBlockMinutes *int `db:"minimum_focus_block_minutes"`
			AutoSchedulingEnabled    bool `db:"auto_scheduling_enabled"`
		}
		if err := tx.GetContext(ctx, &storyTime, `
			SELECT estimated_duration_minutes, minimum_focus_block_minutes, auto_scheduling_enabled
			FROM stories
			WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL
			FOR UPDATE
		`, input.WorkspaceID, *current.StoryID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return calendar.ManualScheduleBlockResult{}, fmt.Errorf("%w: resized calendar block story is unavailable", calendar.ErrInvalidScheduleBlock)
			}
			return calendar.ManualScheduleBlockResult{}, fmt.Errorf("load resized calendar block story time: %w", err)
		}

		estimatedDurationMinutes, err := resizedStoryEstimateMinutes(
			storyTime.EstimatedDurationMinutes,
			storyTime.MinimumFocusBlockMinutes,
			current.StartAt,
			current.EndAt,
			input.StartAt,
			input.EndAt,
		)
		if err != nil {
			return calendar.ManualScheduleBlockResult{}, err
		}
		if storyTime.EstimatedDurationMinutes == nil || *storyTime.EstimatedDurationMinutes != estimatedDurationMinutes {
			if _, err := tx.ExecContext(ctx, `
				UPDATE stories
				SET estimated_duration_minutes = $3,
				    updated_at = CURRENT_TIMESTAMP
				WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL
			`, input.WorkspaceID, *current.StoryID, estimatedDurationMinutes); err != nil {
				return calendar.ManualScheduleBlockResult{}, fmt.Errorf("update resized calendar block story estimate: %w", err)
			}
			// Maya owns the auto-scheduling status state machine. Return a
			// post-commit reconciliation request instead of duplicating its
			// assignee, lock, and status transitions inside this transaction.
			if storyTime.AutoSchedulingEnabled {
				storyID := *current.StoryID
				storyScheduleReconcileID = &storyID
			}
		}
	}
	// A drag is an explicit placement by the user. The calendar renders
	// simultaneous work in lanes and marks meeting overlaps as conflicts; free-slot
	// enforcement remains in create, edit, and automatic scheduling paths.
	if _, err := tx.ExecContext(ctx, `
		UPDATE calendar_schedule_blocks
		SET start_at = $4,
		    end_at = $5,
		    is_locked = TRUE,
		    manual_override_at = CURRENT_TIMESTAMP,
		    manual_override_by = $6,
		    updated_at = CURRENT_TIMESTAMP
		WHERE workspace_id = $1 AND user_id = $2 AND block_id = $3
	`, input.WorkspaceID, input.UserID, input.BlockID, input.StartAt.UTC(), input.EndAt.UTC(), input.ActorID); err != nil {
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("update manually rescheduled calendar block: %w", err)
	}
	// This reschedule event is the canonical audit record for the coupled block
	// and story-estimate change. Its exact before/after ranges retain the duration
	// delta without recording the same user action in a second activity stream.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO calendar_schedule_reschedule_events (
			workspace_id, user_id, story_id, schedule_block_id, action, source, timezone,
			previous_start_at, previous_end_at, next_start_at, next_end_at, client_mutation_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, input.WorkspaceID, input.UserID, current.StoryID, input.BlockID, string(input.Change), "user", input.Timezone,
		current.StartAt.UTC(), current.EndAt.UTC(), input.StartAt.UTC(), input.EndAt.UTC(), input.ClientMutationID); err != nil {
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("record manual calendar reschedule: %w", err)
	}

	if current.StoryID != nil && current.ExternalProvider != nil && current.ExternalCalendarID != nil && current.ExternalEventID != nil {
		event := calendar.ExternalScheduleEventInput{
			CalendarID:  *current.ExternalCalendarID,
			EventID:     *current.ExternalEventID,
			BlockID:     input.BlockID,
			StoryID:     *current.StoryID,
			WorkspaceID: input.WorkspaceID,
			Title:       current.Title,
			StartAt:     input.StartAt.UTC(),
			EndAt:       input.EndAt.UTC(),
			PrivateProperties: map[string]string{
				"fortyone_source": "fortyone",
			},
		}
		syncHash := calendar.ScheduleEventSyncHash(event)
		if err := enqueueScheduleEventOutbox(ctx, tx, input.WorkspaceID, input.UserID, &input.BlockID, calendar.Provider(*current.ExternalProvider), calendar.ScheduleEventOperationUpsert, event, syncHash, true); err != nil {
			return calendar.ManualScheduleBlockResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("commit manual calendar reschedule: %w", err)
	}
	block, err := r.getScheduleBlock(ctx, input.WorkspaceID, input.UserID, input.BlockID)
	if err != nil {
		return calendar.ManualScheduleBlockResult{}, err
	}
	return calendar.ManualScheduleBlockResult{Block: block, StoryScheduleReconcileID: storyScheduleReconcileID}, nil
}

func resizedStoryEstimateMinutes(
	currentEstimate, minimumFocusBlock *int,
	previousStartAt, previousEndAt, nextStartAt, nextEndAt time.Time,
) (int, error) {
	previousDurationMinutes, err := roundedScheduleBlockDurationMinutes(previousStartAt, previousEndAt)
	if err != nil {
		return 0, err
	}
	nextDurationMinutes, err := roundedScheduleBlockDurationMinutes(nextStartAt, nextEndAt)
	if err != nil {
		return 0, err
	}

	nextEstimate := int64(nextDurationMinutes)
	if currentEstimate != nil {
		nextEstimate = int64(*currentEstimate) + int64(nextDurationMinutes-previousDurationMinutes)
	}
	var candidate int
	if nextEstimate < 1 {
		candidate = 0
	} else if nextEstimate > int64(stories.MaximumEstimatedDurationMinutes) {
		candidate = stories.MaximumEstimatedDurationMinutes + 1
	} else {
		candidate = int(nextEstimate)
	}
	if err := stories.ValidateStoryTimeContract(&candidate, minimumFocusBlock); err != nil {
		return 0, fmt.Errorf("%w: %w", calendar.ErrInvalidScheduleBlock, err)
	}
	return candidate, nil
}

func roundedScheduleBlockDurationMinutes(startAt, endAt time.Time) (int, error) {
	if !endAt.After(startAt) {
		return 0, fmt.Errorf("%w: resized calendar block must end after it starts", calendar.ErrInvalidScheduleBlock)
	}
	return int(math.Round(endAt.Sub(startAt).Minutes())), nil
}

func scheduleBlockConflicts(
	ctx context.Context,
	tx interface {
		GetContext(context.Context, any, string, ...any) error
	},
	input calendar.CoreScheduleBlockInput,
	excludeBlockID uuid.UUID,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM calendar_schedule_blocks csb
			WHERE csb.user_id = $1
				AND csb.completed_at IS NULL
				AND csb.start_at < $3
				AND csb.end_at > $2
				AND ($4 = CAST('00000000-0000-0000-0000-000000000000' AS uuid) OR csb.block_id <> $4)
			UNION ALL
			SELECT 1
			FROM calendar_busy_windows cbw
			INNER JOIN calendar_connections cc ON
				cc.connection_id = cbw.connection_id
				AND cc.user_id = cbw.user_id
				AND cc.revoked_at IS NULL
				AND cc.cleanup_pending_at IS NULL
			WHERE cbw.user_id = $1
				AND cbw.start_at < $3
				AND cbw.end_at > $2
		)
	`
	var conflicts bool
	if err := tx.GetContext(
		ctx,
		&conflicts,
		query,
		input.UserID,
		input.StartAt,
		input.EndAt,
		excludeBlockID,
	); err != nil {
		return false, fmt.Errorf("check calendar schedule conflict: %w", err)
	}
	return conflicts, nil
}

func (r *Repo) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete calendar schedule block: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, workspaceID, userID); err != nil {
		return err
	}
	const query = `
		DELETE FROM calendar_schedule_blocks
		WHERE workspace_id = $1
			AND user_id = $2
			AND block_id = $3
			AND source <> 'maya'
	`
	result, err := tx.ExecContext(ctx, query, workspaceID, userID, blockID)
	if err != nil {
		return fmt.Errorf("delete calendar schedule block: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted calendar schedule block count: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarScheduleBlockNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete calendar schedule block: %w", err)
	}
	return nil
}

func (r *Repo) getScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) (calendar.CoreScheduleBlock, error) {
	query := scheduleBlockSelect + `
		WHERE csb.workspace_id = $1
			AND csb.user_id = $2
			AND csb.block_id = $3
		LIMIT 1
	`
	var row dbScheduleBlock
	if err := r.db.GetContext(ctx, &row, query, workspaceID, userID, blockID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreScheduleBlock{}, calendar.ErrCalendarScheduleBlockNotFound
		}
		return calendar.CoreScheduleBlock{}, fmt.Errorf("get calendar schedule block: %w", err)
	}
	return toCoreScheduleBlock(row), nil
}

func (r *Repo) GetScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) (calendar.CoreScheduleBlock, error) {
	return r.getScheduleBlock(ctx, workspaceID, userID, blockID)
}
