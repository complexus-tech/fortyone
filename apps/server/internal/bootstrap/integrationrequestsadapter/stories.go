// Package integrationrequestsadapter contains composition-only bridges from
// integration-request-owned capabilities to sibling module use cases.
package integrationrequestsadapter

import (
	"context"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type StoryBackend interface {
	CreateExternalUserAction(
		context.Context,
		uuid.UUID,
		stories.CoreNewStory,
		uuid.UUID,
	) (stories.CoreSingleStory, error)
}

type StoryCreator struct {
	backend StoryBackend
}

func NewStoryCreator(backend StoryBackend) integrationrequests.StoryCreator {
	if backend == nil {
		return nil
	}
	return StoryCreator{backend: backend}
}

func (adapter StoryCreator) CreateForIntegrationRequest(
	ctx context.Context,
	actorID, workspaceID uuid.UUID,
	input integrationrequests.NewStory,
) (integrationrequests.Story, error) {
	creationKey := input.CreationKey
	story, err := adapter.backend.CreateExternalUserAction(ctx, actorID, stories.CoreNewStory{
		Title:                    input.Title,
		Description:              input.Description,
		Status:                   input.StatusID,
		Reporter:                 input.ReporterID,
		Assignee:                 input.AssigneeID,
		Team:                     input.TeamID,
		Priority:                 input.Priority,
		EstimateValue:            input.EstimateValue,
		EstimatedDurationMinutes: input.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: input.MinimumFocusBlockMinutes,
		Objective:                input.ObjectiveID,
		KeyResult:                input.KeyResultID,
		Sprint:                   input.SprintID,
		StartDate:                input.StartDate,
		EndDate:                  input.EndDate,
		LabelIDs:                 append([]uuid.UUID(nil), input.LabelIDs...),
		CreationKey:              &creationKey,
	}, workspaceID)
	if err != nil {
		return integrationrequests.Story{}, err
	}
	return integrationrequests.Story{
		ID:          story.ID,
		SequenceID:  story.SequenceID,
		TeamID:      story.Team,
		TeamCode:    story.TeamCode,
		Title:       story.Title,
		Description: story.Description,
		StatusID:    story.Status,
		Priority:    story.Priority,
		AssigneeID:  story.Assignee,
		ReporterID:  story.Reporter,
		EndDate:     story.EndDate,
		CreatedAt:   story.CreatedAt,
		UpdatedAt:   story.UpdatedAt,
	}, nil
}

var _ integrationrequests.StoryCreator = StoryCreator{}
