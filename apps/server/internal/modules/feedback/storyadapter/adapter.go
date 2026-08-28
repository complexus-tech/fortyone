// Package storyadapter translates feedback's narrow story-planning capability
// into the stories module's public service contract.
package storyadapter

import (
	"context"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// Adapter deliberately exposes only the three operations feedback needs.
// Authorization remains owned and enforced by the stories service.
type Adapter struct {
	stories *stories.Service
}

func New(storyService *stories.Service) *Adapter {
	if storyService == nil {
		return nil
	}
	return &Adapter{stories: storyService}
}

func (a *Adapter) CreateFromFeedback(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	draft feedback.StoryDraft,
) (feedback.StoryPlan, error) {
	description := draft.Description
	story, err := a.stories.CreateExternal(ctx, actorID, stories.CoreNewStory{
		Title:       draft.Title,
		Description: &description,
		Status:      draft.StatusID,
		Reporter:    &draft.ReporterID,
		Team:        draft.TeamID,
		Priority:    "No Priority",
	}, workspaceID)
	if err != nil {
		return feedback.StoryPlan{}, err
	}
	return toStoryPlan(story), nil
}

func (a *Adapter) GetForFeedback(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
) (feedback.StoryPlan, error) {
	story, err := a.stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return feedback.StoryPlan{}, err
	}
	return toStoryPlan(story), nil
}

func (a *Adapter) DeleteCreatedFromFeedback(
	ctx context.Context,
	workspaceID, storyID, actorID uuid.UUID,
) error {
	return a.stories.Delete(ctx, storyID, workspaceID, stories.BulkDeleteAuthorization{ActorID: actorID})
}

func toStoryPlan(story stories.CoreSingleStory) feedback.StoryPlan {
	return feedback.StoryPlan{
		ID:          story.ID,
		WorkspaceID: story.Workspace,
		TeamID:      story.Team,
		StatusID:    story.Status,
		DeletedAt:   story.DeletedAt,
	}
}
