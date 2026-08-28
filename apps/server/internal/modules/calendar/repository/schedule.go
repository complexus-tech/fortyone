package calendarrepository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ListBusyWindows(ctx context.Context, _ uuid.UUID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreBusyWindow, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCalendarBusyWindows(ctx, calendarsql.ListCalendarBusyWindowsParams{
		UserID:  userID,
		EndAt:   endAt,
		StartAt: startAt,
	})
	if err != nil {
		return nil, fmt.Errorf("list calendar busy windows: %w", err)
	}
	return toCoreBusyWindows(rows), nil
}

func (r *Repo) ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreScheduleBlock, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCalendarScheduleBlocks(ctx, calendarsql.ListCalendarScheduleBlocksParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		EndAt:       endAt,
		StartAt:     startAt,
	})
	if err != nil {
		return nil, fmt.Errorf("list calendar schedule blocks: %w", err)
	}
	return scheduleBlocksFromList(rows), nil
}

func (r *Repo) ListScheduleIssues(ctx context.Context, workspaceID, userID uuid.UUID) ([]calendar.CoreScheduleIssue, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCalendarScheduleIssues(ctx, calendarsql.ListCalendarScheduleIssuesParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list Maya schedule issues: %w", err)
	}
	issues := make([]calendar.CoreScheduleIssue, len(rows))
	for index, row := range rows {
		issues[index] = calendar.CoreScheduleIssue{
			StoryID: row.StoryID, StoryTitle: row.StoryTitle, StoryCode: row.StoryCode,
			TeamID: row.TeamID, TeamName: row.TeamName, TeamCode: row.TeamCode,
			EstimatedDurationMinutes: int32PointerToInt(row.EstimatedDurationMinutes),
			ScheduledDurationMinutes: int(row.ScheduledDurationMinutes),
			RemainingDurationMinutes: int(row.RemainingDurationMinutes),
			AutoSchedulingStatus:     row.AutoSchedulingStatus,
			AutoSchedulingReason:     row.AutoSchedulingReason, UpdatedAt: row.UpdatedAt,
		}
	}
	return issues, nil
}

func (r *Repo) ListSchedulingBlocksForUser(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreScheduleBlock, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSchedulingBlocksForUser(ctx, calendarsql.ListSchedulingBlocksForUserParams{
		UserID:  userID,
		EndAt:   endAt,
		StartAt: startAt,
	})
	if err != nil {
		return nil, fmt.Errorf("list account-wide scheduling blocks: %w", err)
	}
	blocks := scheduleBlocksFromSchedulingList(rows)
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
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListManualScheduleRescheduleEvents(ctx, calendarsql.ListManualScheduleRescheduleEventsParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Since:       since.UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("list manual calendar reschedule events: %w", err)
	}
	events := make([]calendar.CoreScheduleRescheduleEvent, len(rows))
	for index, row := range rows {
		events[index] = calendar.CoreScheduleRescheduleEvent{
			NextStartAt: row.NextStartAt,
			Timezone:    row.Timezone,
			CreatedAt:   row.CreatedAt,
		}
	}
	return events, nil
}

func (r *Repo) ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]calendar.CoreScheduleBlock, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListMayaScheduleBlocksForStory(ctx, calendarsql.ListMayaScheduleBlocksForStoryParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		StoryID:     &storyID,
	})
	if err != nil {
		return nil, fmt.Errorf("list Maya schedule blocks for story: %w", err)
	}
	return scheduleBlocksFromMayaList(rows), nil
}

func (r *Repo) MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	exists, err := r.queries.MayaScheduleOwnershipExists(ctx, calendarsql.MayaScheduleOwnershipExistsParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		StoryID:     storyID,
	})
	if err != nil {
		return false, fmt.Errorf("check Maya schedule ownership: %w", err)
	}
	return exists, nil
}

func (r *Repo) ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	exists, err := r.queries.CalendarScheduleStoryExists(ctx, calendarsql.CalendarScheduleStoryExistsParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return false, fmt.Errorf("check calendar schedule story: %w", err)
	}
	return exists, nil
}

func (r *Repo) CreateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	var created calendar.CoreScheduleBlock
	err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, input.UserID); err != nil {
			return err
		}
		conflicts, err := scheduleBlockConflicts(ctx, queries, input, uuid.Nil)
		if err != nil {
			return err
		}
		if conflicts {
			return calendar.ErrCalendarScheduleConflict
		}
		blockID, err := queries.CreateCalendarScheduleBlock(ctx, calendarsql.CreateCalendarScheduleBlockParams{
			WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: input.StoryID,
			BlockType: string(input.BlockType), Title: input.Title, StartAt: input.StartAt,
			EndAt: input.EndAt, IsLocked: input.IsLocked, Source: string(input.Source),
		})
		if err != nil {
			return fmt.Errorf("create calendar schedule block: %w", err)
		}
		created, err = getScheduleBlockWithQueries(ctx, queries, input.WorkspaceID, input.UserID, blockID)
		return err
	})
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("create calendar schedule block transaction: %w", err)
	}
	return created, nil
}

