package providers

import (
	"context"
	"errors"

	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// FigmaStorySource is the narrow concrete-side contract needed to translate
// between the stories application service and Figma's design-context port.
// Keeping the translation in bootstrap prevents either feature module from
// importing the other's models.
type FigmaStorySource interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error)
	CreateExternal(context.Context, uuid.UUID, stories.CoreNewStory, uuid.UUID) (stories.CoreSingleStory, error)
	RecordActivity(context.Context, stories.CoreActivity) error
}

type FigmaStoryAdapter struct {
	stories FigmaStorySource
}

func NewFigmaStoryAdapter(source FigmaStorySource) (*FigmaStoryAdapter, error) {
	if source == nil {
		return nil, errors.New("figma story adapter: story service is required")
	}
	return &FigmaStoryAdapter{stories: source}, nil
}

func (adapter *FigmaStoryAdapter) Get(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
) (figma.Story, error) {
	story, err := adapter.stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return figma.Story{}, err
	}
	return mapFigmaStory(story), nil
}

func (adapter *FigmaStoryAdapter) CreateExternal(
	ctx context.Context,
	actorID uuid.UUID,
	input figma.NewStory,
	workspaceID uuid.UUID,
) (figma.Story, error) {
	story, err := adapter.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title: input.Title, Description: input.Description,
		DescriptionHTML: input.DescriptionHTML, Team: input.TeamID,
		Status: input.StatusID, Reporter: input.ReporterID, Priority: input.Priority,
	}, workspaceID)
	if err != nil {
		return figma.Story{}, err
	}
	return mapFigmaStory(story), nil
}

func (adapter *FigmaStoryAdapter) RecordActivity(
	ctx context.Context,
	activity figma.StoryActivity,
) error {
	return adapter.stories.RecordActivity(ctx, stories.CoreActivity{
		StoryID: activity.StoryID, UserID: activity.ActorID,
		Type: activity.Type, Field: activity.Field,
		CurrentValue: activity.Previous, NewValue: activity.Current,
		WorkspaceID: activity.WorkspaceID,
	})
}

func mapFigmaStory(story stories.CoreSingleStory) figma.Story {
	return figma.Story{
		ID: story.ID, SequenceID: story.SequenceID,
		TeamCode: story.TeamCode, Title: story.Title,
	}
}

var _ figma.StoryService = (*FigmaStoryAdapter)(nil)
