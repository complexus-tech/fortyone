package objectivesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/google/uuid"
)

// AppObjectiveList represents a list of objectives in the application.
type AppObjectiveList struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	LeadUser    *uuid.UUID     `json:"leadUser"`
	Team        *uuid.UUID     `json:"teamId"`
	Workspace   uuid.UUID      `json:"workspaceId"`
	StartDate   *time.Time     `json:"startDate"`
	EndDate     *time.Time     `json:"endDate"`
	IsPrivate   bool           `json:"isPrivate"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Status      uuid.UUID      `json:"statusId"`
	Priority    *string        `json:"priority"`
	Stats       ObjectiveStats `json:"stats"`
}

type ObjectiveStats struct {
	Total     int `json:"total"`
	Cancelled int `json:"cancelled"`
	Completed int `json:"completed"`
	Started   int `json:"started"`
	Unstarted int `json:"unstarted"`
	Backlog   int `json:"backlog"`
}

type AppFilters struct {
	Team *uuid.UUID `json:"teamId" db:"team_id"`
}

// toAppObjectives converts a list of core objectives to a list of application objectives.
func toAppObjectives(objectives []objectives.CoreObjective) []AppObjectiveList {
	appObjectives := make([]AppObjectiveList, len(objectives))
	for i, objective := range objectives {
		appObjectives[i] = AppObjectiveList{
			ID:          objective.ID,
			Name:        objective.Name,
			Description: objective.Description,
			LeadUser:    objective.LeadUser,
			Team:        objective.Team,
			Workspace:   objective.Workspace,
			StartDate:   objective.StartDate,
			EndDate:     objective.EndDate,
			IsPrivate:   objective.IsPrivate,
			CreatedAt:   objective.CreatedAt,
			UpdatedAt:   objective.UpdatedAt,
			Status:      objective.Status,
			Priority:    objective.Priority,
			Stats: ObjectiveStats{
				Total:     objective.TotalStories,
				Cancelled: objective.CancelledStories,
				Completed: objective.CompletedStories,
				Started:   objective.StartedStories,
				Unstarted: objective.UnstartedStories,
				Backlog:   objective.BacklogStories,
			},
		}
	}
	return appObjectives
}

// AppNewObjective represents the data needed to create a new objective
type AppNewObjective struct {
	Name        string            `json:"name" validate:"required"`
	Description *string           `json:"description"`
	LeadUser    *uuid.UUID        `json:"leadUserId"`
	Team        uuid.UUID         `json:"teamId" validate:"required"`
	StartDate   time.Time         `json:"startDate"`
	EndDate     time.Time         `json:"endDate"`
	IsPrivate   bool              `json:"isPrivate"`
	Status      uuid.UUID         `json:"statusId"`
	Priority    *string           `json:"priority"`
	KeyResults  []AppNewKeyResult `json:"keyResults,omitempty"`
}

type AppNewKeyResult struct {
	Name            string   `json:"name" validate:"required"`
	MeasurementType string   `json:"measurementType" validate:"required,oneof=percentage number boolean"`
	StartValue      *float64 `json:"startValue"`
	TargetValue     *float64 `json:"targetValue"`
}

// toCoreNewObjective converts an AppNewObjective to a CoreNewObjective
func toCoreNewObjective(ano AppNewObjective) objectives.CoreNewObjective {
	return objectives.CoreNewObjective{
		Name:        ano.Name,
		Description: ano.Description,
		LeadUser:    ano.LeadUser,
		Team:        ano.Team,
		StartDate:   ano.StartDate,
		EndDate:     ano.EndDate,
		IsPrivate:   ano.IsPrivate,
		Status:      ano.Status,
		Priority:    ano.Priority,
	}
}

// toAppObjective converts a single core objective to an application objective.
func toAppObjective(objective objectives.CoreObjective) AppObjectiveList {
	return AppObjectiveList{
		ID:          objective.ID,
		Name:        objective.Name,
		Description: objective.Description,
		LeadUser:    objective.LeadUser,
		Team:        objective.Team,
		Workspace:   objective.Workspace,
		StartDate:   objective.StartDate,
		EndDate:     objective.EndDate,
		IsPrivate:   objective.IsPrivate,
		CreatedAt:   objective.CreatedAt,
		UpdatedAt:   objective.UpdatedAt,
		Status:      objective.Status,
		Priority:    objective.Priority,
		Stats: ObjectiveStats{
			Total:     objective.TotalStories,
			Cancelled: objective.CancelledStories,
			Completed: objective.CompletedStories,
			Started:   objective.StartedStories,
			Unstarted: objective.UnstartedStories,
			Backlog:   objective.BacklogStories,
		},
	}
}

// AppUpdateObjective represents the data needed to update an objective
type AppUpdateObjective struct {
	Name        *string    `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	LeadUser    *uuid.UUID `json:"leadUser" db:"lead_user_id"`
	StartDate   *time.Time `json:"startDate" db:"start_date"`
	EndDate     *time.Time `json:"endDate" db:"end_date"`
	IsPrivate   *bool      `json:"isPrivate" db:"is_private"`
	Status      *uuid.UUID `json:"statusId" db:"status_id"`
	Priority    *string    `json:"priority" db:"priority"`
}

// AppKeyResult represents a key result in the application
type AppKeyResult struct {
	ID              uuid.UUID `json:"id"`
	ObjectiveID     uuid.UUID `json:"objectiveId"`
	Name            string    `json:"name"`
	MeasurementType string    `json:"measurementType"`
	StartValue      *float64  `json:"startValue"`
	TargetValue     *float64  `json:"targetValue"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func toAppKeyResult(kr keyresults.CoreKeyResult) AppKeyResult {
	return AppKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		TargetValue:     kr.TargetValue,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
	}
}

func toAppKeyResults(krs []keyresults.CoreKeyResult) []AppKeyResult {
	result := make([]AppKeyResult, len(krs))
	for i, kr := range krs {
		result[i] = toAppKeyResult(kr)
	}
	return result
}
