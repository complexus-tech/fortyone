package slackadapter

import (
	"context"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	"github.com/google/uuid"
)

type ObjectiveBackend interface {
	List(context.Context, uuid.UUID, uuid.UUID, map[string]any) ([]objectives.CoreObjective, error)
}

type ObjectiveReader struct {
	backend ObjectiveBackend
}

func NewObjectiveReader(backend ObjectiveBackend) *ObjectiveReader {
	if backend == nil {
		return nil
	}
	return &ObjectiveReader{backend: backend}
}

func (adapter *ObjectiveReader) ListByID(ctx context.Context, workspaceID, userID, objectiveID uuid.UUID) ([]slack.Objective, error) {
	items, err := adapter.backend.List(ctx, workspaceID, userID, map[string]any{"objective_id": objectiveID})
	if err != nil {
		return nil, err
	}
	result := make([]slack.Objective, 0, len(items))
	for _, item := range items {
		var health *string
		if item.Health != nil {
			value := string(*item.Health)
			health = &value
		}
		result = append(result, slack.Objective{
			ID: item.ID, Name: item.Name, Description: item.Description, LeadUser: item.LeadUser,
			Team: item.Team, Workspace: item.Workspace, StartDate: item.StartDate, EndDate: item.EndDate,
			CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Health: health,
			TotalStories: item.TotalStories, CompletedStories: item.CompletedStories,
		})
	}
	return result, nil
}

type SprintBackend interface {
	List(context.Context, uuid.UUID, uuid.UUID, map[string]any) ([]sprints.CoreSprint, error)
}

type SprintReader struct {
	backend SprintBackend
}

func NewSprintReader(backend SprintBackend) *SprintReader {
	if backend == nil {
		return nil
	}
	return &SprintReader{backend: backend}
}

func (adapter *SprintReader) ListByID(ctx context.Context, workspaceID, userID, sprintID uuid.UUID) ([]slack.Sprint, error) {
	items, err := adapter.backend.List(ctx, workspaceID, userID, map[string]any{"sprint_id": sprintID})
	if err != nil {
		return nil, err
	}
	result := make([]slack.Sprint, 0, len(items))
	for _, item := range items {
		result = append(result, slack.Sprint{
			ID: item.ID, Name: item.Name, Goal: item.Goal, TeamID: item.TeamID,
			WorkspaceID: item.WorkspaceID, StartDate: item.StartDate, EndDate: item.EndDate,
			CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
			TotalStories: item.TotalStories, CompletedStories: item.CompletedStories,
		})
	}
	return result, nil
}

var (
	_ slack.SlackObjectiveReader = (*ObjectiveReader)(nil)
	_ slack.SlackSprintReader    = (*SprintReader)(nil)
)
