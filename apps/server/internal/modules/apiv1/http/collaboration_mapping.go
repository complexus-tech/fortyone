package apiv1http

import (
	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

func labelModel(label labels.CoreLabel) openapiv1.ComponentsResourcesLabel {
	return openapiv1.ComponentsResourcesLabel{
		Id: label.ID, WorkspaceId: label.WorkspaceID, TeamId: label.TeamID,
		Name: label.Name, Color: label.Color,
		CreatedAt: label.CreatedAt.UTC(), UpdatedAt: label.UpdatedAt.UTC(),
	}
}

func workflowStateModel(state states.CoreState) openapiv1.ComponentsResourcesWorkflowState {
	return openapiv1.ComponentsResourcesWorkflowState{
		Id: state.ID, WorkspaceId: state.Workspace, TeamId: state.Team,
		Name: state.Name, Category: openapiv1.ComponentsResourcesWorkflowStateCategory(state.Category),
		Order: state.OrderIndex, Default: state.IsDefault, Color: state.Color,
		CreatedAt: state.CreatedAt.UTC(), UpdatedAt: state.UpdatedAt.UTC(),
	}
}

func sprintModel(sprint sprints.CoreSprint) openapiv1.ComponentsResourcesSprint {
	return openapiv1.ComponentsResourcesSprint{
		Id: sprint.ID, WorkspaceId: sprint.WorkspaceID, TeamId: sprint.TeamID,
		ObjectiveId: sprint.ObjectiveID, Name: sprint.Name, Goal: sprint.Goal,
		StartDate: sprint.StartDate.UTC(), EndDate: sprint.EndDate.UTC(),
		AutomationManaged: sprint.ScheduleManagedByAutomation,
		StoryCounts: storyCounts(
			sprint.TotalStories, sprint.BacklogStories, sprint.UnstartedStories,
			sprint.StartedStories, sprint.CompletedStories, sprint.CancelledStories,
		),
		CreatedAt: sprint.CreatedAt.UTC(), UpdatedAt: sprint.UpdatedAt.UTC(),
	}
}

func objectiveModel(objective objectives.CoreObjective) openapiv1.ComponentsResourcesObjective {
	var health *openapiv1.ComponentsResourcesObjectiveHealth
	if objective.Health != nil {
		value := openapiv1.ComponentsResourcesObjectiveHealth(*objective.Health)
		health = &value
	}
	return openapiv1.ComponentsResourcesObjective{
		Id: objective.ID, WorkspaceId: objective.Workspace, TeamId: objective.Team,
		SequenceId: objective.SequenceID, Name: objective.Name, Description: objective.Description,
		ShortSummary: objective.ShortSummary, LeadUserId: objective.LeadUser,
		StartDate: utcTimePointer(objective.StartDate), EndDate: utcTimePointer(objective.EndDate),
		Private: objective.IsPrivate, StatusId: objective.Status, Priority: objective.Priority,
		Health: health, Color: objective.Color,
		ForecastStartDate: utcTimePointer(objective.ForecastStartDate),
		ForecastEndDate:   utcTimePointer(objective.ForecastEndDate),
		ScheduleStatus:    openapiv1.ComponentsResourcesObjectiveScheduleStatus(objective.ScheduleStatus),
		ForecastDaysDelta: objective.ForecastDaysDelta, KeyResultCount: objective.KeyResultCount,
		StoryCounts: storyCounts(
			objective.TotalStories, objective.BacklogStories, objective.UnstartedStories,
			objective.StartedStories, objective.CompletedStories, objective.CancelledStories,
		),
		CreatedBy: objective.CreatedBy, CreatedAt: objective.CreatedAt.UTC(), UpdatedAt: objective.UpdatedAt.UTC(),
	}
}

func keyResultModel(result keyresults.CoreKeyResultWithObjective) openapiv1.ComponentsResourcesKeyResult {
	objectiveName := result.ObjectiveName
	return openapiv1.ComponentsResourcesKeyResult{
		Id: result.ID, WorkspaceId: result.WorkspaceID, TeamId: result.TeamID,
		ObjectiveId: result.ObjectiveID, ObjectiveName: &objectiveName, SequenceId: result.SequenceID,
		Name:            result.Name,
		MeasurementType: openapiv1.ComponentsResourcesKeyResultMeasurementType(result.MeasurementType),
		StartValue:      result.StartValue, CurrentValue: result.CurrentValue, TargetValue: result.TargetValue,
		LeadUserId: result.Lead, ContributorIds: copyUUIDs(result.Contributors),
		StartDate: utcTimePointer(result.StartDate), EndDate: utcTimePointer(result.EndDate),
		CreatedBy: result.CreatedBy, CreatedAt: result.CreatedAt.UTC(), UpdatedAt: result.UpdatedAt.UTC(),
	}
}

func commentModel(comment stories.CoreComment) openapiv1.ComponentsResourcesComment {
	replies := make([]openapiv1.ComponentsResourcesComment, len(comment.SubComments))
	for index, reply := range comment.SubComments {
		replies[index] = commentModel(reply)
	}
	return openapiv1.ComponentsResourcesComment{
		Id: comment.ID, StoryId: comment.StoryID, ParentId: comment.Parent,
		AuthorId: comment.UserID, Content: comment.Comment, Replies: replies,
		CreatedAt: comment.CreatedAt.UTC(), UpdatedAt: comment.UpdatedAt.UTC(),
	}
}

func storyCounts(total, backlog, unstarted, started, completed, cancelled int) openapiv1.ComponentsResourcesStoryCounts {
	return openapiv1.ComponentsResourcesStoryCounts{
		Total: total, Backlog: backlog, Unstarted: unstarted,
		Started: started, Completed: completed, Cancelled: cancelled,
	}
}
