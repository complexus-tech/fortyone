package calendarrepository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type reconciliationBlock struct {
	ID                 uuid.UUID
	SegmentIndex       int
	Title              string
	StartAt            time.Time
	EndAt              time.Time
	IsLocked           bool
	ExternalProvider   *string
	ExternalCalendarID *string
	ExternalEventID    *string
	ExternalSyncHash   *string
	ManualOverrideAt   *time.Time
	ManualOverrideBy   *uuid.UUID
}

func (r *Repo) ReconcileMayaScheduleBlocks(ctx context.Context, input calendar.MayaScheduleReconcileInput) (calendar.CoreScheduleReconcileResult, error) {
	segments, err := validatedMayaScheduleSegments(input.Segments)
	if err != nil {
		return calendar.CoreScheduleReconcileResult{}, err
	}
	if err := r.configured(); err != nil {
		return calendar.CoreScheduleReconcileResult{}, err
	}

	var result calendar.CoreScheduleReconcileResult
	err = r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, input.UserID); err != nil {
			return err
		}
		if input.ExpectedStoryUpdatedAt != nil {
			_, err := queries.LockMayaScheduleStoryVersion(ctx, calendarsql.LockMayaScheduleStoryVersionParams{
				WorkspaceID:       input.WorkspaceID,
				StoryID:           input.StoryID,
				ExpectedUpdatedAt: input.ExpectedStoryUpdatedAt.UTC(),
			})
			if errors.Is(err, pgx.ErrNoRows) {
				return calendar.ErrCalendarScheduleStalePlan
			}
			if err != nil {
				return fmt.Errorf("lock Maya schedule story version: %w", err)
			}
		}
		if len(segments) > 0 {
			_, err := queries.LockEligibleMayaScheduleStory(ctx, calendarsql.LockEligibleMayaScheduleStoryParams{
				UserID:      input.UserID,
				WorkspaceID: input.WorkspaceID,
				StoryID:     input.StoryID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				return calendar.ErrInvalidScheduleBlock
			}
			if err != nil {
				return fmt.Errorf("lock eligible Maya schedule story: %w", err)
			}
		}

		rows, err := queries.ListExistingMayaScheduleSegments(ctx, calendarsql.ListExistingMayaScheduleSegmentsParams{
			WorkspaceID: input.WorkspaceID,
			UserID:      input.UserID,
			StoryID:     &input.StoryID,
		})
		if err != nil {
			return fmt.Errorf("list existing Maya schedule segments: %w", err)
		}
		current := make([]reconciliationBlock, len(rows))
		for index, row := range rows {
			current[index] = reconciliationBlock{
				ID: row.BlockID, SegmentIndex: int(row.SegmentIndex), Title: row.Title,
				StartAt: row.StartAt, EndAt: row.EndAt, IsLocked: row.IsLocked,
				ExternalProvider: row.ExternalProvider, ExternalCalendarID: row.ExternalCalendarID,
				ExternalEventID: row.ExternalEventID, ExternalSyncHash: row.ExternalSyncHash,
				ManualOverrideAt: row.ManualOverrideAt, ManualOverrideBy: row.ManualOverrideBy,
			}
		}
		currentByIndex, manualByIndex, excludedIDs := indexReconciliationBlocks(current, input.PreemptBlockIDs)
		scheduleProvider, err := preferredScheduleProvider(ctx, queries, input.UserID, current)
		if err != nil {
			return err
		}
		if err := validateMayaScheduleConflicts(ctx, queries, input, segments, currentByIndex, manualByIndex, excludedIDs); err != nil {
			return err
		}

		result = calendar.CoreScheduleReconcileResult{
			Blocks:  make([]calendar.CoreScheduleBlock, 0, len(segments)),
			Actions: make([]calendar.ScheduleReconcileAction, 0, len(segments)+len(current)),
		}
		for _, segment := range segments {
			if manual, exists := manualByIndex[segment.SegmentIndex]; exists {
				result.Blocks = append(result.Blocks, toManualOverrideScheduleBlock(input, manual))
				result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUnchanged)
				delete(manualByIndex, segment.SegmentIndex)
				continue
			}
			block, exists := currentByIndex[segment.SegmentIndex]
			blockID := block.ID
			if !exists {
				blockID = uuid.New()
			}
			event := scheduleEventForSegment(input, segment, block, exists, blockID, scheduleProvider)
			syncHash := calendar.ScheduleEventSyncHash(event)
			providerChanged := reconciliationProviderStateChanged(block, exists, scheduleProvider, event)
			blockChanged := providerChanged || block.IsLocked != input.Locked
			if err := persistMayaScheduleSegment(ctx, queries, input, segment, blockID, scheduleProvider, event.EventID, exists); err != nil {
				return err
			}
			if scheduleBlockNeedsProviderUpsert(block, exists, scheduleProvider, event, syncHash) {
				if err := enqueueScheduleEventOutbox(ctx, queries, input.WorkspaceID, input.UserID, &blockID, scheduleProvider, calendar.ScheduleEventOperationUpsert, event, syncHash, providerChanged); err != nil {
					return err
				}
			}
			switch {
			case !exists:
				result.Actions = append(result.Actions, calendar.ScheduleReconcileActionCreated)
			case blockChanged:
				result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUpdated)
			default:
				result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUnchanged)
			}
			result.Blocks = append(result.Blocks, reconciledScheduleBlock(input, segment, block, blockID, scheduleProvider, event.EventID))
			delete(currentByIndex, segment.SegmentIndex)
		}

		for _, block := range currentByIndex {
			provider := scheduleProvider
			if block.ExternalProvider != nil && strings.TrimSpace(*block.ExternalProvider) != "" {
				provider = calendar.Provider(*block.ExternalProvider)
			}
			eventID := calendar.StableGoogleScheduleEventID(block.ID)
			if block.ExternalEventID != nil && strings.TrimSpace(*block.ExternalEventID) != "" {
				eventID = strings.TrimSpace(*block.ExternalEventID)
			}
			event := calendar.ExternalScheduleEventInput{
				CalendarID: "primary", EventID: eventID, BlockID: block.ID,
				StoryID: input.StoryID, WorkspaceID: input.WorkspaceID,
			}
			if err := enqueueScheduleEventOutbox(ctx, queries, input.WorkspaceID, input.UserID, &block.ID, provider, calendar.ScheduleEventOperationDelete, event, "", true); err != nil {
				return err
			}
			if err := queries.DeleteMayaScheduleSegment(ctx, calendarsql.DeleteMayaScheduleSegmentParams{BlockID: block.ID}); err != nil {
				return fmt.Errorf("delete obsolete Maya schedule segment: %w", err)
			}
			result.Actions = append(result.Actions, calendar.ScheduleReconcileActionDeleted)
		}
		for _, block := range manualByIndex {
			result.Blocks = append(result.Blocks, toManualOverrideScheduleBlock(input, block))
			result.Actions = append(result.Actions, calendar.ScheduleReconcileActionUnchanged)
		}

		if len(segments) > 0 || input.KeepOwnership {
			if err := queries.RetainMayaScheduleOwnership(ctx, calendarsql.RetainMayaScheduleOwnershipParams{
				WorkspaceID: input.WorkspaceID, StoryID: input.StoryID, UserID: input.UserID,
			}); err != nil {
				return fmt.Errorf("retain Maya schedule ownership: %w", err)
			}
		} else if err := queries.ReleaseMayaScheduleOwnership(ctx, calendarsql.ReleaseMayaScheduleOwnershipParams{
			WorkspaceID: input.WorkspaceID, StoryID: input.StoryID, UserID: input.UserID,
		}); err != nil {
			return fmt.Errorf("release Maya schedule ownership: %w", err)
		}
		return nil
	})
	if err != nil {
		return calendar.CoreScheduleReconcileResult{}, fmt.Errorf("reconcile Maya schedule blocks transaction: %w", err)
	}
	return result, nil
}

