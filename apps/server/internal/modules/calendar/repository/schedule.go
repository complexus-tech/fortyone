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
		csb.story_id,
		s.title AS story_title,
		CASE
			WHEN s.id IS NOT NULL AND t.code IS NOT NULL THEN CONCAT(t.code, '-', CAST(s.sequence_id AS text))
			ELSE NULL
		END AS story_code,
		t.team_id,
		t.name AS team_name,
		t.code AS team_code,
		csb.block_type,
		csb.title,
		csb.start_at,
		csb.end_at,
		csb.is_locked,
		csb.source,
		csb.created_at,
		csb.updated_at
	FROM calendar_schedule_blocks csb
	LEFT JOIN stories s ON
		s.id = csb.story_id
		AND s.workspace_id = csb.workspace_id
		AND s.deleted_at IS NULL
	LEFT JOIN teams t ON t.team_id = s.team_id
`

func (r *Repo) ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreBusyWindow, error) {
	const query = `
		SELECT window_id, connection_id, workspace_id, user_id, provider, provider_event_id,
		       calendar_id, title, start_at, end_at, status, transparency,
		       is_private, source_hash, created_at, updated_at
		FROM calendar_busy_windows
		WHERE workspace_id = $1
			AND user_id = $2
			AND start_at < $4
			AND end_at > $3
		ORDER BY start_at ASC
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

func (r *Repo) CreateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
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
	if err := r.db.GetContext(
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
	return r.getScheduleBlock(ctx, input.WorkspaceID, input.UserID, blockID)
}

func (r *Repo) UpdateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
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
	result, err := r.db.ExecContext(
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
	return r.getScheduleBlock(ctx, input.WorkspaceID, input.UserID, input.ID)
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
