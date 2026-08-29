package objectiveshttp

import (
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
)

// AppObjectiveList is the stable HTTP representation of an objective.
type AppObjectiveList struct {
	ID                 uuid.UUID              `json:"id"`
	SequenceID         int                    `json:"sequenceId"`
	Name               string                 `json:"name"`
	Description        *string                `json:"description"`
	ShortSummary       *string                `json:"shortSummary"`
	LeadUser           *uuid.UUID             `json:"leadUser"`
	Team               uuid.UUID              `json:"teamId"`
	Workspace          uuid.UUID              `json:"workspaceId"`
	StartDate          *time.Time             `json:"startDate"`
	EndDate            *time.Time             `json:"endDate"`
	IsPrivate          bool                   `json:"isPrivate"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	Status             uuid.UUID              `json:"statusId"`
	Priority           *string                `json:"priority"`
	Health             *string                `json:"health"`
	Color              string                 `json:"color"`
	ForecastStartDate  *time.Time             `json:"forecastStartDate"`
	ForecastEndDate    *time.Time             `json:"forecastEndDate"`
	ScheduleStatus     string                 `json:"scheduleStatus"`
	ForecastDaysDelta  int                    `json:"forecastDaysDelta"`
	ForecastCauseStory *AppForecastCauseStory `json:"forecastCauseStory"`
	KeyResultCount     int                    `json:"keyResultCount"`
	CreatedBy          uuid.UUID              `json:"createdBy"`
	Stats              ObjectiveStats         `json:"stats"`
}

type AppForecastCauseStory struct {
	ID         uuid.UUID `json:"id"`
	SequenceID int       `json:"sequenceId"`
	Title      string    `json:"title"`
	Source     string    `json:"source"`
}

type ObjectiveStats struct {
	Total     int `json:"total"`
	Cancelled int `json:"cancelled"`
	Completed int `json:"completed"`
	Started   int `json:"started"`
	Unstarted int `json:"unstarted"`
	Backlog   int `json:"backlog"`
}

type AppPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

type AppObjectivesResponse struct {
	Objectives []AppObjectiveList `json:"objectives"`
	Pagination AppPagination      `json:"pagination"`
}

func toAppObjectives(values []objectives.CoreObjective) []AppObjectiveList {
	result := make([]AppObjectiveList, len(values))
	for index, objective := range values {
		result[index] = toAppObjective(objective)
	}
	return result
}

func toAppObjectivesResponse(values []objectives.CoreObjective, page, pageSize int, hasMore bool) AppObjectivesResponse {
	nextPage := 0
	if hasMore {
		nextPage = page + 1
	}
	return AppObjectivesResponse{
		Objectives: toAppObjectives(values),
		Pagination: AppPagination{Page: page, PageSize: pageSize, HasMore: hasMore, NextPage: nextPage},
	}
}

func toAppObjective(objective objectives.CoreObjective) AppObjectiveList {
	var health *string
	if objective.Health != nil {
		value := string(*objective.Health)
		health = &value
	}
	result := AppObjectiveList{
		ID: objective.ID, SequenceID: objective.SequenceID, Name: objective.Name,
		Description: objective.Description, ShortSummary: objective.ShortSummary,
		LeadUser: objective.LeadUser, Team: objective.Team, Workspace: objective.Workspace,
		StartDate: objective.StartDate, EndDate: objective.EndDate, IsPrivate: objective.IsPrivate,
		CreatedAt: objective.CreatedAt, UpdatedAt: objective.UpdatedAt, Status: objective.Status,
		Priority: objective.Priority, Health: health, Color: objective.Color, CreatedBy: objective.CreatedBy,
		ForecastStartDate: objective.ForecastStartDate, ForecastEndDate: objective.ForecastEndDate,
		ScheduleStatus: string(objective.ScheduleStatus), ForecastDaysDelta: objective.ForecastDaysDelta,
		KeyResultCount: objective.KeyResultCount,
		Stats: ObjectiveStats{
			Total: objective.TotalStories, Cancelled: objective.CancelledStories,
			Completed: objective.CompletedStories, Started: objective.StartedStories,
			Unstarted: objective.UnstartedStories, Backlog: objective.BacklogStories,
		},
	}
	if objective.ForecastCauseStory != nil {
		result.ForecastCauseStory = &AppForecastCauseStory{
			ID: objective.ForecastCauseStory.ID, SequenceID: objective.ForecastCauseStory.SequenceID,
			Title: objective.ForecastCauseStory.Title, Source: objective.ForecastCauseStory.Source,
		}
	}
	return result
}
