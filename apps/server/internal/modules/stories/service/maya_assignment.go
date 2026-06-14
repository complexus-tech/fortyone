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

type MayaAssignmentHandler func(ctx context.Context, input MayaAssignmentInput) error

type mayaAssignmentAutomation struct {
	assigneeID uuid.UUID
	handler    MayaAssignmentHandler
}

func (s *Service) ConfigureMayaAssignment(assigneeID uuid.UUID, handler MayaAssignmentHandler) {
	if assigneeID == uuid.Nil || handler == nil {
		s.mayaAssignment = nil
		return
	}

	s.mayaAssignment = &mayaAssignmentAutomation{
		assigneeID: assigneeID,
		handler:    handler,
	}
}

func (s *Service) triggerMayaAssignment(ctx context.Context, story CoreSingleStory, previousAssignee *uuid.UUID, triggeredBy uuid.UUID) {
	if s.mayaAssignment == nil {
		return
	}
	if !shouldTriggerMayaAssignment(previousAssignee, story.Assignee, s.mayaAssignment.assigneeID) {
		return
	}
	if err := s.mayaAssignment.handler(ctx, MayaAssignmentInput{
		Story:       story,
		TriggeredBy: triggeredBy,
	}); err != nil {
		s.log.Error(ctx, "failed to run Maya assignment automation", "story_id", story.ID, "workspace_id", story.Workspace, "error", err)
	}
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
	if previousAssignee == nil {
		return true
	}
	return *previousAssignee != *nextAssignee
}

func storyWithAssignee(story CoreSingleStory, assigneeID *uuid.UUID) (CoreSingleStory, error) {
	if story.ID == uuid.Nil || story.Workspace == uuid.Nil {
		return CoreSingleStory{}, fmt.Errorf("story identity is required for Maya assignment automation")
	}
	story.Assignee = assigneeID
	return story, nil
}
