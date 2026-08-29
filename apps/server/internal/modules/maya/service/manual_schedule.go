package maya

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

const manualScheduleReason = "You chose this time, so Maya locked it on your calendar."

type ManualScheduleStoryInput struct {
	WorkspaceID uuid.UUID
	StoryID     uuid.UUID
	UserID      uuid.UUID
	StartAt     time.Time
	Timezone    string
}

func (s *Service) RetryScheduleIssue(ctx context.Context, workspaceID, storyID, userID uuid.UUID) error {
	if workspaceID == uuid.Nil || storyID == uuid.Nil || userID == uuid.Nil {
		return fmt.Errorf("%w: workspace, story, and user are required", ErrInvalidPlanInput)
	}
	if err := s.validateScheduleIssueOwner(ctx, workspaceID, storyID, userID); err != nil {
		return err
	}
	return s.ReconcileSchedule(ctx, ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID})
}

// ManuallyScheduleStory turns an unresolved placement failure into one exact,
// locked Maya block. The selected start is explicit user intent, so overlap
// checks are bypassed while story eligibility, duration, version, ownership,
// and provider-outbox guarantees remain in force.
func (s *Service) ManuallyScheduleStory(ctx context.Context, input ManualScheduleStoryInput) (calendar.CoreScheduleBlock, error) {
	if err := s.validate(); err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	if input.WorkspaceID == uuid.Nil || input.StoryID == uuid.Nil || input.UserID == uuid.Nil || input.StartAt.IsZero() {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("%w: workspace, story, user, and start time are required", ErrInvalidPlanInput)
	}
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("%w: timezone is invalid", ErrInvalidPlanInput)
	}
	now := s.clock.Now().UTC()
	if input.StartAt.Before(now.Add(-5*time.Minute)) || input.StartAt.After(now.Add(90*24*time.Hour)) {
		return calendar.CoreScheduleBlock{}, fmt.Errorf("%w: start time must be within the next 90 days", ErrInvalidPlanInput)
	}

	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	hasAccess, err := scheduleRepo.WorkspaceCanUseMaya(ctx, input.WorkspaceID)
	if err != nil {
		return calendar.CoreScheduleBlock{}, err
	}
	if !hasAccess {
		return calendar.CoreScheduleBlock{}, ErrMayaAccessDenied
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return calendar.CoreScheduleBlock{}, err
	}

	var scheduledBlock calendar.CoreScheduleBlock
	err = scheduleRepo.WithScheduleStoryLock(ctx, input.WorkspaceID, input.StoryID, func() error {
		story, getErr := s.stories.Get(ctx, input.StoryID, input.WorkspaceID)
		if getErr != nil {
			return getErr
		}
		if story.Assignee == nil || *story.Assignee != input.UserID || !story.AutoSchedulingEnabled {
			return fmt.Errorf("%w: the story is not assigned to this user for auto-scheduling", ErrInvalidPlanInput)
		}
		if story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes <= 0 {
			return fmt.Errorf("%w: the story needs a valid time estimate", ErrInvalidPlanInput)
		}
		existingBlocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, input.WorkspaceID, input.UserID, input.StoryID)
		if listErr != nil {
			return listErr
		}
		scheduledMinutes := 0
		existingSegments := make([]calendar.MayaScheduleSegmentInput, 0, len(existingBlocks))
		for index, block := range existingBlocks {
			scheduledMinutes += int(block.EndAt.Sub(block.StartAt) / time.Minute)
			existingSegments = append(existingSegments, calendar.MayaScheduleSegmentInput{
				SegmentIndex: index,
				Title:        story.Title,
				StartAt:      block.StartAt,
				EndAt:        block.EndAt,
			})
		}
		remainingMinutes := *story.EstimatedDurationMinutes - scheduledMinutes
		if remainingMinutes <= 0 {
			return storydomain.ErrStoryChanged
		}
		endAt := input.StartAt.UTC().Add(time.Duration(remainingMinutes) * time.Minute)
		if story.AutoSchedulingLocked && exactManualScheduleExists(existingBlocks, input.StartAt.UTC(), endAt) {
			scheduledBlock = existingBlocks[0]
			return nil
		}
		if story.AutoSchedulingStatus != storydomain.AutoSchedulingStatusCannotFit {
			return storydomain.ErrStoryChanged
		}
		schedulable, eligibilityErr := scheduleRepo.StoryIsSchedulableForUser(ctx, input.WorkspaceID, input.StoryID, input.UserID)
		if eligibilityErr != nil {
			return eligibilityErr
		}
		if !schedulable {
			return fmt.Errorf("%w: the story is no longer eligible for scheduling", ErrInvalidPlanInput)
		}
		previousOwnership, ownershipErr := scheduleCalendar.MayaScheduleOwnershipExists(ctx, input.WorkspaceID, input.UserID, input.StoryID)
		if ownershipErr != nil {
			return ownershipErr
		}
		segments := append(existingSegments, calendar.MayaScheduleSegmentInput{
			SegmentIndex: len(existingSegments),
			Title:        story.Title,
			StartAt:      input.StartAt.UTC(),
			EndAt:        endAt,
		})
		result, reconcileErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
			WorkspaceID:            input.WorkspaceID,
			UserID:                 input.UserID,
			StoryID:                input.StoryID,
			ExpectedStoryUpdatedAt: &story.UpdatedAt,
			Segments:               segments,
			KeepOwnership:          true,
			Locked:                 true,
			AllowConflicts:         true,
		})
		if reconcileErr != nil {
			return reconcileErr
		}

		locked := true
		reason := manualScheduleReason
		transition := buildStoryScheduleTransition(
			story,
			input.UserID,
			existingBlocks,
			segments,
			input.Timezone,
			storydomain.AutoSchedulingStatusLocked,
			manualScheduleReason,
		)
		if updateErr := s.stories.UpdateAutomationStateIfUnchanged(
			ctx,
			s.mayaActorID,
			story.ID,
			story.Workspace,
			story.UpdatedAt,
			storydomain.AutoSchedulingStatusLocked,
			&reason,
			&locked,
			transition,
		); updateErr != nil {
			_, restoreErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
				WorkspaceID:   input.WorkspaceID,
				UserID:        input.UserID,
				StoryID:       input.StoryID,
				Segments:      existingSegments,
				KeepOwnership: previousOwnership,
			})
			return errors.Join(updateErr, restoreErr)
		}
		if len(result.Blocks) != len(segments) {
			return errors.New("manual Maya schedule did not preserve all focus blocks")
		}
		for _, block := range result.Blocks {
			if block.SegmentIndex == len(existingSegments) {
				scheduledBlock = block
				break
			}
		}
		if scheduledBlock.ID == uuid.Nil {
			return errors.New("manual Maya schedule did not create the remaining focus block")
		}
		return nil
	})
	if err != nil {
		return calendar.CoreScheduleBlock{}, err
	}

	// The local schedule and story state are already durable. Provider delivery
	// remains retryable through the calendar outbox and must not undo the choice.
	_ = scheduleCalendar.DispatchScheduleEventOutbox(ctx, input.UserID)
	return scheduledBlock, nil
}

func (s *Service) validateScheduleIssueOwner(ctx context.Context, workspaceID, storyID, userID uuid.UUID) error {
	if err := s.validate(); err != nil {
		return err
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return err
	}
	hasAccess, err := scheduleRepo.WorkspaceCanUseMaya(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrMayaAccessDenied
	}
	story, err := s.stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}
	if story.Assignee == nil || *story.Assignee != userID || !story.AutoSchedulingEnabled {
		return fmt.Errorf("%w: the story is not assigned to this user for auto-scheduling", ErrInvalidPlanInput)
	}
	if story.AutoSchedulingStatus != storydomain.AutoSchedulingStatusCannotFit {
		return storydomain.ErrStoryChanged
	}
	return nil
}

func exactManualScheduleExists(blocks []calendar.CoreScheduleBlock, startAt, endAt time.Time) bool {
	return len(blocks) == 1 &&
		blocks[0].IsLocked &&
		blocks[0].StartAt.Equal(startAt) &&
		blocks[0].EndAt.Equal(endAt)
}