func validatedMayaScheduleSegments(input []calendar.MayaScheduleSegmentInput) ([]calendar.MayaScheduleSegmentInput, error) {
	segments := append([]calendar.MayaScheduleSegmentInput(nil), input...)
	sort.SliceStable(segments, func(left, right int) bool { return segments[left].SegmentIndex < segments[right].SegmentIndex })
	for index, segment := range segments {
		if segment.SegmentIndex != index || strings.TrimSpace(segment.Title) == "" || !segment.EndAt.After(segment.StartAt) {
			return nil, calendar.ErrInvalidScheduleBlock
		}
		if index > 0 && segments[index-1].EndAt.After(segment.StartAt) {
			return nil, calendar.ErrCalendarScheduleConflict
		}
	}
	return segments, nil
}

func indexReconciliationBlocks(current []reconciliationBlock, preempt []uuid.UUID) (map[int]reconciliationBlock, map[int]reconciliationBlock, []uuid.UUID) {
	currentByIndex := make(map[int]reconciliationBlock, len(current))
	manualByIndex := make(map[int]reconciliationBlock)
	excludedIDs := make([]uuid.UUID, 0, len(current)+len(preempt))
	for _, block := range current {
		if block.ManualOverrideAt != nil {
			manualByIndex[block.SegmentIndex] = block
		} else {
			currentByIndex[block.SegmentIndex] = block
		}
		excludedIDs = append(excludedIDs, block.ID)
	}
	for _, blockID := range preempt {
		if blockID != uuid.Nil {
			excludedIDs = append(excludedIDs, blockID)
		}
	}
	return currentByIndex, manualByIndex, excludedIDs
}