func (r *Repo) UpdateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	var updated calendar.CoreScheduleBlock
	err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, input.UserID); err != nil {
			return err
		}
		conflicts, err := scheduleBlockConflicts(ctx, queries, input, input.ID)
		if err != nil {
			return err
		}
		if conflicts {
			return calendar.ErrCalendarScheduleConflict
		}
		rows, err := queries.UpdateCalendarScheduleBlock(ctx, calendarsql.UpdateCalendarScheduleBlockParams{
			StoryID: input.StoryID, BlockType: string(input.BlockType), Title: input.Title,
			StartAt: input.StartAt, EndAt: input.EndAt, IsLocked: input.IsLocked,
			Source: string(input.Source), WorkspaceID: input.WorkspaceID, UserID: input.UserID, BlockID: input.ID,
		})
		if err != nil {
			return fmt.Errorf("update calendar schedule block: %w", err)
		}
		if rows == 0 {
			return calendar.ErrCalendarScheduleBlockNotFound
		}
		updated, err = getScheduleBlockWithQueries(ctx, queries, input.WorkspaceID, input.UserID, input.ID)
		return err
	})
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("update calendar schedule block transaction: %w", err)
	}
	return updated, nil
}

func (r *Repo) ManuallyRescheduleScheduleBlock(ctx context.Context, input calendar.ManualScheduleBlockInput) (calendar.ManualScheduleBlockResult, error) {
	if err := r.configured(); err != nil {
		return calendar.ManualScheduleBlockResult{}, err
	}
	var result calendar.ManualScheduleBlockResult
	err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, input.UserID); err != nil {
			return err
		}
		existingBlockID, err := queries.GetManualScheduleRescheduleBlockID(ctx, calendarsql.GetManualScheduleRescheduleBlockIDParams{
			ClientMutationID: input.ClientMutationID,
		})
		if err == nil {
			if existingBlockID == nil {
				return fmt.Errorf("%w: manual reschedule record has no schedule block", calendar.ErrInvalidScheduleBlock)
			}
			block, err := getScheduleBlockWithQueries(ctx, queries, input.WorkspaceID, input.UserID, *existingBlockID)
			if err != nil {
				return err
			}
			result = calendar.ManualScheduleBlockResult{Block: block}
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check manual calendar reschedule idempotency: %w", err)
		}

		current, err := queries.LockManualScheduleBlock(ctx, calendarsql.LockManualScheduleBlockParams{
			WorkspaceID: input.WorkspaceID,
			UserID:      input.UserID,
			BlockID:     input.BlockID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return calendar.ErrCalendarScheduleBlockNotFound
		}
		if err != nil {
			return fmt.Errorf("load manual calendar reschedule block: %w", err)
		}
		if input.ExpectedUpdatedAt != nil && !current.UpdatedAt.Equal(input.ExpectedUpdatedAt.UTC()) {
			return calendar.ErrCalendarScheduleStalePlan
		}

		var storyScheduleReconcileID *uuid.UUID
		if input.Change == calendar.ManualScheduleBlockChangeResize && current.StoryID != nil {
			storyTime, err := queries.LockScheduleStoryTime(ctx, calendarsql.LockScheduleStoryTimeParams{
				WorkspaceID: input.WorkspaceID,
				StoryID:     *current.StoryID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: resized calendar block story is unavailable", calendar.ErrInvalidScheduleBlock)
			}
			if err != nil {
				return fmt.Errorf("load resized calendar block story time: %w", err)
			}
			currentEstimate := int32PointerToInt(storyTime.EstimatedDurationMinutes)
			minimumFocus := int32PointerToInt(storyTime.MinimumFocusBlockMinutes)
			estimatedMinutes, err := resizedStoryEstimateMinutes(currentEstimate, minimumFocus, current.StartAt, current.EndAt, input.StartAt, input.EndAt)
			if err != nil {
				return err
			}
			if currentEstimate == nil || *currentEstimate != estimatedMinutes {
				value, err := safecast.Int32(estimatedMinutes)
				if err != nil {
					return fmt.Errorf("validate resized calendar block story estimate: %w", err)
				}
				if err := queries.UpdateScheduleStoryEstimate(ctx, calendarsql.UpdateScheduleStoryEstimateParams{
					EstimatedDurationMinutes: &value,
					WorkspaceID:              input.WorkspaceID,
					StoryID:                  *current.StoryID,
				}); err != nil {
					return fmt.Errorf("update resized calendar block story estimate: %w", err)
				}
				if storyTime.AutoSchedulingEnabled {
					storyID := *current.StoryID
					storyScheduleReconcileID = &storyID
				}
			}
		}

		rows, err := queries.ManuallyRescheduleCalendarBlock(ctx, calendarsql.ManuallyRescheduleCalendarBlockParams{
			StartAt: input.StartAt.UTC(), EndAt: input.EndAt.UTC(), ActorID: &input.ActorID,
			WorkspaceID: input.WorkspaceID, UserID: input.UserID, BlockID: input.BlockID,
		})
		if err != nil {
			return fmt.Errorf("update manually rescheduled calendar block: %w", err)
		}
		if rows == 0 {
			return calendar.ErrCalendarScheduleBlockNotFound
		}
		if err := queries.RecordManualCalendarReschedule(ctx, calendarsql.RecordManualCalendarRescheduleParams{
			WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: current.StoryID,
			ScheduleBlockID: &input.BlockID, Action: string(input.Change), Timezone: input.Timezone,
			PreviousStartAt: current.StartAt.UTC(), PreviousEndAt: current.EndAt.UTC(),
			NextStartAt: input.StartAt.UTC(), NextEndAt: input.EndAt.UTC(), ClientMutationID: input.ClientMutationID,
		}); err != nil {
			return fmt.Errorf("record manual calendar reschedule: %w", err)
		}

		if current.StoryID != nil && current.ExternalProvider != nil && current.ExternalCalendarID != nil && current.ExternalEventID != nil {
			event := calendar.ExternalScheduleEventInput{
				CalendarID: *current.ExternalCalendarID, EventID: *current.ExternalEventID,
				BlockID: input.BlockID, StoryID: *current.StoryID, WorkspaceID: input.WorkspaceID,
				Title: current.Title, StartAt: input.StartAt.UTC(), EndAt: input.EndAt.UTC(),
				PrivateProperties: map[string]string{"fortyone_source": "fortyone"},
			}
			if err := enqueueScheduleEventOutbox(ctx, queries, input.WorkspaceID, input.UserID, &input.BlockID, calendar.Provider(*current.ExternalProvider), calendar.ScheduleEventOperationUpsert, event, calendar.ScheduleEventSyncHash(event), true); err != nil {
				return err
			}
		}
		block, err := getScheduleBlockWithQueries(ctx, queries, input.WorkspaceID, input.UserID, input.BlockID)
		if err != nil {
			return err
		}
		result = calendar.ManualScheduleBlockResult{Block: block, StoryScheduleReconcileID: storyScheduleReconcileID}
		return nil
	})
	if err != nil {
		return calendar.ManualScheduleBlockResult{}, fmt.Errorf("manual calendar reschedule transaction: %w", err)
	}
	return result, nil
}

