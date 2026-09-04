package workspacebootstrap

import (
	"context"
	"errors"
	"fmt"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/google/uuid"
)

type seedContentCreator struct {
	states  *states.Service
	stories *stories.Service
}

func NewSeedContentCreator(statesService *states.Service, storiesService *stories.Service) workspaces.SeedContentCreator {
	return seedContentCreator{states: statesService, stories: storiesService}
}

func NewExampleContentCreator(statesService *states.Service, storiesService *stories.Service) workspaces.ExampleContentCreator {
	return seedContentCreator{states: statesService, stories: storiesService}
}

func (creator seedContentCreator) CreateWorkspaceSeedContent(
	ctx context.Context,
	workspaceID, teamID, userID uuid.UUID,
) error {
	statuses, err := creator.teamStatuses(ctx, workspaceID, teamID)
	if err != nil {
		return err
	}
	seedStories, err := workspaces.BuildSeedStories(teamID, userID, statuses)
	if err != nil {
		return err
	}
	return creator.createStories(ctx, workspaceID, seedStories)
}

func (creator seedContentCreator) CreateWorkspaceExamples(
	ctx context.Context,
	workspaceID, teamID, userID uuid.UUID,
	workType workspaces.WorkType,
) error {
	statuses, err := creator.teamStatuses(ctx, workspaceID, teamID)
	if err != nil {
		return err
	}
	examples, err := workspaces.BuildExampleStories(teamID, userID, statuses, workType)
	if err != nil {
		return err
	}
	return creator.createStories(ctx, workspaceID, examples)
}

func (creator seedContentCreator) teamStatuses(ctx context.Context, workspaceID, teamID uuid.UUID) ([]workspaces.SeedStatus, error) {
	teamStatuses, err := creator.states.TeamList(ctx, workspaceID, teamID)
	if err != nil {
		return nil, fmt.Errorf("list default team statuses: %w", err)
	}
	statuses := make([]workspaces.SeedStatus, len(teamStatuses))
	for index, status := range teamStatuses {
		statuses[index] = workspaces.SeedStatus{ID: status.ID, Category: status.Category}
	}
	return statuses, nil
}

func (creator seedContentCreator) createStories(ctx context.Context, workspaceID uuid.UUID, seedStories []workspaces.SeedStory) error {
	var creationErrors []error
	for _, story := range seedStories {
		reporter, assignee, status := story.Reporter, story.Assignee, story.Status
		_, err := creator.stories.Create(ctx, stories.CoreNewStory{
			Title: story.Title, Description: story.Description, DescriptionHTML: story.DescriptionHTML,
			Reporter: &reporter, Assignee: &assignee, Priority: story.Priority,
			Team: story.Team, Status: &status, StartDate: story.StartDate, EndDate: story.EndDate,
		}, workspaceID)
		if err != nil {
			creationErrors = append(creationErrors, fmt.Errorf("create seed story %q: %w", story.Title, err))
		}
	}
	return errors.Join(creationErrors...)
}
