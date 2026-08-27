package stories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type MayaAssignmentInput struct {
	Story       CoreSingleStory
	TriggeredBy uuid.UUID
}

type MayaAssignmentValidator func(ctx context.Context, input MayaAssignmentInput) error

type mayaAssignmentPolicy struct {
	assigneeID uuid.UUID
	validate   MayaAssignmentValidator
}

// ConfigureMayaActor identifies Maya for story scheduling invariants without
// requiring an assignment side-effect callback. Worker-owned story services use
// this when they can mutate stories but do not initiate background assignment.
func (s *Service) ConfigureMayaActor(assigneeID uuid.UUID) {
	s.mayaActorID = assigneeID
}

func (s *Service) ConfigureMayaAssignment(assigneeID uuid.UUID, validator MayaAssignmentValidator) {
	if assigneeID == uuid.Nil || validator == nil {
		s.mayaAssignment = nil
		return
	}

	s.ConfigureMayaActor(assigneeID)
	s.mayaAssignment = &mayaAssignmentPolicy{
		assigneeID: assigneeID,
		validate:   validator,
	}
}

func (s *Service) validateMayaAssignment(ctx context.Context, story CoreSingleStory, previousAssignee *uuid.UUID, triggeredBy uuid.UUID) error {
	if s.mayaAssignment == nil {
		return nil
	}
	if !shouldTriggerMayaAssignment(previousAssignee, story.Assignee, s.mayaAssignment.assigneeID) {
		return nil
	}
	if err := s.mayaAssignment.validate(ctx, MayaAssignmentInput{
		Story:       story,
		TriggeredBy: triggeredBy,
	}); err != nil {
		s.log.Error(ctx, "Maya assignment validation failed", "story_id", story.ID, "workspace_id", story.Workspace, "error", err)
		return fmt.Errorf("validate Maya assignment: %w", err)
	}
	return nil
}

func mayaAssignmentUpdateAssignee(updates map[string]any) (*uuid.UUID, bool) {
	value, exists := updates["assignee_id"]
	if !exists {
		return nil, false
	}
	if value == nil {
		return nil, true
	}

	switch assigneeID := value.(type) {
	case uuid.UUID:
		return &assigneeID, true
	case *uuid.UUID:
		return assigneeID, true
	case string:
		parsed, err := uuid.Parse(assigneeID)
		if err != nil {
			return nil, true
		}
		return &parsed, true
	default:
		return nil, true
	}
}

func shouldTriggerMayaAssignment(previousAssignee, nextAssignee *uuid.UUID, mayaAssigneeID uuid.UUID) bool {
	if mayaAssigneeID == uuid.Nil || nextAssignee == nil || *nextAssignee != mayaAssigneeID {
		return false
	}
	return previousAssignee == nil || *previousAssignee != mayaAssigneeID
}

func storyWithAssignee(story CoreSingleStory, assigneeID *uuid.UUID) (CoreSingleStory, error) {
	if story.ID == uuid.Nil || story.Workspace == uuid.Nil {
		return CoreSingleStory{}, fmt.Errorf("story identity is required for Maya assignment automation")
	}
	story.Assignee = assigneeID
	return story, nil
}