func resizedStoryEstimateMinutes(currentEstimate, minimumFocusBlock *int, previousStartAt, previousEndAt, nextStartAt, nextEndAt time.Time) (int, error) {
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
	switch {
	case nextEstimate < 1:
		candidate = 0
	case nextEstimate > int64(storydomain.MaximumEstimatedDurationMinutes):
		candidate = storydomain.MaximumEstimatedDurationMinutes + 1
	default:
		candidate = int(nextEstimate)
	}
	if err := storydomain.ValidateScheduling(&candidate, minimumFocusBlock); err != nil {
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

func scheduleBlockConflicts(ctx context.Context, queries calendarsql.Querier, input calendar.CoreScheduleBlockInput, excludeBlockID uuid.UUID) (bool, error) {
	conflicts, err := queries.CalendarScheduleBlockConflicts(ctx, calendarsql.CalendarScheduleBlockConflictsParams{
		UserID:         input.UserID,
		EndAt:          input.EndAt,
		StartAt:        input.StartAt,
		ExcludeBlockID: excludeBlockID,
	})
	if err != nil {
		return false, fmt.Errorf("check calendar schedule conflict: %w", err)
	}
	return conflicts, nil
}

func (r *Repo) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, userID); err != nil {
			return err
		}
		rows, err := queries.DeleteCalendarScheduleBlock(ctx, calendarsql.DeleteCalendarScheduleBlockParams{
			WorkspaceID: workspaceID,
			UserID:      userID,
			BlockID:     blockID,
		})
		if err != nil {
			return fmt.Errorf("delete calendar schedule block: %w", err)
		}
		if rows == 0 {
			return calendar.ErrCalendarScheduleBlockNotFound
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete calendar schedule block transaction: %w", err)
	}
	return nil
}

func getScheduleBlockWithQueries(ctx context.Context, queries calendarsql.Querier, workspaceID, userID, blockID uuid.UUID) (calendar.CoreScheduleBlock, error) {
	row, err := queries.GetCalendarScheduleBlock(ctx, calendarsql.GetCalendarScheduleBlockParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		BlockID:     blockID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return calendar.CoreScheduleBlock{}, calendar.ErrCalendarScheduleBlockNotFound
	}
	if err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("get calendar schedule block: %w", err)
	}
	return toCoreScheduleBlock(scheduleBlockFromGet(row)), nil
}

func (r *Repo) GetScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) (calendar.CoreScheduleBlock, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	return getScheduleBlockWithQueries(ctx, r.queries, workspaceID, userID, blockID)
}

func int32PointerToInt(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}
