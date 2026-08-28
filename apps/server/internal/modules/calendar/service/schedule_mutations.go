package calendar

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error) {
	if s.repo == nil {
		return CoreCalendarEvent{}, ErrCalendarNotConfigured
	}
	if eventID == uuid.Nil {
		return CoreCalendarEvent{}, ErrCalendarEventNotFound
	}
	return s.repo.GetCalendarEvent(ctx, workspaceID, userID, eventID)
}

func (s *Service) CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	normalized, err := normalizeScheduleBlockInput(input, s.now())
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if err := s.validateScheduleStory(ctx, normalized); err != nil {
		return CoreScheduleBlock{}, err
	}
	return s.repo.CreateScheduleBlock(ctx, normalized)
}

func (s *Service) UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	if input.ID == uuid.Nil {
		return CoreScheduleBlock{}, ErrInvalidScheduleBlock
	}
	existing, err := s.repo.GetScheduleBlock(ctx, input.WorkspaceID, input.UserID, input.ID)
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if existing.Source == ScheduleBlockSourceMaya {
		return CoreScheduleBlock{}, ErrManagedScheduleBlock
	}
	normalized, err := normalizeScheduleBlockInput(input, s.now())
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if err := s.validateScheduleStory(ctx, normalized); err != nil {
		return CoreScheduleBlock{}, err
	}
	normalized.ID = input.ID
	return s.repo.UpdateScheduleBlock(ctx, normalized)
}

func (s *Service) ManuallyRescheduleScheduleBlock(ctx context.Context, input ManualScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.ActorID == uuid.Nil || input.BlockID == uuid.Nil || input.ClientMutationID == uuid.Nil {
		return CoreScheduleBlock{}, ErrInvalidScheduleBlock
	}
	if input.Change != ManualScheduleBlockChangeMove && input.Change != ManualScheduleBlockChangeResize {
		return CoreScheduleBlock{}, ErrInvalidScheduleBlock
	}
	if err := validateScheduleRange(input.StartAt, input.EndAt); err != nil {
		return CoreScheduleBlock{}, err
	}
	if strings.TrimSpace(input.Timezone) == "" {
		input.Timezone = "UTC"
	}
	manualRepo, ok := s.repo.(ManualScheduleBlockRepository)
	if !ok {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	result, err := manualRepo.ManuallyRescheduleScheduleBlock(ctx, input)
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if result.StoryScheduleReconcileID != nil {
		s.enqueueStoryScheduleReconciliation(ctx, input.WorkspaceID, *result.StoryScheduleReconcileID)
	}
	if s.cfg.Updates != nil {
		if publishErr := s.cfg.Updates.PublishCalendarUpdated(ctx, input.WorkspaceID, input.UserID, uuid.Nil, s.now().UTC()); publishErr != nil && s.log != nil {
			s.log.Error(ctx, "failed to publish manual calendar update", "error", publishErr, "block_id", input.BlockID)
		}
	}
	return result.Block, nil
}

func (s *Service) enqueueStoryScheduleReconciliation(ctx context.Context, workspaceID, storyID uuid.UUID) {
	tasks, ok := s.cfg.Tasks.(StoryScheduleTaskQueue)
	if !ok {
		return
	}
	enqueueCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := tasks.EnqueueStoryScheduleReconcile(enqueueCtx, workspaceID, storyID); err != nil && s.log != nil {
		s.log.Error(ctx, "failed to enqueue story schedule reconciliation after manual calendar resize", "error", err, "workspace_id", workspaceID, "story_id", storyID)
	}
}

func (s *Service) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	if blockID == uuid.Nil {
		return ErrInvalidScheduleBlock
	}
	existing, err := s.repo.GetScheduleBlock(ctx, workspaceID, userID, blockID)
	if err != nil {
		return err
	}
	if existing.Source == ScheduleBlockSourceMaya {
		return ErrManagedScheduleBlock
	}
	return s.repo.DeleteScheduleBlock(ctx, workspaceID, userID, blockID)
}

func (s *Service) ReconcileMayaScheduleBlocks(ctx context.Context, input MayaScheduleReconcileInput) (CoreScheduleReconcileResult, error) {
	if s.repo == nil {
		return CoreScheduleReconcileResult{}, ErrCalendarNotConfigured
	}
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.StoryID == uuid.Nil {
		return CoreScheduleReconcileResult{}, ErrInvalidScheduleBlock
	}
	if input.AllowConflicts && !input.Locked {
		return CoreScheduleReconcileResult{}, ErrInvalidScheduleBlock
	}
	if len(input.Segments) > 0 {
		if exists, err := s.repo.ScheduleStoryExists(ctx, input.WorkspaceID, input.UserID, input.StoryID); err != nil {
			return CoreScheduleReconcileResult{}, err
		} else if !exists {
			return CoreScheduleReconcileResult{}, ErrCalendarAccessDenied
		}
	}
	for index := range input.Segments {
		input.Segments[index].Title = strings.TrimSpace(input.Segments[index].Title)
		if input.Segments[index].Title == "" || len(input.Segments[index].Title) > 255 {
			return CoreScheduleReconcileResult{}, ErrInvalidScheduleBlock
		}
		if err := validateScheduleRange(input.Segments[index].StartAt, input.Segments[index].EndAt); err != nil {
			return CoreScheduleReconcileResult{}, err
		}
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return CoreScheduleReconcileResult{}, err
	}
	result, err := scheduleRepo.ReconcileMayaScheduleBlocks(ctx, input)
	if err != nil {
		return CoreScheduleReconcileResult{}, err
	}
	if s.cfg.Updates != nil && scheduleReconcileChangesCalendar(result.Actions) {
		if publishErr := s.cfg.Updates.PublishCalendarUpdated(
			ctx,
			input.WorkspaceID,
			input.UserID,
			uuid.Nil,
			s.now().UTC(),
		); publishErr != nil {
			// Reconciliation is already durably committed. Realtime invalidation is
			// advisory and must not make the caller compensate a successful plan.
			if s.log != nil {
				s.log.Error(ctx, "failed to publish reconciled calendar update", "error", publishErr, "story_id", input.StoryID, "user_id", input.UserID)
			}
		}
	}
	return result, nil
}

func scheduleReconcileChangesCalendar(actions []ScheduleReconcileAction) bool {
	for _, action := range actions {
		if action != ScheduleReconcileActionUnchanged {
			return true
		}
	}
	return false
}

func (s *Service) ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]CoreScheduleBlock, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || storyID == uuid.Nil {
		return nil, ErrInvalidScheduleBlock
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return nil, err
	}
	return scheduleRepo.ListMayaScheduleBlocksForStory(ctx, workspaceID, userID, storyID)
}

func (s *Service) MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || storyID == uuid.Nil {
		return false, ErrInvalidScheduleBlock
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return false, err
	}
	return scheduleRepo.MayaScheduleOwnershipExists(ctx, workspaceID, userID, storyID)
}
