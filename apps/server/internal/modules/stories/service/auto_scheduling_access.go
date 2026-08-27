package stories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAutoSchedulingUnavailable          = errors.New("auto-scheduling is not available for this workspace")
	ErrAutoSchedulingAccessCheckFailed    = errors.New("auto-scheduling access could not be verified")
	ErrMayaAssignmentRequiresScheduling   = errors.New("assigning a story to Maya requires auto-scheduling to be enabled")
	ErrMayaAssignmentRequiresDuration     = errors.New("assigning a story to Maya requires time needed")
	ErrMayaAssignmentRequiresDeliveryDate = errors.New("assigning a story to Maya requires a delivery date or sprint")
)

// AutoSchedulingEligibilityChecker verifies the workspace entitlement at the
// moment a story mutation enables auto-scheduling. Callers must not cache this
// result across a proposal/approval boundary.
type AutoSchedulingEligibilityChecker func(ctx context.Context, workspaceID uuid.UUID) (bool, error)

// ConfigureAutoSchedulingEligibility installs the persistence-boundary access
// check used by every story create or update that explicitly enables
// auto-scheduling.
func (s *Service) ConfigureAutoSchedulingEligibility(checker AutoSchedulingEligibilityChecker) {
	s.autoSchedulingEligibility = checker
}

func (s *Service) validateAutoSchedulingEligibility(ctx context.Context, workspaceID uuid.UUID) error {
	if s.autoSchedulingEligibility == nil {
		return ErrAutoSchedulingAccessCheckFailed
	}

	eligible, err := s.autoSchedulingEligibility(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAutoSchedulingAccessCheckFailed, err)
	}
	if !eligible {
		return ErrAutoSchedulingUnavailable
	}
	return nil
}

func (s *Service) validateMayaSchedulingCreate(story CoreSingleStory) error {
	if !s.isMayaActor(story.Assignee) {
		return nil
	}
	return validateCompleteMayaSchedulingIntent(story.AutoSchedulingEnabled, story.EstimatedDurationMinutes, story.EndDate != nil || story.Sprint != nil)
}

func (s *Service) validateMayaSchedulingUpdate(story CoreSingleStory, updates map[string]any) error {
	assigneeID, err := updatedAssignee(story.Assignee, updates)
	if err != nil {
		return err
	}
	if !s.isMayaActor(assigneeID) {
		return nil
	}

	assignmentChanged := !s.isMayaActor(story.Assignee)
	explicitEnable, err := autoSchedulingEnableRequested(story, updates)
	if err != nil {
		return err
	}
	if assignmentChanged && !explicitEnable {
		return ErrMayaAssignmentRequiresScheduling
	}
	if !assignmentChanged && !explicitEnable {
		return nil
	}

	enabled, err := updatedBool(story.AutoSchedulingEnabled, updates, "auto_scheduling_enabled")
	if err != nil {
		return err
	}

	durationMinutes, err := updatedDuration(story.EstimatedDurationMinutes, updates)
	if err != nil {
		return err
	}
	hasDeliveryDate, err := effectiveSchedulingDeliveryDate(story, updates)
	if err != nil {
		return err
	}
	return validateCompleteMayaSchedulingIntent(enabled, durationMinutes, hasDeliveryDate)
}

func validateCompleteMayaSchedulingIntent(enabled bool, durationMinutes *int, hasDeliveryDate bool) error {
	if !enabled {
		return ErrMayaAssignmentRequiresScheduling
	}
	if durationMinutes == nil {
		return ErrMayaAssignmentRequiresDuration
	}
	if !hasDeliveryDate {
		return ErrMayaAssignmentRequiresDeliveryDate
	}
	return nil
}

func effectiveSchedulingDeliveryDate(story CoreSingleStory, updates map[string]any) (bool, error) {
	endDateSet := story.EndDate != nil
	if raw, exists := updates["end_date"]; exists {
		switch value := raw.(type) {
		case nil:
			endDateSet = false
		case *time.Time:
			endDateSet = value != nil
		case time.Time:
			endDateSet = !value.IsZero()
		default:
			return false, fmt.Errorf("invalid end_date type: %T", raw)
		}
	}

	sprintID := story.Sprint
	if raw, exists := updates["sprint_id"]; exists {
		var valid bool
		sprintID, valid = optionalUUIDUpdate(raw)
		if !valid {
			return false, fmt.Errorf("invalid sprint_id type: %T", raw)
		}
	}
	return endDateSet || sprintID != nil, nil
}

func autoSchedulingEnableRequested(story CoreSingleStory, updates map[string]any) (bool, error) {
	raw, exists := updates["auto_scheduling_enabled"]
	if !exists {
		return false, nil
	}
	enabled, err := updatedBool(story.AutoSchedulingEnabled, map[string]any{"auto_scheduling_enabled": raw}, "auto_scheduling_enabled")
	if err != nil {
		return false, err
	}
	return enabled, nil
}
