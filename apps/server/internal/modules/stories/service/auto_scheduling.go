package stories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

const (
	autoSchedulingNeedsOwnerReason = "Choose an assignee so Maya can schedule this story."
	autoSchedulingSelectingReason  = "Maya is selecting the best available teammate."
	autoSchedulingNeedsTimeReason  = "Add time needed so Maya can reserve focused work."
	autoSchedulingPlanningReason   = "Maya is checking availability and scheduling this story."
	autoSchedulingLockedReason     = "Maya will keep the current scheduled time until this story is unlocked."
)

var autoSchedulingReconcileFields = map[string]struct{}{
	"title":                       {},
	"assignee_id":                 {},
	"status_id":                   {},
	"priority":                    {},
	"sprint_id":                   {},
	"start_date":                  {},
	"end_date":                    {},
	"completed_at":                {},
	"estimated_duration_minutes":  {},
	"minimum_focus_block_minutes": {},
	"auto_scheduling_enabled":     {},
	"auto_scheduling_locked":      {},
}

func (s *Service) prepareAutoSchedulingCreate(story *CoreSingleStory) error {
	if story == nil {
		return errors.New("story is required")
	}
	if story.AutoSchedulingLocked {
		return ErrAutoSchedulingLockEmpty
	}
	if s.isMayaActor(story.Assignee) {
		story.AutoSchedulingEnabled = true
	}
	story.AutoSchedulingLocked = false
	if !story.AutoSchedulingEnabled {
		story.AutoSchedulingStatus = AutoSchedulingStatusOff
		story.AutoSchedulingReason = nil
		story.AutoSchedulingUpdatedAt = nil
		return nil
	}

	status, reason := s.initialAutoSchedulingState(story.Assignee, story.EstimatedDurationMinutes)
	now := time.Now().UTC()
	story.AutoSchedulingStatus = status
	story.AutoSchedulingReason = reason
	story.AutoSchedulingUpdatedAt = &now
	return ValidateStoryAutoSchedulingContract(true, false, status)
}

func (s *Service) initialAutoSchedulingState(assigneeID *uuid.UUID, durationMinutes *int) (string, *string) {
	if assigneeID == nil || *assigneeID == uuid.Nil {
		return AutoSchedulingStatusNeedsOwner, stringPointer(autoSchedulingNeedsOwnerReason)
	}
	if s.isMayaActor(assigneeID) {
		return AutoSchedulingStatusNeedsOwner, stringPointer(autoSchedulingSelectingReason)
	}
	if durationMinutes == nil {
		return AutoSchedulingStatusNeedsTime, stringPointer(autoSchedulingNeedsTimeReason)
	}
	return AutoSchedulingStatusPlanning, stringPointer(autoSchedulingPlanningReason)
}