func validateMayaScheduleConflicts(ctx context.Context, queries calendarsql.Querier, input calendar.MayaScheduleReconcileInput, segments []calendar.MayaScheduleSegmentInput, current, manual map[int]reconciliationBlock, excluded []uuid.UUID) error {
	for _, segment := range segments {
		if input.AllowConflicts {
			continue
		}
		if _, exists := manual[segment.SegmentIndex]; exists {
			continue
		}
		if block, exists := current[segment.SegmentIndex]; input.Locked && exists && block.StartAt.Equal(segment.StartAt) && block.EndAt.Equal(segment.EndAt) {
			continue
		}
		conflicts, err := queries.MayaScheduleSegmentConflicts(ctx, calendarsql.MayaScheduleSegmentConflictsParams{
			UserID: input.UserID, EndAt: segment.EndAt, StartAt: segment.StartAt, ExcludedBlockIds: excluded,
		})
		if err != nil {
			return fmt.Errorf("validate Maya schedule segment: %w", err)
		}
		if conflicts {
			return calendar.ErrCalendarScheduleConflict
		}
	}
	return nil
}

func persistMayaScheduleSegment(ctx context.Context, queries calendarsql.Querier, input calendar.MayaScheduleReconcileInput, segment calendar.MayaScheduleSegmentInput, blockID uuid.UUID, provider calendar.Provider, eventID string, exists bool) error {
	segmentIndex, err := safecast.Int32(segment.SegmentIndex)
	if err != nil {
		return fmt.Errorf("validate Maya schedule segment index: %w", err)
	}
	if exists {
		if err := queries.UpdateMayaScheduleSegment(ctx, calendarsql.UpdateMayaScheduleSegmentParams{
			Title: segment.Title, StartAt: segment.StartAt, EndAt: segment.EndAt, IsLocked: input.Locked,
			ExternalProvider: string(provider), ExternalEventID: eventID,
			WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: &input.StoryID,
			SegmentIndex: segmentIndex,
		}); err != nil {
			return fmt.Errorf("update Maya schedule segment: %w", err)
		}
		return nil
	}
	if err := queries.CreateMayaScheduleSegment(ctx, calendarsql.CreateMayaScheduleSegmentParams{
		BlockID: blockID, WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: &input.StoryID,
		Title: segment.Title, StartAt: segment.StartAt, EndAt: segment.EndAt, IsLocked: input.Locked,
		SegmentIndex: segmentIndex, ExternalProvider: string(provider), ExternalEventID: eventID,
	}); err != nil {
		return fmt.Errorf("create Maya schedule segment: %w", err)
	}
	return nil
}

func preferredScheduleProvider(ctx context.Context, queries calendarsql.Querier, userID uuid.UUID, blocks []reconciliationBlock) (calendar.Provider, error) {
	connections, err := queries.ListCalendarWriteDestinations(ctx, calendarsql.ListCalendarWriteDestinationsParams{
		GoogleReadScope: calendar.GoogleCalendarEventsReadonlyScope, GoogleOwnedScope: calendar.GoogleCalendarEventsOwnedScope,
		MicrosoftWriteScope: calendar.MicrosoftCalendarReadWriteScope, UserID: userID,
	})
	if err != nil {
		return "", fmt.Errorf("list calendar write destinations: %w", err)
	}
	active := make(map[string]struct{}, len(connections))
	for _, connection := range connections {
		active[connection.Provider] = struct{}{}
	}
	for _, block := range blocks {
		if block.ExternalProvider != nil {
			if _, ok := active[*block.ExternalProvider]; ok {
				return calendar.Provider(*block.ExternalProvider), nil
			}
		}
	}
	for _, connection := range connections {
		if connection.IsPrimary {
			return calendar.Provider(connection.Provider), nil
		}
	}
	for _, connection := range connections {
		if connection.CanWrite {
			return calendar.Provider(connection.Provider), nil
		}
	}
	return calendar.ProviderGoogle, nil
}

