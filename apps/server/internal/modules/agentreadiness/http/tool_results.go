package agentreadinesshttp

import (
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

func storyToolResult(story stories.CoreSingleStory) map[string]any {
	return map[string]any{
		"id":                       story.ID,
		"sequenceId":               story.SequenceID,
		"title":                    story.Title,
		"description":              story.Description,
		"teamId":                   story.Team,
		"teamCode":                 story.TeamCode,
		"statusId":                 story.Status,
		"assigneeId":               story.Assignee,
		"priority":                 story.Priority,
		"estimateValue":            story.EstimateValue,
		"estimateLabel":            story.EstimateLabel,
		"estimatedDurationMinutes": story.EstimatedDurationMinutes,
		"minimumFocusBlockMinutes": story.MinimumFocusBlockMinutes,
		"autoSchedulingEnabled":    story.AutoSchedulingEnabled,
		"autoSchedulingLocked":     story.AutoSchedulingLocked,
		"autoSchedulingStatus":     story.AutoSchedulingStatus,
		"autoSchedulingReason":     story.AutoSchedulingReason,
		"startDate":                story.StartDate,
		"endDate":                  story.EndDate,
		"sprintId":                 story.Sprint,
		"objectiveId":              story.Objective,
		"keyResultId":              story.KeyResult,
		"parentId":                 story.Parent,
		"labelIds":                 story.Labels,
		"createdAt":                story.CreatedAt,
		"updatedAt":                story.UpdatedAt,
		"completedAt":              story.CompletedAt,
		"archivedAt":               story.ArchivedAt,
	}
}

func sprintToolResult(sprint sprints.CoreSprint) map[string]any {
	return map[string]any{
		"id":                          sprint.ID,
		"name":                        sprint.Name,
		"goal":                        sprint.Goal,
		"objectiveId":                 sprint.Objective,
		"teamId":                      sprint.Team,
		"workspaceId":                 sprint.Workspace,
		"startDate":                   sprint.StartDate,
		"endDate":                     sprint.EndDate,
		"createdAt":                   sprint.CreatedAt,
		"updatedAt":                   sprint.UpdatedAt,
		"scheduleManagedByAutomation": sprint.ScheduleManagedByAutomation,
		"totalStories":                sprint.TotalStories,
		"cancelledStories":            sprint.CancelledStories,
		"completedStories":            sprint.CompletedStories,
		"startedStories":              sprint.StartedStories,
		"unstartedStories":            sprint.UnstartedStories,
		"backlogStories":              sprint.BacklogStories,
	}
}

func objectiveToolResult(objective objectives.CoreObjective) map[string]any {
	return map[string]any{
		"id":                objective.ID,
		"sequenceId":        objective.SequenceID,
		"name":              objective.Name,
		"description":       objective.Description,
		"shortSummary":      objective.ShortSummary,
		"leadUserId":        objective.LeadUser,
		"teamId":            objective.Team,
		"workspaceId":       objective.Workspace,
		"startDate":         objective.StartDate,
		"endDate":           objective.EndDate,
		"isPrivate":         objective.IsPrivate,
		"statusId":          objective.Status,
		"priority":          objective.Priority,
		"health":            objective.Health,
		"color":             objective.Color,
		"forecastStartDate": objective.ForecastStartDate,
		"forecastEndDate":   objective.ForecastEndDate,
		"scheduleStatus":    objective.ScheduleStatus,
		"forecastDaysDelta": objective.ForecastDaysDelta,
		"keyResultCount":    objective.KeyResultCount,
		"totalStories":      objective.TotalStories,
		"cancelledStories":  objective.CancelledStories,
		"completedStories":  objective.CompletedStories,
		"startedStories":    objective.StartedStories,
		"unstartedStories":  objective.UnstartedStories,
		"backlogStories":    objective.BacklogStories,
		"createdAt":         objective.CreatedAt,
		"updatedAt":         objective.UpdatedAt,
	}
}

func keyResultToolResult(keyResult keyresults.CoreKeyResult) map[string]any {
	return map[string]any{
		"id":              keyResult.ID,
		"sequenceId":      keyResult.SequenceID,
		"objectiveId":     keyResult.ObjectiveID,
		"name":            keyResult.Name,
		"measurementType": keyResult.MeasurementType,
		"startValue":      keyResult.StartValue,
		"currentValue":    keyResult.CurrentValue,
		"targetValue":     keyResult.TargetValue,
		"leadUserId":      keyResult.Lead,
		"contributorIds":  keyResult.Contributors,
		"startDate":       keyResult.StartDate,
		"endDate":         keyResult.EndDate,
		"createdAt":       keyResult.CreatedAt,
		"updatedAt":       keyResult.UpdatedAt,
	}
}

func keyResultWithObjectiveToolResult(keyResult keyresults.CoreKeyResultWithObjective) map[string]any {
	result := keyResultToolResult(keyResult.CoreKeyResult)
	result["objectiveId"] = keyResult.ObjectiveID
	result["objectiveName"] = keyResult.ObjectiveName
	result["teamId"] = keyResult.TeamID
	result["teamName"] = keyResult.TeamName
	result["teamCode"] = keyResult.TeamCode
	return result
}
