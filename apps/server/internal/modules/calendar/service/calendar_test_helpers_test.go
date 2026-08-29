package calendar

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (r *fakeRepo) MarkConnectionSynced(
	ctx context.Context,
	workspaceID, connectionID, credentialGeneration uuid.UUID,
	syncedAt time.Time,
) error {
	r.markedGeneration = credentialGeneration
	if r.markSyncedErr != nil {
		return r.markSyncedErr
	}
	r.connection.LastSyncedAt = &syncedAt
	r.connection.SyncStatus = SyncStatusSynced
	return nil
}

func (r *fakeRepo) MarkConnectionSyncFailed(
	ctx context.Context,
	workspaceID, connectionID, credentialGeneration uuid.UUID,
	message string,
) error {
	r.markedGeneration = credentialGeneration
	if r.markFailedErr != nil {
		return r.markFailedErr
	}
	r.connection.SyncStatus = SyncStatusFailed
	r.connection.SyncError = &message
	return nil
}

func (r *fakeRepo) ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error) {
	return r.windows, nil
}

func (r *fakeRepo) ListCalendarEvents(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreCalendarEventSummary, error) {
	return r.events, nil
}

func (r *fakeRepo) GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error) {
	if r.event.ID == uuid.Nil || r.event.ID != eventID {
		return CoreCalendarEvent{}, ErrCalendarEventNotFound
	}
	return r.event, nil
}

func (r *fakeRepo) ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error) {
	r.scheduleBlocksCalls++
	return r.blocks, nil
}

func (r *fakeRepo) GetScheduleBlock(_ context.Context, _, _ uuid.UUID, blockID uuid.UUID) (CoreScheduleBlock, error) {
	for _, block := range r.blocks {
		if block.ID == blockID {
			return block, nil
		}
	}
	return CoreScheduleBlock{}, ErrCalendarScheduleBlockNotFound
}

func (r *fakeRepo) ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	r.storyUserID = userID
	if r.storyAllowed == nil {
		return true, nil
	}
	return *r.storyAllowed, nil
}

func (r *fakeRepo) CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	block := CoreScheduleBlock{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		StoryID:     input.StoryID,
		BlockType:   input.BlockType,
		Title:       input.Title,
		StartAt:     input.StartAt,
		EndAt:       input.EndAt,
		IsLocked:    input.IsLocked,
		Source:      input.Source,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.blocks = append(r.blocks, block)
	return block, nil
}

func (r *fakeRepo) UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	for i := range r.blocks {
		if r.blocks[i].ID == input.ID {
			r.blocks[i].StoryID = input.StoryID
			r.blocks[i].BlockType = input.BlockType
			r.blocks[i].Title = input.Title
			r.blocks[i].StartAt = input.StartAt
			r.blocks[i].EndAt = input.EndAt
			r.blocks[i].IsLocked = input.IsLocked
			r.blocks[i].Source = input.Source
			return r.blocks[i], nil
		}
	}
	return CoreScheduleBlock{}, ErrCalendarScheduleBlockNotFound
}

func (r *fakeRepo) ManuallyRescheduleScheduleBlock(_ context.Context, input ManualScheduleBlockInput) (ManualScheduleBlockResult, error) {
	r.manualRescheduleInputs = append(r.manualRescheduleInputs, input)
	if r.manualRescheduleErr != nil {
		return ManualScheduleBlockResult{}, r.manualRescheduleErr
	}
	return r.manualRescheduleResult, nil
}

func (r *fakeRepo) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	for i := range r.blocks {
		if r.blocks[i].ID == blockID {
			r.blocks = append(r.blocks[:i], r.blocks[i+1:]...)
			return nil
		}
	}
	return ErrCalendarScheduleBlockNotFound
}

func (r *fakeRepo) ListSchedulingBlocksForUser(_ context.Context, _, _ uuid.UUID, _, _ time.Time) ([]CoreScheduleBlock, error) {
	r.accountScheduleBlockCalls++
	return r.blocks, nil
}

func (r *fakeRepo) ListMayaScheduleBlocksForStory(_ context.Context, workspaceID, userID, storyID uuid.UUID) ([]CoreScheduleBlock, error) {
	return r.blocks, nil
}

func (r *fakeRepo) MayaScheduleOwnershipExists(_ context.Context, _, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeRepo) ReconcileMayaScheduleBlocks(_ context.Context, _ MayaScheduleReconcileInput) (CoreScheduleReconcileResult, error) {
	return r.reconcileResult, nil
}

func (r *fakeRepo) ListReadyScheduleEventOutboxUsers(_ context.Context, _ int) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.readyOutboxUsers...), nil
}

func (r *fakeRepo) WithScheduleEventDispatchLock(_ context.Context, _ uuid.UUID, dispatch func(ScheduleEventOutboxStore) error) error {
	r.dispatchLockCalls++
	return dispatch(r)
}

func (r *fakeRepo) ListPendingScheduleEventOutbox(_ context.Context, _ uuid.UUID, _ Provider, _ int) ([]CoreScheduleEventOutbox, error) {
	r.outboxClaimCalls++
	if len(r.pendingOutboxBatches) == 0 {
		return nil, nil
	}
	batch := r.pendingOutboxBatches[0]
	r.pendingOutboxBatches = r.pendingOutboxBatches[1:]
	return batch, nil
}

func (r *fakeRepo) ScheduleEventUpsertIsCurrent(_ context.Context, _ CoreScheduleEventOutbox, _ ExternalScheduleEventInput) (bool, error) {
	r.upsertCurrentChecks++
	if r.upsertCurrent == nil {
		return true, nil
	}
	return *r.upsertCurrent, nil
}

func (r *fakeRepo) MarkScheduleEventOutboxProcessed(_ context.Context, item CoreScheduleEventOutbox, _ string) error {
	r.processedOutbox = append(r.processedOutbox, item.ID)
	r.processedOutboxOperations = append(r.processedOutboxOperations, item.Operation)
	return nil
}

func (r *fakeRepo) MarkScheduleEventOutboxFailed(_ context.Context, item CoreScheduleEventOutbox, _ string, permanent bool) error {
	r.failedOutbox = append(r.failedOutbox, item.ID)
	r.failedOutboxPermanent = append(r.failedOutboxPermanent, permanent)
	return nil
}

func (r *fakeRepo) ReleaseScheduleEventOutbox(_ context.Context, outboxIDs []uuid.UUID) error {
	r.releasedOutbox = append(r.releasedOutbox, outboxIDs...)
	return nil
}

func (r *fakeRepo) DeleteCleanupPendingConnectionIfDrained(_ context.Context, _ uuid.UUID, _ Provider) error {
	r.cleanupFinalizeCalls++
	return nil
}