func scheduleEventForSegment(input calendar.MayaScheduleReconcileInput, segment calendar.MayaScheduleSegmentInput, block reconciliationBlock, exists bool, blockID uuid.UUID, provider calendar.Provider) calendar.ExternalScheduleEventInput {
	eventID := calendar.StableGoogleScheduleEventID(blockID)
	if provider == calendar.ProviderMicrosoft {
		eventID = "pending:" + blockID.String()
		if exists && block.ExternalProvider != nil && *block.ExternalProvider == string(provider) && block.ExternalEventID != nil && strings.TrimSpace(*block.ExternalEventID) != "" {
			eventID = strings.TrimSpace(*block.ExternalEventID)
		}
	}
	return calendar.ExternalScheduleEventInput{
		CalendarID: "primary", EventID: eventID, BlockID: blockID,
		StoryID: input.StoryID, WorkspaceID: input.WorkspaceID, Title: segment.Title,
		StartAt: segment.StartAt.UTC(), EndAt: segment.EndAt.UTC(),
		PrivateProperties: map[string]string{
			"fortyone_source": "maya_schedule", "fortyone_block_id": blockID.String(),
			"fortyone_story_id": input.StoryID.String(), "fortyone_workspace_id": input.WorkspaceID.String(),
		},
	}
}

func reconciliationProviderStateChanged(block reconciliationBlock, exists bool, provider calendar.Provider, event calendar.ExternalScheduleEventInput) bool {
	return !exists || block.Title != event.Title || !block.StartAt.Equal(event.StartAt) || !block.EndAt.Equal(event.EndAt) ||
		block.ExternalProvider == nil || *block.ExternalProvider != string(provider) ||
		block.ExternalCalendarID == nil || *block.ExternalCalendarID != event.CalendarID ||
		block.ExternalEventID == nil || *block.ExternalEventID != event.EventID
}

func scheduleBlockNeedsProviderUpsert(block reconciliationBlock, exists bool, provider calendar.Provider, event calendar.ExternalScheduleEventInput, syncHash string) bool {
	if reconciliationProviderStateChanged(block, exists, provider, event) {
		return true
	}
	return block.ExternalSyncHash == nil || *block.ExternalSyncHash != syncHash
}

func reconciledScheduleBlock(input calendar.MayaScheduleReconcileInput, segment calendar.MayaScheduleSegmentInput, block reconciliationBlock, blockID uuid.UUID, provider calendar.Provider, eventID string) calendar.CoreScheduleBlock {
	storyID, calendarID := input.StoryID, "primary"
	return calendar.CoreScheduleBlock{
		ID: blockID, WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: &storyID,
		BlockType: calendar.ScheduleBlockTypeWork, Title: segment.Title,
		StartAt: segment.StartAt, EndAt: segment.EndAt, IsLocked: input.Locked,
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: segment.SegmentIndex,
		ExternalProvider: &provider, ExternalCalendarID: &calendarID, ExternalEventID: &eventID,
		ExternalSyncHash: block.ExternalSyncHash,
	}
}

func toManualOverrideScheduleBlock(input calendar.MayaScheduleReconcileInput, block reconciliationBlock) calendar.CoreScheduleBlock {
	storyID, provider, calendarID := input.StoryID, calendar.ProviderGoogle, "primary"
	if block.ExternalProvider != nil && strings.TrimSpace(*block.ExternalProvider) != "" {
		provider = calendar.Provider(*block.ExternalProvider)
	}
	return calendar.CoreScheduleBlock{
		ID: block.ID, WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: &storyID,
		BlockType: calendar.ScheduleBlockTypeWork, Title: block.Title, StartAt: block.StartAt, EndAt: block.EndAt,
		IsLocked: true, Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: block.SegmentIndex,
		ExternalProvider: &provider, ExternalCalendarID: &calendarID, ExternalEventID: block.ExternalEventID,
		ExternalSyncHash: block.ExternalSyncHash, ManualOverrideAt: block.ManualOverrideAt, ManualOverrideBy: block.ManualOverrideBy,
	}
}