func (s *Service) prepareAutoSchedulingUpdate(ctx context.Context, story CoreSingleStory, updates map[string]any) (bool, error) {
	relevant := false
	for field := range updates {
		if _, ok := autoSchedulingReconcileFields[field]; ok {
			relevant = true
			break
		}
	}
	if !relevant {
		return false, nil
	}

	_, assigneeWasUpdated := updates["assignee_id"]
	assigneeID, err := updatedAssignee(story.Assignee, updates)
	if err != nil {
		return false, err
	}
	enabled, err := updatedBool(story.AutoSchedulingEnabled, updates, "auto_scheduling_enabled")
	if err != nil {
		return false, err
	}
	if assigneeWasUpdated && s.isMayaActor(assigneeID) {
		enabled = true
		updates["auto_scheduling_enabled"] = true
	}
	locked, err := updatedBool(story.AutoSchedulingLocked, updates, "auto_scheduling_locked")
	if err != nil {
		return false, err
	}
	if assigneeWasUpdated && !sameOptionalUUID(story.Assignee, assigneeID) && story.AutoSchedulingLocked && enabled {
		rawUnlock, unlockRequested := updates["auto_scheduling_locked"]
		if !unlockRequested {
			return false, ErrAutoSchedulingOwnerLocked
		}
		unlocking, unlockErr := updatedBool(story.AutoSchedulingLocked, map[string]any{"auto_scheduling_locked": rawUnlock}, "auto_scheduling_locked")
		if unlockErr != nil || unlocking {
			return false, ErrAutoSchedulingOwnerLocked
		}
	}
	terminal, err := s.autoSchedulingUpdateIsTerminal(ctx, story, updates)
	if err != nil {
		return false, err
	}
	if terminal {
		updates["auto_scheduling_locked"] = false
		updates["auto_scheduling_status"] = AutoSchedulingStatusOff
		updates["auto_scheduling_reason"] = stringPointer("Auto-scheduling stopped because this story is complete or cancelled.")
		updates["auto_scheduling_updated_at"] = time.Now().UTC()
		return story.AutoSchedulingEnabled || story.AutoSchedulingLocked || terminalLifecycleUpdateRequested(updates), nil
	}
	if !enabled {
		_, enabledWasRequested := updates["auto_scheduling_enabled"]
		_, lockedWasRequested := updates["auto_scheduling_locked"]
		if !story.AutoSchedulingEnabled && !story.AutoSchedulingLocked && !enabledWasRequested && !lockedWasRequested {
			return false, nil
		}
		// Pausing is one operation. Callers do not need to know that a locked
		// schedule must first be unlocked before Maya can retire its blocks.
		updates["auto_scheduling_locked"] = false
		updates["auto_scheduling_status"] = AutoSchedulingStatusOff
		updates["auto_scheduling_reason"] = nil
		updates["auto_scheduling_updated_at"] = time.Now().UTC()
		return story.AutoSchedulingEnabled || story.AutoSchedulingLocked, nil
	}

	if locked && !story.AutoSchedulingLocked {
		if story.AutoSchedulingStatus != AutoSchedulingStatusScheduled && story.AutoSchedulingStatus != AutoSchedulingStatusAtRisk {
			return false, ErrAutoSchedulingLockEmpty
		}
		autoRepo, ok := s.repo.(autoSchedulingRepository)
		if !ok {
			return false, errors.New("story repository does not support auto-scheduling state")
		}
		exists, err := autoRepo.MayaScheduleBlocksExist(ctx, story.ID, story.Workspace)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, ErrAutoSchedulingLockEmpty
		}
	}

	status := story.AutoSchedulingStatus
	reason := story.AutoSchedulingReason
	if locked {
		status = AutoSchedulingStatusLocked
		reason = stringPointer(autoSchedulingLockedReason)
	} else if schedulingConstraintsChanged(updates) || !story.AutoSchedulingEnabled || story.AutoSchedulingLocked || status == AutoSchedulingStatusOff {
		durationMinutes, err := updatedDuration(story.EstimatedDurationMinutes, updates)
		if err != nil {
			return false, err
		}
		status, reason = s.initialAutoSchedulingState(assigneeID, durationMinutes)
	}
	if err := ValidateStoryAutoSchedulingContract(enabled, locked, status); err != nil {
		return false, err
	}
	updates["auto_scheduling_status"] = status
	updates["auto_scheduling_reason"] = reason
	updates["auto_scheduling_updated_at"] = time.Now().UTC()
	return true, nil
}

func (s *Service) autoSchedulingUpdateIsTerminal(ctx context.Context, story CoreSingleStory, updates map[string]any) (bool, error) {
	completedAt := story.CompletedAt
	if raw, exists := updates["completed_at"]; exists {
		switch value := raw.(type) {
		case nil:
			completedAt = nil
		case time.Time:
			completedAt = &value
		case *time.Time:
			completedAt = value
		default:
			return false, fmt.Errorf("invalid completed_at type: %T", raw)
		}
	}
	if completedAt != nil {
		return true, nil
	}
	rawStatus, changed := updates["status_id"]
	if !changed {
		return false, nil
	}
	statusID, valid := optionalUUIDUpdate(rawStatus)
	if !valid || statusID == nil {
		return false, fmt.Errorf("invalid status_id type: %T", rawStatus)
	}
	category, err := s.repo.GetStatusCategory(ctx, statusID.String())
	if err != nil {
		return false, err
	}
	return category == "completed" || category == "cancelled", nil
}

func terminalLifecycleUpdateRequested(updates map[string]any) bool {
	_, statusChanged := updates["status_id"]
	_, completedAtChanged := updates["completed_at"]
	return statusChanged || completedAtChanged
}

func schedulingConstraintsChanged(updates map[string]any) bool {
	for _, field := range []string{
		"assignee_id", "status_id", "priority", "sprint_id", "start_date", "end_date", "completed_at",
		"estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked",
	} {
		if _, exists := updates[field]; exists {
			return true
		}
	}
	return false
}

func updatedBool(current bool, updates map[string]any, field string) (bool, error) {
	raw, exists := updates[field]
	if !exists {
		return current, nil
	}
	switch value := raw.(type) {
	case bool:
		return value, nil
	case *bool:
		if value == nil {
			return false, fmt.Errorf("%s cannot be null", field)
		}
		return *value, nil
	default:
		return false, fmt.Errorf("invalid %s type: %T", field, raw)
	}
}

func updatedAssignee(current *uuid.UUID, updates map[string]any) (*uuid.UUID, error) {
	raw, exists := updates["assignee_id"]
	if !exists {
		return current, nil
	}
	value, valid := optionalUUIDUpdate(raw)
	if !valid {
		return nil, fmt.Errorf("invalid assignee_id type: %T", raw)
	}
	return value, nil
}

