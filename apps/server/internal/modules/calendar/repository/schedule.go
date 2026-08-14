package calendarrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
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
		EXISTS (
			SELECT 1
			FROM calendar_busy_windows conflict_window
			INNER JOIN calendar_connections conflict_connection ON
				conflict_connection.connection_id = conflict_window.connection_id
				AND conflict_connection.user_id = conflict_window.user_id
				AND conflict_connection.revoked_at IS NULL
			WHERE conflict_window.user_id = csb.user_id
				AND conflict_window.start_at < csb.end_at
				AND conflict_window.end_at > csb.start_at
		) AS has_conflict,
		csb.is_locked,
		csb.source,
		csb.created_at,
		csb.updated_at
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
`

func (r *Repo) ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreBusyWindow, error) {
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
		WHERE cbw.user_id = $2
			AND cbw.start_at < $4
			AND cbw.end_at > $3
		ORDER BY cbw.start_at ASC
	`
	rows := []dbBusyWindow{}
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID, userID, startAt, endAt); err != nil {
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
			WHERE csb.workspace_id = $1
				AND csb.user_id = $2
				AND csb.start_at < $4
				AND csb.end_at > $3
				AND ($5 = CAST('00000000-0000-0000-0000-000000000000' AS uuid) OR csb.block_id <> $5)
			UNION ALL
			SELECT 1
			FROM calendar_busy_windows cbw
			INNER JOIN calendar_connections cc ON
				cc.connection_id = cbw.connection_id
				AND cc.user_id = cbw.user_id
				AND cc.revoked_at IS NULL
			WHERE cbw.user_id = $2
				AND cbw.start_at < $4
				AND cbw.end_at > $3
		)
	`
	var conflicts bool
	if err := tx.GetContext(
		ctx,
		&conflicts,
		query,
		input.WorkspaceID,
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
	const query = `
		DELETE FROM calendar_schedule_blocks
		WHERE workspace_id = $1
			AND user_id = $2
			AND block_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, workspaceID, userID, blockID)
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