func updatedDuration(current *int, updates map[string]any) (*int, error) {
	raw, exists := updates["estimated_duration_minutes"]
	if !exists {
		return current, nil
	}
	if raw == nil {
		return nil, nil
	}
	value, ok := raw.(*int)
	if !ok {
		return nil, fmt.Errorf("invalid estimated_duration_minutes type: %T", raw)
	}
	return value, nil
}

func (s *Service) isMayaActor(userID *uuid.UUID) bool {
	return userID != nil && s.mayaAssignment != nil && *userID == s.mayaAssignment.assigneeID
}

func stringPointer(value string) *string {
	return &value
}

// UpdateAutomationStateIfUnchanged stores the result of a committed schedule
// reconciliation without changing the story input version. A structured
// transition is published only after the state write succeeds.
func (s *Service) UpdateAutomationStateIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	status string,
	reason *string,
	locked *bool,
	schedule *events.StoryScheduleTransition,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected story update time is required")
	}
	if err := ValidateAutoSchedulingStatus(status); err != nil {
		return err
	}
	if schedule != nil {
		if schedule.UserID == uuid.Nil {
			return errors.New("schedule transition user is required")
		}
		if schedule.State != events.StoryScheduleState(status) {
			return errors.New("schedule transition state does not match persisted auto-scheduling status")
		}
	}
	story, err := s.repo.Get(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}
	if !story.UpdatedAt.Equal(expectedUpdatedAt) {
		return ErrStoryChanged
	}
	effectiveLocked := story.AutoSchedulingLocked
	if locked != nil {
		effectiveLocked = *locked
	}
	if effectiveLocked && !story.AutoSchedulingEnabled {
		return ErrLockedAutoSchedulingOff
	}
	if status == AutoSchedulingStatusLocked && !effectiveLocked {
		return ErrLockedAutoSchedulingOff
	}
	lockChanged := locked != nil && story.AutoSchedulingLocked != *locked
	if story.AutoSchedulingStatus == status && equalOptionalString(story.AutoSchedulingReason, reason) && !lockChanged && schedule == nil {
		return nil
	}

	stateUpdatedAt := time.Now().UTC()
	audienceIDs, audienceErr := s.repo.GetNotificationAudience(ctx, storyID, workspaceID)
	audienceResolved := audienceErr == nil
	if audienceErr != nil {
		s.log.Error(ctx, "failed to load story notification audience", "error", audienceErr, "story_id", storyID)
	}
	updates := map[string]any{
		"auto_scheduling_status":     status,
		"auto_scheduling_reason":     reason,
		"auto_scheduling_updated_at": stateUpdatedAt,
	}
	if lockChanged {
		updates["auto_scheduling_locked"] = *locked
	}
	eventReason := ""
	if reason != nil {
		eventReason = *reason
	}
	event := events.Event{
		Type: events.StoryUpdated,
		Payload: events.StoryUpdatedPayload{
			StoryID:          storyID,
			WorkspaceID:      workspaceID,
			Updates:          updates,
			Source:           events.StoryUpdateSourceMaya,
			Reason:           eventReason,
			Schedule:         schedule,
			AssigneeID:       story.Assignee,
			AudienceIDs:      audienceIDs,
			AudienceResolved: audienceResolved,
		},
		Timestamp: stateUpdatedAt,
		ActorID:   actorID,
	}

	if schedule != nil {
		outboxRepo, ok := s.repo.(scheduleTransitionOutboxWriter)
		if !ok {
			return errors.New("story repository does not support durable schedule transitions")
		}
		outbox, err := buildScheduleTransitionOutboxInput(
			event,
			expectedUpdatedAt,
			status,
			reason,
			locked,
			schedule,
			s.publisher != nil,
		)
		if err != nil {
			return err
		}
		updated, claim, err := outboxRepo.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
			ctx,
			storyID,
			workspaceID,
			expectedUpdatedAt,
			status,
			reason,
			stateUpdatedAt,
			locked,
			outbox,
		)
		if err != nil {
			return err
		}
		if !updated {
			return ErrStoryChanged
		}
		if claim != nil && s.publisher != nil {
			s.dispatchImmediateScheduleTransition(ctx, outboxRepo, *claim)
		}
		return nil
	}

	autoRepo, ok := s.repo.(autoSchedulingRepository)
	if !ok {
		return errors.New("story repository does not support auto-scheduling state")
	}
	updated, err := autoRepo.UpdateAutoSchedulingStateIfUnchanged(
		ctx,
		storyID,
		workspaceID,
		expectedUpdatedAt,
		status,
		reason,
		stateUpdatedAt,
		locked,
	)
	if err != nil {
		return err
	}
	if !updated {
		return ErrStoryChanged
	}
	if s.publisher == nil {
		return nil
	}
	if err := s.publisher.Publish(context.WithoutCancel(ctx), event); err != nil {
		s.log.Error(ctx, "failed to publish Maya auto-scheduling state", "error", err, "story_id", storyID)
	}
	return nil
}

func equalOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
